// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package defaultaggregator

import (
	"reflect"
	"sync"
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/stretchr/testify/assert"
)

func startTask(workerNum int, task func(wg *sync.WaitGroup)) {
	wg := sync.WaitGroup{}
	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		go task(&wg)
	}
	wg.Wait()
}

func assertIntTest(t *testing.T, aggValues aggregatedValues, workerNum int, expectedNum int64,
	task func(wg *sync.WaitGroup)) {
	startTask(workerNum, task)
	got := aggValues.get()
	if expectedNum != aggValues.get().GetInt().Value {
		t.Errorf("The result is expected to be %d, but got %d", expectedNum, got.GetInt().Value)
	}
}

func assertHistogramTest(t *testing.T, aggValues aggregatedValues, workerNum int, expected *model.Histogram,
	task func(wg *sync.WaitGroup)) {
	startTask(workerNum, task)
	got := aggValues.get().GetHistogram()
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("The result is expected to be %v, but got %v", expected, got)
	}
}

func Test_aggValues_sum(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	expectedNum := int64(workerNum * loopNum)
	aggValues := &sumValue{}
	assertIntTest(t, aggValues, workerNum, expectedNum, func(wg *sync.WaitGroup) {
		for j := 0; j < loopNum; j++ {
			aggValues.calculate(1)
		}
		wg.Done()
	})
}

func Test_aggValues_sum_histogram(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	expectedNum := int64(workerNum * loopNum)
	aggValues := &sumValue{}
	assertIntTest(t, aggValues, workerNum, expectedNum, func(wg *sync.WaitGroup) {
		for j := 0; j < loopNum; j++ {
			aggValues.merge(&model.Histogram{Sum: 1})
		}
		wg.Done()
	})
}

func Test_aggValues_max(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	expectedNum := int64(loopNum - 1)
	aggValues := &maxValue{}
	assertIntTest(t, aggValues, workerNum, expectedNum, func(wg *sync.WaitGroup) {
		for j := 0; j < loopNum; j++ {
			aggValues.calculate(int64(j))
		}
		wg.Done()
	})
}

func Test_aggValues_max_histogram(t *testing.T) {
	aggValues := &maxValue{}
	assert.EqualError(t, aggValues.merge(&model.Histogram{Sum: 1}), "can not use max on a histogram metric")
}

func Test_aggValues_avg(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	var expectedNum int64 = 1
	aggValues := &avgValue{}
	assertIntTest(t, aggValues, workerNum, expectedNum, func(wg *sync.WaitGroup) {
		for j := 0; j < loopNum; j++ {
			aggValues.calculate(1)
		}
		wg.Done()
	})
}

func Test_aggValues_avg_histogram(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	var expectedNum int64 = 100
	aggValues := &avgValue{}
	assertIntTest(t, aggValues, workerNum, expectedNum, func(wg *sync.WaitGroup) {
		for j := 0; j < loopNum; j++ {
			aggValues.merge(&model.Histogram{Sum: int64(loopNum * 100), Count: uint64(loopNum)})
		}
		wg.Done()
	})
}

func Test_aggValues_last(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	var expectedNum int64 = 1
	aggValues := &lastValue{}
	assertIntTest(t, aggValues, workerNum, expectedNum, func(wg *sync.WaitGroup) {
		for j := loopNum; j > 0; j-- {
			aggValues.calculate(int64(j))
		}
		wg.Done()
	})
}

func Test_aggValues_last_histogram(t *testing.T) {
	aggValues := &lastValue{}
	assert.EqualError(t, aggValues.merge(&model.Histogram{Sum: 1}), "can not use lastValue on a histogram metric")
}

func Test_aggValues_count(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	var expectedNum = int64(workerNum * loopNum)
	aggValues := &countValue{}
	assertIntTest(t, aggValues, workerNum, expectedNum, func(wg *sync.WaitGroup) {
		for j := loopNum; j > 0; j-- {
			aggValues.calculate(int64(j))
		}
		wg.Done()
	})
}

