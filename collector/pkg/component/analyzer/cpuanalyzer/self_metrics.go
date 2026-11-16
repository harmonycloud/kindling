// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cpuanalyzer

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/metric"
)

var onceMetric sync.Once

const (
	goroutineSize = "kindling_telemetry_cpuanalyzer_routine_size"
)

func newSelfMetrics(meterProvider metric.MeterProvider, analyzer *CpuAnalyzer) {
	onceMetric.Do(func() {
		meter := metric.Must(meterProvider.Meter("kindling"))
		meter.NewInt64GaugeObserver(goroutineSize,
			func(ctx context.Context, result metric.Int64ObserverResult) {
				result.Observe(int64(analyzer.routineSize.Load()))
			})
	})
}
