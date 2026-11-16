// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package prometheusexporter

type Config struct {
	PromCfg      *PrometheusConfig `mapstructure:"prometheus"`
	CustomLabels map[string]string `mapstructure:"custom_labels"`
	// MetricAggregationMap map[string]MetricAggregationKind `mapstructure:"metric_aggregation_map"`
	AdapterConfig *AdapterConfig `mapstructure:"adapter_config"`
}

type PrometheusConfig struct {
	Endpoint string `mapstructure:"endpoint,omitempty"`
}

type AdapterConfig struct {
	NeedTraceAsResourceSpan bool `mapstructure:"need_trace_as_span"`
	NeedTraceAsMetric       bool `mapstructure:"need_trace_as_metric"`
	NeedPodDetail           bool `mapstructure:"need_pod_detail"`
	StoreExternalSrcIP      bool `mapstructure:"store_external_src_ip"`
}
