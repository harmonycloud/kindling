package symbolizer

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/traceagg"
	"go.opentelemetry.io/ebpf-profiler/host"
	"go.opentelemetry.io/ebpf-profiler/libpf"
)

const (
	fileLibc  = 0x11
	fileGoApp = 0x22
	kernelPC  = 0xffffffff81000000
)

// fakeFiles names Go frames: addresses of fileGoApp listed in names.
type fakeFiles map[uint64]string

func (f fakeFiles) SymbolizeAddr(fileID host.FileID, va libpf.Address) (string, bool) {
	if fileID != fileGoApp {
		return "", false
	}
	n, ok := f[uint64(va)]
	return n, ok
}

// fakeProcs maps every frame to procPC = fileID<<32 | va, except unmapped ones.
type fakeProcs map[uint64]bool

func (f fakeProcs) ResolveProcPC(_ libpf.PID, fileID host.FileID, va libpf.Address) (uint64, bool) {
	if f[uint64(va)] {
		return 0, false
	}
	return uint64(fileID)<<32 | uint64(va), true
}

type fakeSymtab struct {
	kernel  map[uint64]string
	process map[uint64]string
	// short drops names from this index on, like resolveProcess' break.
	short    int
	calls    int
	lastAddr []uint64
	err      error
}

func (f *fakeSymtab) KernelNames(addrs []uint64) ([]string, error) {
	out := make([]string, len(addrs))
	for i, a := range addrs {
		out[i] = f.kernel[a]
	}
	return out, nil
}

func (f *fakeSymtab) ProcessNames(_ uint32, pcs []uint64) ([]string, error) {
	f.calls++
	f.lastAddr = append([]uint64(nil), pcs...)
	if f.err != nil {
		return nil, f.err
	}
	out := make([]string, 0, len(pcs))
	for i, pc := range pcs {
		if f.short > 0 && i >= f.short {
			break
		}
		out = append(out, f.process[pc])
	}
	return out, nil
}

func pc(fileID, va uint64) uint64 { return fileID<<32 | va }

// callbackStack is the cgo callback stack from the design doc, leaf first:
// kernel, libc, cgo C, Go, Go, libc, cgo C, Go, Go, Go.
func callbackStack() ([]traceagg.Frame, fakeFiles, *fakeSymtab) {
	frames := []traceagg.Frame{
		traceagg.NewKernelFrame(kernelPC),
		traceagg.NewNativeFrame(fileLibc, 0x100, 0),
		traceagg.NewNativeFrame(fileGoApp, 0x900, 0),
		traceagg.NewNativeFrame(fileGoApp, 0x200, 0),
		traceagg.NewNativeFrame(fileGoApp, 0x210, 0),
		traceagg.NewNativeFrame(fileLibc, 0x300, 0),
		traceagg.NewNativeFrame(fileGoApp, 0x910, 0),
		traceagg.NewNativeFrame(fileGoApp, 0x220, 0),
		traceagg.NewNativeFrame(fileGoApp, 0x230, 0),
		traceagg.NewNativeFrame(fileGoApp, 0x240, 0),
	}
	files := fakeFiles{
		0x200: "main.goCompare",
		0x210: "runtime.cgocallback",
		0x220: "runtime.asmcgocall",
		0x230: "main.sortViaC",
		0x240: "main.main",
	}
	symtab := &fakeSymtab{
		kernel: map[uint64]string{kernelPC: "do_syscall_64"},
		process: map[uint64]string{
			pc(fileLibc, 0x100):  "__libc_write",
			pc(fileGoApp, 0x900): "my_c_log",
			pc(fileLibc, 0x300):  "qsort",
			pc(fileGoApp, 0x910): "_cgo_Cfunc_qsort",
		},
	}
	return frames, files, symtab
}

func TestInterleavedCgoStackKeepsOrder(t *testing.T) {
	frames, files, symtab := callbackStack()
	s := New(files, fakeProcs{}, symtab)

	got, err := s.Symbolize(1, &traceagg.Entry{Frames: frames, Comm: "go_app"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"do_syscall_64", "__libc_write", "my_c_log",
		"main.goCompare", "runtime.cgocallback",
		"qsort", "_cgo_Cfunc_qsort",
		"runtime.asmcgocall", "main.sortViaC", "main.main",
		"go_app",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("stack\n got %q\nwant %q", got, want)
	}
	if symtab.calls != 1 {
		t.Fatalf("symtab calls = %d, want one batch", symtab.calls)
	}
	wantBatch := []uint64{pc(fileLibc, 0x100), pc(fileGoApp, 0x900), pc(fileLibc, 0x300), pc(fileGoApp, 0x910)}
	if !reflect.DeepEqual(symtab.lastAddr, wantBatch) {
		t.Fatalf("symtab batch = %x, want %x", symtab.lastAddr, wantBatch)
	}
	if st := s.Stats(); st.File != 5 || st.Symtab != 4 || st.SymtabShort != 0 {
		t.Fatalf("stats = %+v", st)
	}
}

func TestShortSymtabFillsUnknownAndKeepsRoot(t *testing.T) {
	frames, files, symtab := callbackStack()
	symtab.short = 2 // qsort and _cgo_Cfunc_qsort come back missing
	s := New(files, fakeProcs{}, symtab)

	got, err := s.Symbolize(1, &traceagg.Entry{Frames: frames, Comm: "go_app"})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"do_syscall_64", "__libc_write", "my_c_log",
		"main.goCompare", "runtime.cgocallback",
		UnknownName, UnknownName,
		"runtime.asmcgocall", "main.sortViaC", "main.main",
		"go_app",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("stack\n got %q\nwant %q", got, want)
	}
	if s.Stats().SymtabShort != 1 {
		t.Fatalf("SymtabShort = %d", s.Stats().SymtabShort)
	}
}

