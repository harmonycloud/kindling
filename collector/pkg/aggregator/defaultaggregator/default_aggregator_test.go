// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package defaultaggregator

import (
	"sync"
	"testing"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/aggregator"
	"github.com/Kindling-project/kindling/collector/pkg/model"
)

func TestConcurrentAggregator(t *testing.T) {
	aggKindMap := &AggregatedConfig{KindMap: map[string][]KindConfig{
		"duration": {
			{Kind: SumKind, OutputName: "duration_sum"},
			{Kind: CountKind, OutputName: "request_count"},
		},
	}}

	aggregatorInstance := NewDefaultAggregator(aggKindMap)
	labels := model.NewAttributeMap()
	labels.AddStringValue("key", "value")

	labelSelectors := aggregator.NewLabelSelectors(
		aggregator.LabelSelector{Name: "key", VType: aggregator.StringType},
	)

	wg := sync.WaitGroup{}
	wg.Add(1)
	stopCh := make(chan bool)

	var runLoop = 100000
	var duration int64 = 100
	go func() {
		for i := 0; i < runLoop; i++ {
			metricValues := []*model.Metric{
				{Name: "duration", Data: &model.Int{Value: duration}},
			}
			dataGroup := model.NewDataGroup("testMetric", labels, 0, metricValues...)
			aggregatorInstance.Aggregate(dataGroup, labelSelectors)
			time.Sleep(time.Microsecond)
		}
		stopCh <- true
		wg.Done()
	}()

	wg.Add(1)
	var durationSum int64
	var requestCount int64
	go func() {
		ticker := time.NewTicker(500 * time.Microsecond)
		for {
			select {
			case <-ticker.C:
				ret := aggregatorInstance.Dump()
				for _, g := range ret {
					for _, v := range g.Metrics {
						if v.Name == "duration_sum" {
							durationSum += v.GetInt().Value
						} else if v.Name == "request_count" {
							requestCount += v.GetInt().Value
						}
					}
				}
			case <-stopCh:
				ret := aggregatorInstance.Dump()
				for _, g := range ret {
					for _, v := range g.Metrics {
						if v.Name == "duration_sum" {
							durationSum += v.GetInt().Value
						} else if v.Name == "request_count" {
							requestCount += v.GetInt().Value
						}
					}
				}
				wg.Done()
				return
			}
		}
	}()
	wg.Wait()
	expectedDurationSum := int64(runLoop) * duration
	expectedRequestCount := int64(runLoop)
	if durationSum != expectedDurationSum || requestCount != expectedRequestCount {
		t.Errorf("request_count: exptected %v, got %v", expectedRequestCount, requestCount)
		t.Errorf("duration_sum: expected %v, got %v", expectedDurationSum, durationSum)
	}
}
