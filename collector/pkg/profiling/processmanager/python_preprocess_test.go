//go:build linux

package processmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/bits"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter/interpreterconfig"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/process"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager/ebpfapi"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager/execinfomanager"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/support"
	"go.opentelemetry.io/ebpf-profiler/host"
	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.opentelemetry.io/ebpf-profiler/libpf/pfelf"
	"go.opentelemetry.io/ebpf-profiler/lpm"
)

// Record writes at the system boundary. ELF parsing, file IDs, libc extraction,
// Attach/Detach and shared-file refcounts use the real preprocessing path.
type pythonPreprocessMaps struct {
	ranges      map[uint64]ebpfapi.InterpreterOffsets
	procs       map[uint32]support.ProfilePyProcInfo
	rangeWrites int
}

func (m *pythonPreprocessMaps) UpdateInterpreterOffsets(id uint64, r ebpfapi.InterpreterOffsets) error {
	m.ranges[id] = r
	m.rangeWrites++
	return nil
}
func (m *pythonPreprocessMaps) DeleteInterpreterOffsets(id uint64) error {
	delete(m.ranges, id)
	return nil
}
func (m *pythonPreprocessMaps) UpdatePyProcInfo(pid uint32, p support.ProfilePyProcInfo) error {
	m.procs[pid] = p
	return nil
}
func (m *pythonPreprocessMaps) DeletePyProcInfo(pid uint32) error               { delete(m.procs, pid); return nil }
func (*pythonPreprocessMaps) UpdateUnwindInfo(uint16, ebpfapi.UnwindInfo) error { return nil }
func (*pythonPreprocessMaps) UpdateExeIDToStackDeltas(_ uint64, d []ebpfapi.StackDeltaEBPF) (uint16, error) {
	return uint16(max(8, bits.Len(uint(len(d))))), nil
}
func (*pythonPreprocessMaps) DeleteExeIDToStackDeltas(uint64, uint16) error { return nil }
func (*pythonPreprocessMaps) UpdateStackDeltaPages(uint64, []uint16, uint16, uint64) error {
	return nil
}
func (*pythonPreprocessMaps) DeleteStackDeltaPage(uint64, uint64) error { return nil }
func (*pythonPreprocessMaps) UpdatePidPageMappingInfo(uint32, lpm.Prefix, uint64, uint64) error {
	return nil
}
func (*pythonPreprocessMaps) DeletePidPageMappingInfo(_ uint32, p []lpm.Prefix) (uint64, error) {
	return uint64(len(p)), nil
}
func (*pythonPreprocessMaps) DeleteReportedPID(uint32) error { return nil }

func startPythonTarget(t *testing.T) (libpf.PID, uint64) {
	t.Helper()
	binary := os.Getenv("KINDLING_TEST_PYTHON")
	if binary == "" {
		var err error
		binary, err = exec.LookPath("python3")
		if err != nil {
			t.Skip("python3 not installed; set KINDLING_TEST_PYTHON for another version")
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	cmd := exec.CommandContext(ctx, binary, "-u", "-c", `
import json, sys

def leaf():
    print(json.dumps({"version": list(sys.version_info[:2]), "code": id(leaf.__code__)}), flush=True)
    sys.stdin.read(1)

def root():
    leaf()

root()
`)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = stdin.Close(); cancel(); _ = cmd.Wait() })
	var ready struct {
		Version []int
		Code    uint64
	}
	if err := json.NewDecoder(stdout).Decode(&ready); err != nil {
		t.Fatal(err)
	}
	if len(ready.Version) != 2 || ready.Version[0] != 3 || ready.Version[1] < 6 || ready.Version[1] > 14 {
		t.Skipf("unsupported Python version %v", ready.Version)
	}
	return libpf.PID(cmd.Process.Pid), ready.Code
}

