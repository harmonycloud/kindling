// Package flamestats periodically logs flamegraph self-monitoring counters.
//
// Counting always runs at the data owners (BPF, pid sync, aggregator,
// symbolizer, sender).
// This package only snapshots them through Sources, diffs against the
// previous snapshot and prints one line per section as cumulative(+delta).
// Printing is off unless Config.Enable.
package flamestats

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

const defaultInterval = 60 * time.Second

// Config controls printing only. Counters are kept either way.
type Config struct {
	Enable   bool          `mapstructure:"enable"`
	Interval time.Duration `mapstructure:"interval"`
}

// BPF is the CPU-summed profile_metrics map. Cumulative.
type BPF struct {
	Samples            uint64
	PerfOutputFail     uint64
	ErrKernelPC        uint64
	ErrNoMapping       uint64
	ErrZeroPC          uint64
	ErrFramesFull      uint64
	PythonAttempts     uint64
	PythonFrames       uint64
	PythonNoProcInfo   uint64
	PythonThreadState  uint64
	PythonFrameRead    uint64
	PythonCodeReadFail uint64
}

// PIDSync counts pids drained by pidSyncLoop. Cumulative.
type PIDSync struct {
	PIDs uint64
}

// Sync mirrors processmanager.MappingStats. Cumulative. Attempts only feeds
// avg_parse_us.
type Sync struct {
	Attempts       uint32
	ErrPerm        uint32
	ErrNotExist    uint32
	ErrESRCH       uint32
	ParseErrors    uint32
	TotalParseUsec uint32
}

// Agg is the CPU trace aggregator. PeakStacks is the largest single flush
// since the previous read, MaxStacks the per-flush cap, Dropped cumulative.
type Agg struct {
	PeakStacks int
	MaxStacks  int
	Dropped    uint64
}

// Sym mirrors symbolizer.Stats. Cumulative, in frames unless noted.
type Sym struct {
	File        uint64
	Symtab      uint64
	SymtabShort uint64
	NoMapping   uint64
}

// Python counts early symbolization requests for newly admitted stacks. Cached
// object lookups are included in OK; repeated whole stacks do not request names.
type Python struct{ OK, Failed uint64 }

// Push mirrors pyroscopeexporter.SenderStats. Cumulative.
type Push struct {
	Enqueued  uint64
	QueueFull uint64
	PushOK    uint64
	Retries   uint64
	Failed    map[string]uint64
}

// Sources are read once per interval. A nil source is skipped.
type Sources struct {
	BPF     func() (BPF, error)
	PIDSync func() PIDSync
	Sync    func() Sync
	Agg     func() Agg
	Sym     func() Sym
	Python  func() Python
	Push    func() Push
}

// Logger is the subset Reporter needs.
type Logger interface {
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
}

type snapshot struct {
	bpf    BPF
	bpfOK  bool
	pid    PIDSync
	sync   Sync
	agg    Agg
	sym    Sym
	python Python
	push   Push
}

// Reporter prints counter deltas every Config.Interval.
type Reporter struct {
	cfg Config
	src Sources
	log Logger

	prev      snapshot
	bpfWarned bool

	stop chan struct{}
	wg   sync.WaitGroup
}

func New(cfg Config, src Sources, log Logger) *Reporter {
	if cfg.Interval <= 0 {
		cfg.Interval = defaultInterval
	}
	return &Reporter{cfg: cfg, src: src, log: log, stop: make(chan struct{})}
}

// Start does nothing when printing is disabled.
func (r *Reporter) Start() {
	if r == nil || !r.cfg.Enable {
		return
	}
	r.prev = r.collect()
	r.wg.Add(1)
	go r.loop()
}

func (r *Reporter) Stop() {
	if r == nil {
		return
	}
	select {
	case <-r.stop:
	default:
		close(r.stop)
	}
	r.wg.Wait()
}

func (r *Reporter) loop() {
	defer r.wg.Done()
	ticker := time.NewTicker(r.cfg.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-r.stop:
			return
		case <-ticker.C:
			cur := r.collect()
			prefix := "flamegraph stats(" + r.cfg.Interval.String() + ") "
			for _, l := range r.lines(r.prev, cur) {
				r.log.Info(prefix + l)
			}
			r.prev = cur
		}
	}
}

func (r *Reporter) collect() snapshot {
	var s snapshot
	if r.src.BPF != nil {
		b, err := r.src.BPF()
		if err != nil {
			if !r.bpfWarned {
				r.log.Warn("flamegraph stats: read profile_metrics failed, probe.o may be too old", zap.Error(err))
				r.bpfWarned = true
			}
		} else {
			s.bpf, s.bpfOK = b, true
		}
	}
	if r.src.PIDSync != nil {
		s.pid = r.src.PIDSync()
	}
	if r.src.Sync != nil {
		s.sync = r.src.Sync()
	}
	if r.src.Agg != nil {
		s.agg = r.src.Agg()
	}
	if r.src.Sym != nil {
		s.sym = r.src.Sym()
	}
	if r.src.Python != nil {
		s.python = r.src.Python()
	}
	if r.src.Push != nil {
		s.push = r.src.Push()
	}
	return s
}

