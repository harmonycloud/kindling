package execinfomanager

import (
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/processmanager/ebpfapi"
	"github.com/Kindling-project/kindling/collector/pkg/profiling/support"
	sdtypes "go.opentelemetry.io/ebpf-profiler/nativeunwind/stackdeltatypes"
	upstream "go.opentelemetry.io/ebpf-profiler/support"
)

type unwindWrites struct {
	ebpfapi.EbpfHandler
	writes []ebpfapi.UnwindInfo
}

func (m *unwindWrites) UpdateUnwindInfo(_ uint16, info ebpfapi.UnwindInfo) error {
	m.writes = append(m.writes, info)
	return nil
}

func TestRuleIndexUsesEncodedIdentity(t *testing.T) {
	maps := &unwindWrites{}
	s := executableInfoManagerState{
		ebpf: maps, unwindInfoIndex: make(map[ebpfapi.UnwindInfo]uint16),
	}
	first := sdtypes.UnwindInfo{
		BaseReg: upstream.UnwindRegSp | upstream.UnwindRegX86RDI<<4, Param: 8,
		Flags: upstream.UnwindFlagLeafOnly,
	}
	// The fork can include FP recovery metadata that Kindling's register-RA
	// encoding cannot use. Both forms must share one BPF rule/index.
	second := first
	second.AuxBaseReg, second.AuxParam = upstream.UnwindRegCfa, -16
	a, err := s.getUnwindInfoIndex(first)
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.getUnwindInfoIndex(second)
	if err != nil {
		t.Fatal(err)
	}
	if a != b || len(maps.writes) != 1 || len(s.unwindInfoIndex) != 1 {
		t.Fatalf("duplicate encoded rule: indices=%d/%d writes=%d", a, b, len(maps.writes))
	}
	if maps.writes[0].BaseReg != support.UnwindRegSp ||
		maps.writes[0].Flags&support.UnwindFlagRegisterRA == 0 ||
		maps.writes[0].AuxBaseReg != support.UnwindRegX86RDI {
		t.Fatalf("wrong BPF encoding: %+v", maps.writes[0])
	}
	_, err = s.getUnwindInfoIndex(sdtypes.UnwindInfo{
		Flags: upstream.UnwindFlagCommand, Param: upstream.UnwindCommandGoMorestack,
	})
	if err == nil || len(maps.writes) != 1 {
		t.Fatal("unsupported command must fail without a map write")
	}
}
