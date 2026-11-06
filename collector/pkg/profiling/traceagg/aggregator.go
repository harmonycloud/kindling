package traceagg

import (
	"encoding/binary"
	"hash/fnv"
	"sync"
	"sync/atomic"
)

// MaxEntries caps distinct stacks per interval. This is an entry cap, not a
// byte budget: memory also depends on stack depth and retained interpreter names.
const MaxEntries = 100000

// Add and AddWithNames are called by a single sampling goroutine. Swap and
// metric reads may run concurrently with that writer.
type Aggregator struct {
	mu  sync.Mutex
	m   map[Key]*Entry
	max int

	dropped atomic.Uint64
	peak    atomic.Int64
}

// New returns an empty Aggregator.
func New() *Aggregator {
	return &Aggregator{m: make(map[Key]*Entry), max: MaxEntries}
}

func hashFrames(frames []Frame) (hi, lo uint64) {
	h := fnv.New128a()
	var buf [18]byte
	for _, f := range frames {
		binary.LittleEndian.PutUint64(buf[0:8], f.payload0)
		payload1 := f.payload1
		if f.kind == KindPython {
			// Function-level profiles intentionally merge different bytecode
			// offsets. The full lasti remains in the retained first-seen frame,
			// so enabling line-level aggregation only changes this policy.
			payload1 = uint64(uint32(payload1))
		}
		binary.LittleEndian.PutUint64(buf[8:16], payload1)
		buf[16] = byte(f.kind)
		buf[17] = f.flags
		_, _ = h.Write(buf[:])
	}
	sum := h.Sum(make([]byte, 0, 16))
	hi = binary.BigEndian.Uint64(sum[0:8])
	lo = binary.BigEndian.Uint64(sum[8:16])
	return hi, lo
}

// Add increments pid+whole-stack without interpreter names.
func (a *Aggregator) Add(pid uint32, frames []Frame, comm string) {
	a.AddWithNames(pid, frames, comm, nil)
}

// AddWithNames calls resolve only for a new, admitted stack. Remote reads run
// outside the aggregation lock. A concurrent Swap cannot create an entry that
// lacks names: insertion and the owned name slice are published together.
func (a *Aggregator) AddWithNames(pid uint32, frames []Frame, comm string, resolve func() []string) {
	if len(frames) == 0 {
		return
	}
	hi, lo := hashFrames(frames)
	key := Key{PID: pid, HashHi: hi, HashLo: lo}
	a.mu.Lock()
	if e, ok := a.m[key]; ok {
		e.Count++
		a.mu.Unlock()
		return
	}
	if len(a.m) >= a.max {
		a.dropped.Add(1)
		a.mu.Unlock()
		return
	}
	// Native samples need no remote lookup and can insert under this lock.
	if resolve == nil {
		a.m[key] = &Entry{Frames: append([]Frame(nil), frames...), Comm: comm, Count: 1}
		a.mu.Unlock()
		return
	}
	a.mu.Unlock()
	names := resolve()
	a.mu.Lock()
	defer a.mu.Unlock()
	var ownedNames []string
	if len(names) == len(frames) {
		ownedNames = append([]string(nil), names...)
	}
	a.m[key] = &Entry{Frames: append([]Frame(nil), frames...), InterpNames: ownedNames, Comm: comm, Count: 1}
}

// Swap replaces the live map with an empty one and returns the previous table.
// The sampling goroutine can keep adding to the new table while the caller
// processes the returned table without holding the aggregation lock.
func (a *Aggregator) Swap() map[Key]*Entry {
	a.mu.Lock()
	defer a.mu.Unlock()
	old := a.m
	a.m = make(map[Key]*Entry)
	if n := int64(len(old)); n > a.peak.Load() {
		a.peak.Store(n)
	}
	return old
}

// Len returns the number of distinct stacks currently aggregated (for tests/metrics).
func (a *Aggregator) Len() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.m)
}

// Dropped returns the cumulative number of samples dropped by the entry cap.
func (a *Aggregator) Dropped() uint64 {
	return a.dropped.Load()
}

// TakePeak returns the largest swapped table size since the previous call and
// resets it. Meant for a single periodic reader.
func (a *Aggregator) TakePeak() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return int(a.peak.Swap(0))
}
