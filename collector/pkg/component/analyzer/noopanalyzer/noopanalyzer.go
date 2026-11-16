// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package noopanalyzer

import (
	"fmt"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"go.uber.org/zap/zapcore"
)

const Type analyzer.Type = "noopanalyzer"

type NoopAnalyzer struct {
	cfg           *Config
	nextConsumers []consumer.Consumer
	telemetry     *component.TelemetryTools
}

func New(cfg interface{}, telemetry *component.TelemetryTools, consumer []consumer.Consumer) analyzer.Analyzer {
	config, ok := cfg.(*Config)
	if !ok {
		telemetry.Logger.Panic("Cannot convert noopanalyzer config")
	}
	return &NoopAnalyzer{
		cfg:           config,
		nextConsumers: consumer,
		telemetry:     telemetry,
	}
}

func (a *NoopAnalyzer) Start() error {
	return nil
}

func (a *NoopAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
	if ce := a.telemetry.Logger.Check(zapcore.InfoLevel, ""); ce != nil {
		a.telemetry.Logger.Debug(fmt.Sprintf("Receive event: %+v", event))
	}
	for _, nextConsumer := range a.nextConsumers {
		nextConsumer.Consume(&model.DataGroup{})
	}
	return nil
}

func (a *NoopAnalyzer) Shutdown() error {
	return nil
}

func (a *NoopAnalyzer) Type() analyzer.Type {
	return Type
}

func (a *NoopAnalyzer) ConsumableEvents() []string {
	return []string{analyzer.ConsumeAllEvents}
}

type Config struct {
}
