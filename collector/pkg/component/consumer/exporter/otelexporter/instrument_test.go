// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package otelexporter

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter/tools/adapter"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
)

func Test_instrumentFactory_recordLastValue(t *testing.T) {
	cfg := &Config{
		ExportKind:   StdoutKindExporter,
		PromCfg:      nil,
		OtlpGrpcCfg:  nil,
		StdoutCfg:    &StdoutConfig{CollectPeriod: 10 * time.Second},
		CustomLabels: nil,
		MetricAggregationMap: map[string]MetricAggregationKind{
			"kindling_entity_request_total":                          MACounterKind,
			"kindling_entity_request_duration_nanoseconds_total":     MACounterKind,
			"kindling_entity_request_average_duration_nanoseconds":   MAHistogramKind,
			"kindling_entity_request_send_bytes_total":               MACounterKind,
			"kindling_entity_request_receive_bytes_total":            MACounterKind,
			"kindling_topology_request_total":                        MACounterKind,
			"kindling_topology_request_duration_nanoseconds_total":   MACounterKind,
			"kindling_topology_request_average_duration_nanoseconds": MAHistogramKind,
			"kindling_topology_request_request_bytes_total":          MACounterKind,
			"kindling_topology_request_response_bytes_total":         MACounterKind,
			"kindling_trace_request_duration_nanoseconds":            MAGaugeKind,
			"kindling_tcp_srtt_microseconds":                         MAGaugeKind,
			"kindling_tcp_retransmit_total":                          MACounterKind,
			"kindling_tcp_packet_loss_total":                         MACounterKind,
		},
		AdapterConfig: &AdapterConfig{
			NeedTraceAsResourceSpan: true,
			NeedTraceAsMetric:       true,
			NeedPodDetail:           true,
			StoreExternalSrcIP:      false,
		},
	}

	loggerInstance := component.NewDefaultTelemetryTools()
	exporter, _ := newExporters(context.Background(), cfg, loggerInstance)

	cont := controller.New(
		otelprocessor.NewFactory(simple.NewWithHistogramDistribution(
			histogram.WithExplicitBoundaries(exponentialInt64NanosecondsBoundaries),
		), exporter.metricExporter),
		controller.WithExporter(exporter.metricExporter),
		controller.WithCollectPeriod(cfg.StdoutCfg.CollectPeriod),
		controller.WithResource(nil),
	)

	_ = cont.Start(context.Background())

	ins := newInstrumentFactory(cont.Meter("test"), component.NewDefaultTelemetryTools(), nil)

	for i := 0; i < 10000; i++ {
		time.Sleep(1 * time.Second)
		group := makeTcpGroup(int64(i))
		err := ins.recordLastValue(constnames.TcpRttMetricName, group)
		if err != nil {
			t.Errorf("record last value failed %e", err)
		}
	}
}

func makeTcpGroup(rttLatency int64) *model.DataGroup {
	return model.NewDataGroup(
		constnames.TcpRttMetricGroupName,
		model.NewAttributeMapWithValues(
			map[string]model.AttributeValue{
				constlabels.SrcIp:           model.NewStringValue("src-ip"),
				constlabels.SrcPort:         model.NewIntValue(33333),
				constlabels.DstIp:           model.NewStringValue("dst-ip"),
				constlabels.DstPort:         model.NewIntValue(8080),
				constlabels.DstPod:          model.NewStringValue("dst-pod"),
				constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
				constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
				constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
				constlabels.SrcPod:          model.NewStringValue("src-pod"),
				constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
				constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
				constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
				constlabels.SrcService:      model.NewStringValue("src-service"),
				constlabels.DstService:      model.NewStringValue("dst-service"),
				constlabels.SrcNode:         model.NewStringValue("src-node"),
				constlabels.DstNode:         model.NewStringValue("dst-node"),
			}),
		123,
		[]*model.Metric{
			model.NewIntMetric(constnames.TcpRttMetricName, rttLatency),
		}...)
}

