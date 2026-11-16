// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package observability

import "time"

type Config struct {
	ExportKind  string            `mapstructure:"export_kind"`
	PromCfg     *PrometheusConfig `mapstructure:"prometheus"`
	OtlpGrpcCfg *OtlpGrpcConfig   `mapstructure:"otlp"`
	StdoutCfg   *StdoutConfig     `mapstructure:"stdout"`
}

type PrometheusConfig struct {
	Port         string   `mapstructure:"port,omitempty"`
	ExtraMetrics []string `mapstructure:"extra_metrics"`
}

type OtlpGrpcConfig struct {
	CollectPeriod time.Duration `mapstructure:"collect_period,omitempty"`
	Endpoint      string        `mapstructure:"endpoint,omitempty"`
}

type StdoutConfig struct {
	CollectPeriod time.Duration `mapstructure:"collect_period,omitempty"`
}

var DefaultConfig = Config{
	ExportKind:  "stdout",
	PromCfg:     nil,
	OtlpGrpcCfg: nil,
	StdoutCfg:   &StdoutConfig{CollectPeriod: 10 * time.Second},
}
