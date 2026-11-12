package flamestats

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
)

type nopLogger struct{ warns int }

func (l *nopLogger) Info(string, ...zap.Field) {}
func (l *nopLogger) Warn(string, ...zap.Field) { l.warns++ }

func TestFieldsAreDeltas(t *testing.T) {
	r := New(Config{Enable: true, Interval: time.Second}, Sources{
		BPF:     func() (BPF, error) { return BPF{}, nil },
		PIDSync: func() PIDSync { return PIDSync{} },
		Sync:    func() Sync { return Sync{} },
		Agg:     func() Agg { return Agg{} },
		Sym:     func() Sym { return Sym{} },
		Push:    func() Push { return Push{} },
	}, &nopLogger{})

	prev := snapshot{
		bpf:   BPF{Samples: 100, ErrNoMapping: 5},
		bpfOK: true,
		pid:   PIDSync{PIDs: 3},
		sync:  Sync{Attempts: 2, TotalParseUsec: 100},
		agg:   Agg{PeakStacks: 900, MaxStacks: 100000, Dropped: 10},
		sym:   Sym{File: 50, Symtab: 20, NoMapping: 1},
		push:  Push{Enqueued: 2, PushOK: 2, Failed: map[string]uint64{"unavailable": 1}},
	}
	cur := snapshot{
		bpf:   BPF{Samples: 160, PerfOutputFail: 3, ErrNoMapping: 9},
		bpfOK: true,
		pid:   PIDSync{PIDs: 10},
		sync:  Sync{Attempts: 6, TotalParseUsec: 500},
		agg:   Agg{PeakStacks: 1200, MaxStacks: 100000, Dropped: 25},
		sym:   Sym{File: 80, Symtab: 35, SymtabShort: 2, NoMapping: 1},
		push:  Push{Enqueued: 5, PushOK: 3, Retries: 4, Failed: map[string]uint64{"unavailable": 3, "deadline_exceeded": 1}},
	}
	got := r.lines(prev, cur)
	want := []string{
		"[bpf] samples=160(+60) perf_output_fail=3(+3) err_kernel_pc=0(+0) err_no_mapping=9(+4) err_zero_pc=0(+0) err_frames_full=0(+0)",
		"[sync] pids=10(+7) err_perm=0(+0) err_not_exist=0(+0) err_esrch=0(+0) parse_errors=0(+0) avg_parse_us=100",
		"[agg] peak_stacks=1200/100000 dropped=25(+15)",
		"[sym] go_ok=80(+30) symtab=35(+15) symtab_short=2(+2) no_mapping=1(+0)",
		"[push] enqueued=5(+3) queue_full=0(+0) ok=3(+1) retries=4(+4) failed=deadline_exceeded:1(+1),unavailable:3(+2)",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d lines, want %d: %q", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d\n got: %s\nwant: %s", i, got[i], want[i])
		}
	}
}

func TestBPFUnavailableWarnsOnce(t *testing.T) {
	log := &nopLogger{}
	r := New(Config{Enable: true}, Sources{
		BPF: func() (BPF, error) { return BPF{}, errors.New("no map") },
	}, log)
	s1 := r.collect()
	r.collect()
	if log.warns != 1 {
		t.Fatalf("warns = %d, want 1", log.warns)
	}
	if got := r.lines(snapshot{}, s1); len(got) != 1 || got[0] != "[bpf] n/a" {
		t.Fatalf("lines = %q, want [bpf] n/a", got)
	}
}

func TestDisabledDoesNotStart(t *testing.T) {
	called := false
	r := New(Config{}, Sources{Python: func() Python { called = true; return Python{} }}, &nopLogger{})
	r.Start()
	r.Stop()
	if called {
		t.Fatal("disabled reporter collected")
	}
}

func TestD64SourceRestart(t *testing.T) {
	if got := d64(100, 7); got != 7 {
		t.Fatalf("d64 restart = %d, want 7", got)
	}
}

func TestPythonCountersAreCumulative(t *testing.T) {
	r := New(Config{}, Sources{Python: func() Python { return Python{} }}, &nopLogger{})
	prev := snapshot{bpfOK: true, bpf: BPF{PythonFrames: 10}, python: Python{OK: 3, Failed: 1}}
	cur := snapshot{bpfOK: true, bpf: BPF{PythonAttempts: 2, PythonFrames: 15, PythonNoProcInfo: 1}, python: Python{OK: 5, Failed: 2}}
	lines := r.lines(prev, cur)
	want := "[py] attempts=2(+2) frames=15(+5) err_no_proc=1(+1) err_tstate=0(+0) err_frame=0(+0) code_read_fail=0(+0) sym_ok=5(+2) sym_fail=2(+1)"
	if len(lines) != 2 || lines[1] != want {
		t.Fatalf("%q", lines)
	}
}
