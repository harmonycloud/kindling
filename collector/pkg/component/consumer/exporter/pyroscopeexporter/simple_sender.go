package pyroscopeexporter

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/grafana/dskit/backoff"
	pushv1 "github.com/grafana/pyroscope/api/gen/proto/go/push/v1"
	"github.com/grafana/pyroscope/api/gen/proto/go/push/v1/pushv1connect"
	typesv1 "github.com/grafana/pyroscope/api/gen/proto/go/types/v1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
)

// RawSample holds a gzip-compressed pprof payload.
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
		q:          make(chan *connect.Request[pushv1.PushRequest], queueCap),
		client:     c,
		timeout:    timeout,
		minBackoff: minB,
		maxBackoff: maxB,
		maxRetries: maxRetries,
		workers:    workers,
		ctx:        ctx,
		cancel:     cancel,
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
		return nil
	default:
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
			for bo.Ongoing() {
				ctx, cancel := context.WithTimeout(s.ctx, s.timeout)
				_, err := s.client.Push(ctx, req)
				cancel()
				if err == nil || !shouldRetry(err) {
					break
				}
				bo.Wait()
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
