package processmanager

import (
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter"
	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.opentelemetry.io/ebpf-profiler/util"
)

type namedInterpreter struct {
	interpreter.InstanceStubs
	expected uint64
}

func (p *namedInterpreter) SymbolizeFrame(codeObj, checksum uint64) (string, bool) {
	return "leaf (app.py)", codeObj == p.expected && checksum == 99
}
func (*namedInterpreter) Detach(interpreter.EbpfHandler, libpf.PID) error { return nil }
func TestInterpreterRouteAndDetach(t *testing.T) {
	id := util.OnDiskFileIdentifier{InodeNum: 1}
	instance := interpreter.NewMultiInstance([]interpreter.Instance{&namedInterpreter{expected: 0x1000}})
	pm := &ProcessManager{interpreterTracerEnabled: true, interpreters: map[libpf.PID]map[util.OnDiskFileIdentifier]interpreter.Instance{42: {id: instance}}}
	if name, ok := pm.SymbolizeInterp(42, 0x1000, 99); !ok || name != "leaf (app.py)" {
		t.Fatalf("%q %v", name, ok)
	}
	if _, ok := pm.SymbolizeInterp(43, 0x1000, 99); ok {
		t.Fatal("routed unknown PID")
	}
	if _, ok := pm.SymbolizeInterp(42, 0x1000, 98); ok {
		t.Fatal("accepted bad checksum")
	}
	pm.mu.Lock()
	pm.processRemovedInterpreters(42, libpf.Set[util.OnDiskFileIdentifier]{})
	pm.mu.Unlock()
	if _, ok := pm.SymbolizeInterp(42, 0x1000, 99); ok {
		t.Fatal("routed detached instance")
	}
}
