package cgoreceiver

import (
	"github.com/Kindling-project/kindling/collector/pkg/profiling"
	"go.uber.org/zap"
)

func (r *CgoReceiver) startProfiling() {
	cfg := r.cfg.EbpfProfiling
	if cfg.Endpoint.URL == "" || len(cfg.Pids) == 0 {
		return
	}
	service, err := profiling.New(profiling.Config{
		EndpointURL: cfg.Endpoint.URL, RemoteTimeout: cfg.Endpoint.RemoteTimeout,
		MinBackoff: cfg.Endpoint.MinBackoff, MaxBackoff: cfg.Endpoint.MaxBackoff,
		MaxBackoffRetries: cfg.Endpoint.MaxBackoffRetries,
		SenderWorkerCount: cfg.SenderWorkerCount, SenderQueueCapacity: cfg.SenderQueueCapacity,
		Pids: cfg.Pids, Metrics: cfg.Metrics,
	}, r.telemetry.Logger)
	if err != nil {
		r.telemetry.Logger.Error("Failed to initialize CPU profiling", zap.Error(err))
		return
	}
	r.profiling = service
	service.Start()
}
