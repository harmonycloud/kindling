// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"log"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Kindling-project/kindling/collector/pkg/observability"
	"github.com/Kindling-project/kindling/collector/pkg/observability/logger"
)

const (
	ObservabilityConfig = "observability"
	LogKey              = "logger"
	MetricKey           = "opentelemetry"
)

type TelemetryManager struct {
	MeterProvider metric.MeterProvider
	Logger        *zap.Logger
	Selector      []string
}

func NewTelemetryManager() *TelemetryManager {
	return &TelemetryManager{
		MeterProvider: metric.NewNoopMeterProvider(),
		Logger:        logger.CreateDefaultLogger(),
		Selector:      make([]string, 0),
	}
}

func (t *TelemetryManager) ConstructConfig(viper *viper.Viper) {
	t.initLogger(viper)
	t.initProvider(viper)
}

func (t *TelemetryManager) initLogger(viper *viper.Viper) {
	var loggerConfig = logger.Config{}
	key := ObservabilityConfig + "." + LogKey
	err := viper.UnmarshalKey(key, &loggerConfig, func(config *mapstructure.DecoderConfig) {
		config.ZeroFields = true
	})
	if err != nil {
		log.Printf("Error happened when reading logger config, and default config will be used: %v", err)
	}
	t.Selector = loggerConfig.Selector
	t.Logger = logger.InitLogger(loggerConfig)
}

func (t *TelemetryManager) initProvider(viper *viper.Viper) {
	var config = &observability.DefaultConfig
	key := ObservabilityConfig + "." + MetricKey
	err := viper.UnmarshalKey(key, config, func(config *mapstructure.DecoderConfig) {
		config.ZeroFields = true
	})
	if err != nil {
		log.Printf("Error happened when reading observability config, and default config will be used: %v", err)
	}
	meterProvider, err := observability.InitTelemetry(t.Logger, config)
	if err != nil {
		log.Printf("Error happened when initializing meter provider: %v", err)
		return
	}
	t.MeterProvider = meterProvider
	// Here we set a global meter provider to make it accessible to all components
	global.SetMeterProvider(meterProvider)
}

func (t *TelemetryManager) GetGlobalTelemetryTools() *TelemetryTools {
	return t.getToolsWithOption(
		WithDebug(true),
	)
}

func (t *TelemetryManager) GetTelemetryTools(component string) *TelemetryTools {
	options := make([]TelemetryOption, 0)
	// DebugSelector
	options = append(options, t.SelectorOption(component))

	return t.getToolsWithOption(options...)
}

func (t *TelemetryManager) SelectorOption(component string) TelemetryOption {
	if len(t.Selector) == 0 {
		return WithDebug(true)
	}
	for _, selectedComponent := range t.Selector {
		if selectedComponent == component {
			return WithDebug(true)
		}
	}
	return WithDebug(false)
}

func (t *TelemetryManager) getToolsWithOption(opts ...TelemetryOption) *TelemetryTools {
	newSubTool := &TelemetryTools{
		MeterProvider: t.MeterProvider,
		Logger: &TelemetryLogger{
			logger: t.Logger,
			sugar:  t.Logger.Sugar(),
		},
	}

	for _, opt := range opts {
		opt(newSubTool)
	}

	return newSubTool
}

type TelemetryLogger struct {
	logger      *zap.Logger
	sugar       *zap.SugaredLogger
	EnableDebug bool
}

type TelemetryTools struct {
	MeterProvider metric.MeterProvider
	Logger        *TelemetryLogger
}

func (t *TelemetryTools) GetZapLogger() *zap.Logger {
	return t.Logger.logger
}

func (t *TelemetryLogger) Debug(msg string, fields ...zap.Field) {
	if t.EnableDebug {
		t.logger.Debug(msg, fields...)
	}
}

func (t *TelemetryLogger) Info(msg string, fields ...zap.Field) {
	t.logger.Info(msg, fields...)
}

func (t *TelemetryLogger) Warn(msg string, fields ...zap.Field) {
	t.logger.Warn(msg, fields...)
}

func (t *TelemetryLogger) Error(msg string, fields ...zap.Field) {
	t.logger.Error(msg, fields...)
}

func (t *TelemetryLogger) Panic(msg string, fields ...zap.Field) {
	t.logger.Panic(msg, fields...)
}

func (t *TelemetryLogger) Debugf(template string, args ...interface{}) {
	if t.EnableDebug {
		t.sugar.Debugf(template, args...)
	}
}

func (t *TelemetryLogger) Infof(template string, args ...interface{}) {
	t.sugar.Infof(template, args...)
}

func (t *TelemetryLogger) Warnf(template string, args ...interface{}) {
	t.sugar.Warnf(template, args...)
}

func (t *TelemetryLogger) Errorf(template string, args ...interface{}) {
	t.sugar.Errorf(template, args...)
}

func (t *TelemetryLogger) Panicf(template string, args ...interface{}) {
	t.sugar.Panicf(template, args...)
}

func (t *TelemetryLogger) Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	if t.EnableDebug {
		return t.logger.Check(lvl, msg)
	} else {
		return nil
	}
}

type TelemetryOption func(*TelemetryTools)

func WithDebug(enableDebug bool) TelemetryOption {
	return func(tt *TelemetryTools) {
		tt.Logger.EnableDebug = enableDebug
	}
}

func NewDefaultTelemetryTools() *TelemetryTools {
	logger := logger.CreateDefaultLogger()
	return &TelemetryTools{
		Logger: &TelemetryLogger{
			logger:      logger,
			sugar:       logger.Sugar(),
			EnableDebug: true,
		},
		MeterProvider: metric.NewNoopMeterProvider(),
	}
}
