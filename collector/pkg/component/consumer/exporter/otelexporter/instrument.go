// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package otelexporter

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/Kindling-project/kindling/collector/pkg/aggregator"
	"github.com/Kindling-project/kindling/collector/pkg/aggregator/defaultaggregator"
	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter/tools/adapter"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
	"go.opentelemetry.io/otel/attribute"

	"go.opentelemetry.io/otel/metric"
)

var traceAsMetricHelp = "Describe a single request which is abnormal. " +
	"For status labels, number '1', '2' and '3' stands for Green, Yellow and Red respectively."

type instrumentFactory struct {
	instruments  sync.Map
	meter        metric.Meter
	customLabels []attribute.KeyValue
	telemetry    *component.TelemetryTools

	aggregator *defaultaggregator.DefaultAggregator

	traceAsMetricSelector *aggregator.LabelSelectors
	TcpRttMillsSelector   *aggregator.LabelSelectors
	K8sWorkloadSelector   *aggregator.LabelSelectors
}

func newInstrumentFactory(meter metric.Meter, telemetry *component.TelemetryTools, customLabels []attribute.KeyValue) *instrumentFactory {
	return &instrumentFactory{
		instruments:  sync.Map{},
		meter:        meter,
		customLabels: customLabels,
		telemetry:    telemetry,
		aggregator: defaultaggregator.NewDefaultAggregator(&defaultaggregator.AggregatedConfig{
			KindMap: map[string][]defaultaggregator.KindConfig{
				constnames.TcpRttMetricName: {
					{Kind: defaultaggregator.LastKind, OutputName: constnames.TcpRttMetricName},
				},
				constnames.TraceAsMetric: {
					{Kind: defaultaggregator.LastKind, OutputName: constnames.TraceAsMetric},
				},
				constnames.K8sWorkLoadMetricName: {
					{Kind: defaultaggregator.LastKind, OutputName: constnames.K8sWorkLoadMetricName},
				},
			},
		}),

		traceAsMetricSelector: newTraceAsMetricSelectors(),
		TcpRttMillsSelector:   newTcpRttMicroSecondsSelectors(),
		K8sWorkloadSelector:   newK8sWorkloadSelector(),
	}
}
func (i *instrumentFactory) getInstrument(metricName string, kind MetricAggregationKind) instrument {
	if ins, ok := i.instruments.Load(metricName); ok {
		return ins.(instrument)
	} else {
		newIns := i.createNewInstrument(metricName, kind)
		i.instruments.Store(metricName, newIns)
		return newIns
	}
}

func (i *instrumentFactory) createNewInstrument(metricName string, kind MetricAggregationKind) instrument {
	switch kind {
	case MACounterKind:
		ins := metric.Must(i.meter).NewInt64Counter(metricName)
		return &counterInstrument{instrument: &ins}
	case MAHistogramKind:
		ins := metric.Must(i.meter).NewInt64Histogram(metricName)
		return &histogramInstrument{instrument: &ins}
	default:
		ins := metric.Must(i.meter).NewInt64Counter(metricName)
		return &counterInstrument{instrument: &ins}
	}
}

// recordLastValue Only support TraceAsMetric and TcpRttMills now
func (i *instrumentFactory) recordLastValue(metricName string, singleMetric *model.DataGroup) error {
	if !i.aggregator.CheckExist(singleMetric.Name) {
		metric.Must(i.meter).NewInt64GaugeObserver(metricName, func(ctx context.Context, result metric.Int64ObserverResult) {
			dumps := i.aggregator.DumpSingle(singleMetric.Name)
			if dumps == nil {
				return
			}
			for s := 0; s < len(dumps); s++ {
				if len(dumps[s].Metrics) > 0 {
					result.Observe(dumps[s].Metrics[0].GetInt().Value, adapter.GetLabels(dumps[s].Labels, i.customLabels)...)
				}
			}
		}, WithDescription(metricName))
	}

	if selector := i.getSelector(metricName); selector != nil {
		i.aggregator.Aggregate(singleMetric, selector)
		return nil
	} else {
		return errors.New(fmt.Sprintf("no matched Selector has been be defined for %s", metricName))
	}
}

func WithDescription(metricName string) metric.InstrumentOption {
	var option metric.InstrumentOption
	switch metricName {
	case constnames.TraceAsMetric:
		option = metric.WithDescription(traceAsMetricHelp)
	default:
		option = metric.WithDescription("")
	}
	return option
}

func (i *instrumentFactory) getSelector(metricName string) *aggregator.LabelSelectors {
	switch metricName {
	case constnames.TraceAsMetric:
		return i.traceAsMetricSelector
	case constnames.TcpRttMetricName:
		return i.TcpRttMillsSelector
	case constnames.K8sWorkLoadMetricName:
		return i.K8sWorkloadSelector
	default:
		return nil
	}
}

func newTraceAsMetricSelectors() *aggregator.LabelSelectors {
	return aggregator.NewLabelSelectors(
		aggregator.LabelSelector{Name: constlabels.Protocol, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.IsServer, VType: aggregator.BooleanType},
		aggregator.LabelSelector{Name: constlabels.ContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcContainer, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DnatIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DnatPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DstContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstContainer, VType: aggregator.StringType},

		aggregator.LabelSelector{Name: constlabels.RequestReqxferStatus, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.RequestProcessingStatus, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.ResponseRspxferStatus, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.RequestDurationStatus, VType: aggregator.StringType},

		aggregator.LabelSelector{Name: constlabels.RequestContent, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.ResponseContent, VType: aggregator.StringType},
	)
}

func newTcpRttMicroSecondsSelectors() *aggregator.LabelSelectors {
	return aggregator.NewLabelSelectors(
		aggregator.LabelSelector{Name: constlabels.SrcNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.SrcContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.SrcContainer, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNode, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNodeIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstNamespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPod, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadName, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstWorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstService, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstIp, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstPort, VType: aggregator.IntType},
		aggregator.LabelSelector{Name: constlabels.DstContainerId, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.DstContainer, VType: aggregator.StringType},
	)
}

func newK8sWorkloadSelector() *aggregator.LabelSelectors {
	return aggregator.NewLabelSelectors(
		aggregator.LabelSelector{Name: constlabels.Namespace, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.WorkloadKind, VType: aggregator.StringType},
		aggregator.LabelSelector{Name: constlabels.WorkloadName, VType: aggregator.StringType},
	)
}

type instrument interface {
	Measurement(value int64) metric.Measurement
}

type counterInstrument struct {
	instrument *metric.Int64Counter
}

func (c *counterInstrument) Measurement(value int64) metric.Measurement {
	return c.instrument.Measurement(value)
}

type histogramInstrument struct {
	instrument *metric.Int64Histogram
}

func (h *histogramInstrument) Measurement(value int64) metric.Measurement {
	return h.instrument.Measurement(value)
}

type AsyncMetricGroup struct {
}
