package profiling

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	exporter "github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter/pyroscopeexporter"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/cgoapi"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/flamepid"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/flamestats"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter/interpreterconfig"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager"
	pmebpf "github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager/ebpf"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager/ebpfapi"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/symbolizer"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/traceagg"
	"go.uber.org/zap"
)

// Logger is the subset Service needs (matches component.TelemetryLogger).
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
}

// Service owns CPU flame-graph profiling.
// New builds everything. Start/Stop is the lifecycle. Mapping sync follows pid_events.
type Service struct {
	cfg      Config
	log      Logger
	pm       *processmanager.ProcessManager
	sym      *symbolizer.Symbolizer
	agg      *traceagg.Aggregator
	sender   *exporter.AsyncSender
	stopCh   chan struct{}
	wg       sync.WaitGroup
	cancel   context.CancelFunc
	stopOnce sync.Once

	flame *flamepid.Switch
	stats *flamestats.Reporter

	pidDrained                       atomic.Uint64
	pythonSymbolOK, pythonSymbolFail atomic.Uint64
}

// New constructs ProcessManager, aggregator, BCC, and the Pyroscope sender.
// Call after libkindling probe is up. Does not start goroutines.
func New(cfg Config, log Logger) (*Service, error) {
	cfg.applyDefaults()
	if err := cfg.validateTargets(); err != nil {
		return nil, err
	}
	if cfg.EndpointURL == "" {
		return nil, fmt.Errorf("profiling endpoint url is empty")
	}
	if err := cgoapi.InitCPUProfiling(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	sender := exporter.NewAsyncSender(
		cfg.EndpointURL, cfg.RemoteTimeout,
		cfg.SenderWorkerCount, cfg.SenderQueueCapacity,
		cfg.MinBackoff, cfg.MaxBackoff, cfg.MaxBackoffRetries,
	)
	if err := cgoapi.InitSymbolizer(); err != nil {
		sender.Stop()
		cancel()
		cgoapi.ShutdownCPUProfiling()
		return nil, err
	}
	// Collector and probe are deployed together. Register both loaders by default;
	// target PIDs are controlled by the shared CPU profiling switch.
	interps := interpreterconfig.Config{}
	pm, err := processmanager.New(ctx, processmanager.Config{
		InterpretersConfig: interps,
		EbpfHandler:        ebpfapi.NewAdapter(pmebpf.NewMaps()),
		StackDeltaProvider: ebpfapi.NewStackDeltaProvider(),
		// Go names are looked up by fileID after the flush, possibly after the
		// process exited; the entry has to outlive the aggregation interval.
		ExecutableUnloadDelay: max(time.Minute, 2*cfg.Interval),
	})
	if err != nil {
		cgoapi.ShutdownSymbolizer()
		sender.Stop()
		cancel()
		cgoapi.ShutdownCPUProfiling()
		return nil, err
	}
	s := &Service{
		cfg:    cfg,
		cancel: cancel,
		log:    log,
		pm:     pm,
		agg:    traceagg.New(),
		sender: sender,
		stopCh: make(chan struct{}),
	}
	s.sym = symbolizer.New(pm, pm, cgoapi.SymtabSymbolizer{})
	s.flame = flamepid.New(s.flameHooks())
	s.stats = flamestats.New(cfg.Metrics, s.statsSources(), log)
	return s, nil
}

// Start trace polling, periodic export and process synchronization.
func (s *Service) Start() {
	s.flame.Start()
	s.applyLocalTargets()
	if s.cfg.captureAll() {
		cgoapi.StartCPUSamplingAll()
		s.log.Info("CPU profiling: CAPTURE_ALL (pids: [0])")
	} else if len(s.cfg.Pids) > 0 {
		cgoapi.StartCPUSampling()
	}
	s.log.Info("native CPU sampling enabled")
	s.wg.Add(3)
	go func() {
		defer s.wg.Done()
		s.pollCPULoop()
	}()
	go func() {
		defer s.wg.Done()
		s.sendCPULoop()
	}()

	go func() {
		defer s.wg.Done()
		s.pidSyncLoop()
	}()
	s.stats.Start()
}

// Stop ends loops, then closes sender and BCC.
func (s *Service) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() {
		s.stats.Stop()
		if s.flame != nil {
			s.flame.Stop()
			s.flame.ApplyCPU(flamepid.Targets{})
		}
		cgoapi.StopCPUSampling()
		select {
		case <-s.stopCh:
		default:
			close(s.stopCh)
		}
		s.wg.Wait()
		s.cancel()
		s.pm.Close()
		cgoapi.ShutdownCPUProfiling()
		if s.sender != nil {
			s.sender.Stop()
		}
		cgoapi.ShutdownSymbolizer()
	})
}
