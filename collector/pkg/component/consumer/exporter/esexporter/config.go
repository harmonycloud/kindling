// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package esexporter

type Config struct {
	EsHost  string `mapstructure:"es_host"`
	EsIndex string `mapstructure:"es_index"`
}

func (cfg *Config) GetEsHost() string {
	if cfg.EsHost == "" {
		return "http://39.103.171.51:8080"
	} else {
		return cfg.EsHost
	}
}

func (cfg *Config) GetEsIndexName() string {
	if cfg.EsIndex == "" {
		return "kindling_trace"
	} else {
		return cfg.EsIndex
	}
}
