// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package network

const (
	defaultFdReuseTimeout        = 15
	defaultNoResponseThreshold   = 120
	defaultConnectTimeout        = 1
	defaultResponseSlowThreshold = 500
)

type Config struct {
	// This option is set only for testing. We enable it by default otherwise the function will not work.
	EnableTimeoutCheck  bool
	EventChannelSize    int `mapstructure:"event_channel_size"`
	ConnectTimeout      int `mapstructure:"connect_timeout"`
	FdReuseTimeout      int `mapstructure:"fd_reuse_timeout"`
	NoResponseThreshold int `mapstructure:"no_response_threshold"`
	// unit is ms
	ResponseSlowThreshold int `mapstructure:"response_slow_threshold"`

	EnableConntrack       bool   `mapstructure:"enable_conntrack"`
	IgnoreDnsRcode3Error  bool   `mapstructure:"ignore_dns_rcode3_error"`
	ConntrackMaxStateSize int    `mapstructure:"conntrack_max_state_size"`
	ConntrackRateLimit    int    `mapstructure:"conntrack_rate_limit"`
	ProcRoot              string `mapstructure:"proc_root"`

	ProtocolParser      []string         `mapstructure:"protocol_parser"`
	ProtocolConfigs     []ProtocolConfig `mapstructure:"protocol_config,omitempty"`
	UrlClusteringMethod string           `mapstructure:"url_clustering_method"`
}

func NewDefaultConfig() *Config {
	return &Config{
		EventChannelSize:      10000,
		EnableTimeoutCheck:    true,
		ConnectTimeout:        100,
		FdReuseTimeout:        15,
		NoResponseThreshold:   120,
		ResponseSlowThreshold: 500,
		EnableConntrack:       true,
		IgnoreDnsRcode3Error:  false,
		ConntrackMaxStateSize: 131072,
		ConntrackRateLimit:    500,
		ProcRoot:              "/proc",
		ProtocolParser:        []string{"http", "mysql", "dns", "redis", "kafka", "dubbo"},
		ProtocolConfigs: []ProtocolConfig{
			{
				Key:           "http",
				PayloadLength: 200,
			},
			{
				Key:           "dubbo",
				PayloadLength: 200,
			},
			{
				Key:       "mysql",
				Ports:     []uint32{3306},
				Threshold: 100,
			},
			{
				Key:       "kafka",
				Ports:     []uint32{9092},
				Threshold: 100,
			},
			{
				Key:       "dns",
				Ports:     []uint32{53},
				Threshold: 100,
			},
			{
				Key:       "cassandra",
				Ports:     []uint32{9042},
				Threshold: 100,
			},
			{
				Key:       "s3",
				Ports:     []uint32{9190},
				Threshold: 100,
			},
		},
		UrlClusteringMethod: "alphabet",
	}
}

type ProtocolConfig struct {
	Key            string   `mapstructure:"key,omitempty"`
	Ports          []uint32 `mapstructure:"ports,omitempty"`
	PayloadLength  int      `mapstructure:"payload_length"`
	DisableDiscern bool     `mapstructure:"disable_discern,omitempty"`
	Threshold      int      `mapstructure:"slow_threshold,omitempty"`
}

func (cfg *Config) GetConnectTimeout() int {
	if cfg.ConnectTimeout > 0 {
		return cfg.ConnectTimeout
	} else {
		return defaultConnectTimeout
	}
}

func (cfg *Config) GetFdReuseTimeout() int {
	if cfg.FdReuseTimeout > 0 {
		return cfg.FdReuseTimeout
	} else {
		return defaultFdReuseTimeout
	}
}

func (cfg *Config) getResponseSlowThreshold() int {
	if cfg.ResponseSlowThreshold > 0 {
		return cfg.ResponseSlowThreshold
	} else {
		return defaultResponseSlowThreshold
	}
}

func (cfg *Config) getNoResponseThreshold() int {
	if cfg.NoResponseThreshold > 0 {
		return cfg.NoResponseThreshold
	} else {
		return defaultNoResponseThreshold
	}
}
