package golang

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.opentelemetry.io/ebpf-profiler/libpf/pfelf"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter"
)

// Exercise the imported pclntab parser through our loader and symbolizer, using
// a real stripped executable rather than ELF symbols supplied by the test.
func TestStrippedGoExecutable(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "main.go")
	binary := filepath.Join(dir, "target")
	if err := os.WriteFile(source, []byte(`package main
import "fmt"
//go:noinline
func hot(n int) int { return n * n }
func main() { fmt.Println(hot(7)) }
`), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build",
		"-ldflags=-s -w", "-o", binary, source)
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build stripped Go target: %v\n%s", err, output)
	}
	ref := pfelf.NewReference(binary, os.DirFS("/"))
	defer ref.Close()
	loaded, err := load(nil, interpreter.NewLoaderInfo(1, ref, nil))
	if err != nil || loaded == nil {
		t.Fatalf("load Go target: %v", err)
	}
	d := loaded.(*goData)
	defer d.Unload(nil)
	sym, err := d.pclntab.LookupSymbol("main.hot")
	if err != nil {
		t.Fatal(err)
	}
	name, ok := d.SymbolizeAddr(libpf.Address(sym.Address))
	if !ok || name != "main.hot" {
		t.Fatalf("symbolize hot: name=%q found=%v", name, ok)
	}
	d.ReleaseResources()
}
