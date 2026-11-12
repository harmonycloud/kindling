package cgoapi

/*
#cgo LDFLAGS: -L${SRCDIR} -L${SRCDIR}/../processmanager/ebpf -lkindling -lstdc++ -ldl
#cgo CFLAGS: -I${SRCDIR}
#include <stdlib.h>
#include <stdint.h>
#include "profiling_cgo.h"
*/
import "C"
import "fmt"

func EnableNativeProfile(pid uint32) int {
	return int(C.enableNativeProfile(C.uint32_t(pid)))
}

func DisableNativeProfile(pid uint32) int {
	return int(C.disableNativeProfile(C.uint32_t(pid)))
}

func StartCPUSampling() {
	C.startCpuSampeling()
}

func StartCPUSamplingAll() {
	C.startCpuSampelingAll()
}

func StopCPUSampling() {
	C.stopCpuSampeling()
}

func InitCPUProfiling() error {
	if rc := int(C.initCPUProfiling()); rc != 0 {
		return fmt.Errorf("init CPU profiling: %d", rc)
	}
	return nil
}
func ShutdownCPUProfiling() { C.shutdownCPUProfiling() }
