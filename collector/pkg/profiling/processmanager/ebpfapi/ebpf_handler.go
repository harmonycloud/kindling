package ebpfapi

import (
	"fmt"
	"unsafe"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/support"
	"go.opentelemetry.io/ebpf-profiler/host"
	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.opentelemetry.io/ebpf-profiler/lpm"
	"go.opentelemetry.io/ebpf-profiler/util"
)

// Native and Python preprocessing map operations are forwarded through Handler.
type EbpfHandler interface {
	interpreter.EbpfHandler

	RemoveReportedPID(pid libpf.PID)

	UpdateUnwindInfo(index uint16, info UnwindInfo) error

	UpdateExeIDToStackDeltas(fileID host.FileID, deltas []StackDeltaEBPF) (uint16, error)

	DeleteExeIDToStackDeltas(fileID host.FileID, mapID uint16) error

	UpdateStackDeltaPages(fileID host.FileID, numDeltasPerPage []uint16,
		mapID uint16, firstPageAddr uint64) error

	DeleteStackDeltaPage(fileID host.FileID, page uint64) error

	UpdatePidPageMappingInfo(pid libpf.PID, prefix lpm.Prefix, fileID, bias uint64) error

	DeletePidPageMappingInfo(pid libpf.PID, prefixes []lpm.Prefix) (uint64, error)

	SupportsLPMTrieBatchOperations() bool
}

type Adapter struct {
	H Handler
}

var _ EbpfHandler = (*Adapter)(nil)

func NewAdapter(h Handler) *Adapter {
	return &Adapter{H: h}
}

func (a *Adapter) RemoveReportedPID(pid libpf.PID) {
	if a == nil || a.H == nil {
		return
	}
	_ = a.H.DeleteReportedPID(uint32(pid))
}

func (a *Adapter) UpdateUnwindInfo(index uint16, info UnwindInfo) error {
	return a.H.UpdateUnwindInfo(index, info)
}

func (a *Adapter) UpdateExeIDToStackDeltas(fileID host.FileID, deltas []StackDeltaEBPF) (uint16, error) {
	return a.H.UpdateExeIDToStackDeltas(uint64(fileID), deltas)
}

func (a *Adapter) DeleteExeIDToStackDeltas(fileID host.FileID, mapID uint16) error {
	return a.H.DeleteExeIDToStackDeltas(uint64(fileID), mapID)
}

func (a *Adapter) UpdateStackDeltaPages(fileID host.FileID, numDeltasPerPage []uint16,
	mapID uint16, firstPageAddr uint64,
) error {
	return a.H.UpdateStackDeltaPages(uint64(fileID), numDeltasPerPage, mapID, firstPageAddr)
}

func (a *Adapter) DeleteStackDeltaPage(fileID host.FileID, page uint64) error {
	return a.H.DeleteStackDeltaPage(uint64(fileID), page)
}

func (a *Adapter) UpdatePidPageMappingInfo(pid libpf.PID, prefix lpm.Prefix, fileID, bias uint64) error {
	return a.H.UpdatePidPageMappingInfo(uint32(pid), prefix, fileID, bias)
}

func (a *Adapter) DeletePidPageMappingInfo(pid libpf.PID, prefixes []lpm.Prefix) (uint64, error) {
	return a.H.DeletePidPageMappingInfo(uint32(pid), prefixes)
}

func (a *Adapter) SupportsLPMTrieBatchOperations() bool { return true }

func (a *Adapter) UpdateInterpreterOffsets(program uint16, fileID host.FileID, ranges []util.Range) error {
	offsets, err := EncodeInterpreterOffsets(program, ranges)
	if err != nil {
		return err
	}
	return a.H.UpdateInterpreterOffsets(uint64(fileID), offsets)
}

func (a *Adapter) DeleteInterpreterOffsets(fileID host.FileID) error {
	return a.H.DeleteInterpreterOffsets(uint64(fileID))
}

func (a *Adapter) UpdateProcData(typ libpf.InterpreterType, pid libpf.PID, data unsafe.Pointer) error {
	if typ != libpf.Python {
		return fmt.Errorf("unsupported interpreter proc data: %v", typ)
	}
	if data == nil {
		return fmt.Errorf("nil Python proc data")
	}
	return a.H.UpdatePyProcInfo(uint32(pid), *(*support.ProfilePyProcInfo)(data))
}

func (a *Adapter) DeleteProcData(typ libpf.InterpreterType, pid libpf.PID) error {
	if typ != libpf.Python {
		return fmt.Errorf("unsupported interpreter proc data: %v", typ)
	}
	return a.H.DeletePyProcInfo(uint32(pid))
}

// EncodeInterpreterOffsets translates half-open ELF ranges to BPF inclusive ends.
func EncodeInterpreterOffsets(program uint16, ranges []util.Range) (InterpreterOffsets, error) {
	if program != support.ProfileProgUnwindPython {
		return InterpreterOffsets{}, fmt.Errorf("unsupported interpreter program %d", program)
	}
	if len(ranges) == 0 || len(ranges) > 2 {
		return InterpreterOffsets{}, fmt.Errorf("need 1 or 2 interpreter ranges, got %d", len(ranges))
	}
	for _, r := range ranges {
		if r.Start >= r.End {
			return InterpreterOffsets{}, fmt.Errorf("empty or reversed interpreter range [%x,%x)", r.Start, r.End)
		}
	}
	result := InterpreterOffsets{LowerOffset1: ranges[0].Start, UpperOffset1: ranges[0].End - 1, ProgramIndex: program}
	if len(ranges) == 2 {
		result.LowerOffset2 = ranges[1].Start
		result.UpperOffset2 = ranges[1].End - 1
	}
	return result, nil
}

func (a *Adapter) UpdatePidInterpreterMapping(_ libpf.PID, _ lpm.Prefix, _ uint8, _ host.FileID, _ uint64) error {
	return nil
}

func (a *Adapter) DeletePidInterpreterMapping(_ libpf.PID, _ lpm.Prefix) error { return nil }
