package cgoapi

/*
#cgo LDFLAGS: -L${SRCDIR} -L${SRCDIR}/../processmanager/ebpf -lkindling -lstdc++ -ldl
#cgo CFLAGS: -I${SRCDIR}
#include <stdint.h>
#include "profiling_cgo.h"
*/
import "C"
import "fmt"

const pidEventBatch = 64

// DrainPidEvents takes pids out of profile_pid_events and deletes them.
// An empty map returns nil, nil. A missing map returns an error.
func DrainPidEvents() ([]uint32, error) {
	buf := make([]C.uint32_t, pidEventBatch)
	n := int(C.drainProfilePidEvents(&buf[0], C.int(len(buf))))
	if n < 0 {
		return nil, fmt.Errorf("drainProfilePidEvents: %d", n)
	}
	if n == 0 {
		return nil, nil
	}
	out := make([]uint32, n)
	for i := 0; i < n; i++ {
		out[i] = uint32(buf[i])
	}
	return out, nil
}

// DeleteReportedPID removes one rate-limit entry. Missing key is success.
func DeleteReportedPID(pid uint32) error {
	if pid == 0 {
		return nil
	}
	rc := int(C.deleteProfileReportedPid(C.uint32_t(pid)))
	if rc < 0 {
		return fmt.Errorf("deleteProfileReportedPid: %d", rc)
	}
	return nil
}
