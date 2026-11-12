// Package interpreterconfig aggregates per-interpreter configuration.
package interpreterconfig

import (
	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter/python"
)

type interpreterCfg struct {
	interpreter.BaseConfig
}

// Config holds configuration for all interpreters. The zero value enables all
// registered loaders; currently only Go and Python are registered.
type Config struct {
	Python python.Config
	Go     interpreterCfg
}

// NoInterpreters returns a Config with every interpreter disabled.
func NoInterpreters() Config {
	disabled := interpreterCfg{BaseConfig: interpreter.BaseConfig{Disabled: true}}
	return Config{
		Python: python.Config{BaseConfig: disabled.BaseConfig},
		Go:     disabled,
	}
}
