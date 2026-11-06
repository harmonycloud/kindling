package cgoapi

/*
#cgo LDFLAGS: -L${SRCDIR} -L${SRCDIR}/../processmanager/ebpf -lkindling -lstdc++ -ldl
#cgo CFLAGS: -I${SRCDIR}
#include <stdint.h>
#include "profiling_cgo.h"
*/
import "C"
import "fmt"

// profile_metrics slots. Must match profile_metric_id in agent-libs profile_types.h.
const (
	MetricSamples = iota
	MetricPerfOutputFail
	MetricUnwindErrKernelPC
	MetricUnwindErrNoMapping
	MetricUnwindErrZeroPC
	MetricUnwindErrFramesFull
	MetricPythonAttempts
	MetricPythonFrames
	MetricPythonErrNoProcInfo
	MetricPythonErrThreadState
	MetricPythonErrFrameRead
	MetricPythonCodeReadFail
	MetricCount
)

// ReadProfileMetrics returns the CPU-summed profile_metrics counters.
func ReadProfileMetrics() ([MetricCount]uint64, error) {
	var out [MetricCount]uint64
	var buf [MetricCount]C.uint64_t
	n := int(C.readProfileMetrics(&buf[0], C.int(MetricCount)))
	if n != MetricCount {
		return out, fmt.Errorf("readProfileMetrics: %d", n)
	}
	for i := 0; i < n && i < MetricCount; i++ {
		out[i] = uint64(buf[i])
	}
	return out, nil
}
