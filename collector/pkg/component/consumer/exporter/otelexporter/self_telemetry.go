// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package otelexporter

import (
	"sync"

	"go.opentelemetry.io/otel/metric"
)

var otelexporterMetricgroupsReceivedTotal = "kindling_telemetry_otelexporter_metricgroups_received_total"

var once sync.Once

var dataGroupReceiverCounter metric.Int64Counter

func newSelfMetrics(meterProvider metric.MeterProvider) {
	once.Do(func() {
		dataGroupReceiverCounter = metric.Must(meterProvider.Meter("kindling")).NewInt64Counter(
			otelexporterMetricgroupsReceivedTotal, metric.WithDescription("The total count of the data received by otelexporter"))
	})
}
