package pyroscopeexporter

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"connectrpc.com/connect"
	"github.com/grafana/dskit/backoff"
	pushv1 "github.com/grafana/pyroscope/api/gen/proto/go/push/v1"
	"github.com/grafana/pyroscope/api/gen/proto/go/push/v1/pushv1connect"
	typesv1 "github.com/grafana/pyroscope/api/gen/proto/go/types/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
)

type RawSample struct {
	RawProfile []byte `json:"raw_profile"`
	ID         string `json:"id"`
}

func BulkSend(series []*pushv1.RawProfileSeries, sender *AsyncSender) error {
	req := connect.NewRequest(&pushv1.PushRequest{Series: series})
	return sender.Enqueue(req)
}

func PrepareDataForTransmission(lbs labels.Labels, samples []*RawSample) ([]*typesv1.LabelPair, []*pushv1.RawSample, error) {
	b := labels.NewBuilder(nil)
	for _, l := range lbs {
		if strings.HasPrefix(l.Name, model.ReservedLabelPrefix) && l.Name != labels.MetricName && l.Name != "__delta__" {
			continue
		}
		b.Set(l.Name, l.Value)
	}
	var protoLabels []*typesv1.LabelPair
	for _, l := range b.Labels() {
		protoLabels = append(protoLabels, &typesv1.LabelPair{Name: l.Name, Value: l.Value})
	}
	var protoSamples []*pushv1.RawSample
	for _, s := range samples {
		protoSamples = append(protoSamples, &pushv1.RawSample{RawProfile: s.RawProfile})
	}
	return protoLabels, protoSamples, nil
}

type AsyncSender struct {
	q          chan *connect.Request[pushv1.PushRequest]
	client     pushv1connect.PusherServiceClient
	timeout    time.Duration
	minBackoff time.Duration
	maxBackoff time.Duration
	maxRetries int
	workers    int
	ctx        context.Context
	cancel     context.CancelFunc

	enqueued  atomic.Uint64
	queueFull atomic.Uint64
	pushOK    atomic.Uint64
	retries   atomic.Uint64
	failedMu  sync.Mutex
	failed    map[string]uint64
}

// SenderStats are cumulative push counters. Failed is keyed by connect code.
type SenderStats struct {
	Enqueued  uint64
	QueueFull uint64
	PushOK    uint64
	Retries   uint64
	Failed    map[string]uint64
}

// Stats returns a copy of the cumulative counters.
func (s *AsyncSender) Stats() SenderStats {
	st := SenderStats{
		Enqueued:  s.enqueued.Load(),
		QueueFull: s.queueFull.Load(),
		PushOK:    s.pushOK.Load(),
		Retries:   s.retries.Load(),
	}
	s.failedMu.Lock()
	st.Failed = make(map[string]uint64, len(s.failed))
	for k, v := range s.failed {
		st.Failed[k] = v
	}
	s.failedMu.Unlock()
	return st
}

func (s *AsyncSender) recordFailed(err error) {
	code := connect.CodeOf(err).String()
	s.failedMu.Lock()
	s.failed[code]++
	s.failedMu.Unlock()
}

func NewAsyncSender(endpointURL string, timeout time.Duration, workers, queueCap int,
	minB, maxB time.Duration, maxRetries int,
) *AsyncSender {
	tr := &http.Transport{
		MaxIdleConns:        128,
		MaxIdleConnsPerHost: workers * 4,
		MaxConnsPerHost:     workers * 4,
		IdleConnTimeout:     90 * time.Second,
	}
	httpClient := &http.Client{Timeout: timeout, Transport: tr}
	c := pushv1connect.NewPusherServiceClient(
		httpClient,
		endpointURL,
		connect.WithSendCompression("gzip"),
		connect.WithCompressMinBytes(1024),
	)
	ctx, cancel := context.WithCancel(context.Background())
	s := &AsyncSender{
		q:      make(chan *connect.Request[pushv1.PushRequest], queueCap),
		client: c, timeout: timeout,
		minBackoff: minB, maxBackoff: maxB, maxRetries: maxRetries,
		workers: workers, ctx: ctx, cancel: cancel,
		failed: map[string]uint64{},
	}
	for i := 0; i < workers; i++ {
		go s.worker()
	}
	return s
}

func (s *AsyncSender) Stop() { s.cancel(); close(s.q) }

func (s *AsyncSender) Enqueue(req *connect.Request[pushv1.PushRequest]) error {
	select {
	case s.q <- req:
		s.enqueued.Add(1)
		return nil
	default:
		s.queueFull.Add(1)
		return fmt.Errorf("flamegraph send queue full")
	}
}

func (s *AsyncSender) worker() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case req, ok := <-s.q:
			if !ok {
				return
			}
			bo := backoff.New(s.ctx, backoff.Config{
				MinBackoff: s.minBackoff, MaxBackoff: s.maxBackoff, MaxRetries: s.maxRetries,
			})
			err := s.ctx.Err()
			for attempt := 0; bo.Ongoing(); attempt++ {
				if attempt > 0 {
					s.retries.Add(1)
				}
				ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
				_, err = s.client.Push(ctx, req)
				cancel()
				if err == nil || !shouldRetry(err) {
					break
				}
				bo.Wait()
			}
			if err == nil {
				s.pushOK.Add(1)
			} else {
				s.recordFailed(err)
			}
		}
	}
}

func shouldRetry(err error) bool {
	code := connect.CodeOf(err)
	switch code {
	case connect.CodeDeadlineExceeded,
		connect.CodeUnknown,
		connect.CodeResourceExhausted,
		connect.CodeInternal,
		connect.CodeUnavailable,
		connect.CodeDataLoss,
		connect.CodeAborted:
		return true
	default:

		return false
	}
}
