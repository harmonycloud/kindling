package cgoreceiver

import "time"

type Config struct {
	SubscribeInfo     []SubEvent          `mapstructure:"subscribe"`
	ProcessFilterInfo ProcessFilter       `mapstructure:"process_filter"`
	EbpfProfiling     EbpfProfilingConfig `mapstructure:"ebpf_profiling"`
}

type SubEvent struct {
	Category string            `mapstructure:"category"`
	Name     string            `mapstructure:"name"`
	Params   map[string]string `mapstructure:"params"`
}

type ProcessFilter struct {
	Comms []string `mapstructure:"comms"`
}

// EbpfEndpointOptions is the Pyroscope push endpoint configuration.
type EbpfEndpointOptions struct {
	URL               string        `mapstructure:"url"`
	RemoteTimeout     time.Duration `mapstructure:"remote_timeout"`
	MinBackoff        time.Duration `mapstructure:"min_backoff_period"`
	MaxBackoff        time.Duration `mapstructure:"max_backoff_period"`
	MaxBackoffRetries int           `mapstructure:"max_backoff_retries"`
}

// EbpfProfilingConfig configures eBPF CPU stack sampling → Pyroscope export.
// Empty endpoint.url disables the feature.
// Empty Pids means CAPTURE_ALL; non-empty enables whitelist sampling only.
type EbpfProfilingConfig struct {
	Endpoint            EbpfEndpointOptions `mapstructure:"endpoint"`
	SenderWorkerCount   int                 `mapstructure:"sender_worker_count"`
	SenderQueueCapacity int                 `mapstructure:"sender_queue_capacity"`
	Pids                []uint32            `mapstructure:"pids"`
}

func NewDefaultConfig() *Config {
	return &Config{
		SubscribeInfo: []SubEvent{
			{
				Name:     "syscall_exit-writev",
				Category: "net",
			},
			{
				Name:     "syscall_exit-readv",
				Category: "net",
			},
			{
				Name:     "syscall_exit-write",
				Category: "net",
			},
			{
				Name:     "syscall_exit-read",
				Category: "net",
			},
			{
				Name:     "syscall_exit-sendto",
				Category: "net",
			},
			{
				Name:     "syscall_exit-recvfrom",
				Category: "net",
			},
			{
				Name:     "syscall_exit-sendmsg",
				Category: "net",
			},
			{
				Name:     "syscall_exit-recvmsg",
				Category: "net",
			},
			{
				Name:     "syscall_exit-sendmmsg",
				Category: "net",
			},
			{
				Name: "kprobe-tcp_close",
			},
			{
				Name: "kprobe-tcp_rcv_established",
			},
			{
				Name: "kprobe-tcp_drop",
			},
			{
				Name: "kprobe-tcp_retransmit_skb",
			},
			{
				Name: "syscall_exit-connect",
			},
			{
				Name: "kretprobe-tcp_connect",
			},
			{
				Name: "kprobe-tcp_set_state",
			},
			{
				Name: "tracepoint-procexit",
			},
		},
		ProcessFilterInfo: ProcessFilter{
			Comms: []string{"kindling-collec", "containerd", "dockerd", "containerd-shim"},
		},
		EbpfProfiling: EbpfProfilingConfig{
			Endpoint: EbpfEndpointOptions{
				URL:               "",
				RemoteTimeout:     10 * time.Second,
				MinBackoff:        500 * time.Millisecond,
				MaxBackoff:        30 * time.Second,
				MaxBackoffRetries: 10,
			},
			SenderWorkerCount:   2,
			SenderQueueCapacity: 200,
			Pids:                nil,
		},
	}
}
