// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package otelexporter

import (
	"context"

	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter/tools/adapter"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constvalues"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	apitrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (e *OtelExporter) Consume(dataGroup *model.DataGroup) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	if dataGroup == nil {
		// no need consume
		return nil
	}

	dataGroupReceiverCounter.Add(context.Background(), 1, attribute.String("name", dataGroup.Name))
	if ce := e.telemetry.Logger.Check(zap.DebugLevel, ""); ce != nil {
		e.telemetry.Logger.Debug("exporter receives a dataGroup: \n" + dataGroup.String())
	}

	for i := 0; i < len(e.adapters); i++ {
		results, err := e.adapters[i].Adapt(dataGroup, adapter.AttributeList)
		if err != nil {
			e.telemetry.Logger.Error("Failed to adapt dataGroup", zap.Error(err))
		}
		if len(results) > 0 {
			e.Export(results)
		}
	}
	return nil
}

func (e *OtelExporter) Export(results []*adapter.AdaptedResult) {
	for i := 0; i < len(results); i++ {
		result := results[i]
		switch result.ResultType {
		case adapter.Metric:
			e.exportMetric(result)
		case adapter.Trace:
			e.exportTrace(result)
		default:
			e.telemetry.Logger.Error("Unexpected ResultType", zap.String("type", string(result.ResultType)))
		}
		result.Free()
	}
}

func (e *OtelExporter) exportTrace(result *adapter.AdaptedResult) {
	if e.defaultTracer == nil {
		e.telemetry.Logger.Error("Send span failed: this exporter doesn't support Span Data", zap.String("exporter", e.cfg.ExportKind))
		return
	}
	_, span := e.defaultTracer.Start(
		context.Background(),
		constvalues.SpanInfo,
		apitrace.WithAttributes(result.AttrsList...),
	)
	span.End()
}

func (e *OtelExporter) exportMetric(result *adapter.AdaptedResult) {
	measurements := make([]metric.Measurement, 0, len(result.Metrics))
	for s := 0; s < len(result.Metrics); s++ {
		metric := result.Metrics[s]
		if metricKind, ok := e.findInstrumentKind(metric.Name); ok &&
			metricKind == MAGaugeKind &&
			metric.DataType() == model.IntMetricType {
			if result.AttrsMap == nil {
				e.telemetry.Logger.Error("Unexpected Error: no labels find for MAGaugeKind", zap.String("MetricName", metric.Name))
			}
			err := e.instrumentFactory.recordLastValue(metric.Name, &model.DataGroup{
				Name:      metric.Name,
				Metrics:   []*model.Metric{model.NewIntMetric(metric.Name, metric.GetInt().Value)},
				Labels:    result.AttrsMap,
				Timestamp: result.Timestamp,
			})
			if err != nil {
				e.telemetry.Logger.Error("Failed to record lastValue of Metric", zap.String("MetricName", metric.Name), zap.Error(err))
			}
		} else if ok && metric.DataType() == model.IntMetricType {
			measurements = append(measurements, e.instrumentFactory.getInstrument(metric.Name, metricKind).Measurement(metric.GetInt().Value))
		} else if metric.DataType() == model.HistogramMetricType {
			e.telemetry.Logger.Warn("Failed to exporter Metric: can not use otlp-exporter to export histogram Data", zap.String("MetricName", metric.Name))
		} else {
			if ce := e.telemetry.Logger.Check(zapcore.DebugLevel, "Undefined metricKind for this Metric in metric_aggregation_map"); ce != nil {
				ce.Write(zap.String("MetricName", metric.Name))
			}
		}
	}
	if len(measurements) > 0 {
		e.instrumentFactory.meter.RecordBatch(context.Background(), result.AttrsList, measurements...)
	}
}