func makeTraceAsMetricGroup(requestLatency int64, timestamp uint64, dstIp string) *model.DataGroup {
	return model.NewDataGroup(
		constnames.TraceAsMetric,
		model.NewAttributeMapWithValues(
			map[string]model.AttributeValue{
				// instanceInfo *Need to remove dstIp and dstPort from internal agg topology*
				constlabels.SrcIp:   model.NewStringValue("src-ip"),
				constlabels.DstIp:   model.NewStringValue(dstIp),
				constlabels.DstPort: model.NewIntValue(8080),

				// protocolInfo
				constlabels.Protocol:        model.NewStringValue("http"),
				constlabels.RequestContent:  model.NewStringValue("/test"),
				constlabels.ResponseContent: model.NewStringValue("200"),

				// k8sInfo
				constlabels.DstPod:          model.NewStringValue("dst-pod"),
				constlabels.DstWorkloadName: model.NewStringValue("dst-workloadName"),
				constlabels.DstNamespace:    model.NewStringValue("dst-Namespace"),
				constlabels.DstWorkloadKind: model.NewStringValue("dst-workloadKind"),
				constlabels.SrcPod:          model.NewStringValue("src-pod"),
				constlabels.SrcWorkloadName: model.NewStringValue("src-workloadName"),
				constlabels.SrcNamespace:    model.NewStringValue("src-Namespace"),
				constlabels.SrcWorkloadKind: model.NewStringValue("src-workloadKind"),
				constlabels.SrcService:      model.NewStringValue("src-service"),
				constlabels.DstService:      model.NewStringValue("dst-service"),
				constlabels.SrcNode:         model.NewStringValue("src-node"),
				constlabels.DstNode:         model.NewStringValue("dst-node"),
				constlabels.DnatIp:          model.NewStringValue("dnat-ip"),
				constlabels.DnatPort:        model.NewIntValue(80),

				constlabels.IsServer:                model.NewIntValue(0),
				constlabels.RequestDurationStatus:   model.NewStringValue(adapter.GreenStatus),
				constlabels.RequestReqxferStatus:    model.NewStringValue(adapter.GreenStatus),
				constlabels.RequestProcessingStatus: model.NewStringValue(adapter.GreenStatus),
				constlabels.ResponseRspxferStatus:   model.NewStringValue(adapter.GreenStatus),

				"const-labels1": model.NewStringValue("const-values1"),
			}),
		timestamp,
		[]*model.Metric{
			model.NewIntMetric(constnames.TraceAsMetric, requestLatency),
		}...)
}

func Test_instrumentFactory_recordTraceAsMetric(t *testing.T) {
	ins := newInstrumentFactory(metric.Meter{}, component.NewDefaultTelemetryTools(), nil)
	metricName := constnames.TraceAsMetric
	var randTime int64
	var timestamp uint64

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		select {
		case <-ticker.C:
			lastTraceAsMetric := ins.aggregator.DumpSingle(metricName)
			if len(lastTraceAsMetric) != 2 {
				t.Errorf("Labels Check failed,expected 2 groups , got %d", len(lastTraceAsMetric))
			}

			var t1 *model.DataGroup

			for i := 0; i < len(lastTraceAsMetric); i++ {
				if lastTraceAsMetric[i].Labels.GetStringValue(constlabels.DstIp) == "1.1.1.1" {
					t1 = lastTraceAsMetric[i]

					// value check
					if m, ok := t1.GetMetric(constnames.TraceAsMetric); ok {
						if m.GetInt().Value != randTime {
							t.Errorf("Value check failed")
						}
					} else {
						t.Errorf("Value check failed")
					}

					// timestamp check
					if t1.Timestamp != timestamp {
						t.Errorf("Timestamp check failed")
					}
				}
			}
		}
	}()

	go func() {
		for i := 0; i < 10; i++ {
			randTime2 := rand.Int63n(1000)
			timestamp2 := rand.Uint64()
			singleMetric := makeTraceAsMetricGroup(randTime2, timestamp2, "2.2.2.2")
			ins.aggregator.Aggregate(singleMetric, ins.getSelector(metricName))
			time.Sleep(1 * time.Second)
		}
	}()

	for i := 0; i < 10; i++ {
		randTime = rand.Int63n(1000)
		timestamp = rand.Uint64()
		singleMetric := makeTraceAsMetricGroup(randTime, timestamp, "1.1.1.1")
		ins.aggregator.Aggregate(singleMetric, ins.getSelector(metricName))
		time.Sleep(1 * time.Second)
	}
}
