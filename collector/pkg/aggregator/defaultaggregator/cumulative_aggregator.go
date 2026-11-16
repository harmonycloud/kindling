// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package defaultaggregator

import (
	"sync/atomic"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/aggregator"
	"github.com/Kindling-project/kindling/collector/pkg/model"
)

// DataGroup
// Name:
//   labels:
//      DataGroup

type CumulativeAggregator struct {
	*DefaultAggregator
	MetricExpiration time.Duration
}

type cumulativeValueRecorder struct {
	*valueRecorder
}

func (r *cumulativeValueRecorder) RecordWithDefaultSumKindAgg(key *aggregator.LabelKeys, metricValues []*model.Metric, timestamp uint64, now time.Time) {
	if key == nil {
		return
	}
	aggValues, ok := r.labelValues.Load(*key)
	if !ok {
		// double check to avoid double writing
		aggValues, _ = r.labelValues.LoadOrStore(*key, newExpiredAggValuesMapWithDefaultSumKind(metricValues, r.aggKindMap))
	}
	for _, metric := range metricValues {
		aggValues.(*expiredValuesMap).calculateWithExpired(metric, timestamp, now)
	}
}

func remove(s []aggregatedValues, i int) []aggregatedValues {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func (r *cumulativeValueRecorder) removeExpired(expirationTime time.Time) {
	expiredLabelValues := make([]interface{}, 0)
	r.labelValues.Range(func(key, value interface{}) bool {
		valueMap := value.(*expiredValuesMap)
		if expirationTime.After(valueMap.update) {
			expiredLabelValues = append(expiredLabelValues, key)
		}
		return true
	})
	for i := 0; i < len(expiredLabelValues); i++ {
		// TODO log debug expired
		r.labelValues.Delete(expiredLabelValues[i])
	}
}

func (c *CumulativeAggregator) AggregatorWithAllLabelsAndMetric(g *model.DataGroup, now time.Time) {
	name := g.Name
	c.mut.RLock()
	defer c.mut.RUnlock()
	recorder, ok := c.recordersMap.Load(name)
	// Won't enter this branch too many times, as the recordersMap
	// will become stable after running a period of time.
	if !ok {
		// double check to avoid double writing
		recorder, _ = c.recordersMap.LoadOrStore(name, &cumulativeValueRecorder{newValueRecorder(name, c.config.KindMap)})
	}
	key := aggregator.GetLabelsKeys(g.Labels)
	recorder.(*cumulativeValueRecorder).RecordWithDefaultSumKindAgg(key, g.Metrics, g.Timestamp, now)
}

func (c *CumulativeAggregator) DumpAndRemoveExpired(now time.Time) []*model.DataGroup {
	expirationTime := now.Add(-c.MetricExpiration)
	ret := make([]*model.DataGroup, 0)
	c.mut.Lock()
	defer c.mut.Unlock()
	c.recordersMap.Range(func(_, value interface{}) bool {
		recorder := value.(*cumulativeValueRecorder)
		ret = append(ret, recorder.dump()...)
		// Assume that the recordersMap will become stable after running a period of time.
		recorder.removeExpired(expirationTime)
		return true
	})
	return ret
}

func NewCumulativeAggregator(config *AggregatedConfig, metricExpiration time.Duration) *CumulativeAggregator {
	ret := &CumulativeAggregator{
		DefaultAggregator: NewDefaultAggregator(config),
		MetricExpiration:  metricExpiration,
	}
	return ret
}

type expiredValuesMap struct {
	*defaultValuesMap
	update time.Time
}

func (e *expiredValuesMap) calculateWithExpired(metric *model.Metric, timestamp uint64, now time.Time) {
	vSlice, ok := e.values[metric.Name]
	if !ok {
		return
	}
	for _, v := range vSlice {
		if metric.DataType() == model.IntMetricType {
			v.calculate(metric.GetInt().Value)
		} else if metric.DataType() == model.HistogramMetricType {
			v.merge(metric.GetHistogram())
		}
	}
	e.update = now
	atomic.StoreUint64(&e.timestamp, timestamp)
}

func newExpiredAggValuesMapWithDefaultSumKind(metrics []*model.Metric, kindMap map[string][]KindConfig) aggValuesMap {
	ret := &expiredValuesMap{defaultValuesMap: &defaultValuesMap{values: make(map[string][]aggregatedValues)}}
	for _, metric := range metrics {
		kindSlice, found := kindMap[metric.Name]
		var aggValuesSlice []aggregatedValues
		if found {
			aggValuesSlice = make([]aggregatedValues, len(kindSlice))
			for i, kind := range kindSlice {
				aggValuesSlice[i] = newAggValue(kind)
			}
		} else {
			aggValuesSlice = make([]aggregatedValues, 1)
			aggValuesSlice[0] = newAggValue(KindConfig{
				OutputName: metric.Name,
				Kind:       SumKind,
			})
		}
		ret.values[metric.Name] = aggValuesSlice
	}
	return ret
}
