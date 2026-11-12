// Package ebpfapi defines the native and Python map operations used by the process manager.
package ebpfapi

import (
	"github.com/Kindling-project/kindling/collector/pkg/profiling/support"
	"go.opentelemetry.io/ebpf-profiler/lpm"
)

type StackDeltaEBPF struct {
	AddressLow uint16
	UnwindInfo uint16
}

// UnwindInfo is the Go form of BPF ProfileUnwindInfo.
// Layout must match C ProfileMapUnwindInfo / ProfileUnwindInfo.
type UnwindInfo struct {
	Flags       uint8
	BaseReg     uint8
	AuxBaseReg  uint8
	MergeOpcode uint8
	Param       int32
	AuxParam    int32
}

// Handler writes native-unwind and Python preprocessing data to Kindling maps.
type Handler interface {
	UpdateInterpreterOffsets(fileID uint64, offsets InterpreterOffsets) error
	DeleteInterpreterOffsets(fileID uint64) error
	UpdatePyProcInfo(pid uint32, info support.ProfilePyProcInfo) error
	DeletePyProcInfo(pid uint32) error

	// UpdateUnwindInfo writes UnwindInfo at the given unwind_info_array index.
	UpdateUnwindInfo(index uint16, info UnwindInfo) error

	// UpdateExeIDToStackDeltas creates/fills the inner ARRAY and links it under
	// exe_id_to_{mapID}_stack_deltas. Returns the chosen mapID (8..23).
	UpdateExeIDToStackDeltas(fileID uint64, deltas []StackDeltaEBPF) (uint16, error)

	// DeleteExeIDToStackDeltas removes the outer HASH_OF_MAPS entry for fileID.
	DeleteExeIDToStackDeltas(fileID uint64, mapID uint16) error

	// UpdateStackDeltaPages writes stack_delta_page_to_info for consecutive

	UpdateStackDeltaPages(fileID uint64, numDeltasPerPage []uint16,
		mapID uint16, firstPageAddr uint64) error

	// DeleteStackDeltaPage removes one page index entry.
	DeleteStackDeltaPage(fileID uint64, page uint64) error

	// UpdatePidPageMappingInfo writes one LPM entry (pid + VA prefix → fileID + bias).
	UpdatePidPageMappingInfo(pid uint32, prefix lpm.Prefix, fileID, bias uint64) error

	// DeletePidPageMappingInfo removes LPM entries for the given prefixes.
	// Returns how many keys were deleted.
	DeletePidPageMappingInfo(pid uint32, prefixes []lpm.Prefix) (uint64, error)

	// DeleteReportedPID removes one profile_reported_pids entry. A missing key is success.
	DeleteReportedPID(pid uint32) error
}

// InterpreterOffsets mirrors ProfileMapOffsetRange; upper bounds are inclusive.
type InterpreterOffsets struct {
	LowerOffset1 uint64
	UpperOffset1 uint64
	LowerOffset2 uint64
	UpperOffset2 uint64
	ProgramIndex uint16
	Pad          [3]uint16
}
