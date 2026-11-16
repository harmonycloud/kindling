// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package aggregateprocessor

type Config struct {
	// The unit is second.
	TickerInterval int `mapstructure:"ticker_interval"`

	AggregateKindMap map[string][]AggregatedKindConfig `mapstructure:"aggregate_kind_map"`
	SamplingRate     *SampleConfig                     `mapstructure:"sampling_rate"`
}

type AggregatedKindConfig struct {
	OutputName         string  `mapstructure:"output_name"`
	Kind               string  `mapstructure:"kind"`
	ExplicitBoundaries []int64 `mapstructure:"explicit_boundaries"`
}

type SampleConfig struct {
	NormalData int `mapstructure:"normal_data"`
	SlowData   int `mapstructure:"slow_data"`
	ErrorData  int `mapstructure:"error_data"`
}

func NewDefaultConfig() *Config {
	ret := &Config{
		TickerInterval: 5,
		AggregateKindMap: map[string][]AggregatedKindConfig{
			"request_total_time": {{Kind: "sum"}, {Kind: "avg", OutputName: "request_total_time_avg"},
				{Kind: "count", OutputName: "request_count"}},
			"request_io":  {{Kind: "sum"}},
			"response_io": {{Kind: "sum"}},
			// tcp
			"kindling_tcp_srtt_microseconds": {{Kind: "last"}},
			"kindling_tcp_retransmit_total":  {{Kind: "sum"}},
			"kindling_tcp_packet_loss_total": {{Kind: "sum"}},
		},
		SamplingRate: &SampleConfig{
			NormalData: 0,
			SlowData:   100,
			ErrorData:  100,
		},
	}
	return ret
}