func Test_aggValues_count_histogram(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	// only when loopNum is even
	var expectedNum = int64((loopNum*loopNum/2 + loopNum/2) * workerNum)
	aggValues := &countValue{}
	assertIntTest(t, aggValues, workerNum, expectedNum, func(wg *sync.WaitGroup) {
		for j := 0; j <= loopNum; j++ {
			aggValues.merge(&model.Histogram{Sum: int64(loopNum * 100), Count: uint64(j)})
		}
		wg.Done()
	})
}

func Test_aggValues_histogram(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	aggValues := &histogramValue{
		sum:                0,
		count:              0,
		explicitBoundaries: []int64{0, 100, 200, 500, 1000, 2000, 5000, 10000},
		bucketCounts:       []uint64{0, 0, 0, 0, 0, 0, 0, 0},
	}
	expected := &model.Histogram{
		Sum:                int64((10000*4999 + 15000) * workerNum),
		Count:              uint64(10000 * workerNum),
		ExplicitBoundaries: []int64{0, 100, 200, 500, 1000, 2000, 5000, 10000},
		BucketCounts:       []uint64{0, 500, 1000, 2500, 5000, 10000, 25000, 50000},
	}
	assertHistogramTest(t, aggValues, workerNum, expected, func(wg *sync.WaitGroup) {
		for j := loopNum; j > 0; j-- {
			aggValues.calculate(int64(j))
		}
		wg.Done()
	})
}

func Test_aggValues_histogram_histogram(t *testing.T) {
	workerNum := 5
	loopNum := 10000
	aggValues := &histogramValue{
		sum:                0,
		count:              0,
		explicitBoundaries: []int64{0, 100, 200, 500, 1000, 2000, 5000, 10000},
		bucketCounts:       []uint64{0, 0, 0, 0, 0, 0, 0, 0},
	}

	tmp := loopNum * workerNum
	expected := &model.Histogram{
		Sum:                int64((10000*4999 + 15000) * workerNum),
		Count:              uint64(10000 * workerNum * 100),
		ExplicitBoundaries: []int64{0, 100, 200, 500, 1000, 2000, 5000, 10000},
		BucketCounts:       []uint64{uint64(tmp), uint64(tmp * 2), uint64(tmp * 3), uint64(tmp * 4), uint64(tmp * 5), uint64(tmp * 6), uint64(tmp * 7), uint64(tmp * 8)},
	}
	assertHistogramTest(t, aggValues, workerNum, expected, func(wg *sync.WaitGroup) {
		for j := loopNum; j > 0; j-- {
			histogram := &model.Histogram{
				Sum:                int64(j),
				Count:              1 * 100,
				ExplicitBoundaries: []int64{0, 100, 200, 500, 1000, 2000, 5000, 10000},
				BucketCounts:       []uint64{1, 2, 3, 4, 5, 6, 7, 8},
			}
			aggValues.merge(histogram)
		}
		wg.Done()
	})
}

func Test_defaultValuesMap_sum(t *testing.T) {
	kindMap := make(map[string][]KindConfig)
	kindMap["sum_value"] = []KindConfig{{OutputName: "sum_value_sum", Kind: SumKind}}
	metrics := []*model.Metric{{Name: "sum_value"}}
	m := newAggValuesMap(metrics, kindMap)
	for i := 0; i < 10000; i++ {
		m.calculate(model.NewIntMetric("sum_value", int64(1)), 0)
	}
	got := m.get("sum_value")
	if got[0].Name != "sum_value_sum" || got[0].GetInt().Value != 10000 {
		t.Errorf("sum result is %v, expected %v", got[0].GetInt().Value, 10000)
	}
}

func Test_defaultValuesMap_avg(t *testing.T) {
	kindMap := make(map[string][]KindConfig)
	kindMap["avg_value"] = []KindConfig{{OutputName: "avg_value_avg", Kind: AvgKind}}
	metrics := []*model.Metric{{Name: "avg_value"}}
	m := newAggValuesMap(metrics, kindMap)
	for i := 0; i < 10000; i++ {
		m.calculate(model.NewIntMetric("avg_value", int64(1)), 0)
	}
	got := m.get("avg_value")
	if got[0].Name != "avg_value_avg" || got[0].GetInt().Value != 1 {
		t.Errorf("avg result is %v, expected %v", got[0].GetInt().Value, 1)
	}
}

