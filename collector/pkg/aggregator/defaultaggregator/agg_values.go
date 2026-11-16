// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package defaultaggregator

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/Kindling-project/kindling/collector/pkg/model"
)

type AggregatorKind int

const (
	SumKind AggregatorKind = iota
	MaxKind
	AvgKind
	LastKind
	CountKind
	HistogramKind
)

func (k AggregatorKind) name() string {
	switch k {
	case SumKind:
		return "sum"
	case MaxKind:
		return "max"
	case AvgKind:
		return "avg"
	case LastKind:
		return "last"
	case CountKind:
		return "count"
	case HistogramKind:
		return "histogram"
	default:
		return ""
	}
}

func GetAggregatorKind(kind string) AggregatorKind {
	switch kind {
	case "sum":
		return SumKind
	case "max":
		return MaxKind
	case "avg":
		return AvgKind
	case "last":
		return LastKind
	case "count":
		return CountKind
	case "histogram":
		return HistogramKind
	default:
		return SumKind
	}
}

type (
	AggregatedConfig struct {
		KindMap map[string][]KindConfig
	}

	KindConfig struct {
		OutputName string
		Kind       AggregatorKind
		// Only HistogramKind has this value
		ExplicitBoundaries []int64
	}
)

type aggValuesMap interface {
	// calculate should be thread-safe to use
	calculate(metric *model.Metric, timestamp uint64)
	get(name string) []*model.Metric
	getAll() []*model.Metric
	getTimestamp() uint64
}

type defaultValuesMap struct {
	values    map[string][]aggregatedValues
	timestamp uint64
}

func newAggValuesMap(metrics []*model.Metric, kindMap map[string][]KindConfig) aggValuesMap {
	ret := &defaultValuesMap{values: make(map[string][]aggregatedValues)}
	for _, metric := range metrics {
		kindSlice, found := kindMap[metric.Name]
		if !found {
			continue
		}

		aggValuesSlice := make([]aggregatedValues, len(kindSlice))
		for i, kind := range kindSlice {
			aggValuesSlice[i] = newAggValue(kind)
		}
		ret.values[metric.Name] = aggValuesSlice
	}
	return ret
}

// calculate returns the result value
func (m *defaultValuesMap) calculate(metric *model.Metric, timestamp uint64) {
	vSlice, ok := m.values[metric.Name]
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
	atomic.StoreUint64(&m.timestamp, timestamp)
}

func (m *defaultValuesMap) get(name string) []*model.Metric {
	vSlice, ok := m.values[name]
	if !ok {
		return nil
	}
	ret := make([]*model.Metric, len(vSlice))
	for i, v := range vSlice {
		ret[i] = v.get()
	}
	return ret
}

func (m *defaultValuesMap) getAll() []*model.Metric {
	ret := make([]*model.Metric, 0)
	for k := range m.values {
		ret = append(ret, m.get(k)...)
	}
	return ret
}

func (m *defaultValuesMap) getTimestamp() uint64 {
	return atomic.LoadUint64(&m.timestamp)
}

func newAggValue(cfg KindConfig) aggregatedValues {
	name := cfg.OutputName
	switch cfg.Kind {
	case SumKind:
		return &sumValue{name: name}
	case MaxKind:
		return &maxValue{name: name}
	case AvgKind:
		return &avgValue{name: name}
	case LastKind:
		return &lastValue{name: name}
	case CountKind:
		return &countValue{name: name}
	case HistogramKind:
		return &histogramValue{name: name, explicitBoundaries: cfg.ExplicitBoundaries, bucketCounts: make([]uint64, len(cfg.ExplicitBoundaries))}
	default:
		return &lastValue{name: name}
	}
}

type aggregatedValues interface {
	merge(metric *model.Histogram) error
	calculate(value int64) int64
	// get returns the value
	get() *model.Metric
	// getName returns the value's name
	getName() string
}

type maxValue struct {
	name  string
	value int64
	mut   sync.RWMutex
}

func (v *maxValue) calculate(value int64) int64 {
	v.mut.Lock()
	defer v.mut.Unlock()
	if v.value < value {
		v.value = value
	}
	return v.value
}
func (v *maxValue) get() *model.Metric {
	v.mut.RLock()
	defer v.mut.RUnlock()
	return model.NewIntMetric(v.name, v.value)
}

