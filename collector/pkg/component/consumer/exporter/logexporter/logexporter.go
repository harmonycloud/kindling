// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package logexporter

import (
	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/exporter"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"go.uber.org/zap/zapcore"
)

const Type = "logexporter"

type LogExporter struct {
	telemetry *component.TelemetryTools
}

func New(config interface{}, telemetry *component.TelemetryTools) exporter.Exporter {
	return &LogExporter{
		telemetry: telemetry,
	}
}

func (e *LogExporter) Consume(dataGroup *model.DataGroup) error {
	if ce := e.telemetry.Logger.Check(zapcore.DebugLevel, "Receiver DataGroup"); ce != nil {
		e.telemetry.Logger.Debug("Receiver DataGroup:\n " + dataGroup.String())
	}
	return nil
}

type Config struct {
}
