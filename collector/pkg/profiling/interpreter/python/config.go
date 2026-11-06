// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package python // import "github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter/python"

import "github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter"

const BPFMapName = "profile_py_procs"

type Config struct {
	interpreter.BaseConfig `mapstructure:",squash"`
}

var _ interpreter.Config = Config{}
