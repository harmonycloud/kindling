// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package prometheusexporter

import (
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/aggregator/defaultaggregator"
	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
	"github.com/Kindling-project/kindling/collector/pkg/model/constvalues"
	"github.com/prometheus/client_golang/prometheus"
)

type collector struct {
	aggregator *defaultaggregator.CumulativeAggregator
}

func (c *collector) Collect(metrics chan<- prometheus.Metric) {
	// TODO debug
	dataGroups := c.aggregator.DumpAndRemoveExpired(time.Now())
	for i := 0; i < len(dataGroups); i++ {
		dataGroup := dataGroups[i]
		labelMap := dataGroup.Labels.GetValues()
		ts := getTimestamp(dataGroup.Timestamp)
		keys := make([]string, 0, len(labelMap))
		values := make([]string, 0, len(labelMap))
		for k, v := range labelMap {
			keys = append(keys, k)
			values = append(values, v.ToString())
		}
		for s := 0; s < len(dataGroup.Metrics); s++ {
			metric := dataGroup.Metrics[s]
			switch metric.DataType() {
			case model.IntMetricType:
				metric, error := prometheus.NewConstMetric(prometheus.NewDesc(
					sanitize(metric.Name, true),
					"",
					keys,
					nil,
					// TODO not all IntMetric are Counter, they can also be a Metric
				), prometheus.CounterValue, float64(metric.GetInt().Value), values...)
				if error == nil {
					tm := prometheus.NewMetricWithTimestamp(ts, metric)
					metrics <- tm
				}
			case model.HistogramMetricType:
				histogram := metric.GetHistogram()
				buckets := make(map[float64]uint64, len(histogram.ExplicitBoundaries))
				for x := 0; x < len(histogram.ExplicitBoundaries); x++ {
					bound := histogram.ExplicitBoundaries[x]
					buckets[float64(bound)] = histogram.BucketCounts[x]
				}
				metric, error := prometheus.NewConstHistogram(prometheus.NewDesc(
					sanitize(metric.Name, true),
					"",
					keys,
					nil,
				), histogram.Count, float64(histogram.Sum), buckets, values...)
				if error == nil {
					tm := prometheus.NewMetricWithTimestamp(ts, metric)
					metrics <- tm
				}
			}
		}
	}
}

func (c *collector) recordMetricGroups(group *model.DataGroup) {
	c.aggregator.AggregatorWithAllLabelsAndMetric(group, time.Now())
}

func newCollector(config *Config, _ *component.TelemetryLogger) *collector {
	// TODO Do this in config later !!!!
	requestTimeHistogramTopologyMetric := constnames.ToKindlingNetMetricName(constvalues.RequestTimeHistogram, false)
	requestTimeHistogramEntityMetric := constnames.ToKindlingNetMetricName(constvalues.RequestTimeHistogram, true)
	return &collector{
		aggregator: defaultaggregator.NewCumulativeAggregator(
			&defaultaggregator.AggregatedConfig{
				KindMap: map[string][]defaultaggregator.KindConfig{
					constnames.TcpRttMetricName: {{Kind: defaultaggregator.LastKind, OutputName: constnames.TcpRttMetricName}},
					requestTimeHistogramTopologyMetric: {{
						Kind:               defaultaggregator.HistogramKind,
						OutputName:         requestTimeHistogramTopologyMetric,
						ExplicitBoundaries: []int64{10e6, 20e6, 50e6, 80e6, 130e6, 200e6, 300e6, 400e6, 500e6, 700e6, 1000e6, 2000e6, 5000e6, 30000e6},
					}},
					requestTimeHistogramEntityMetric: {{
						Kind:               defaultaggregator.HistogramKind,
						OutputName:         requestTimeHistogramEntityMetric,
						ExplicitBoundaries: []int64{10e6, 20e6, 50e6, 80e6, 130e6, 200e6, 300e6, 400e6, 500e6, 700e6, 1000e6, 2000e6, 5000e6, 30000e6},
					}},
				},
			}, time.Minute*5)}
}

func getTimestamp(ts uint64) time.Time {
	return time.UnixMicro(int64(ts / 1000))
}

// Describe is a no-op, because the collector dynamically allocates metrics.
// https://github.com/prometheus/client_golang/blob/v1.9.0/prometheus/collector.go#L28-L40
func (c *collector) Describe(_ chan<- *prometheus.Desc) {}
