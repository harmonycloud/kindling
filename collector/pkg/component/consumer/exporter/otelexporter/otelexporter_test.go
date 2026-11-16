// Copyright (C) 2022-2024 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package otelexporter

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.uber.org/zap"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter/tools/adapter"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
	"github.com/Kindling-project/kindling/collector/pkg/model/constvalues"
)

func InitOtelExporter(t *testing.T) exporter.Exporter {
	configPath := "testdata/kindling-collector-config.yml"
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		t.Fatalf("error happened when reading config: %v", err)
	}
	config := &Config{}
	err = viper.UnmarshalKey("exporters.otelexporter", config)
	if err != nil {
		t.Fatalf("error happened when unmarshaling config: %v", err)
	}
	return NewExporter(config, component.NewDefaultTelemetryTools())
}

func TestConsumeAggNetMetricGroup(t *testing.T) {
	latencyArray := []int64{1e6, 10e6, 20e6, 50e6, 100e6, 500e6}
	exp := InitOtelExporter(t)
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
	exp := InitOtelExporter(t)
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
	metricsGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService"+strconv.Itoa(i))

	metricsGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId"+strconv.Itoa(i))

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
			model.NewIntMetric(constvalues.RequestTotalTime, int64(i)),
			model.NewIntMetric(constvalues.RequestIo, 4500),
			model.NewIntMetric(constvalues.RequestCount, 4500),
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 19900909090,
	}
	metricsGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService"+strconv.Itoa(i))

	metricsGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId"+strconv.Itoa(i))

	metricsGroup.Labels.AddStringValue(constlabels.Protocol, "http")
	metricsGroup.Labels.AddStringValue(constlabels.StatusCode, "200")

	// Topology data preferentially use D Nat Ip and D Nat Port
	metricsGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")
	metricsGroup.Labels.AddIntValue(constlabels.DstPort, 8081)
	return metricsGroup
}

func appendToFile(filename, text string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(text)
	return err
}

