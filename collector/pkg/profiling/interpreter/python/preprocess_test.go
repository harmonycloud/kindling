package python

import (
	"errors"
	"testing"
	"unsafe"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/support"
	"go.opentelemetry.io/ebpf-profiler/host"
	"go.opentelemetry.io/ebpf-profiler/libc"
	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.opentelemetry.io/ebpf-profiler/remotememory"
)

type recordingEBPF struct {
	interpreter.EbpfHandler
	info                     support.ProfilePyProcInfo
	writes, deletes, unloads int
	writeErr                 error
	deleteErr                error
	fileID                   host.FileID
}

func (r *recordingEBPF) UpdateProcData(typ libpf.InterpreterType, _ libpf.PID, ptr unsafe.Pointer) error {
	if typ != libpf.Python {
		panic("unexpected interpreter")
	}
	r.writes++
	if r.writeErr != nil {
		return r.writeErr
	}
	r.info = *(*support.ProfilePyProcInfo)(ptr)
	return nil
}

func (r *recordingEBPF) DeleteProcData(_ libpf.InterpreterType, _ libpf.PID) error {
	r.deletes++
	return r.deleteErr
}

func (r *recordingEBPF) DeleteInterpreterOffsets(id host.FileID) error {
	r.unloads++
	r.fileID = id
	return nil
}

func testPythonData(version uint16) *pythonData {
	d := &pythonData{fileID: 123, version: version, autoTLSKey: 0x200, noneStruct: 0x300}
	d.vmStructs.PyCodeObject.Sizeof = 320 // Must survive the uint16 cookbook field.
	d.vmStructs.PyThreadState.Frame = 72
	d.vmStructs.PyFrameObject.Code = 32
	d.vmStructs.PyFrameObject.Back = 48
	d.vmStructs.PyFrameObject.LastI = 56
	d.vmStructs.PyCodeObject.ArgCount = 16
	d.vmStructs.PyCodeObject.KwOnlyArgCount = 24
	d.vmStructs.PyCodeObject.Flags = 36
	d.vmStructs.PyCodeObject.FirstLineno = 40
	return d
}

func TestProcInfoLifecycle(t *testing.T) {
	ebpf := &recordingEBPF{}
	d := testPythonData(pythonVer(3, 11))
	instance, err := d.Attach(ebpf, 42, 0x10000, remotememory.RemoteMemory{})
	if err != nil {
		t.Fatal(err)
	}
	if ebpf.writes != 0 {
		t.Fatal("TSD-based Python must wait for libc")
	}
	tsd := libc.TSDInfo{Offset: 1296, Multiplier: 16, KeyLimit: 1024, BlockEntries: 32, DataOffset: 8}
	info := libc.LibcInfo{TSDInfo: tsd}
	if err = instance.UpdateLibcInfo(ebpf, 42, info); err != nil {
		t.Fatal(err)
	}
	got := ebpf.info
	if got.AutoTLSKeyAddr != 0x10200 || got.NoneStructAddr != 0x10300 || got.CodeSizeof != 320 || got.TsdInfo != support.TSDInfo(tsd) || got.FrameIsCFrame != 1 || got.LastIIsCodeunit != 1 {
		t.Fatalf("unexpected Python cookbook: %+v", got)
	}
	if err = instance.UpdateLibcInfo(ebpf, 42, info); err != nil {
		t.Fatal(err)
	}
	if ebpf.writes != 1 {
		t.Fatal("duplicate libc callback must not rewrite proc data")
	}
	ebpf.deleteErr = errors.New("delete failed")
	if err = instance.Detach(ebpf, 42); err == nil {
		t.Fatal("lost delete failure")
	}
	ebpf.deleteErr = nil
	if err = instance.Detach(ebpf, 42); err != nil {
		t.Fatal(err)
	}
	if err = instance.Detach(ebpf, 42); err != nil {
		t.Fatal(err)
	}
	if ebpf.deletes != 2 {
		t.Fatalf("deletes=%d", ebpf.deletes)
	}
	d.Unload(ebpf)
	if ebpf.unloads != 1 || ebpf.fileID != d.fileID {
		t.Fatal("file-level range was not removed")
	}
}

func TestDirectTLSAttachWithoutLibc(t *testing.T) {
	ebpf := &recordingEBPF{}
	d := testPythonData(pythonVer(3, 13))
	d.staticTLSOffset = -16
	instance, err := d.Attach(ebpf, 42, 0, remotememory.RemoteMemory{})
	if err != nil {
		t.Fatal(err)
	}
	if ebpf.writes != 1 || ebpf.info.TLSOffset != -16 || ebpf.info.FrameIsCFrame != 0 {
		t.Fatalf("direct TLS not inserted at attach: %+v", ebpf.info)
	}
	if err := instance.Detach(ebpf, 42); err != nil {
		t.Fatal(err)
	}
}

func TestProcInfoFailedWriteCanRetry(t *testing.T) {
	ebpf := &recordingEBPF{writeErr: errors.New("map full")}
	d := testPythonData(pythonVer(3, 10))
	instance, err := d.Attach(ebpf, 42, 0, remotememory.RemoteMemory{})
	if err != nil {
		t.Fatal(err)
	}
	info := libc.LibcInfo{TSDInfo: libc.TSDInfo{Multiplier: 16, KeyLimit: 32}}
	if err := instance.UpdateLibcInfo(ebpf, 42, info); err == nil {
		t.Fatal("lost map write failure")
	}
	ebpf.writeErr = nil
	if err := instance.UpdateLibcInfo(ebpf, 42, info); err != nil {
		t.Fatal(err)
	}
	if ebpf.writes != 2 {
		t.Fatalf("writes=%d", ebpf.writes)
	}
	if ebpf.info.FrameIsCFrame != 0 || ebpf.info.LastIIsCodeunit != 0 {
		t.Fatal("3.10 must use PyFrameObject")
	}
}

func TestRejectOutOfRangeOffsets(t *testing.T) {
	d := testPythonData(pythonVer(3, 13))
	d.staticTLSOffset = 1 << 16
	if err := d.validateOffsets(); err == nil {
		t.Fatal("TLS offset truncated")
	}
	d.staticTLSOffset = -16
	d.vmStructs.PyFrameObject.Code = 121
	if err := d.validateOffsets(); err == nil {
		t.Fatal("frame access outside scratch accepted")
	}
	d.vmStructs.PyFrameObject.Code = 32
	d.vmStructs.PyCodeObject.ArgCount = 189
	if err := d.validateOffsets(); err == nil {
		t.Fatal("code access outside scratch accepted")
	}
}
