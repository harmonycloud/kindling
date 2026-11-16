// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var consoleLogLevel zapcore.Level
var lumberJackLogLevel zapcore.Level

type Config struct {
	ConsoleLogLevel  string             `mapstructure:"console_level"`
	FileLogLevel     string             `mapstructure:"file_level"`
	Selector         []string           `mapstructure:"debug_selector"`
	LumberJackConfig *lumberjack.Logger `mapstructure:"file_rotation"`
}

// CreateCombineLogger Create a merged Logger instance
func CreateCombineLogger(config *lumberjack.Logger) *zap.Logger {
	core := zapcore.NewTee(
		ConfigConsoleOutput(ConsoleEncodingConfig, consoleLogLevel),
		ConfigFileRotationOutput(lumberJackEncodingConfig, config, lumberJackLogLevel),
	)
	return zap.New(core).WithOptions(zap.AddCaller(), zap.AddCallerSkip(1))
}

func CreateConsoleLogger() *zap.Logger {
	core := ConfigConsoleOutput(ConsoleEncodingConfig, consoleLogLevel)
	return zap.New(core).WithOptions(zap.AddCaller(), zap.AddCallerSkip(1))
}

func CreateFileRotationLogger(config *lumberjack.Logger) *zap.Logger {
	core := ConfigFileRotationOutput(lumberJackEncodingConfig, config, lumberJackLogLevel)
	return zap.New(core).WithOptions(zap.AddCaller(), zap.AddCallerSkip(1))
}

func InitLogger(config Config) *zap.Logger {
	if config.ConsoleLogLevel != "none" && config.FileLogLevel != "none" {
		consoleLogLevel = toLogLevel(config.ConsoleLogLevel)
		lumberJackLogLevel = toLogLevel(config.FileLogLevel)
		logger := CreateCombineLogger(config.LumberJackConfig)
		logger.Sugar().Infof("Log Initialize Success! ConsoleLevel: %s,FileRotationLevel: %s\nFileRotationConfig: \t%s", consoleLogLevel.String(), lumberJackLogLevel.String(), ToString(config.LumberJackConfig))
		return logger
	} else if config.FileLogLevel != "none" {
		lumberJackLogLevel = toLogLevel(config.FileLogLevel)
		logger := CreateFileRotationLogger(config.LumberJackConfig)
		logger.Sugar().Infof("Log Initialize Success! ConsoleLevel: %s,FileRotationLevel: %s,FileRotationConfig: \t%s", "none", lumberJackLogLevel.String(), ToString(config.LumberJackConfig))
		return logger
	} else if config.ConsoleLogLevel != "none" {
		consoleLogLevel = toLogLevel(config.ConsoleLogLevel)
		logger := CreateConsoleLogger()
		logger.Sugar().Infof("Log Initialize Success! ConsoleLevel: %s,FileRotationLevel: %s", consoleLogLevel.String(), "none")
		return logger
	} else {
		// Both none, only record the panic log to console
		consoleLogLevel = zapcore.DPanicLevel
		logger := CreateConsoleLogger()
		logger.Sugar().Infof("Log Initialize Success! ConsoleLevel: %s,FileRotationLevel: %s", consoleLogLevel.String(), "none")
		return logger
	}
}

func CreateDefaultLogger() *zap.Logger {
	core := ConfigConsoleOutput(ConsoleEncodingConfig, zapcore.InfoLevel)
	return zap.New(core).WithOptions(zap.AddCaller(), zap.AddCallerSkip(1))
}

func toLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	}
	return zapcore.InfoLevel
}
