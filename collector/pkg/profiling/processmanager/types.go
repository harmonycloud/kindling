// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package processmanager // import "github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager"

import (
	"sync"
	"sync/atomic"

	lru "github.com/elastic/go-freelru"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/process"
	pmebpf "github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager/ebpfapi"
	eim "github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager/execinfomanager"
	"go.opentelemetry.io/ebpf-profiler/libc"
	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.opentelemetry.io/ebpf-profiler/libpf/pfelf"
	"go.opentelemetry.io/ebpf-profiler/times"
	"go.opentelemetry.io/ebpf-profiler/util"
)

// elfInfo contains cached data from an executable needed for processing mappings.
// A negative cache entry may also be recorded with err set to indicate permanent
// error. This avoids inspection of non-ELF or corrupted files again and again.
type elfInfo struct {
	err           error
	lastModified  int64
	mappingFile   libpf.FrameMappingFile
	addressMapper pfelf.AddressMapper
}

// ProcessManager is responsible for managing the events happening throughout the lifespan of a
// process.
type ProcessManager struct {
	// A mutex to synchronize access to internal data within this struct.
	mu sync.RWMutex

	// interpreterTracerEnabled indicates if at last one non-native tracer is loaded.
	interpreterTracerEnabled bool

	// eim stores per executable (file ID) information.
	eim *eim.ExecutableInfoManager

	// interpreters records the interpreter.Instance interface which contains hooks for
	// process exits, and various other situations needing interpreter specific attention.
	// The key of the first map is a process ID, while the key of the second map is
	// the unique on-disk identifier of the interpreter DSO.
	interpreters map[libpf.PID]map[util.OnDiskFileIdentifier]interpreter.Instance

	// pidToProcessInfo keeps track of the executable memory mappings.
	pidToProcessInfo map[libpf.PID]*processInfo

	// exitEvents records the pid exit time and is a list of pending exit events to be handled.
	exitEvents map[libpf.PID]times.KTime

	// ebpf contains the interface to manipulate ebpf maps
	ebpf pmebpf.EbpfHandler

	// mappingStats are statistics for parsing process mappings
	mappingStats struct {
		errProcNotExist    atomic.Uint32
		errProcESRCH       atomic.Uint32
		errProcPerm        atomic.Uint32
		numProcAttempts    atomic.Uint32
		maxProcParseUsec   atomic.Uint32
		totalProcParseUsec atomic.Uint32
		numProcParseErrors atomic.Uint32
	}

	// elfInfoCache provides a cache to quickly retrieve the ELF info and fileID for a particular
	// executable. It caches results based on iNode number and device ID. Locked LRU.
	elfInfoCache *lru.LRU[util.OnDiskFileIdentifier, elfInfo]

	metaEnrichers []process.MetaEnricher
}

// Mapping represents an executable memory mapping of a process.
type Mapping struct {
	// Vaddr represents the starting virtual address of the mapping.
	Vaddr libpf.Address

	// Length is the length of the mapping
	Length uint64

	// Device number of the backing file
	Device uint64

	// Inode number of the backing file
	Inode uint64

	// FrameMapping data for this mapping.
	FrameMapping libpf.FrameMapping
}

// GetOnDiskFileIdentifier returns the OnDiskFileIdentifier for the mapping
func (m *Mapping) GetOnDiskFileIdentifier() util.OnDiskFileIdentifier {
	return util.OnDiskFileIdentifier{
		DeviceID: m.Device,
		InodeNum: m.Inode,
	}
}

// processInfo contains information about the executable mappings
// and Thread Specific Data of a process.
type processInfo struct {
	// process metadata, updated on executable changes
	meta process.Meta
	// executable mappings sorted by FileID and mapping start address
	mappings []Mapping
	// C-library Thread Specific Data information
	libcInfo *libc.LibcInfo
}
