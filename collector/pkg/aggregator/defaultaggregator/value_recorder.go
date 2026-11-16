// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package defaultaggregator

import (
	"sync"

	"github.com/Kindling-project/kindling/collector/pkg/aggregator"
	"github.com/Kindling-project/kindling/collector/pkg/model"
)

type valueRecorder struct {
	name string
	// aggValuesMap is responsible for its own thread-safe access.
	labelValues sync.Map
	aggKindMap  map[string][]KindConfig
}

func newValueRecorder(recorderName string, aggKindMap map[string][]KindConfig) *valueRecorder {
	return &valueRecorder{
		name:        recorderName,
		labelValues: sync.Map{},
		aggKindMap:  aggKindMap,
	}
}

// Record is thread-safe, and return the result value.
// A recorder can record only the metrics that are the same as the initial ones when using the same key.
// But it can record different metrics with different keys.
func (r *valueRecorder) Record(key *aggregator.LabelKeys, metricValues []*model.Metric, timestamp uint64) {
	if key == nil {
		return
	}
	aggValues, ok := r.labelValues.Load(*key)
	if !ok {
		// double check to avoid double writing
		aggValues, _ = r.labelValues.LoadOrStore(*key, newAggValuesMap(metricValues, r.aggKindMap))
	}
	for _, metric := range metricValues {
		aggValues.(aggValuesMap).calculate(metric, timestamp)
	}
}

// dump a set of metric from counter cache.
// The return value holds the reference to the metric, not the copied one.
func (r *valueRecorder) dump() []*model.DataGroup {
	ret := make([]*model.DataGroup, 0)
	r.labelValues.Range(func(key, value interface{}) bool {
		k := key.(aggregator.LabelKeys)
		v := value.(aggValuesMap)
		dataGroup := model.NewDataGroup(r.name, k.GetLabels(), v.getTimestamp(), v.getAll()...)
		ret = append(ret, dataGroup)
		return true
	})
	return ret
}

// reset the labelValues for further aggregation.
// This method is not thread safe.
func (r *valueRecorder) reset() {
	r.labelValues = sync.Map{}
}
