// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestTelemetryManager_DebugSelector(t *testing.T) {
	tm := NewTelemetryManager()
	type args struct {
		options []TelemetryOption
	}
	type want struct {
		entry []observer.LoggedEntry
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "EnableDebug",
			args: args{
				options: []TelemetryOption{WithDebug(true)},
			},
			want: want{
				entry: []observer.LoggedEntry{
					{
						Entry:   zapcore.Entry{Level: zap.DebugLevel, Message: "Debug Test"},
						Context: []zap.Field{zap.Int("int", 1)},
					},
				},
			},
		},
		{
			name: "DisableDebug",
			args: args{
				options: []TelemetryOption{WithDebug(false)},
			},
			want: want{
				entry: []observer.LoggedEntry{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withLogger(t, zap.DebugLevel, nil, tt.args.options, tm, func(logger *TelemetryLogger, obs *observer.ObservedLogs) {
				logger.Debug("Debug Test", zap.Int("int", 1))
				assert.Equal(
					t,
					tt.want.entry,
					obs.AllUntimed(),
					"Unexpected log output from Debug log",
				)
			})
		})
	}
}

func TestTelemetryLogger_TestInfo(t *testing.T) {
	tm := NewTelemetryManager()
	withLogger(t, zap.InfoLevel, nil, nil, tm, func(logger *TelemetryLogger, obs *observer.ObservedLogs) {
		logger.Info("Info Test", zap.Int("field", 1))
		assert.Equal(
			t,
			[]observer.LoggedEntry{
				{
					Entry:   zapcore.Entry{Level: zap.InfoLevel, Message: "Info Test"},
					Context: []zap.Field{zap.Int("field", 1)},
				},
			},
			obs.AllUntimed(),
			"Unexpected log output from Info log",
		)
	})
}

func TestTelemetryLogger_TestInfof(t *testing.T) {
	tm := NewTelemetryManager()
	withLogger(t, zap.InfoLevel, nil, nil, tm, func(logger *TelemetryLogger, obs *observer.ObservedLogs) {
		logger.Infof("Info format Test: int field \"%d\" , string field \"%s\"", 1, "testStr")
		assert.Equal(
			t,
			[]observer.LoggedEntry{
				{
					Entry:   zapcore.Entry{Level: zap.InfoLevel, Message: "Info format Test: int field \"1\" , string field \"testStr\""},
					Context: []zap.Field{},
				},
			},
			obs.AllUntimed(),
			"Unexpected log output from Infof log",
		)
	})
}

func withLogger(t testing.TB, e zapcore.LevelEnabler, zapOpt []zap.Option, opts []TelemetryOption, tm *TelemetryManager, f func(*TelemetryLogger, *observer.ObservedLogs)) {
	fac, logs := observer.New(e)
	tm.Logger = zap.New(fac, zapOpt...)
	tools := tm.getToolsWithOption(opts...)
	f(tools.Logger, logs)
}