func (v *maxValue) getName() string {
	return v.name
}

func (v *maxValue) merge(metric *model.Histogram) error {
	return errors.New("can not use max on a histogram metric")
}

type sumValue struct {
	name  string
	value int64
}

func (v *sumValue) calculate(value int64) int64 {
	return atomic.AddInt64(&v.value, value)
}
func (v *sumValue) get() *model.Metric {
	return model.NewIntMetric(v.name, atomic.LoadInt64(&v.value))
}
func (v *sumValue) getName() string {
	return v.name
}
func (v *sumValue) merge(metric *model.Histogram) error {
	atomic.AddInt64(&v.value, metric.Sum)
	return nil
}

type avgValue struct {
	name  string
	value int64
	count int64
	mut   sync.RWMutex
}

// calculate returns the total count of data.
// Note it will not return the current average value which could be got by calling get()
func (v *avgValue) calculate(value int64) int64 {
	v.mut.Lock()
	defer v.mut.Unlock()
	v.value += value
	v.count++
	return v.count
}
func (v *avgValue) get() *model.Metric {
	v.mut.RLock()
	defer v.mut.RUnlock()
	return model.NewIntMetric(v.name, v.value/v.count)
}
func (v *avgValue) getName() string {
	return v.name
}

func (v *avgValue) merge(metric *model.Histogram) error {
	v.mut.Lock()
	defer v.mut.Unlock()
	v.count += int64(metric.Count)
	v.value += metric.Sum
	return nil
}

type lastValue struct {
	name  string
	value int64
}

func (v *lastValue) calculate(value int64) int64 {
	return atomic.SwapInt64(&v.value, value)
}
func (v *lastValue) get() *model.Metric {
	return model.NewIntMetric(v.name, atomic.LoadInt64(&v.value))
}
func (v *lastValue) getName() string {
	return v.name
}
func (v *lastValue) merge(metric *model.Histogram) error {
	return errors.New("can not use lastValue on a histogram metric")
}

type countValue struct {
	name  string
	value int64
}

// calculate add 1 to its own value
func (v *countValue) calculate(_ int64) int64 {
	return atomic.AddInt64(&v.value, 1)
}

// calculate add 1 to its own value
func (v *countValue) merge(metric *model.Histogram) error {
	atomic.AddInt64(&v.value, int64(metric.Count))
	return nil
}

func (v *countValue) get() *model.Metric {
	return model.NewIntMetric(v.name, atomic.LoadInt64(&v.value))
}
func (v *countValue) getName() string {
	return v.name
}

type histogramValue struct {
	name               string
	sum                int64
	count              uint64
	explicitBoundaries []int64
	bucketCounts       []uint64
	mut                sync.RWMutex
}

func (v *histogramValue) calculate(value int64) int64 {
	v.mut.Lock()
	defer v.mut.Unlock()
	v.sum += value
	v.count++
	for i := 0; i < len(v.explicitBoundaries); i++ {
		if value <= v.explicitBoundaries[i] {
			v.bucketCounts[i] += 1
		}
	}
	return int64(v.count)
}

func (v *histogramValue) get() *model.Metric {
	v.mut.RLock()
	defer v.mut.RUnlock()
	return model.NewHistogramMetric(v.name, &model.Histogram{
		Sum:                v.sum,
		Count:              v.count,
		ExplicitBoundaries: v.explicitBoundaries,
		BucketCounts:       v.bucketCounts,
	})
}

func (v *histogramValue) getName() string {
	return v.name
}

func (v *histogramValue) merge(metric *model.Histogram) error {
	v.mut.Lock()
	defer v.mut.Unlock()
	if len(v.explicitBoundaries) != len(metric.ExplicitBoundaries) {
		return errors.New("merge histogram failed when aggregating,the incoming metric's explicitBoundaries is different from memory")
	}
	v.sum += metric.Sum
	v.count += metric.Count
	for i := 0; i < len(metric.BucketCounts); i++ {
		v.bucketCounts[i] += metric.BucketCounts[i]
	}
	return nil
}
