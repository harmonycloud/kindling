// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package tcpconnectanalyzer

type Config struct {
	ChannelSize     int  `mapstructure:"channel_size"`
	WaitEventSecond int  `mapstructure:"wait_event_second"`
	NeedProcessInfo bool `mapstructure:"need_process_info"`
}

func NewDefaultConfig() *Config {
	return &Config{
		ChannelSize:     2000,
		WaitEventSecond: 10,
		NeedProcessInfo: false,
	}
}
