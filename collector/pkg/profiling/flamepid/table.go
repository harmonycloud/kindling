package flamepid

import (
	"os"
	"strconv"
	"sync"
)

// Snapshot is container id (12 hex chars) to main-thread pids, and the reverse.
type Snapshot struct {
	ByContainer map[string][]uint32
	ByPID       map[uint32]string
}

// Table is refreshed off the event thread. Current is false until the first successful scan.
type Table struct {
	root string

	mu    sync.RWMutex
	snap  Snapshot
	ready bool
}

func newTable(procRoot string) *Table {
	if procRoot == "" {
		procRoot = "/proc"
	}
	return &Table{root: procRoot}
}

func (t *Table) Refresh() bool {
	snap, ok := scanProc(t.root)
	if !ok {
		return false
	}
	t.mu.Lock()
	t.snap = snap
	t.ready = true
	t.mu.Unlock()
	return true
}

func (t *Table) Current() (Snapshot, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if !t.ready {
		return Snapshot{}, false
	}
	return t.snap, true
}

func (t *Table) ContainerID(pid uint32) string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if !t.ready {
		return ""
	}
	return t.snap.ByPID[pid]
}

func scanProc(root string) (Snapshot, bool) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return Snapshot{}, false
	}
	snap := Snapshot{
		ByContainer: make(map[string][]uint32),
		ByPID:       make(map[uint32]string),
	}
	for _, ent := range entries {
		if !ent.IsDir() {
			continue
		}
		n, err := strconv.ParseUint(ent.Name(), 10, 32)
		if err != nil || n == 0 {
			continue
		}
		pid := uint32(n)
		raw, err := os.ReadFile(procPath(root, pid, "cgroup"))
		if err != nil {
			continue
		}
		id := containerIDFromCgroupFile(string(raw))
		if id == "" {
			continue
		}
		snap.ByContainer[id] = append(snap.ByContainer[id], pid)
		snap.ByPID[pid] = id
	}
	return snap, true
}
