package profiling

import (
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/cgoapi"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/flamepid"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/process"
	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.uber.org/zap"
)

func (s *Service) flameHooks() flamepid.Hooks {
	return flamepid.Hooks{
		EnableNative:  func(pid uint32) { cgoapi.EnableNativeProfile(pid) },
		DisableNative: func(pid uint32) { cgoapi.DisableNativeProfile(pid) },

		DropProcess: s.dropPID,
	}
}

func (s *Service) applyLocalTargets() {
	pids := s.cfg.localPIDs()
	if len(pids) == 0 || s.flame == nil {
		return
	}
	s.log.Info("configured CPU profiling targets", zap.Uint32s("pids", pids))
	s.flame.ApplyCPU(flamepid.Targets{PIDs: pids})
}

// pidSyncLoop drains profile_pid_events and synchronizes one pid at a time.
// It is not the switch's 30s container-table loop, and it is not the trace poll loop.
func (s *Service) pidSyncLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	warned := false
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			pids, err := cgoapi.DrainPidEvents()
			if err != nil {
				if !warned {
					s.log.Debug("drain profile_pid_events", zap.Error(err))
					warned = true
				}
				continue
			}
			if n := uint64(len(pids)); n > 0 {
				s.pidDrained.Add(n)
			}
			for _, pid := range pids {
				if pid == 0 || s.pm == nil {
					continue
				}
				s.pm.SynchronizeProcess(process.New(libpf.PID(pid), libpf.PID(pid)))
			}
		}
	}
}

// dropPID runs when pid leaves the native roster. The process manager keeps
// its mappings until ProcessedUntil after the next CPU flush.
func (s *Service) dropPID(pid uint32) {
	if pid == 0 || s.pm == nil {
		return
	}
	s.pm.DropPID(libpf.PID(pid))
	_ = cgoapi.DeleteReportedPID(pid)
}
