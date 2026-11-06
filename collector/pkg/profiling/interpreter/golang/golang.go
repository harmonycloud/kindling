// Package golang names Go frames from .gopclntab, which survives strip.

package golang // import "github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter/golang"

import (
	"sync/atomic"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter"
	"go.opentelemetry.io/ebpf-profiler/host"
	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.opentelemetry.io/ebpf-profiler/nativeunwind/elfunwindinfo"
	"go.opentelemetry.io/ebpf-profiler/remotememory"
)

var (
	_ interpreter.Data           = &goData{}
	_ interpreter.AddrSymbolizer = &goData{}
	_ interpreter.Instance       = &goInstance{}
)

type goData struct {
	fileID  host.FileID
	pclntab *elfunwindinfo.Gopclntab
	// refs counts Unload plus one per attached Instance.
	refs atomic.Int32
}

type goInstance struct {
	interpreter.InstanceStubs
	d *goData
}

// Loader recognizes Go executables. Go plugins (.so) share the main binary's
// runtime and are left to the native symbolizer.
var Loader = interpreter.NewLoader(load, nil)

func load(_ interpreter.EbpfHandler, info *interpreter.LoaderInfo) (interpreter.Data, error) {
	ef, err := info.GetELF()
	if err != nil {
		return nil, err
	}
	if !ef.IsGolang() || !ef.IsExecutable() {
		return nil, nil
	}
	pclntab, err := elfunwindinfo.NewGopclntab(ef)
	if err != nil {
		return nil, err
	}
	d := &goData{fileID: info.FileID(), pclntab: pclntab}
	d.refs.Store(1)
	return d, nil
}

func (d *goData) String() string {
	return "Go " + d.fileID.StringNoQuotes()
}

func (d *goData) unref() {
	if d.refs.Add(-1) == 0 {
		_ = d.pclntab.Close()
	}
}

func (d *goData) Attach(_ interpreter.EbpfHandler, _ libpf.PID, _ libpf.Address,
	_ remotememory.RemoteMemory) (interpreter.Instance, error) {
	d.refs.Add(1)
	return &goInstance{d: d}, nil
}

func (d *goData) Unload(_ interpreter.EbpfHandler) {
	d.unref()
}

// SymbolizeAddr resolves an ELF-relative program counter using the Go pclntab.
func (d *goData) SymbolizeAddr(addr libpf.Address) (string, bool) {
	_, _, fn := d.pclntab.Symbolize(uintptr(addr))
	if fn == "" {
		return "", false
	}
	// fn points into the pclntab mapping; copy it out before Close can unmap it.
	return libpf.Intern(fn).String(), true
}

func (d *goData) ReleaseResources() {
	_ = d.pclntab.SetDontNeed()
}

func (i *goInstance) Detach(_ interpreter.EbpfHandler, _ libpf.PID) error {
	i.d.unref()
	return nil
}
