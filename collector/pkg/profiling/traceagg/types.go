// Package traceagg aggregates native-unwind ProfileTrace samples by whole-stack hash.
package traceagg

// FrameKind identifies the meaning of a Frame payload. Values used by BPF
// frames intentionally match the wire-format marker values.
type FrameKind uint8

const (
	KindKernel FrameKind = 0
	KindPython FrameKind = 1
	KindNative FrameKind = 3
)

// NativeFrame is the typed payload of a native user frame.
type NativeFrame struct {
	FileID uint64
	ElfVA  uint64
}

// PythonFrame is the typed payload of a CPython frame.
type PythonFrame struct {
	CodeObject uint64
	LastI      uint32
	Checksum   uint32
}

// Frame is the compact, owned representation retained by the aggregator.
// Payload words are private so runtime-specific values cannot accidentally be
// interpreted as native FileID/ElfVA fields. Use the constructors and typed
// accessors below when adding another runtime.
type Frame struct {
	payload0 uint64
	payload1 uint64
	kind     FrameKind
	flags    uint8
}

func NewKernelFrame(pc uint64) Frame {
	return Frame{kind: KindKernel, payload0: pc}
}

func NewNativeFrame(fileID, elfVA uint64, flags uint8) Frame {
	return Frame{kind: KindNative, flags: flags, payload0: fileID, payload1: elfVA}
}

func NewPythonFrame(codeObject uint64, lastI, checksum uint32, flags uint8) Frame {
	packed := uint64(lastI)<<32 | uint64(checksum)
	return Frame{kind: KindPython, flags: flags, payload0: codeObject, payload1: packed}
}

func (f Frame) Kind() FrameKind {
	return f.kind
}

func (f Frame) Flags() uint8 {
	return f.flags
}

func (f Frame) KernelPC() (uint64, bool) {
	if f.kind != KindKernel {
		return 0, false
	}
	return f.payload0, true
}

func (f Frame) Native() (NativeFrame, bool) {
	if f.kind != KindNative {
		return NativeFrame{}, false
	}
	return NativeFrame{FileID: f.payload0, ElfVA: f.payload1}, true
}

func (f Frame) Python() (PythonFrame, bool) {
	if f.kind != KindPython {
		return PythonFrame{}, false
	}
	return PythonFrame{
		CodeObject: f.payload0,
		LastI:      uint32(f.payload1>>32) & 0x0fffffff,
		Checksum:   uint32(f.payload1),
	}, true
}

// Key identifies an aggregated stack within a process.
type Key struct {
	PID    uint32
	HashHi uint64
	HashLo uint64
}

// Entry holds the first-seen frame list / comm and accumulated count.
type Entry struct {
	Frames []Frame
	// InterpNames aligns with Frames. Only Python entries allocate this slice;
	// native/Go/kernel names remain empty until the periodic flush.
	InterpNames []string
	Comm        string
	Count       uint64
}
