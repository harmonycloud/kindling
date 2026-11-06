package ebpfapi

import (
	"errors"
	"testing"
	"unsafe"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/support"
	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.opentelemetry.io/ebpf-profiler/util"
)

func TestInterpreterRangeBounds(t *testing.T) {
	got, err := EncodeInterpreterOffsets(support.ProfileProgUnwindPython,
		[]util.Range{{Start: 0x1000, End: 0x2000}, {Start: 0x500, End: 0x600}})
	if err != nil {
		t.Fatal(err)
	}
	if got.LowerOffset1 != 0x1000 || got.UpperOffset1 != 0x1fff || got.UpperOffset2 != 0x5ff || got.ProgramIndex != 2 {
		t.Fatalf("%+v", got)
	}
	for _, ranges := range [][]util.Range{nil, {{Start: 2, End: 2}}, {{Start: 2, End: 1}}, {{End: 1}, {End: 2}, {End: 3}}} {
		if _, err := EncodeInterpreterOffsets(2, ranges); err == nil {
			t.Fatalf("accepted %v", ranges)
		}
	}
	if _, err := EncodeInterpreterOffsets(4, []util.Range{{End: 1}}); err == nil {
		t.Fatal("accepted upstream program slot 4")
	}
}

type failedPythonMap struct{ Handler }

var errMapWrite = errors.New("map write failed")

func (failedPythonMap) UpdatePyProcInfo(uint32, support.ProfilePyProcInfo) error { return errMapWrite }

func TestPythonAdapterReportsErrors(t *testing.T) {
	a := NewAdapter(failedPythonMap{})
	info := support.ProfilePyProcInfo{Version: 0x30a}
	if err := a.UpdateProcData(libpf.Python, 123, unsafe.Pointer(&info)); !errors.Is(err, errMapWrite) {
		t.Fatalf("%v", err)
	}
	if err := a.UpdateProcData(libpf.Python, 123, nil); err == nil {
		t.Fatal("accepted nil proc data")
	}
	if err := a.UpdateProcData(libpf.Ruby, 123, unsafe.Pointer(&info)); err == nil {
		t.Fatal("accepted unsupported interpreter")
	}
}