func TestUnmappedFrameIsDropped(t *testing.T) {
	frames, files, symtab := callbackStack()
	s := New(files, fakeProcs{0x300: true}, symtab)

	got, err := s.Symbolize(1, &traceagg.Entry{Frames: frames, Comm: "go_app"})
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range got {
		if n == "qsort" {
			t.Fatalf("unmapped frame kept: %q", got)
		}
	}
	if len(got) != len(frames) || got[5] != "_cgo_Cfunc_qsort" {
		t.Fatalf("stack = %q", got)
	}
	if s.Stats().NoMapping != 1 {
		t.Fatalf("NoMapping = %d", s.Stats().NoMapping)
	}
}

func TestAllGoNeedsNoSymtab(t *testing.T) {
	frames := []traceagg.Frame{
		traceagg.NewNativeFrame(fileGoApp, 0x230, 0),
		traceagg.NewNativeFrame(fileGoApp, 0x240, 0),
	}
	symtab := &fakeSymtab{}
	s := New(fakeFiles{0x230: "main.sortViaC", 0x240: "main.main"}, fakeProcs{}, symtab)

	got, err := s.Symbolize(1, &traceagg.Entry{Frames: frames, Comm: "go_app"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"main.sortViaC", "main.main", "go_app"}) {
		t.Fatalf("stack = %q", got)
	}
	if symtab.calls != 0 {
		t.Fatalf("symtab called %d times", symtab.calls)
	}
}

func TestSymtabErrorDropsStack(t *testing.T) {
	frames, files, symtab := callbackStack()
	symtab.err = errors.New("boom")
	s := New(files, fakeProcs{}, symtab)
	if _, err := s.Symbolize(1, &traceagg.Entry{Frames: frames, Comm: "go_app"}); err == nil {
		t.Fatal("want error")
	}
}

func TestNothingResolvable(t *testing.T) {
	frames := []traceagg.Frame{traceagg.NewNativeFrame(fileLibc, 0x100, 0)}
	s := New(fakeFiles{}, fakeProcs{0x100: true}, &fakeSymtab{})
	got, err := s.Symbolize(1, &traceagg.Entry{Frames: frames, Comm: "x"})
	if err != nil || got != nil {
		t.Fatalf("got %q, %v", got, err)
	}
}

func TestMixedPythonNamesSurviveProcessExit(t *testing.T) {
	frames := []traceagg.Frame{
		traceagg.NewPythonFrame(0x1000, 1, 42, 4),
		traceagg.NewNativeFrame(fileLibc, 0x100, 0),
		traceagg.NewPythonFrame(0x2000, 1, 43, 4),
	}
	s := New(fakeFiles{}, fakeProcs{}, &fakeSymtab{process: map[uint64]string{pc(fileLibc, 0x100): "c_callback"}})
	entry := &traceagg.Entry{Frames: frames, InterpNames: []string{"inner (app.py)", "", "outer (app.py)"}, Comm: "python", Count: 2}
	got, err := s.Symbolize(1, entry)
	want := []string{"inner (app.py)", "c_callback", "outer (app.py)", "python"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("%q %v", got, err)
	}
	if s.Stats().File != 0 {
		t.Fatal("Python frames counted as Go frames")
	}
	// Process mappings can disappear before flush; Python slots still survive.
	s.procs = fakeProcs{0x100: true}
	got, err = s.Symbolize(1, entry)
	want = []string{"inner (app.py)", "outer (app.py)", "python"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("%q %v", got, err)
	}
}

func TestMissingPythonNameIsUnknown(t *testing.T) {
	s := New(fakeFiles{}, fakeProcs{}, &fakeSymtab{})
	got, err := s.Symbolize(1, &traceagg.Entry{Frames: []traceagg.Frame{traceagg.NewPythonFrame(1, 1, 1, 4)}, Comm: "python"})
	if err != nil || !reflect.DeepEqual(got, []string{UnknownName, "python"}) {
		t.Fatalf("%q %v", got, err)
	}
}

func TestPythonStackSurvivesNativeLookupFailure(t *testing.T) {
	s := New(fakeFiles{}, fakeProcs{}, &fakeSymtab{err: errors.New("target exited")})
	e := &traceagg.Entry{
		Frames:      []traceagg.Frame{traceagg.NewPythonFrame(1, 1, 1, 4), traceagg.NewNativeFrame(fileLibc, 0x100, 0), traceagg.NewPythonFrame(2, 1, 1, 4)},
		InterpNames: []string{"leaf (app.py)", "", "main (app.py)"}, Comm: "python",
	}
	got, err := s.Symbolize(1, e)
	want := []string{"leaf (app.py)", UnknownName, "main (app.py)", "python"}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("%v %v", got, err)
	}
}