func TestPythonPreprocessRealProcesses(t *testing.T) {
	pid, code := startPythonTarget(t)
	pid2, _ := startPythonTarget(t)
	// Opening map_files can require CAP_SYS_ADMIN/CAP_CHECKPOINT_RESTORE even
	// for a child. Keep this integration test separate from ELF/offset tests.
	access := process.New(pid, pid)
	defer access.Close()
	var accessErr error
	_, errAccess := access.IterateMappings(func(m process.RawMapping) bool {
		if !m.IsExecutable() || m.IsAnonymous() || m.IsVDSO() {
			return true
		}
		f, err := access.OpenMappingFile(&m)
		if err != nil {
			accessErr = err
		} else {
			_ = f.Close()
		}
		return false
	})
	if errors.Is(accessErr, syscall.EPERM) || errors.Is(accessErr, syscall.EACCES) {
		t.Skipf("map_files requires additional capabilities: %v", accessErr)
	}
	if accessErr != nil {
		t.Fatal(accessErr)
	}
	if errAccess != nil && !errors.Is(errAccess, process.ErrCallbackStopped) {
		t.Fatal(errAccess)
	}

	m := &pythonPreprocessMaps{ranges: make(map[uint64]ebpfapi.InterpreterOffsets), procs: make(map[uint32]support.ProfilePyProcInfo)}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg := interpreterconfig.NoInterpreters()
	cfg.Python.Disabled = false
	pm, err := New(ctx, Config{InterpretersConfig: cfg, EbpfHandler: ebpfapi.NewAdapter(m), StackDeltaProvider: ebpfapi.NewStackDeltaProvider()})
	if err != nil {
		t.Fatal(err)
	}
	pr := process.New(pid, pid)
	defer pr.Close()
	pr2 := process.New(pid2, pid2)
	defer pr2.Close()
	pm.SynchronizeProcess(pr)
	pm.SynchronizeProcess(pr2)
	if len(m.procs) != 2 || len(m.ranges) != 1 || m.rangeWrites != 1 {
		t.Fatalf("Python preprocessing not populated/deduplicated: procs=%v ranges=%v writes=%d stats=%+v", m.procs, m.ranges, m.rangeWrites, pm.MappingStats())
	}
	info := m.procs[uint32(pid)]
	t.Logf("Python %d.%d: pid=%d code=%#x cookbook=%+v ranges=%+v", info.Version>>8, info.Version&255, pid, code, info, m.ranges)
	if info.TLSOffset == 0 {
		key, err := pr.GetRemoteMemory().ReadUint32(libpf.Address(info.AutoTLSKeyAddr))
		if err != nil || key >= uint32(info.TsdInfo.KeyLimit) {
			t.Fatalf("invalid runtime TSD key: %d err=%v", key, err)
		}
	}
	// Verify the cookbook against stopped target memory. This is not a BPF
	// verifier test: it independently checks offsets and pointer relocations.
	t.Run("cookbook follows real Python frame", func(t *testing.T) {
		target, err := process.NewPtrace(pid)
		if err != nil {
			t.Skipf("ptrace unavailable: %v", err)
		}
		defer target.Close()
		threads, err := target.GetThreads()
		if err != nil {
			t.Fatal(err)
		}
		current, err := currentPythonCode(target, info, threads[0].TPBase)
		if err != nil {
			t.Fatal(err)
		}
		if current != code {
			t.Fatalf("code object %#x, expected Python id(leaf.__code__)=%#x", current, code)
		}
	})
	pm.DropPID(pid)
	if len(m.procs) != 1 || len(m.ranges) != 1 {
		t.Fatal("first detach removed shared file ranges or retained proc")
	}
	if err := pm.eim.CleanupUnused(0); err != nil {
		t.Fatal(err)
	}
	if len(m.ranges) != 1 {
		t.Fatal("shared interpreter ranges removed before last process detached")
	}
	pm.DropPID(pid2)
	if len(m.procs) != 0 {
		t.Fatal("per-PID cookbook leaked")
	}
	if err := pm.eim.CleanupUnused(0); err != nil {
		t.Fatal(err)
	}
	if len(m.ranges) != 0 {
		t.Fatal("file-level interpreter ranges leaked")
	}
}

