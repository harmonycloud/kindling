// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package interpreter // import "github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter"

import (
	"github.com/Kindling-project/kindling/collector/pkg/profiling/process"
	"go.opentelemetry.io/ebpf-profiler/libc"
	"go.opentelemetry.io/ebpf-profiler/libpf"
)

// InstanceStubs provides empty implementations of Instance hooks that are
// not mandatory for a Instance implementation.
type InstanceStubs struct {
}

func (is *InstanceStubs) UsesAnonymousMappings() bool {
	return false
}

func (is *InstanceStubs) SynchronizeMappings(EbpfHandler,
	process.Process, []process.RawMapping) error {
	return nil
}

func (is *InstanceStubs) UpdateLibcInfo(EbpfHandler, libpf.PID, libc.LibcInfo) error {
	return nil
}

func (is *InstanceStubs) Symbolize(libpf.EbpfFrame, *libpf.Frames, libpf.FrameMapping) error {
	return ErrMismatchInterpreterType
}

func (is *InstanceStubs) ReleaseResources() error {
	return nil
}
