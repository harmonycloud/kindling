// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package prometheusexporter

type MetricAggregationKind string

const (
	MAGaugeKind     MetricAggregationKind = "gauge"
	MACounterKind   MetricAggregationKind = "counter"
	MAHistogramKind MetricAggregationKind = "histogram"
)