// lines turns two snapshots into one log line per section. Counters print as
// cumulative(+this interval).
func (r *Reporter) lines(prev, cur snapshot) []string {
	var out []string
	add := func(section string, kvs ...string) {
		out = append(out, "["+section+"] "+strings.Join(kvs, " "))
	}

	if cur.bpfOK {
		p := prev.bpf
		if !prev.bpfOK {
			p = BPF{}
		}
		c := cur.bpf
		add("bpf",

			kv("samples", cum64(p.Samples, c.Samples)),

			kv("perf_output_fail", cum64(p.PerfOutputFail, c.PerfOutputFail)),

			kv("err_kernel_pc", cum64(p.ErrKernelPC, c.ErrKernelPC)),

			kv("err_no_mapping", cum64(p.ErrNoMapping, c.ErrNoMapping)),

			kv("err_zero_pc", cum64(p.ErrZeroPC, c.ErrZeroPC)),

			kv("err_frames_full", cum64(p.ErrFramesFull, c.ErrFramesFull)),
		)
	} else if r.src.BPF != nil {
		add("bpf", "n/a")
	}

	if r.src.PIDSync != nil || r.src.Sync != nil {
		var kvs []string
		if r.src.PIDSync != nil {
			p, c := prev.pid, cur.pid
			kvs = append(kvs,

				kv("pids", cum64(p.PIDs, c.PIDs)),
			)
		}
		if r.src.Sync != nil {
			p, c := prev.sync, cur.sync
			attempts := d32(p.Attempts, c.Attempts)
			var avg uint32
			if attempts > 0 {
				avg = d32(p.TotalParseUsec, c.TotalParseUsec) / attempts
			}
			kvs = append(kvs,

				kv("err_perm", cum32(p.ErrPerm, c.ErrPerm)),

				kv("err_not_exist", cum32(p.ErrNotExist, c.ErrNotExist)),

				kv("err_esrch", cum32(p.ErrESRCH, c.ErrESRCH)),

				kv("parse_errors", cum32(p.ParseErrors, c.ParseErrors)),

				kv("avg_parse_us", avg),
			)
		}
		add("sync", kvs...)
	}

	if r.src.Agg != nil {
		p, c := prev.agg, cur.agg
		add("agg",

			kv("peak_stacks", strconv.Itoa(c.PeakStacks)+"/"+strconv.Itoa(c.MaxStacks)),

			kv("dropped", cum64(p.Dropped, c.Dropped)),
		)
	}

	if r.src.Sym != nil {
		p, c := prev.sym, cur.sym
		add("sym",

			kv("go_ok", cum64(p.File, c.File)),

			kv("symtab", cum64(p.Symtab, c.Symtab)),

			kv("symtab_short", cum64(p.SymtabShort, c.SymtabShort)),

			kv("no_mapping", cum64(p.NoMapping, c.NoMapping)),
		)
	}

	if r.src.Python != nil {
		p, c := prev.bpf, cur.bpf
		fields := []string{}
		if cur.bpfOK {
			if !prev.bpfOK {
				p = BPF{}
			}
			fields = append(fields,

				kv("attempts", cum64(p.PythonAttempts, c.PythonAttempts)),

				kv("frames", cum64(p.PythonFrames, c.PythonFrames)),

				kv("err_no_proc", cum64(p.PythonNoProcInfo, c.PythonNoProcInfo)),

				kv("err_tstate", cum64(p.PythonThreadState, c.PythonThreadState)),

				kv("err_frame", cum64(p.PythonFrameRead, c.PythonFrameRead)),

				kv("code_read_fail", cum64(p.PythonCodeReadFail, c.PythonCodeReadFail)),
			)
		} else {
			fields = append(fields, "bpf=n/a")
		}
		fields = append(fields,

			kv("sym_ok", cum64(prev.python.OK, cur.python.OK)),
			kv("sym_fail", cum64(prev.python.Failed, cur.python.Failed)),
		)
		add("py", fields...)
	}

	if r.src.Push != nil {
		p, c := prev.push, cur.push
		failed := failedCum(p.Failed, c.Failed)
		add("push",

			kv("enqueued", cum64(p.Enqueued, c.Enqueued)),

			kv("queue_full", cum64(p.QueueFull, c.QueueFull)),

			kv("ok", cum64(p.PushOK, c.PushOK)),

			kv("retries", cum64(p.Retries, c.Retries)),

			kv("failed", failed),
		)
	}

	return out
}

func kv(k string, v interface{}) string {
	return k + "=" + fmt.Sprint(v)
}

// d64 is cur-prev. A smaller cur means the source restarted (probe reload).
func d64(prev, cur uint64) uint64 {
	if cur < prev {
		return cur
	}
	return cur - prev
}

// d32 relies on uint32 wraparound; TotalParseUsec overflows after ~71 min of parsing.
func d32(prev, cur uint32) uint32 {
	return cur - prev
}

func cum64(prev, cur uint64) string {
	return strconv.FormatUint(cur, 10) + "(+" + strconv.FormatUint(d64(prev, cur), 10) + ")"
}

func cum32(prev, cur uint32) string {
	return strconv.FormatUint(uint64(cur), 10) + "(+" + strconv.FormatUint(uint64(d32(prev, cur)), 10) + ")"
}

// failedCum prints code:cumulative(+delta) for every code seen, or "-".
func failedCum(prev, cur map[string]uint64) string {
	if len(cur) == 0 {
		return "-"
	}
	keys := make([]string, 0, len(cur))
	for k := range cur {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = k + ":" + cum64(prev[k], cur[k])
	}
	return strings.Join(parts, ",")
}
