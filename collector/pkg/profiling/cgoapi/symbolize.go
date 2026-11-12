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
	"time"
	"unsafe"
)

func InitSymbolizer() error {
	if C.symbolizerInit() != 0 {
		return fmt.Errorf("kindling_symbolizer_init failed")
	}
	return nil
}

func ShutdownSymbolizer() {
	C.symbolizerShutdown()
}

// ExpireSymbolizerCaches frees process symbol caches not used for longer than
// ttl. Caches of pids still being sampled are touched every flush and stay.
func ExpireSymbolizerCaches(ttl time.Duration) {
	C.maintainBccCaches(C.int64_t(ttl / time.Second))
}

// SymtabSymbolizer names addresses from symbol tables through libkindling's
// BCC symbolizer: kallsyms for kernel addresses, ELF .symtab/.dynsym and
// debuginfo files for process addresses. It satisfies
// symbolizer.SymtabSymbolizer.
type SymtabSymbolizer struct{}

func (SymtabSymbolizer) KernelNames(addrs []uint64) ([]string, error) {
	return SymbolizeKernelNames(addrs)
}

func (SymtabSymbolizer) ProcessNames(pid uint32, procPCs []uint64) ([]string, error) {
	return SymbolizeProcessNames(pid, procPCs)
}

func SymbolizeProcessNames(pid uint32, addrs []uint64) ([]string, error) {
	if len(addrs) == 0 {
		return nil, fmt.Errorf("kindling_symbolize_process_names addrs == 0")
	}
	var cNames **C.char
	var outN C.uint32_t
	rc := C.symbolizeProcessNamesWithFallback(
		C.uint32_t(pid),
		(*C.uint64_t)(unsafe.Pointer(&addrs[0])), C.uint32_t(len(addrs)),
		1, 1, 1,
		&cNames, &outN,
	)
	if rc != 0 {
		return nil, fmt.Errorf("kindling_symbolize_process_names failed")
	}
	defer C.freeNames(cNames, outN)
	return cStringsToSlice(cNames, outN), nil
}

func SymbolizeKernelNames(addrs []uint64) ([]string, error) {
	if len(addrs) == 0 {
		return nil, nil
	}
	var cNames **C.char
	var outN C.uint32_t
	rc := C.symbolizeKernelNames(
		(*C.uint64_t)(unsafe.Pointer(&addrs[0])), C.uint32_t(len(addrs)),
		&cNames, &outN,
	)
	if rc != 0 {
		return nil, fmt.Errorf("kindling_symbolize_kernel_names failed")
	}
	defer C.freeNames(cNames, outN)
	return cStringsToSlice(cNames, outN), nil
}

func cStringsToSlice(cNames **C.char, n C.uint32_t) []string {
	count := int(n)
	if count == 0 || cNames == nil {
		return nil
	}
	arr := (*[1 << 30]*C.char)(unsafe.Pointer(cNames))[:count:count]
	out := make([]string, count)
	for i := 0; i < count; i++ {
		out[i] = C.GoString(arr[i])
	}
	return out
}
