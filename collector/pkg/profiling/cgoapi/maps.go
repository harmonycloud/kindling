package cgoapi

/*
#cgo CFLAGS: -I${SRCDIR}
#include <stdlib.h>
#include <stdint.h>
#include "profiling_cgo.h"
*/
import "C"
import (
	"fmt"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/support"
	"syscall"
	"unsafe"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager/ebpfapi"
)

func checkBatch(op string, want, got int) error {
	if got < 0 {
		return fmt.Errorf("%s: errno %d", op, -got)
	}
	if got != want {
		return fmt.Errorf("%s: updated %d/%d entries", op, got, want)
	}
	return nil
}

func UpdateUnwindInfos(indices []uint32, infos []ebpfapi.UnwindInfo) error {
	n := len(indices)
	if n != len(infos) {
		return fmt.Errorf("updateUnwindInfos: len mismatch %d vs %d", n, len(infos))
	}
	if n == 0 {
		return nil
	}
	got := int(C.profileUpdateUnwindInfos(
		(*C.uint32_t)(unsafe.Pointer(&indices[0])),
		(*C.ProfileMapUnwindInfo)(unsafe.Pointer(&infos[0])),
		C.uint32_t(n),
	))
	return checkBatch("UpdateUnwindInfo", n, got)
}

type StackDeltaPageKey struct {
	FileID uint64
	Page   uint64
}

type StackDeltaPageInfo struct {
	FirstDelta uint32
	NumDeltas  uint16
	MapID      uint16
}

type PidPage struct {
	PrefixLen uint32
	PID       uint32
	Page      uint64
}

type PidPageMappingInfo struct {
	FileID               uint64
	BiasAndUnwindProgram uint64
}

func UpdateStackDeltaPages(keys []StackDeltaPageKey, values []StackDeltaPageInfo) error {
	n := len(keys)
	if n != len(values) {
		return fmt.Errorf("updateStackDeltaPages: len mismatch")
	}
	if n == 0 {
		return nil
	}
	got := int(C.profileUpdateStackDeltaPages(
		(*C.ProfileMapStackDeltaPageKey)(unsafe.Pointer(&keys[0])),
		(*C.ProfileMapStackDeltaPageInfo)(unsafe.Pointer(&values[0])),
		C.uint32_t(n),
	))
	return checkBatch("UpdateStackDeltaPages", n, got)
}

func DeleteStackDeltaPages(keys []StackDeltaPageKey) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	got := int(C.profileDeleteStackDeltaPages(
		(*C.ProfileMapStackDeltaPageKey)(unsafe.Pointer(&keys[0])),
		C.uint32_t(len(keys)),
	))
	if got < 0 {
		return 0, fmt.Errorf("DeleteStackDeltaPage: errno %d", -got)
	}
	return got, nil
}

func UpdatePidPageMappings(keys []PidPage, values []PidPageMappingInfo) error {
	n := len(keys)
	if n != len(values) {
		return fmt.Errorf("updatePidPageMappings: len mismatch")
	}
	if n == 0 {
		return nil
	}
	got := int(C.profileUpdatePidPageMappings(
		(*C.ProfileMapPidPage)(unsafe.Pointer(&keys[0])),
		(*C.ProfileMapPidPageMappingInfo)(unsafe.Pointer(&values[0])),
		C.uint32_t(n),
	))
	return checkBatch("UpdatePidPageMappingInfo", n, got)
}

func DeletePidPageMappings(keys []PidPage) (int, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	got := int(C.profileDeletePidPageMappings(
		(*C.ProfileMapPidPage)(unsafe.Pointer(&keys[0])),
		C.uint32_t(len(keys)),
	))
	if got < 0 {
		return 0, fmt.Errorf("DeletePidPageMappingInfo: errno %d", -got)
	}
	return got, nil
}

func UpdateExeIDToStackDeltas(fileID uint64, deltas []ebpfapi.StackDeltaEBPF) (uint16, error) {
	got := int(C.profileUpdateExeIDToStackDeltas(
		C.uint64_t(fileID),
		(*C.ProfileMapStackDelta)(unsafe.Pointer(&deltas[0])),
		C.uint32_t(len(deltas)),
	))
	if got < 0 {
		return 0, fmt.Errorf("UpdateExeIDToStackDeltas: errno %d", -got)
	}
	return uint16(got), nil
}

func DeleteExeIDToStackDeltas(fileID uint64, mapID uint16) error {
	got := int(C.profileDeleteExeIDToStackDeltas(C.uint64_t(fileID), C.uint16_t(mapID)))
	if got < 0 {
		return fmt.Errorf("DeleteExeIDToStackDeltas: errno %d", -got)
	}
	return nil
}

// Python preprocessing uses synchronous single-entry writes (0 or -errno),
// unlike the native batch API above. No Go memory is retained by C.
func UpdateInterpreterOffsets(fileID uint64, offsets ebpfapi.InterpreterOffsets) error {
	ret := int(C.profileUpdateInterpreterOffsets(C.uint64_t(fileID),
		(*C.ProfileMapOffsetRange)(unsafe.Pointer(&offsets))))
	return checkMapResult("UpdateInterpreterOffsets", ret)
}

func DeleteInterpreterOffsets(fileID uint64) error {
	return checkMapResult("DeleteInterpreterOffsets", int(C.profileDeleteInterpreterOffsets(C.uint64_t(fileID))))
}

func UpdatePyProcInfo(pid uint32, info support.ProfilePyProcInfo) error {
	ret := int(C.profileUpdatePyProcInfo(C.uint32_t(pid),
		(*C.ProfileMapPyProcInfo)(unsafe.Pointer(&info))))
	return checkMapResult("UpdatePyProcInfo", ret)
}

func DeletePyProcInfo(pid uint32) error {
	return checkMapResult("DeletePyProcInfo", int(C.profileDeletePyProcInfo(C.uint32_t(pid))))
}

func checkMapResult(op string, ret int) error {
	if ret < 0 {
		return fmt.Errorf("%s: %w", op, syscall.Errno(-ret))
	}
	return nil
}

// Fail the build if Go/C map values disagree in size or key field position.
var (
	_ [C.sizeof_ProfileMapOffsetRange - unsafe.Sizeof(ebpfapi.InterpreterOffsets{})]byte
	_ [unsafe.Sizeof(ebpfapi.InterpreterOffsets{}) - C.sizeof_ProfileMapOffsetRange]byte
	_ [C.sizeof_ProfileMapPyProcInfo - unsafe.Sizeof(support.ProfilePyProcInfo{})]byte
	_ [unsafe.Sizeof(support.ProfilePyProcInfo{}) - C.sizeof_ProfileMapPyProcInfo]byte
	_ [unsafe.Offsetof(C.ProfileMapPyProcInfo{}.tsd_info) - unsafe.Offsetof(support.ProfilePyProcInfo{}.TsdInfo)]byte
	_ [unsafe.Offsetof(support.ProfilePyProcInfo{}.TsdInfo) - unsafe.Offsetof(C.ProfileMapPyProcInfo{}.tsd_info)]byte
	_ [unsafe.Offsetof(C.ProfileMapPyProcInfo{}.code_sizeof) - unsafe.Offsetof(support.ProfilePyProcInfo{}.CodeSizeof)]byte
	_ [unsafe.Offsetof(support.ProfilePyProcInfo{}.CodeSizeof) - unsafe.Offsetof(C.ProfileMapPyProcInfo{}.code_sizeof)]byte
)
