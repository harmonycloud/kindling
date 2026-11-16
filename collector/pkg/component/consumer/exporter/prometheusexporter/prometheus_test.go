// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package prometheusexporter

import (
	"testing"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
	"github.com/Kindling-project/kindling/collector/pkg/model/constvalues"
	"github.com/spf13/viper"
)

func InitPrometheusExporter(t *testing.T) exporter.Exporter {
	configPath := "testdata/kindling-collector-config.yml"
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		t.Fatalf("error happened when reading config: %v", err)
	}
	config := &Config{}
	err = viper.UnmarshalKey("exporters.prometheusexporter", config)
	if err != nil {
		t.Fatalf("error happened when unmarshaling config: %v", err)
	}
	return NewExporter(config, component.NewDefaultTelemetryTools())
}

func TestConsumeAggNetMetricGroup(t *testing.T) {
	latencyArray := []int64{1e6, 10e6, 20e6, 50e6, 100e6, 500e6}
	exp := InitPrometheusExporter(t)
	for {
		for _, latency := range latencyArray {
			_ = exp.Consume(makePreAggNetMetricGroup(int(latency)))
			time.Sleep(1 * time.Second)
		}
		time.Sleep(30 * time.Second)
	}
}

func TestConsumeSingleNetMetricGroup(t *testing.T) {
	latencyArray := []int64{1e6, 10e6, 20e6, 50e6, 100e6, 500e6}
	exp := InitPrometheusExporter(t)
	for {
		for _, latency := range latencyArray {
			_ = exp.Consume(makeSingleMetricGroup(int(latency)))
			time.Sleep(1 * time.Second)
		}
		time.Sleep(30 * time.Second)
	}
}

func makeSingleMetricGroup(i int) *model.DataGroup {
	metricsGroup := &model.DataGroup{
		Name: constnames.SingleNetRequestMetricGroup,
		Metrics: []*model.Metric{
			model.NewIntMetric(constvalues.ResponseIo, 1234567891),
			model.NewIntMetric(constvalues.RequestTotalTime, int64(i)),
			model.NewIntMetric(constvalues.RequestIo, 4500),
			model.NewIntMetric(constvalues.RequestCount, 4500),
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 19900909090,
	}
	metricsGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode")
	metricsGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace")
	metricsGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod")
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName")
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind")
	metricsGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService")
	metricsGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp")
	metricsGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode")
	metricsGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace")
	metricsGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod")
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName")
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind")
	metricsGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService")

	metricsGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer")
	metricsGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId")

	metricsGroup.Labels.AddStringValue(constlabels.Protocol, "http")
	metricsGroup.Labels.AddStringValue(constlabels.StatusCode, "200")

	// Topology data preferentially use D Nat Ip and D Nat Port
	metricsGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")
	metricsGroup.Labels.AddIntValue(constlabels.DstPort, 8081)
	return metricsGroup
}

func makePreAggNetMetricGroup(i int) *model.DataGroup {
	metricsGroup := &model.DataGroup{
		Name: constnames.AggregatedNetRequestMetricGroup,
		Metrics: []*model.Metric{
			model.NewIntMetric(constvalues.ResponseIo, 1234567891),
			model.NewHistogramMetric(constvalues.RequestTotalTime, &model.Histogram{
				Sum:                int64(i),
				Count:              100,
				ExplicitBoundaries: []int64{1e6, 2e6, 5e6, 1e7, 2e7, 5e7, 1e8, 2e8, 5e8, 1e9, 2e9, 5e9},
				BucketCounts:       []uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			}),
			model.NewIntMetric(constvalues.RequestTotalTime, int64(i)),
			model.NewIntMetric(constvalues.RequestIo, 4500),
			model.NewIntMetric(constvalues.RequestCount, 100),
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: uint64(time.Now().UnixNano()),
	}
	metricsGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode")
	metricsGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace")
	metricsGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod")
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName")
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind")
	metricsGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService")
	metricsGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp")
	metricsGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode")
	metricsGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace")
	metricsGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod")
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName")
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind")
	metricsGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService")

	metricsGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer")
	metricsGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId")

	metricsGroup.Labels.AddStringValue(constlabels.Protocol, "http")
	metricsGroup.Labels.AddStringValue(constlabels.StatusCode, "200")

	// Topology data preferentially use D Nat Ip and D Nat Port
	metricsGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")
	metricsGroup.Labels.AddIntValue(constlabels.DstPort, 8081)
	return metricsGroup
}