func BenchmarkOtelExporter_Consume(b *testing.B) {
	dimension := 100000

	cfg := &Config{
		ExportKind:   StdoutKindExporter,
		PromCfg:      nil,
		OtlpGrpcCfg:  nil,
		StdoutCfg:    &StdoutConfig{CollectPeriod: 30 * time.Second},
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

	telemetry := component.NewDefaultTelemetryTools()
	myExporter, _ := newExporters(context.Background(), cfg, telemetry)

	cont := controller.New(
		otelprocessor.NewFactory(simple.NewWithHistogramDistribution(
			histogram.WithExplicitBoundaries(exponentialInt64NanosecondsBoundaries),
		), myExporter.metricExporter),
		controller.WithExporter(myExporter.metricExporter),
		controller.WithCollectPeriod(cfg.StdoutCfg.CollectPeriod),
		controller.WithResource(nil),
	)

	otelexporter := &OtelExporter{
		cfg:                  cfg,
		metricController:     cont,
		traceProvider:        nil,
		defaultTracer:        nil,
		customLabels:         nil,
		instrumentFactory:    newInstrumentFactory(cont.Meter(MeterName), telemetry, nil),
		metricAggregationMap: cfg.MetricAggregationMap,
		telemetry:            component.NewDefaultTelemetryTools(),
		adapters: []adapter.Adapter{
			adapter.NewNetAdapter(nil, &adapter.NetAdapterConfig{
				StoreTraceAsMetric: cfg.AdapterConfig.NeedTraceAsMetric,
				StoreTraceAsSpan:   cfg.AdapterConfig.NeedTraceAsResourceSpan,
				StorePodDetail:     cfg.AdapterConfig.NeedPodDetail,
				StoreExternalSrcIP: cfg.AdapterConfig.StoreExternalSrcIP,
			}),
			adapter.NewSimpleAdapter([]string{constnames.TcpRttMetricGroupName, constnames.TcpRetransmitMetricGroupName,
				constnames.TcpDropMetricGroupName}, nil),
		},
	}

	if err := cont.Start(context.Background()); err != nil {
		telemetry.Logger.Panic("failed to start controller:", zap.Error(err))
	}
	newSelfMetrics(otelexporter.telemetry.MeterProvider)

	//mockMetric := make(chan *model.DataGroup, 100)
	//MockMetric(mockMetric, 800, 1000, 10*time.Minute)
	recordCounter := 0

	metricsGroupsSlice := make([]*model.DataGroup, dimension)

	for i := 0; i < dimension; i++ {
		metricsGroup := &model.DataGroup{
			Name: constnames.AggregatedNetRequestMetricGroup,
			Metrics: []*model.Metric{
				model.NewIntMetric(constvalues.ResponseIo, 1234567891),
				model.NewIntMetric(constvalues.RequestTotalTime, int64(i)),
				model.NewIntMetric(constvalues.RequestIo, 4500),
				model.NewIntMetric(constvalues.RequestCount, 4500),
			},
			Labels:    model.NewAttributeMap(),
			Timestamp: 19900909090,
		}
		metricsGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService"+strconv.Itoa(i))

		metricsGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer"+strconv.Itoa(i))
		metricsGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId"+strconv.Itoa(i))

		metricsGroup.Labels.AddStringValue(constlabels.Protocol, "http")
		metricsGroup.Labels.AddStringValue(constlabels.StatusCode, "200")

		// Topology data preferentially use D Nat Ip and D Nat Port
		metricsGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")
		metricsGroup.Labels.AddIntValue(constlabels.DstPort, 8081)

		metricsGroupsSlice[i] = metricsGroup
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = otelexporter.Consume(metricsGroupsSlice[recordCounter%dimension])
		recordCounter++
	}

	log.Printf("Test Finished!")
}

func BenchmarkMemTest(b *testing.B) {
	testDuration := time.Minute
	endTime := time.Now().Add(30 * testDuration)
	metricCounter := 4
	recordCounter := 0
	memTicker := time.NewTicker(1 * time.Second)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	go func() {
		for range memTicker.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			stats := fmt.Sprintf("Alloc = %v MiB, Metric Count: %d, Record Count: %d\n", m.Alloc/1024/1024, metricCounter, recordCounter)
			log.Printf(stats)
			err := appendToFile("metrics_data.txt", stats)
			if err != nil {
				log.Println("Error writing to file:", err)
			}
		}
	}()

	cfg := &Config{
		ExportKind: PrometheusKindExporter,
		PromCfg: &PrometheusConfig{
			Port:       ":9500",
			WithMemory: false,
			MemCleanUpConfig: &MemCleanUpConfig{
				RestartPeriod: 12,
			},
		},
		OtlpGrpcCfg:  &OtlpGrpcConfig{CollectPeriod: 15 * time.Second, Endpoint: "10.10.10.10:8080"},
		StdoutCfg:    &StdoutConfig{CollectPeriod: 15 * time.Second},
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
			NeedTraceAsResourceSpan: false,
			NeedTraceAsMetric:       true,
			NeedPodDetail:           true,
			StoreExternalSrcIP:      false,
		},
	}

	telemetry := component.NewDefaultTelemetryTools()
	otelexporter := NewExporter(cfg, telemetry)

	i := 0

	metricsGroup := &model.DataGroup{
		Name: constnames.AggregatedNetRequestMetricGroup,
		Metrics: []*model.Metric{
			model.NewIntMetric(constvalues.ResponseIo, 1234567891),
			model.NewIntMetric(constvalues.RequestTotalTime, int64(i)),
			model.NewIntMetric(constvalues.RequestIo, 4500),
			model.NewIntMetric(constvalues.RequestCount, 4500),
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 19900909090,
	}

	metricsGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcNamespace, "test-SrcNamespace"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcPod, "test-SrcPod"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcService, "test-SrcService"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcIp, "test-SrcIp"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstNode, "test-DstNode"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstNamespace, "test-DstNamespace"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstPod, "test-DstPod"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.DstService, "test-DstService"+strconv.Itoa(i))

	metricsGroup.Labels.AddStringValue(constlabels.SrcContainer, "test-SrcContainer"+strconv.Itoa(i))
	metricsGroup.Labels.AddStringValue(constlabels.SrcContainerId, "test-SrcContainerId"+strconv.Itoa(i))

	metricsGroup.Labels.AddStringValue(constlabels.Protocol, "http")
	metricsGroup.Labels.AddStringValue(constlabels.StatusCode, "200")

	// Topology data preferentially use D Nat Ip and D Nat Port
	metricsGroup.Labels.AddStringValue(constlabels.DstIp, "test-DnatIp")

	for time.Now().Before(endTime) {
		metricsGroup.Labels.UpdateAddStringValue(constlabels.SrcNode, "test-SrcNode"+strconv.Itoa(i))
		metricsGroup.Labels.UpdateAddStringValue(constlabels.SrcNamespace, "test-SrcNamespace"+strconv.Itoa(i))
		metricsGroup.Labels.UpdateAddStringValue(constlabels.SrcPod, "test-SrcPod"+strconv.Itoa(i))
		metricsGroup.Labels.UpdateAddStringValue(constlabels.SrcWorkloadName, "test-SrcWorkloadName"+strconv.Itoa(i))
		metricsGroup.Labels.UpdateAddStringValue(constlabels.SrcWorkloadKind, "test-SrcWorkloadKind"+strconv.Itoa(i))
		metricsGroup.Labels.UpdateAddStringValue(constlabels.SrcService, "test-SrcService"+strconv.Itoa(i))
		metricsGroup.Labels.UpdateAddStringValue(constlabels.SrcIp, "test-SrcIp"+strconv.Itoa(i))
		metricsGroup.Labels.UpdateAddStringValue(constlabels.DstNode, "test-DstNode"+strconv.Itoa(i))
		metricsGroup.Labels.UpdateAddStringValue(constlabels.DstNamespace, "test-DstNamespace"+strconv.Itoa(i))
		metricsGroup.Labels.UpdateAddStringValue(constlabels.DstPod, "test-DstPod"+strconv.Itoa(i))
		metricsGroup.Labels.UpdateAddStringValue(constlabels.DstWorkloadName, "test-DstWorkloadName"+strconv.Itoa(i))
		metricsGroup.Labels.UpdateAddStringValue(constlabels.DstWorkloadKind, "test-DstWorkloadKind"+strconv.Itoa(i))
		metricsGroup.Labels.UpdateAddStringValue(constlabels.DstService, "test-DstService"+strconv.Itoa(i))

		metricsGroup.Labels.UpdateAddStringValue(constlabels.SrcContainer, "test-SrcContainer"+strconv.Itoa(i))
		metricsGroup.Labels.UpdateAddStringValue(constlabels.SrcContainerId, "test-SrcContainerId"+strconv.Itoa(i))

		metricsGroup.Labels.UpdateAddStringValue(constlabels.Protocol, "http")
		metricsGroup.Labels.UpdateAddStringValue(constlabels.StatusCode, "200")

		newMetricsGroup := &model.DataGroup{
			Name: constnames.AggregatedNetRequestMetricGroup,
			Metrics: []*model.Metric{
				model.NewIntMetric(constvalues.ResponseIo, 1234567891),
				model.NewIntMetric(constvalues.RequestTotalTime, int64(i)),
				model.NewIntMetric(constvalues.RequestIo, 4500),
				model.NewIntMetric(constvalues.RequestCount, 4500),
			},
			Labels:    metricsGroup.Labels,
			Timestamp: 19900909090,
		}

		_ = otelexporter.Consume(newMetricsGroup)
		recordCounter++
		i++
		time.Sleep(time.Second) //event amount control
	}

	log.Printf("Test Finished!")
}

