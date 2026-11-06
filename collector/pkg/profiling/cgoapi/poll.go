package cgoapi

/*
#cgo CFLAGS: -I${SRCDIR}
#include <stdlib.h>
#include <stdint.h>
#include "profiling_cgo.h"
*/
import "C"
import (
	"bytes"
	"unsafe"
)

const (
	NativeMaxFrames = 128
	FrameWords      = 512
)

// TraceSample is the Go form of C ProfileTraceSample.
type TraceSample struct {
	PID             uint32
	TID             uint32
	Ktime           uint64
	CPUId           uint16
	NumFrames       uint16
	FrameDataLen    uint32
	NumKernelFrames uint16
	Comm            string
	Frames          []uint64
}

// TracePoller owns the raw batch buffer for one polling loop. It is not shared
// between goroutines; returned samples own their frames independently of it.
type TracePoller struct {
	buf []C.ProfileTraceSample
}

func NewTracePoller(capacity int) *TracePoller {
	return &TracePoller{buf: make([]C.ProfileTraceSample, max(0, capacity))}
}

// Poll reads at most the buffer capacity. Returns count, 0 if empty, <0 on error.
func (p *TracePoller) Poll() ([]TraceSample, int) {
	if len(p.buf) == 0 {
		return nil, 0
	}
	n := int(C.pollProfileTraces(&p.buf[0], C.int(len(p.buf))))
	if n <= 0 {
		return nil, n
	}
	out := make([]TraceSample, n)
	for i := 0; i < n; i++ {
		out[i] = traceFromC(&p.buf[i])
	}
	return out, n
}

func traceFromC(s *C.ProfileTraceSample) TraceSample {
	words := int(s.frame_data_len)
	maxWords := FrameWords
	if words > maxWords {
		words = maxWords
	}
	frames := make([]uint64, words)
	for i := 0; i < words; i++ {
		frames[i] = uint64(s.frames[i])
	}
	return TraceSample{
		PID:             uint32(s.pid),
		TID:             uint32(s.tid),
		Ktime:           uint64(s.ktime),
		CPUId:           uint16(s.cpu_id),
		NumFrames:       uint16(s.num_frames),
		FrameDataLen:    uint32(s.frame_data_len),
		NumKernelFrames: uint16(s.num_kernel_frames),
		Comm:            string(bytes.TrimRight(C.GoBytes(unsafe.Pointer(&s.comm[0]), C.PROFILE_COMM_LEN), "\x00")),
		Frames:          frames,
	}
}