func Test_defaultValuesMap_max(t *testing.T) {
	kindMap := make(map[string][]KindConfig)
	kindMap["max_value"] = []KindConfig{{OutputName: "max_value_max", Kind: MaxKind}}
	metrics := []*model.Metric{{Name: "max_value"}}
	m := newAggValuesMap(metrics, kindMap)
	for i := 0; i < 10000; i++ {
		m.calculate(model.NewIntMetric("max_value", int64(i)), 0)
	}
	got := m.get("max_value")
	var expected int64 = 9999
	if got[0].Name != "max_value_max" || got[0].GetInt().Value != expected {
		t.Errorf("max result is %v, expected %v", got[0].GetInt().Value, expected)
	}

	kindMap["reserve_max_value"] = []KindConfig{{OutputName: "reserve_max_value_max", Kind: MaxKind}}
	metrics = []*model.Metric{{Name: "reserve_max_value"}}
	m = newAggValuesMap(metrics, kindMap)
	for i := 10000; i > 0; i-- {
		m.calculate(model.NewIntMetric("reserve_max_value", int64(i)), 0)
	}
	got = m.get("reserve_max_value")
	if got[0].Name != "reserve_max_value_max" || got[0].GetInt().Value != 10000 {
		t.Errorf("max result is %v, expected %v", got[0].GetInt().Value, 10000)
	}
}

func Test_defaultValuesMap_lastValue(t *testing.T) {
	kindMap := make(map[string][]KindConfig)
	kindMap["last_value"] = []KindConfig{{OutputName: "last_value_last", Kind: LastKind}}
	metrics := []*model.Metric{{Name: "last_value"}}
	m := newAggValuesMap(metrics, kindMap)
	for i := 10000; i > 0; i-- {
		m.calculate(model.NewIntMetric("last_value", int64(i)), 0)
	}
	got := m.get("last_value")
	if got[0].Name != "last_value_last" || got[0].GetInt().Value != 1 {
		t.Errorf("lastValue result is %v, expected %v", got[0].GetInt().Value, 1)
	}
}

func Test_defaultValuesMap_countValue(t *testing.T) {
	kindMap := make(map[string][]KindConfig)
	kindMap["count_value"] = []KindConfig{{OutputName: "count_value_count", Kind: CountKind}}
	metrics := []*model.Metric{{Name: "count_value"}}
	m := newAggValuesMap(metrics, kindMap)
	for i := 10000; i > 0; i-- {
		m.calculate(model.NewIntMetric("count_value", int64(i)), 0)
	}
	got := m.get("count_value")
	if got == nil || got[0].Name != "count_value_count" || got[0].GetInt().Value != 10000 {
		t.Errorf("count_value result is %v, expected %v", got[0].GetInt().Value, 10000)
	}
}

func Test_defaultValuesMap_histogramValue(t *testing.T) {
	kindMap := make(map[string][]KindConfig)
	kindMap["histogram_value"] = []KindConfig{{OutputName: "histogram_value", Kind: HistogramKind, ExplicitBoundaries: []int64{0, 100, 200, 500, 1000, 2000, 5000, 10000}}}
	metrics := []*model.Metric{{Name: "histogram_value"}}
	m := newAggValuesMap(metrics, kindMap)
	for i := 10000; i > 0; i-- {
		m.calculate(model.NewIntMetric("histogram_value", int64(i)), 0)
	}
	got := m.get("histogram_value")
	expected := model.NewHistogramMetric("histogram_value", &model.Histogram{
		Sum:                10000*4999 + 15000,
		Count:              10000,
		ExplicitBoundaries: []int64{0, 100, 200, 500, 1000, 2000, 5000, 10000},
		BucketCounts:       []uint64{0, 100, 200, 500, 1000, 2000, 5000, 10000},
	})
	if got == nil || got[0].Name != "histogram_value" || got[0].GetHistogram() == nil ||
		!reflect.DeepEqual(got[0].GetHistogram(), expected.GetHistogram()) {
		t.Errorf("lastValue result is %v, expected %v", got[0].GetHistogram(), expected.GetHistogram())
	}
}
