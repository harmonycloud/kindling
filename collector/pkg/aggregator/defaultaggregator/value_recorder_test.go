// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package defaultaggregator

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Kindling-project/kindling/collector/pkg/aggregator"
	"github.com/Kindling-project/kindling/collector/pkg/model"
)

func TestRecord(t *testing.T) {
	aggKindMap := AggregatedConfig{KindMap: map[string][]KindConfig{
		"duration": {
			{Kind: SumKind, OutputName: "duration_sum"},
			{Kind: CountKind, OutputName: "request_count"},
			{Kind: MaxKind, OutputName: "duration_max"},
			{Kind: AvgKind, OutputName: "duration_avg"},
			{Kind: HistogramKind, OutputName: "duration_histogram", ExplicitBoundaries: []int64{0, 100, 200, 500, 1000}},
		},
		"last": {{Kind: LastKind, OutputName: "last"}},
	}}

	recorder := newValueRecorder("testRecorder", aggKindMap.KindMap)
	keys := aggregator.NewLabelKeys([]aggregator.LabelKey{
		{
			Name:  "stringKey",
			Value: "stringValue",
			VType: aggregator.StringType,
		},
		{
			Name:  "booleanKey",
			Value: "true",
			VType: aggregator.BooleanType,
		},
		{
			Name:  "intKey",
			Value: "100",
			VType: aggregator.IntType,
		},
	}...)

	for i := 0; i < 100; i++ {
		metricValues := []*model.Metric{
			model.NewIntMetric("duration", int64(100)),
			model.NewIntMetric("last", int64(i)),
		}
		recorder.Record(keys, metricValues, 0)
	}
	retMetricGroup := recorder.dump()
	sumValue, _ := retMetricGroup[0].GetMetric("duration_sum")
	if sumValue.GetInt().Value != 10000 {
		t.Errorf("sum expected %v, got %v", 10000, sumValue.GetInt().Value)
	}
	countValue, _ := retMetricGroup[0].GetMetric("request_count")
	if countValue.GetInt().Value != 100 {
		t.Errorf("expected %v, got %v", 100, countValue.GetInt().Value)
	}
	maxValue, _ := retMetricGroup[0].GetMetric("duration_max")
	if maxValue.GetInt().Value != 100 {
		t.Errorf("expected %v, got %v", 100, maxValue.GetInt().Value)
	}
	avgValue, _ := retMetricGroup[0].GetMetric("duration_avg")
	if avgValue.GetInt().Value != 100 {
		t.Errorf("expected %v, got %v", 100, avgValue.GetInt().Value)
	}
	lastValue, _ := retMetricGroup[0].GetMetric("last")
	if lastValue.GetInt().Value != 99 {
		t.Errorf("expected %v, got %v", 99, lastValue.GetInt().Value)
	}
	histogramValue, _ := retMetricGroup[0].GetMetric("duration_histogram")
	expectedValue := &model.Histogram{
		Sum:                10000,
		Count:              100,
		ExplicitBoundaries: []int64{0, 100, 200, 500, 1000},
		BucketCounts:       []uint64{0, 100, 100, 100, 100},
	}
	if !reflect.DeepEqual(histogramValue.GetHistogram(), expectedValue) {
		t.Errorf("expected %+v, got %+v", expectedValue, histogramValue.GetHistogram())
	}
}

// TestRecordDiffMetricsWithSameKey validates that a recorder can record only the metrics that are same as the initial ones.
// But it allows to record different metrics with different keys.
func TestRecordDiffMetricsWithSameKey(t *testing.T) {
	aggKindMap := AggregatedConfig{KindMap: map[string][]KindConfig{
		"duration": {
			{Kind: SumKind, OutputName: "duration_sum"},
		},
		"last": {{Kind: LastKind, OutputName: "last"}},
	}}
	recorder := newValueRecorder("testRecorder", aggKindMap.KindMap)
	keys := aggregator.NewLabelKeys([]aggregator.LabelKey{
		{
			Name:  "stringKey",
			Value: "stringValue",
			VType: aggregator.StringType,
		},
	}...)
	for i := 0; i < 100; i++ {
		metricValues := []*model.Metric{
			model.NewIntMetric("duration", int64(100)),
		}
		recorder.Record(keys, metricValues, 0)
		metricValues = []*model.Metric{
			model.NewIntMetric("last", int64(i)),
		}
		recorder.Record(keys, metricValues, 0)
	}
	retMetricGroup := recorder.dump()
	sumValue, _ := retMetricGroup[0].GetMetric("duration_sum")
	assert.Equal(t, int64(10000), sumValue.GetInt().Value)
	// last is not aggregated because it is not the initial one using the current key
	_, ok := retMetricGroup[0].GetMetric("last")
	assert.Equal(t, false, ok)
}
