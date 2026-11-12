package profiling

import (
	"fmt"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/flamestats"
)

// Config is the profiling Service setup. Filled from cgoreceiver ebpf_profiling.
type Config struct {
	EndpointURL         string
	RemoteTimeout       time.Duration
	MinBackoff          time.Duration
	MaxBackoff          time.Duration
	MaxBackoffRetries   int
	SenderWorkerCount   int
	SenderQueueCapacity int
	Interval            time.Duration

	Pids []uint32

	Metrics flamestats.Config
}

func (c *Config) applyDefaults() {
	if c.RemoteTimeout == 0 {
		c.RemoteTimeout = 10 * time.Second
	}
	if c.MinBackoff == 0 {
		c.MinBackoff = 500 * time.Millisecond
	}
	if c.MaxBackoff == 0 {
		c.MaxBackoff = 30 * time.Second
	}
	if c.MaxBackoffRetries == 0 {
		c.MaxBackoffRetries = 10
	}
	if c.SenderWorkerCount <= 0 {
		c.SenderWorkerCount = 2
	}
	if c.SenderQueueCapacity <= 0 {
		c.SenderQueueCapacity = 200
	}
	if c.Interval <= 0 {
		c.Interval = 15 * time.Second
	}
}

func (c Config) localPIDs() []uint32 {
	seen := make(map[uint32]struct{}, len(c.Pids))
	out := make([]uint32, 0, len(c.Pids))
	add := func(pid uint32) {
		if pid == 0 {
			return
		}
		if _, ok := seen[pid]; ok {
			return
		}
		seen[pid] = struct{}{}
		out = append(out, pid)
	}
	for _, pid := range c.Pids {
		add(pid)
	}
	return out
}

// PID zero is reserved for explicitly selecting every process.
func (c Config) captureAll() bool {
	return len(c.Pids) == 1 && c.Pids[0] == 0
}

func (c Config) validateTargets() error {
	for _, pid := range c.Pids {
		if pid == 0 && len(c.Pids) != 1 {
			return fmt.Errorf("profiling pids: 0 selects all processes and must be used alone")
		}
	}
	return nil
}
