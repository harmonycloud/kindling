package profiling

import (
	"bytes"
	"fmt"
	"os"
	"strconv"

	exporter "github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter/pyroscopeexporter"
	"github.com/Kindling-project/kindling/collector/pkg/ebpf/pprof"
	pushv1 "github.com/grafana/pyroscope/api/gen/proto/go/push/v1"
	"github.com/prometheus/prometheus/model/labels"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func (s *Service) sendPprof(builders *pprof.ProfileBuilders, profileName string, pidSet map[int64]struct{}) error {
	labeler := newLabeler(profileName, s.flame.ContainerID)
	perPidLabels, err := labeler.build(pidSet)
	if err != nil {
		s.log.Error("BuildLabelsForPIDs err: ", zap.Error(err))
		return nil
	}
	return sendPprofSeries(s.sender, builders, perPidLabels)
}

type labeler struct {
	base        labels.Labels
	containerID func(uint32) string
}

func newLabeler(profileName string, containerID func(uint32) string) *labeler {
	return &labeler{containerID: containerID, base: labels.Labels{
		{Name: "__name__", Value: profileName}, {Name: "__delta__", Value: "false"},
		{Name: "service_name", Value: "kindling"}, {Name: "host_ip", Value: os.Getenv("MY_NODE_IP")},
	}}
}

func (l *labeler) build(pidSet map[int64]struct{}) (map[int64]labels.Labels, error) {
	result := make(map[int64]labels.Labels, len(pidSet))
	pids := make([]int64, 0, len(pidSet))
	for pid := range pidSet {
		pids = append(pids, pid)
	}
	for _, pid := range pids {
		containerID := ""
		if pid > 0 && l.containerID != nil {
			containerID = l.containerID(uint32(pid))
		}
		values := append(labels.Labels(nil), l.base...)
		values = append(values, labels.Label{Name: "pid", Value: strconv.FormatInt(pid, 10)}, labels.Label{Name: "container_id", Value: containerID})
		result[pid] = values
	}
	return result, nil
}

func sendPprofSeries(sender *exporter.AsyncSender, builders *pprof.ProfileBuilders, perPidLabels map[int64]labels.Labels) error {
	const (
		maxSeriesPerReq = 80
		maxBytesPerReq  = 1 << 17
	)
	curBytes := 0
	protoSeries := make([]*pushv1.RawProfileSeries, 0, maxSeriesPerReq)
	flush := func() {
		if len(protoSeries) == 0 {
			return
		}
		cp := make([]*pushv1.RawProfileSeries, len(protoSeries))
		copy(cp, protoSeries)
		_ = exporter.BulkSend(cp, sender)
		protoSeries = protoSeries[:0]
		curBytes = 0
	}
	for pid, builder := range builders.Builders {
		buf := bytes.NewBuffer(nil)
		if _, err := builder.Write(buf); err != nil {
			return fmt.Errorf("pprof encode %w", err)
		}
		lbs := append(labels.Labels(nil), perPidLabels[int64(pid)]...)
		protoLabels, rawSample, _ := exporter.PrepareDataForTransmission(lbs, []*exporter.RawSample{
			{RawProfile: buf.Bytes()},
		})
		if len(protoSeries) >= maxSeriesPerReq || curBytes >= maxBytesPerReq {
			flush()
		}
		series := &pushv1.RawProfileSeries{Labels: protoLabels, Samples: rawSample}
		protoSeries = append(protoSeries, series)
		curBytes += proto.Size(series)
	}
	flush()
	return nil
}
