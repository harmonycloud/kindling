package profiling

import (
	"runtime"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/ebpf/pprof"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/cgoapi"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/framedecode"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/symbolizer"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/traceagg"
	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.opentelemetry.io/ebpf-profiler/times"
	"go.uber.org/zap"
)

const pollBatch = 64

func (s *Service) pollCPULoop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	poller := cgoapi.NewTracePoller(pollBatch)
	for {
		select {
		case <-s.stopCh:
			return
		default:
		}
		samples, n := poller.Poll()
		if n <= 0 {
			if n < 0 {
				s.log.Debug("pollProfileTraces", zap.Int("ret", n))
			}
			// Empty ring: MiniSinsp/scap_next already usleeps in
			// refill_read_buffers (500us … 30ms).
			continue
		}
		for i := 0; i < n; i++ {
			s.addCPUSample(&samples[i])
		}
	}
}

func (s *Service) addCPUSample(sample *cgoapi.TraceSample) {
	if sample.PID == 0 {
		return
	}
	words := len(sample.Frames)
	kframes := int(sample.NumKernelFrames)
	if kframes > words {
		kframes = words
	}
	userWords := sample.Frames[kframes:]
	if len(userWords) == 0 && kframes == 0 {
		return
	}
	frames := make([]traceagg.Frame, 0, kframes+int(sample.NumFrames))
	// Leading words are raw kernel PCs (fileID 0). The symbolizer routes
	// them to the kernel symbolizer via IsKernelPC.
	for i := 0; i < kframes; i++ {
		frames = append(frames, traceagg.NewKernelFrame(sample.Frames[i]))
	}
	if len(userWords) > 0 {
		err := framedecode.WalkUserFrames(userWords, int(sample.NumFrames), func(fr framedecode.RawFrame) {
			switch fr.Kind() {
			case framedecode.MarkerNative:
				frames = append(frames, traceagg.NewNativeFrame(
					fr.Variable(0), fr.Data(), fr.Flags()))
			case framedecode.MarkerPython:
				packed := fr.Variable(1)
				frames = append(frames, traceagg.NewPythonFrame(
					fr.Variable(0), uint32(packed>>32), uint32(packed), fr.Flags()))
			default:
				// The generic decoder has already advanced over this complete
				// frame. Add a typed conversion here when the runtime is supported.
			}
		})
		if err != nil {
			s.log.Debug("decode user frames",
				zap.Uint32("pid", sample.PID), zap.Error(err))
			return
		}
	}
	if len(frames) == 0 {
		return
	}
	s.aggregateCPUFrames(sample.PID, frames, sample.Comm)
}

// aggregateCPUFrames resolves only Python frames of a newly admitted stack.
// Existing stacks count immediately; native symbols remain batched at flush.
func (s *Service) aggregateCPUFrames(pid uint32, frames []traceagg.Frame, comm string) {
	hasPython := false
	for _, f := range frames {
		if f.Kind() == traceagg.KindPython {
			hasPython = true
			break
		}
	}
	if !hasPython {
		s.agg.Add(pid, frames, comm)
		return
	}
	s.agg.AddWithNames(pid, frames, comm, func() []string {
		names := make([]string, len(frames))
		for i, f := range frames {
			py, ok := f.Python()
			if !ok {
				continue
			}
			name, found := s.pm.SymbolizeInterp(libpf.PID(pid), py.CodeObject, uint64(py.Checksum))
			if found {
				names[i] = name
				s.pythonSymbolOK.Add(1)
			} else {
				names[i] = symbolizer.UnknownName
				s.pythonSymbolFail.Add(1)
			}
		}
		return names
	})
}

func (s *Service) sendCPULoop() {
	for {
		timer := time.NewTimer(s.cfg.Interval)
		select {
		case <-s.stopCh:
			timer.Stop()
			s.log.Info("Periodic native CPU sampling stopped")
			return
		case <-timer.C:
		}
		start := time.Now()
		// Pids dropped before this point have all their samples in this flush,
		// so their mappings can go once it is done.
		ktime := times.GetKTime()
		if err := s.sendCPUProfile(); err != nil {
			s.log.Error("Failed native CPU profile flush", zap.Error(err))
		}
		s.pm.ProcessedUntil(ktime)
		cgoapi.ExpireSymbolizerCaches(3 * s.cfg.Interval)
		s.log.Info("Native CPU FlamegraphSampling once cost",
			zap.Duration("cost", time.Since(start)))
	}
}

func (s *Service) sendCPUProfile() error {
	table := s.agg.Swap()
	if len(table) == 0 {
		s.log.Info("native cpu: no aggregated traces")
		return nil
	}
	builders := pprof.NewProfileBuilders(97)
	pidSet := make(map[int64]struct{}, len(table))
	for key, entry := range table {
		if entry == nil {
			continue
		}
		stack, err := s.sym.Symbolize(key.PID, entry)
		if err != nil || len(stack) == 0 {
			if err != nil {
				s.log.Debug("symbolize native stack",
					zap.Uint32("pid", key.PID), zap.Error(err))
			}
			continue
		}
		pidSet[int64(key.PID)] = struct{}{}
		builders.BuilderForTarget(uint64(key.PID), nil).AddSample(stack, entry.Count)
	}
	s.pm.ReleaseResources()
	s.log.Info("Native CPU profiling",
		zap.Int("stacks", len(table)),
		zap.Int("builders", len(builders.Builders)))
	return s.sendPprof(builders, "process_cpu", pidSet)
}
