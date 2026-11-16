// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"os"

	"go.uber.org/zap/zapcore"
)

var ConsoleEncodingConfig = zapcore.EncoderConfig{
	MessageKey:     "msg",
	LevelKey:       "level",
	TimeKey:        "ts",
	NameKey:        "name",
	CallerKey:      "caller",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.CapitalColorLevelEncoder,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

// ConfigConsoleOutput ConsoleEncoding Configuration
func ConfigConsoleOutput(config zapcore.EncoderConfig, level zapcore.Level) zapcore.Core {
	return zapcore.NewCore(zapcore.NewConsoleEncoder(config), zapcore.AddSync(os.Stdout), level)
}
