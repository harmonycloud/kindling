package flamepid

import (
	"strings"
	"sync"
	"time"
)

const refreshInterval = 30 * time.Second

// Targets selects the processes to profile.
// ContainerIDs and PIDs are not combined: container ids use the table, pids are explicit.
// Both empty means the switch is off.
type Targets struct {
	ContainerIDs []string
	PIDs         []uint32
}

// Hooks are the actions Switch calls after the diff. Nil hooks are skipped.
type Hooks struct {
	EnableNative  func(pid uint32)
	DisableNative func(pid uint32)
	DropProcess   func(pid uint32)
}

// Switch owns the CPU target roster.
type Switch struct {
	hooks Hooks
	table *Table

	mu         sync.Mutex
	cpuTargets Targets
	native     map[uint32]struct{}
	procRoot   string

	stop chan struct{}
	wg   sync.WaitGroup
}

func New(h Hooks) *Switch {
	return newSwitch(h, "/proc")
}

func newSwitch(h Hooks, procRoot string) *Switch {
	return &Switch{
		hooks:    h,
		table:    newTable(procRoot),
		procRoot: procRoot,
		native:   map[uint32]struct{}{},
	}
}

// Start refreshes the container table every 30s and reapplies the last targets.
func (s *Switch) Start() {
	s.stop = make(chan struct{})
	s.wg.Add(1)
	go s.loop()
}

// Stop ends the refresh loop.
func (s *Switch) Stop() {
	if s.stop == nil {
		return
	}
	close(s.stop)
	s.wg.Wait()
	s.stop = nil
}

func (s *Switch) loop() {
	defer s.wg.Done()
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			s.table.Refresh()
			s.mu.Lock()
			s.applyCPULocked()
			s.mu.Unlock()
		}
	}
}

// ContainerID returns the 12-char id for a main-thread pid, or "" when unknown.
func (s *Switch) ContainerID(pid uint32) string {
	return s.table.ContainerID(pid)
}

// ApplyCPU stores the CPU targets and diffs them against the previous roster.
func (s *Switch) ApplyCPU(t Targets) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cpuTargets = cloneTargets(t)
	s.applyCPULocked()
}

func (s *Switch) applyCPULocked() {
	native, ok := s.resolveCPU(s.cpuTargets)
	if !ok {
		return
	}
	for pid := range native {
		if _, on := s.native[pid]; on {
			continue
		}
		call(s.hooks.EnableNative, pid)
	}
	for pid := range s.native {
		if _, on := native[pid]; on {
			continue
		}
		call(s.hooks.DisableNative, pid)
		call(s.hooks.DropProcess, pid)
	}
	s.native = native
}

// resolveCPU returns false when container ids were requested and the table is not ready.
func (s *Switch) resolveCPU(t Targets) (native map[uint32]struct{}, ok bool) {
	native = map[uint32]struct{}{}
	if len(t.ContainerIDs) > 0 {
		snap, ready := s.readySnapshot()
		if !ready {
			return nil, false
		}
		for _, id := range t.ContainerIDs {
			id = strings.TrimSpace(id)
			if len(id) != reportedIDLen && len(id) != containerIDLen {
				continue
			}
			if !isHex(id) {
				continue
			}
			pids := snap.ByContainer[strings.ToLower(id[:reportedIDLen])]
			if len(pids) == 0 {
				continue
			}
			if containerHasJava(s.procRoot, pids) {
				continue
			}
			for _, pid := range pids {
				if shouldDropCmd(s.procRoot, pid) {
					continue
				}
				native[pid] = struct{}{}
			}
		}
		return native, true
	}
	for _, pid := range t.PIDs {
		if pid == 0 {
			continue
		}
		if isJavaProcess(s.procRoot, pid) {
			continue
		}
		native[pid] = struct{}{}
	}
	return native, true
}

func (s *Switch) readySnapshot() (Snapshot, bool) {
	if snap, ok := s.table.Current(); ok {
		return snap, true
	}
	if !s.table.Refresh() {
		return Snapshot{}, false
	}
	return s.table.Current()
}

func containerHasJava(procRoot string, pids []uint32) bool {
	for _, pid := range pids {
		if isJavaProcess(procRoot, pid) {
			return true
		}
	}
	return false
}

func cloneTargets(t Targets) Targets {
	return Targets{
		ContainerIDs: append([]string(nil), t.ContainerIDs...),
		PIDs:         append([]uint32(nil), t.PIDs...),
	}
}

func call(fn func(uint32), pid uint32) {
	if fn != nil {
		fn(pid)
	}
}
