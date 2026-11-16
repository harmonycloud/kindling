// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package model

const (
	IntMetricType MetricType = iota
	HistogramMetricType
	NoneMetricType
)

type MetricType int

type Metric struct {
	Name string
	//	Data can be assigned by:
	//	Int
	//	Histogram
	Data isMetricData
}

func (i *Metric) GetData() isMetricData {
	if i != nil {
		return i.Data
	}
	return nil
}

func (i *Metric) GetInt() *Int {
	if x, ok := i.GetData().(*Int); ok {
		return x
	}
	return nil
}

func (i *Metric) GetHistogram() *Histogram {
	if x, ok := i.GetData().(*Histogram); ok {
		return x
	}
	return nil
}

func (i *Metric) DataType() MetricType {
	switch i.GetData().(type) {
	case *Int:
		return IntMetricType
	case *Histogram:
		return HistogramMetricType
	default:
		return NoneMetricType
	}
}

func (i *Metric) Clear() {
	switch i.DataType() {
	case IntMetricType:
		i.GetInt().Value = 0
	case HistogramMetricType:
		histogram := i.GetHistogram()
		histogram.BucketCounts = nil
		histogram.Count = 0
		histogram.Sum = 0
		histogram.ExplicitBoundaries = nil
	}
}

func (i *Metric) Clone() *Metric {
	ret := &Metric{
		Name: i.Name,
		Data: nil,
	}
	switch i.DataType() {
	case IntMetricType:
		ret.Data = &Int{Value: i.GetInt().Value}
	case HistogramMetricType:
		histogram := i.GetHistogram()
		ret.Data = &Histogram{
			Sum:                histogram.Sum,
			Count:              histogram.Count,
			ExplicitBoundaries: histogram.ExplicitBoundaries,
			BucketCounts:       histogram.BucketCounts,
		}
	}
	return ret
}

type Int struct {
	Value int64
}

func NewIntMetric(name string, value int64) *Metric {
	return &Metric{Name: name, Data: &Int{Value: value}}
}

func NewMetric(name string, data isMetricData) *Metric {
	return &Metric{Name: name, Data: data}
}

type Histogram struct {
	Sum                int64
	Count              uint64
	ExplicitBoundaries []int64
	BucketCounts       []uint64
}

func NewHistogramMetric(name string, histogram *Histogram) *Metric {
	return &Metric{Name: name, Data: histogram}
}

type isMetricData interface {
	isMetricData()
}

func (*Int) isMetricData()       {}
func (*Histogram) isMetricData() {}
