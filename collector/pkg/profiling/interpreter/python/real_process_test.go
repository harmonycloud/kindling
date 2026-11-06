//go:build linux

package python

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager/ebpfapi"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/process"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/symbolizer"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/traceagg"
	"go.opentelemetry.io/ebpf-profiler/host"
	"go.opentelemetry.io/ebpf-profiler/libc"
	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.opentelemetry.io/ebpf-profiler/libpf/pfelf"
	"go.opentelemetry.io/ebpf-profiler/util"
)

type realELFMaps struct {
	recordingEBPF
	ranges  []util.Range
	program uint16
}

func (m *realELFMaps) UpdateInterpreterOffsets(program uint16, id host.FileID, ranges []util.Range) error {
	m.program = program
	m.fileID = id
	m.ranges = append([]util.Range(nil), ranges...)
	return nil
}

// allows this test to run without map_files capabilities; production mapping
// file access is separately covered by processmanager's integration test.
func TestPythonLoaderRealELF(t *testing.T) {
	binary := os.Getenv("KINDLING_TEST_PYTHON")
	if binary == "" {
		var err error
		binary, err = exec.LookPath("python3")
		if err != nil {
			t.Skip("python3 not installed")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "-u", "-c", `
import ctypes,json,sys

ctypes.pythonapi.PyThreadState_Get.restype = ctypes.c_void_p
exec(compile("""
def 处理_é():
    c = 处理_é.__code__
    checksum = ((c.co_argcount << 25) + (c.co_kwonlyargcount << 18) + (c.co_flags << 10) + c.co_firstlineno) & 0xffffffff
    print(json.dumps({"version":list(sys.version_info[:2]),"code":id(c),"line":c.co_firstlineno,"tstate":ctypes.pythonapi.PyThreadState_Get(),"checksum":checksum,"name":getattr(c,"co_qualname",c.co_name),"filename":c.co_filename}),flush=True)
    sys.stdin.read(1)
""", "模型/训练🙂.py", "exec"))
处理_é()

`)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stdin.Close(); cancel(); _ = cmd.Wait() }()
	var ready struct {
		Version        []int
		Code           uint64
		Line           uint32
		Tstate         uint64
		Checksum       uint64
		Name, Filename string
	}
	if err := json.NewDecoder(bufio.NewReader(out)).Decode(&ready); err != nil {
		t.Fatal(err)
	}
	if len(ready.Version) != 2 || ready.Version[0] != 3 || ready.Version[1] < 6 || ready.Version[1] > 14 {
		t.Skipf("Python version %v", ready.Version)
	}
	pid := libpf.PID(cmd.Process.Pid)
	pr := process.New(pid, pid)
	defer pr.Close()
	var mappings []process.RawMapping
	_, err = pr.IterateMappings(func(m process.RawMapping) bool {
		if m.IsExecutable() && m.IsFileBacked() {
			m.Path = strings.Clone(m.Path)
			mappings = append(mappings, m)
		}
		return true
	})
	if err != nil {
		t.Fatal(err)
	}
	maps := &realELFMaps{}
	var instance interpreter.Instance
	var data *pythonData
	var libcInfo libc.LibcInfo
	for _, m := range mappings {
		if libc.IsPotentialLibcDSO(m.Path) {
			ef, e := pfelf.Open(m.Path)
			if e != nil {
				t.Fatal(e)
			}
			info, e := libc.ExtractLibcInfo(ef)
			_ = ef.Close()
			if e == nil {
				libcInfo.Merge(*info)
			}
		}
		// An interpreter ELF may have multiple executable mappings (e.g. the
		// standalone 3.13 build). Attach once, but keep scanning for libc info.
		if instance != nil || !strings.Contains(filepath.Base(m.Path), "python") {
			continue
		}
		ref := pfelf.NewReference(m.Path, os.DirFS("/"))
		ef, e := ref.GetELF()
		if e != nil {
			ref.Close()
			t.Fatal(e)
		}
		intervals, e := ebpfapi.NewStackDeltaProvider().GetIntervalDataForFile(ref)
		if e != nil {
			ref.Close()
			t.Fatal(e)
		}
		loaded, e := GetLoader(Config{}).Load(maps, interpreter.NewLoaderInfo(123, ref, intervals))
		if e != nil {
			ref.Close()
			t.Fatal(e)
		}
		if loaded != nil {
			data = loaded.(*pythonData)
			mapper := ef.GetAddressMapper()
			elfVA, ok := mapper.FileOffsetToVirtualAddress(m.FileOffset)
			if !ok {
				ref.Close()
				t.Fatal("cannot convert executable mapping offset to ELF VA")
			}
			instance, e = data.Attach(maps, pid, libpf.Address(m.Vaddr-elfVA), pr.GetRemoteMemory())
			if e != nil {
				ref.Close()
				t.Fatal(e)
			}
		}
		ref.Close()
	}
	if instance == nil {
		t.Fatal("did not recognize real Python interpreter")
	}
	if err := instance.UpdateLibcInfo(maps, pid, libcInfo); err != nil {
		t.Fatal(err)
	}
	if maps.writes != 1 || maps.program != 2 || len(maps.ranges) == 0 {
		t.Fatalf("missing cookbook/ranges: %+v", maps)
	}
	info := maps.info
	if info.Version != pythonVer(uint16(ready.Version[0]), uint16(ready.Version[1])) {
		t.Fatalf("version %#x != %v", info.Version, ready.Version)
	}
	t.Logf("Python %v: pid=%d code=%#x cookbook=%+v ranges=%+v", ready.Version, pid, ready.Code, info, maps.ranges)
	// This field comes from PyCode_Type's introspection table; compare the
	// preprocessed offset to Python's own view of the live code object.
	line, err := pr.GetRemoteMemory().ReadUint32(libpf.Address(ready.Code + uint64(info.CodeFirstLineno)))
	if errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.EACCES) {
		t.Skipf("cross-process memory access unavailable: %v", err)
	}
	if err != nil {
		t.Fatal(err)
	}
	if line != ready.Line {
		t.Fatalf("code first line=%d, Python reports %d", line, ready.Line)
	}
	frame, err := pr.GetRemoteMemory().ReadPtr(libpf.Address(ready.Tstate + uint64(info.ThreadStateFrame)))
	if err != nil {
		t.Fatal(err)
	}
	if info.FrameIsCFrame != 0 {
		frame, err = pr.GetRemoteMemory().ReadPtr(frame + libpf.Address(info.CFrameCurrentFrame))
		if err != nil {
			t.Fatal(err)
		}
	}
	code, err := pr.GetRemoteMemory().ReadUint64(frame + libpf.Address(info.FrameCode))
	if err != nil || code&^1 != ready.Code {
		t.Fatalf("current code=%#x, expected %#x: %v", code, ready.Code, err)
	}
	if info.LastIIsCodeunit == 0 {
		lastI, err := pr.GetRemoteMemory().ReadUint32(frame + libpf.Address(info.FrameLastI))
		if err != nil || lastI == 0 || lastI > 1<<20 {
			t.Fatalf("bad frame lasti=%d: %v", lastI, err)
		}
	}
	if info.TLSOffset == 0 {
		key, err := pr.GetRemoteMemory().ReadUint32(libpf.Address(info.AutoTLSKeyAddr))
		if err != nil || key >= uint32(info.TsdInfo.KeyLimit) {
			t.Fatalf("bad relocated TSD key %d: %v", key, err)
		}
	}
	symbols := instance.(interpreter.InterpSymbolizer)
	name, ok := symbols.SymbolizeFrame(ready.Code, ready.Checksum)
	want := ready.Name + " (" + ready.Filename + ")"
	if !ok || name != want {
		t.Fatalf("Python symbol=%q, want %q (ok=%v)", name, want, ok)
	}
	if _, ok := symbols.SymbolizeFrame(ready.Code, ready.Checksum+1); ok {
		t.Fatal("stale checksum accepted")
	}
	if got, ok := symbols.SymbolizeFrame(ready.Code, ready.Checksum); !ok || got != want {
		t.Fatal("cache miss")
	}
	if stats := instance.(*pythonInstance).SymbolizationStats(); stats.CacheHits != 1 {
		t.Fatalf("cache stats: %+v", stats)
	}
	// A new stack reads Python immediately. Later flush has no live-memory
	// dependency, even after the target and its interpreter instance are gone.
	agg := traceagg.New()
	agg.AddWithNames(uint32(pid), []traceagg.Frame{traceagg.NewPythonFrame(ready.Code, 1, uint32(ready.Checksum), 4)}, "python", func() []string { return []string{name} })
	_ = cmd.Process.Kill()
	if err := instance.Detach(maps, pid); err != nil {
		t.Fatal(err)
	}
	if _, ok := symbols.SymbolizeFrame(ready.Code, ready.Checksum); ok {
		t.Fatal("detached symbolizer still resolves")
	}
	for key, entry := range agg.Swap() {
		stack, err := symbolizer.New(nil, nil, nil).Symbolize(key.PID, entry)
		if err != nil || len(stack) != 2 || stack[0] != want {
			t.Fatalf("post-exit stack=%v err=%v", stack, err)
		}
	}
	data.Unload(maps)
	if maps.deletes != 1 || maps.unloads != 1 {
		t.Fatal("Python lifecycle did not remove both map entries")
	}
}