func currentPythonCode(pr process.Process, p support.ProfilePyProcInfo, tp uint64) (uint64, error) {
	rm := pr.GetRemoteMemory()
	read := func(a uint64) (uint64, error) { return rm.ReadUint64(libpf.Address(a)) }
	var tstate uint64
	var err error
	if p.TLSOffset != 0 {
		tstate, err = read(tp + uint64(int64(p.TLSOffset)))
	} else {
		key, e := rm.ReadUint32(libpf.Address(p.AutoTLSKeyAddr))
		if e != nil {
			return 0, e
		}
		tsd := p.TsdInfo
		if key >= uint32(tsd.KeyLimit) {
			return 0, fmt.Errorf("TSD key %d out of range", key)
		}
		base := tp + uint64(int64(tsd.Offset))
		if tsd.BlockEntries != 0 {
			base, err = read(base + uint64(key/uint32(tsd.BlockEntries))*8)
			base += uint64(tsd.DataOffset)
			key %= uint32(tsd.BlockEntries)
		} else if tsd.Indirect != 0 {
			base, err = read(base)
		}
		if err != nil {
			return 0, err
		}
		tstate, err = read(base + uint64(key)*uint64(tsd.Multiplier))
	}
	if err != nil {
		return 0, err
	}
	if tstate == 0 {
		return 0, fmt.Errorf("nil Python thread state")
	}
	frame, err := read(tstate + uint64(p.ThreadStateFrame))
	if err != nil {
		return 0, err
	}
	if p.FrameIsCFrame != 0 {
		frame, err = read(frame + uint64(p.CFrameCurrentFrame))
		if err != nil {
			return 0, err
		}
	}
	code, err := read(frame + uint64(p.FrameCode))
	return code &^ 1, err
}

// Exercise the actual file cache/refcount lifecycle without requiring map_files
// or remote memory access. Only the map syscalls are replaced by the recorder.
func TestPythonSharedELFReferences(t *testing.T) {
	pid, _ := startPythonTarget(t)
	pr := process.New(pid, pid)
	defer pr.Close()
	var executable, library string
	_, err := pr.IterateMappings(func(m process.RawMapping) bool {
		if !m.IsExecutable() {
			return true
		}
		base := filepath.Base(m.Path)
		if strings.HasPrefix(base, "libpython") {
			library = strings.Clone(m.Path)
		}
		if strings.HasPrefix(base, "python") {
			executable = strings.Clone(m.Path)
		}
		return true
	})
	if err != nil {
		t.Fatal(err)
	}
	if library != "" {
		executable = library
	}
	if executable == "" {
		t.Fatal("no Python executable mapping")
	}
	ref := pfelf.NewReference(executable, os.DirFS("/"))
	defer ref.Close()
	m := &pythonPreprocessMaps{ranges: make(map[uint64]ebpfapi.InterpreterOffsets), procs: make(map[uint32]support.ProfilePyProcInfo)}
	cfg := interpreterconfig.NoInterpreters()
	cfg.Python.Disabled = false
	manager, err := execinfomanager.NewExecutableInfoManager(ebpfapi.NewStackDeltaProvider(), ebpfapi.NewAdapter(m), cfg)
	if err != nil {
		t.Fatal(err)
	}
	const fileID host.FileID = 123
	first, err := manager.AddOrIncRef(fileID, ref)
	if err != nil {
		t.Fatal(err)
	}
	second, err := manager.AddOrIncRef(fileID, ref)
	if err != nil {
		t.Fatal(err)
	}
	if first.Data == nil || first.Data != second.Data || m.rangeWrites != 1 {
		t.Fatal("shared Python ELF was reloaded or not detected")
	}
	if err := manager.DecRef(fileID); err != nil {
		t.Fatal(err)
	}
	if err := manager.CleanupUnused(0); err != nil {
		t.Fatal(err)
	}
	if len(m.ranges) != 1 {
		t.Fatal("interpreter ranges removed while a file reference remains")
	}
	if err := manager.DecRef(fileID); err != nil {
		t.Fatal(err)
	}
	if err := manager.CleanupUnused(0); err != nil {
		t.Fatal(err)
	}
	if len(m.ranges) != 0 {
		t.Fatal("interpreter range leaked after last file reference")
	}
}
