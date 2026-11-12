package profiling

import (
	"github.com/Kindling-project/kindling/collector/pkg/profiling/cgoapi"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/flamestats"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/traceagg"
)

func (s *Service) statsSources() flamestats.Sources {
	return flamestats.Sources{
		BPF: func() (flamestats.BPF, error) {
			m, err := cgoapi.ReadProfileMetrics()
			if err != nil {
				return flamestats.BPF{}, err
			}
			return flamestats.BPF{
				Samples:        m[cgoapi.MetricSamples],
				PerfOutputFail: m[cgoapi.MetricPerfOutputFail],
				ErrKernelPC:    m[cgoapi.MetricUnwindErrKernelPC],
				ErrNoMapping:   m[cgoapi.MetricUnwindErrNoMapping],
				ErrZeroPC:      m[cgoapi.MetricUnwindErrZeroPC],
				ErrFramesFull:  m[cgoapi.MetricUnwindErrFramesFull],
				PythonAttempts: m[cgoapi.MetricPythonAttempts], PythonFrames: m[cgoapi.MetricPythonFrames],
				PythonNoProcInfo: m[cgoapi.MetricPythonErrNoProcInfo], PythonThreadState: m[cgoapi.MetricPythonErrThreadState],
				PythonFrameRead: m[cgoapi.MetricPythonErrFrameRead], PythonCodeReadFail: m[cgoapi.MetricPythonCodeReadFail],
			}, nil
		},
		PIDSync: func() flamestats.PIDSync {
			return flamestats.PIDSync{
				PIDs: s.pidDrained.Load(),
			}
		},
		Sync: func() flamestats.Sync {
			m := s.pm.MappingStats()
			return flamestats.Sync{
				Attempts:       m.Attempts,
				ErrPerm:        m.ErrPerm,
				ErrNotExist:    m.ErrNotExist,
				ErrESRCH:       m.ErrESRCH,
				ParseErrors:    m.ParseErrors,
				TotalParseUsec: m.TotalParseUsec,
			}
		},
		Agg: func() flamestats.Agg {
			return flamestats.Agg{
				PeakStacks: s.agg.TakePeak(),
				MaxStacks:  traceagg.MaxEntries,
				Dropped:    s.agg.Dropped(),
			}
		},
		Sym: func() flamestats.Sym {
			st := s.sym.Stats()
			return flamestats.Sym{
				File:        st.File,
				Symtab:      st.Symtab,
				SymtabShort: st.SymtabShort,
				NoMapping:   st.NoMapping,
			}
		},
		Python: func() flamestats.Python {
			return flamestats.Python{OK: s.pythonSymbolOK.Load(), Failed: s.pythonSymbolFail.Load()}
		},
		Push: func() flamestats.Push {
			st := s.sender.Stats()
			return flamestats.Push{
				Enqueued:  st.Enqueued,
				QueueFull: st.QueueFull,
				PushOK:    st.PushOK,
				Retries:   st.Retries,
				Failed:    st.Failed,
			}
		},
	}
}
