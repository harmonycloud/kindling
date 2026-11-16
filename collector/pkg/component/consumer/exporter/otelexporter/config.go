// Copyright (C) 2022-2024 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package otelexporter

import (
	"time"
)

type Config struct {
	ExportKind           string                           `mapstructure:"export_kind"`
	PromCfg              *PrometheusConfig                `mapstructure:"prometheus"`
	OtlpGrpcCfg          *OtlpGrpcConfig                  `mapstructure:"otlp"`
	StdoutCfg            *StdoutConfig                    `mapstructure:"stdout"`
	CustomLabels         map[string]string                `mapstructure:"custom_labels"`
	MetricAggregationMap map[string]MetricAggregationKind `mapstructure:"metric_aggregation_map"`
	AdapterConfig        *AdapterConfig                   `mapstructure:"adapter_config"`
}

type PrometheusConfig struct {
	Port             string            `mapstructure:"port,omitempty"`
	WithMemory       bool              `mapstructure:"with_memory,omitempty"`
	MemCleanUpConfig *MemCleanUpConfig `mapstructure:"memcleanup"`
}

type OtlpGrpcConfig struct {
	CollectPeriod time.Duration `mapstructure:"collect_period,omitempty"`
	Endpoint      string        `mapstructure:"endpoint,omitempty"`
}

type StdoutConfig struct {
	CollectPeriod time.Duration `mapstructure:"collect_period,omitempty"`
}

type AdapterConfig struct {
	NeedTraceAsResourceSpan bool `mapstructure:"need_trace_as_span"`
	NeedTraceAsMetric       bool `mapstructure:"need_trace_as_metric"`
	NeedPodDetail           bool `mapstructure:"need_pod_detail"`
	StoreExternalSrcIP      bool `mapstructure:"store_external_src_ip"`
}

type MemCleanUpConfig struct {
	Enabled       bool `mapstructure:"enable,omitempty"`
	RestartPeriod int  `mapstructure:"restart_period,omitempty"`
}
