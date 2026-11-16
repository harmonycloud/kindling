// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cameraexporter

import (
	"fmt"
	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/pkg/esclient"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"go.uber.org/zap"
)

const Type = "cameraexporter"

type CameraExporter struct {
	config *Config
	writer writer

	telemetry *component.TelemetryTools
}

func New(config interface{}, telemetry *component.TelemetryTools) exporter.Exporter {
	cfg, _ := config.(*Config)
	ret := &CameraExporter{
		config:    cfg,
		telemetry: telemetry,
	}
	switch cfg.Storage {
	case storageElasticsearch:
		writer, err := newEsWriter(cfg.EsConfig)
		if err != nil {
			telemetry.Logger.Panicf("Can't create new cameraexporter with eswriter: %v", err)
		}
		ret.writer = writer
	case storageFile:
		writer, err := newFileWriter(cfg.FileConfig, telemetry.Logger)
		if err != nil {
			telemetry.Logger.Panicf("Can't create new cameraexporter with filewriter: %v", err)
		}
		ret.writer = writer
	}
	return ret
}

func (e *CameraExporter) Consume(dataGroup *model.DataGroup) error {
	if ce := e.telemetry.Logger.Check(zap.DebugLevel, ""); ce != nil {
		e.telemetry.Logger.Debug(dataGroup.String())
	}
	e.writer.write(dataGroup)
	return nil
}

type writer interface {
	write(dataGroup *model.DataGroup)
	name() string
}

type esWriter struct {
	config   *esConfig
	esClient *esclient.EsClient
}

func newEsWriter(cfg *esConfig) (*esWriter, error) {
	client, err := esclient.NewEsClient(cfg.EsHost)
	if err != nil {
		return nil, fmt.Errorf("fail to create elasticsearch client: %w", err)
	}
	ret := &esWriter{
		config:   cfg,
		esClient: client,
	}
	return ret, nil
}

func (ew *esWriter) write(group *model.DataGroup) {
	isSent := group.Labels.GetIntValue(constlabels.IsSent)
	// The data has been sent before, so esExporter will not index it again.
	// But fileExporter will.
	if isSent == 1 {
		return
	}
	index := group.Name
	if ew.config.IndexSuffix != "" {
		index = index + "_" + ew.config.IndexSuffix
	}
	ew.esClient.AddIndexRequestWithParams(index, group)
}

func (ew *esWriter) name() string {
	return storageElasticsearch
}
