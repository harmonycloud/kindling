// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package conntracker

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	cacheSizeMetric     = "kindling_telemetry_conntracker_cache_size"
	cacheMaxSizeMetric  = "kindling_telemetry_conntracker_cache_max_size"
	operationTimesTotal = "kindling_telemetry_conntracker_operation_times_total"
	errorsTotal         = "kindling_telemetry_conntracker_errors_total"
	samplingRate        = "kindling_telemetry_conntracker_sampling_rate"
	throttlesTotal      = "kindling_telemetry_conntracker_throttles_total"
)

var (
	selfTelemetryOnce        sync.Once
	cacheSizeInstrument      metric.Int64GaugeObserver
	cacheMaxSizeInstrument   metric.Int64GaugeObserver
	operationTimesInstrument metric.Int64CounterObserver
	errorsTotalInstrument    metric.Int64CounterObserver
	samplingRateInstrument   metric.Int64GaugeObserver
	throttlesTotalInstrument metric.Int64CounterObserver
)

// To avoid getting statistics multiple times, we cache the result in a global variable.
// This can be done because the observation functions are executed in the order as they were
// registered. See go.opentelemetry.io/otel/internal/metric/async/AsyncInstrumentState.runners.
var conntrackerStaticStates map[string]int64

func newSelfMetrics(meterProvider metric.MeterProvider, conntracker Conntracker) {
	selfTelemetryOnce.Do(func() {
		meter := metric.Must(meterProvider.Meter("kindling"))
		cacheSizeInstrument = meter.NewInt64GaugeObserver(cacheSizeMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				conntrackerStaticStates = conntracker.GetStats()
				result.Observe(conntrackerStaticStates["state_size"], attribute.String("type", "general"))
				result.Observe(conntrackerStaticStates["orphan_size"], attribute.String("type", "orphan"))
			}, metric.WithDescription("The current number of the conntrack records stored in the map"))
		cacheMaxSizeInstrument = meter.NewInt64GaugeObserver(cacheMaxSizeMetric,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(conntrackerStaticStates["cache_max_size"])
			}, metric.WithDescription("The maximum size of the cache map"))
		operationTimesInstrument = meter.NewInt64CounterObserver(operationTimesTotal,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(conntrackerStaticStates["registers_total"], attribute.String("op", "add"))
				result.Observe(conntrackerStaticStates["registers_dropped"], attribute.String("op", "drop"))
				result.Observe(conntrackerStaticStates["unregisters_total"], attribute.String("op", "remove"))
				result.Observe(conntrackerStaticStates["gets_total"], attribute.String("op", "get"))
				result.Observe(conntrackerStaticStates["evicts_total"], attribute.String("op", "evict"))
			}, metric.WithDescription("The total operation times the conntracker does to the cache map"))
		errorsTotalInstrument = meter.NewInt64CounterObserver(errorsTotal,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(conntrackerStaticStates["enobufs"], attribute.String("type", "enobuf"))
				result.Observe(conntrackerStaticStates["read_errors"], attribute.String("type", "read_errors"))
				result.Observe(conntrackerStaticStates["msg_errors"], attribute.String("type", "msg_errors"))
			}, metric.WithDescription("The total count of errors the conntracker encounters"))
		samplingRateInstrument = meter.NewInt64GaugeObserver(samplingRate,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(conntrackerStaticStates["sampling_pct"])
			}, metric.WithDescription("The sampling rate of the conntracker module"))
		throttlesTotalInstrument = meter.NewInt64CounterObserver(throttlesTotal,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(conntrackerStaticStates["throttles"])
			}, metric.WithDescription("The total count of the records being throttled due to the high load"))
		// Suppress warnings of unused variables
		_ = cacheSizeInstrument
		_ = cacheMaxSizeInstrument
		_ = operationTimesInstrument
		_ = errorsTotalInstrument
		_ = samplingRateInstrument
		_ = throttlesTotalInstrument
	})
}