func BenchmarkLabelTest(b *testing.B) {
	testDuration := time.Minute
	endTime := time.Now().Add(30 * testDuration)
	metricCounter := 4
	recordCounter := 0
	memTicker := time.NewTicker(1 * time.Second)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	go func() {
		for range memTicker.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			stats := fmt.Sprintf("Alloc = %v MiB, Metric Count: %d, Record Count: %d\n", m.Alloc/1024/1024, metricCounter, recordCounter)
			log.Printf(stats)
			err := appendToFile("metrics_data.txt", stats)
			if err != nil {
				log.Println("Error writing to file:", err)
			}
		}
	}()

	cfg := &Config{
		ExportKind: PrometheusKindExporter,
		PromCfg: &PrometheusConfig{
			Port:       ":9500",
			WithMemory: false,
			MemCleanUpConfig: &MemCleanUpConfig{
				RestartPeriod: 12,
			},
		},
		OtlpGrpcCfg:  &OtlpGrpcConfig{CollectPeriod: 15 * time.Second, Endpoint: "10.10.10.10:8080"},
		StdoutCfg:    &StdoutConfig{CollectPeriod: 15 * time.Second},
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
			NeedTraceAsResourceSpan: false,
			NeedTraceAsMetric:       true,
			NeedPodDetail:           true,
			StoreExternalSrcIP:      false,
		},
	}

	telemetry := component.NewDefaultTelemetryTools()
	otelexporter := NewExporter(cfg, telemetry)

	i := 0

	metricsGroup := &model.DataGroup{
		Name: constnames.AggregatedNetRequestMetricGroup,
		Metrics: []*model.Metric{
			model.NewIntMetric(constvalues.RequestTotalTime, int64(i)),
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 19900909090,
	}

	metricsGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode"+strconv.Itoa(i))

	for time.Now().Before(endTime) {
		metricsGroup.Labels.UpdateAddStringValue(constlabels.SrcNode, "test-SrcNode"+strconv.Itoa(i))

		newMetricsGroup := &model.DataGroup{
			Name: constnames.AggregatedNetRequestMetricGroup,
			Metrics: []*model.Metric{
				model.NewIntMetric(constvalues.RequestTotalTime, int64(i)),
			},
			Labels:    metricsGroup.Labels,
			Timestamp: 19900909090,
		}

		_ = otelexporter.Consume(newMetricsGroup)
		recordCounter++
		i++
		time.Sleep(time.Second) //event amount control
	}

	log.Printf("Test Finished!")
}

