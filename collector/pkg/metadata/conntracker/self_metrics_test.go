// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package conntracker

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

func TestNewSelfMetric(t *testing.T) {
	// Initialize opentelemetry
	exp, err := stdoutmetric.New()
	if err != nil {
		t.Fatal(err)
	}
	cont := controller.New(
		otelprocessor.NewFactory(selector.NewWithInexpensiveDistribution(), exp),
		controller.WithExporter(exp),
		controller.WithCollectPeriod(1*time.Second),
	)

	_ = cont.Start(context.Background())
	// Create conntracker self-telemetry
	noopConntracker := NewNoopConntracker(nil)
	newSelfMetrics(cont, noopConntracker)

	metricNames := []string{"state_size", "cache_max_size", "registers_total", "gets_total", "evicts_total",
		"enobufs", "read_errors", "msg_errors", "sampling_pct", "throttles"}
	// Cannot assert the result of opentelemetry-exporter,
	// so here we just wait to see the log.
	var stats int64 = 1
	// Wait first to see whether the first output is zero
	time.Sleep(time.Millisecond * 1000)
	go func() {
		for _, metric := range metricNames {
			noopConntracker.stats[metric] = stats
		}
		stats++
		time.Sleep(time.Millisecond * 500)
	}()

	time.Sleep(time.Second * 5)
	_ = cont.Stop(context.Background())
}
