// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package esexporter

import (
	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/pkg/esclient"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const Type = "esexporter"

type EsExporter struct {
	esClient *esclient.EsClient

	telemetry *component.TelemetryTools
	config    *Config
}

func New(config interface{}, telemetry *component.TelemetryTools) exporter.Exporter {
	cfg, _ := config.(*Config)
	ret := &EsExporter{
		config:    cfg,
		telemetry: telemetry,
	}
	client, err := esclient.NewEsClient(cfg.GetEsHost())
	if err != nil {
		telemetry.Logger.Errorf("Fail to create elasticsearch client: ", zap.Error(err))
	}
	ret.esClient = client
	return ret
}

func (e *EsExporter) Consume(dataGroup *model.DataGroup) error {
	if ce := e.telemetry.Logger.Check(zapcore.DebugLevel, "Receiver DataGroup"); ce != nil {
		ce.Write(
			zap.String("dataGroup", dataGroup.String()),
		)
	}
	if dataGroup == nil {
		// no need consume
		return nil
	}

	switch dataGroup.Name {
	case constnames.AggregatedNetRequestMetricGroup:
		// We don't care about metrics now.
		return nil
	case constnames.SingleNetRequestMetricGroup:
		e.sendTrace(dataGroup)
		return nil
	}
	return nil
}

func (e *EsExporter) sendTrace(dataGroup *model.DataGroup) {
	e.telemetry.Logger.Info("Will send a trace to ElasticSearch")
	if e.esClient == nil {
		return
	}
	trace := TraceData{
		Name:      dataGroup.Name,
		Metrics:   make(map[string]int64),
		Labels:    dataGroup.Labels.ToStringMap(),
		Timestamp: dataGroup.Timestamp,
	}
	for _, metric := range dataGroup.Metrics {
		trace.Metrics[metric.Name] = metric.GetInt().Value
	}
	if e.esClient != nil {
		_ = e.esClient.IndexJson(e.config.GetEsIndexName(), trace)
	} else {
		e.telemetry.Logger.Infof("EsClient is nil, the trace should have been sent: %v", trace)
	}
}

type TraceData struct {
	Name      string            `json:"name"`
	Metrics   map[string]int64  `json:"metrics"`
	Labels    map[string]string `json:"labels"`
	Timestamp uint64            `json:"timestamp"`
}