func BenchmarkMetricTest(b *testing.B) {
	testDuration := time.Minute
	endTime := time.Now().Add(30 * testDuration)
	metricCounter := 4
	recordCounter := 0
	memTicker := time.NewTicker(1 * time.Second)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	go func() {
		for range memTicker.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			stats := fmt.Sprintf("Alloc = %v MiB, Metric Count: %d, Record Count: %d\n", m.Alloc/1024/1024, metricCounter, recordCounter)
			log.Printf(stats)
			err := appendToFile("metrics_data.txt", stats)
			if err != nil {
				log.Println("Error writing to file:", err)
			}
		}
	}()

	cfg := &Config{
		ExportKind: PrometheusKindExporter,
		PromCfg: &PrometheusConfig{
			Port:       ":9500",
			WithMemory: false,
			MemCleanUpConfig: &MemCleanUpConfig{
				RestartPeriod: 30,
			},
		},
		OtlpGrpcCfg:  &OtlpGrpcConfig{CollectPeriod: 15 * time.Second, Endpoint: "10.10.10.10:8080"},
		StdoutCfg:    &StdoutConfig{CollectPeriod: 15 * time.Second},
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
			NeedTraceAsResourceSpan: false,
			NeedTraceAsMetric:       true,
			NeedPodDetail:           true,
			StoreExternalSrcIP:      false,
		},
	}

	telemetry := component.NewDefaultTelemetryTools()
	otelexporter := NewExporter(cfg, telemetry)

	i := 0

	metricsGroup := &model.DataGroup{
		Name: constnames.AggregatedNetRequestMetricGroup,
		Metrics: []*model.Metric{
			model.NewIntMetric(constvalues.RequestCount, 1),
		},
		Labels:    model.NewAttributeMap(),
		Timestamp: 19900909090,
	}

	metricsGroup.Labels.AddStringValue(constlabels.SrcNode, "test-SrcNode")

	for time.Now().Before(endTime) {
		metricsGroup.Labels.UpdateAddStringValue(constlabels.SrcNode, "test-SrcNode")

		newMetricsGroup := &model.DataGroup{
			Name: constnames.AggregatedNetRequestMetricGroup,
			Metrics: []*model.Metric{
				model.NewIntMetric(constvalues.RequestCount, 1),
			},
			Labels:    metricsGroup.Labels,
			Timestamp: 19900909090,
		}

		_ = otelexporter.Consume(newMetricsGroup)
		recordCounter++
		i++
		time.Sleep(time.Second) //event amount control
	}

	log.Printf("Test Finished!")
}
