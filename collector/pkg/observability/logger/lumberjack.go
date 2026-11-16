// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"fmt"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var lumberJackEncodingConfig = zapcore.EncoderConfig{
	MessageKey:     "msg",
	LevelKey:       "level",
	TimeKey:        "ts",
	NameKey:        "name",
	CallerKey:      "caller",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.CapitalLevelEncoder,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

// ConfigFileRotationOutput RotationFileEncoding Configuration
func ConfigFileRotationOutput(config zapcore.EncoderConfig, jackConfig *lumberjack.Logger, level zapcore.Level) zapcore.Core {
	fileWriteSyncer := zapcore.AddSync(jackConfig)
	return zapcore.NewCore(zapcore.NewConsoleEncoder(config), fileWriteSyncer, level)
}

func ToString(logger *lumberjack.Logger) string {
	return fmt.Sprintf("LogFile: %s\tMaxSize: %dm\tbackup: %d\tMaxAge: %dday", logger.Filename, logger.MaxSize, logger.MaxBackups, logger.MaxAge)
}
