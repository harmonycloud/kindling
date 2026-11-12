// Package ebpf implements ebpfapi.Handler for Kindling profile_* BPF maps.
// Each update uses batch CGO calls into libkindling's map writer.
package ebpf

import (
	"fmt"
	"math/bits"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/cgoapi"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager/ebpfapi"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/support"
	"go.opentelemetry.io/ebpf-profiler/lpm"
)

const (
	// BitWidthPID is the high part of the LPM prefixLen. The page bits come from prefix.Length.
	BitWidthPID = 32

	ProgUnwindNative uint8 = 1

	ProgUnwindPython uint8 = support.ProfileProgUnwindPython

	stackDeltaBucketSmallest = 8
	stackDeltaBucketLargest  = 23
	stackDeltaPageBits       = 16
)

// Maps is the concrete ebpfapi.Handler.
type Maps struct{}

// NewMaps returns a Handler backed by Kindling profile map writer CGO.
func NewMaps() *Maps {
	return &Maps{}
}

var _ ebpfapi.Handler = (*Maps)(nil)

// UpdateUnwindInfo implements ebpfapi.Handler.
func (m *Maps) UpdateUnwindInfo(index uint16, info ebpfapi.UnwindInfo) error {
	return cgoapi.UpdateUnwindInfos([]uint32{uint32(index)}, []ebpfapi.UnwindInfo{info})
}

// UpdateExeIDToStackDeltas implements ebpfapi.Handler.
func (m *Maps) UpdateExeIDToStackDeltas(fileID uint64, deltas []ebpfapi.StackDeltaEBPF) (uint16, error) {
	if len(deltas) == 0 {
		return 0, fmt.Errorf("UpdateExeIDToStackDeltas: empty deltas")
	}
	return cgoapi.UpdateExeIDToStackDeltas(fileID, deltas)
}

// DeleteExeIDToStackDeltas implements ebpfapi.Handler.
func (m *Maps) DeleteExeIDToStackDeltas(fileID uint64, mapID uint16) error {
	return cgoapi.DeleteExeIDToStackDeltas(fileID, mapID)
}

func (m *Maps) UpdateStackDeltaPages(fileID uint64, numDeltasPerPage []uint16,
	mapID uint16, firstPageAddr uint64,
) error {
	if len(numDeltasPerPage) == 0 {
		return nil
	}
	keys, values := buildStackDeltaPageEntries(fileID, numDeltasPerPage, mapID, firstPageAddr)
	return cgoapi.UpdateStackDeltaPages(keys, values)
}

// DeleteStackDeltaPage implements ebpfapi.Handler.
func (m *Maps) DeleteStackDeltaPage(fileID uint64, page uint64) error {
	_, err := cgoapi.DeleteStackDeltaPages([]cgoapi.StackDeltaPageKey{{FileID: fileID, Page: page}})
	return err
}

// UpdatePidPageMappingInfo implements ebpfapi.Handler.
func (m *Maps) UpdatePidPageMappingInfo(pid uint32, prefix lpm.Prefix, fileID, bias uint64) error {
	packed, err := EncodeBiasAndUnwindProgram(bias, ProgUnwindNative)
	if err != nil {
		return err
	}
	key := cgoapi.PidPage{
		PrefixLen: BitWidthPID + prefix.Length,
		PID:       pid,
		Page:      prefix.Key,
	}
	val := cgoapi.PidPageMappingInfo{
		FileID:               fileID,
		BiasAndUnwindProgram: packed,
	}
	return cgoapi.UpdatePidPageMappings([]cgoapi.PidPage{key}, []cgoapi.PidPageMappingInfo{val})
}

// DeleteReportedPID implements ebpfapi.Handler.
func (m *Maps) DeleteReportedPID(pid uint32) error {
	return cgoapi.DeleteReportedPID(pid)
}

// DeletePidPageMappingInfo implements ebpfapi.Handler.
func (m *Maps) DeletePidPageMappingInfo(pid uint32, prefixes []lpm.Prefix) (uint64, error) {
	if len(prefixes) == 0 {
		return 0, nil
	}
	keys := make([]cgoapi.PidPage, len(prefixes))
	for i, p := range prefixes {
		keys[i] = cgoapi.PidPage{
			PrefixLen: BitWidthPID + p.Length,
			PID:       pid,
			Page:      p.Key,
		}
	}
	n, err := cgoapi.DeletePidPageMappings(keys)
	return uint64(n), err
}

// --- helpers (exported for tests / future batch flush) ---

// EncodeBiasAndUnwindProgram packs bias (low 56) and unwind program (high 8).

func EncodeBiasAndUnwindProgram(bias uint64, unwindProgram uint8) (uint64, error) {
	if (bias >> 56) > 0 {
		return 0, fmt.Errorf("unsupported bias value (too large): 0x%x", bias)
	}
	return bias | (uint64(unwindProgram) << 56), nil
}

// DecodeBiasAndUnwindProgram unpacks bias_and_unwind_program.
func DecodeBiasAndUnwindProgram(v uint64) (bias uint64, unwindProgram uint8) {
	bias = v & 0x00ffffffffffffff
	unwindProgram = uint8(v >> 56)
	return bias, unwindProgram
}

// Chooses mapID such that 2^mapID > numDeltas, clamped to [8, 23].
func MapIDForNumDeltas(numDeltas uint32) (uint16, error) {
	if numDeltas == 0 {
		return 0, fmt.Errorf("empty stack deltas")
	}
	significant := 32 - bits.LeadingZeros32(numDeltas)
	if significant <= stackDeltaBucketSmallest {
		return stackDeltaBucketSmallest, nil
	}
	if significant > stackDeltaBucketLargest {
		return 0, fmt.Errorf("no map available for %d stack deltas", numDeltas)
	}
	return uint16(significant), nil
}

func buildStackDeltaPageEntries(fileID uint64, numDeltasPerPage []uint16,
	mapID uint16, firstPageAddr uint64,
) (keys []cgoapi.StackDeltaPageKey, values []cgoapi.StackDeltaPageInfo) {
	keys = make([]cgoapi.StackDeltaPageKey, len(numDeltasPerPage))
	values = make([]cgoapi.StackDeltaPageInfo, len(numDeltasPerPage))
	var firstDelta uint32
	for pageNumber, numDeltas := range numDeltasPerPage {
		pageAddr := firstPageAddr + uint64(pageNumber)<<stackDeltaPageBits
		keys[pageNumber] = cgoapi.StackDeltaPageKey{FileID: fileID, Page: pageAddr}
		values[pageNumber] = cgoapi.StackDeltaPageInfo{
			FirstDelta: firstDelta,
			NumDeltas:  numDeltas,
			MapID:      mapID,
		}
		firstDelta += uint32(numDeltas)
	}
	return keys, values
}

func (m *Maps) UpdateInterpreterOffsets(fileID uint64, offsets ebpfapi.InterpreterOffsets) error {
	return cgoapi.UpdateInterpreterOffsets(fileID, offsets)
}
func (m *Maps) DeleteInterpreterOffsets(fileID uint64) error {
	return cgoapi.DeleteInterpreterOffsets(fileID)
}
func (m *Maps) UpdatePyProcInfo(pid uint32, info support.ProfilePyProcInfo) error {
	return cgoapi.UpdatePyProcInfo(pid, info)
}
func (m *Maps) DeletePyProcInfo(pid uint32) error {
	return cgoapi.DeletePyProcInfo(pid)
}
