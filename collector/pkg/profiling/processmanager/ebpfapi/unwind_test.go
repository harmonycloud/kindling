package ebpfapi

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/support"
	"go.opentelemetry.io/ebpf-profiler/libpf/pfelf"
	sdtypes "go.opentelemetry.io/ebpf-profiler/nativeunwind/stackdeltatypes"
	upstream "go.opentelemetry.io/ebpf-profiler/support"
)

func TestNativeRulesFromELF(t *testing.T) {
	compiler, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc is required to build the native ELF fixtures")
	}
	for _, section := range []string{".eh_frame", ".debug_frame"} {
		t.Run(section, func(t *testing.T) {
			dir := t.TempDir()
			source, binary := filepath.Join(dir, "target.c"), filepath.Join(dir, "target")
			if err := os.WriteFile(source, []byte(`
__attribute__((noinline)) int hot(int n) {
    volatile char data[192];
    for (int i = 0; i < 192; ++i) data[i] = (char)(i + n);
    return data[n & 127];
}
int main(void) { return hot(7); }
`), 0600); err != nil {
				t.Fatal(err)
			}
			args := []string{"-O2", "-rdynamic", "-fomit-frame-pointer", "-o", binary, source}
			if section == ".debug_frame" {
				args = append(args, "-g", "-fno-asynchronous-unwind-tables", "-fno-unwind-tables")
			} else {
				args = append(args, "-fasynchronous-unwind-tables")
			}
			if output, err := exec.Command(compiler, args...).CombinedOutput(); err != nil {
				t.Fatalf("build native fixture: %v\n%s", err, output)
			}
			ref := pfelf.NewReference(binary, os.DirFS("/"))
			defer ref.Close()
			ef, err := ref.GetELF()
			if err != nil {
				t.Fatal(err)
			}
			if ef.Section(section) == nil {
				t.Fatalf("compiler did not produce %s", section)
			}
			symbol, err := ef.LookupSymbol("hot")
			if err != nil {
				t.Fatal(err)
			}
			provider := NewStackDeltaProvider()
			rules, err := provider.GetIntervalDataForFile(ref)
			if err != nil {
				t.Fatal(err)
			}
			block := rules.Find(uint64(symbol.Address))
			if block == nil || len(block.Deltas) < 2 {
				t.Fatalf("no stack-allocation rules for hot from %s: %+v", section, block)
			}
			for _, delta := range block.Deltas {
				if delta.Info.Flags&upstream.UnwindFlagCommand != 0 {
					if _, err := EncodeUnwindCommand(delta.Info.Param); err != nil {
						t.Fatal(err)
					}
				} else if info := EncodeUnwindInfo(delta.Info); info.BaseReg == support.UnwindRegInvalid {
					t.Fatalf("invalid encoded CFA from %s: %+v", section, info)
				}
			}
			if stats := provider.GetAndResetStatistics(); stats.Success != 1 || stats.ExtractionErrors != 0 {
				t.Fatalf("provider statistics: %+v", stats)
			}
		})
	}
}

func TestEncodeUnwindInfo(t *testing.T) {
	cases := []struct {
		name string
		in   sdtypes.UnwindInfo
		want UnwindInfo
	}{
		{
			name: "CFA and saved frame pointer",
			in: sdtypes.UnwindInfo{BaseReg: upstream.UnwindRegSp, Param: 32,
				AuxBaseReg: upstream.UnwindRegCfa, AuxParam: -16, MergeOpcode: 9},
			want: UnwindInfo{BaseReg: support.UnwindRegSp, Param: 32,
				AuxBaseReg: support.UnwindRegCfa, AuxParam: -16, MergeOpcode: 9},
		},
		{
			name: "vfork return address in register",
			in: sdtypes.UnwindInfo{BaseReg: upstream.UnwindRegSp | upstream.UnwindRegX86RDI<<4,
				Param: 8, Flags: upstream.UnwindFlagLeafOnly,
				AuxBaseReg: upstream.UnwindRegCfa, AuxParam: -16},
			want: UnwindInfo{BaseReg: support.UnwindRegSp, Param: 8,
				Flags:      support.UnwindFlagLeafOnly | support.UnwindFlagRegisterRA,
				AuxBaseReg: support.UnwindRegX86RDI},
		},
		{
			name: "return address in frame pointer",
			in:   sdtypes.UnwindInfo{BaseReg: upstream.UnwindRegSp | upstream.UnwindRegFp<<4, Param: 8},
			want: UnwindInfo{BaseReg: support.UnwindRegSp, Param: 8,
				Flags: support.UnwindFlagRegisterRA, AuxBaseReg: support.UnwindRegFp},
		},
		{
			name: "arm64 leaf LR is not a packed x86 return address",
			in:   sdtypes.UnwindInfoLR,
			want: UnwindInfo{BaseReg: support.UnwindRegSp, AuxBaseReg: support.UnwindRegLr},
		},
		{
			name: "dereferenced CFA",
			in: sdtypes.UnwindInfo{Flags: upstream.UnwindFlagDerefCfa,
				BaseReg: upstream.UnwindRegFp, Param: 24, AuxBaseReg: upstream.UnwindRegFp},
			want: UnwindInfo{Flags: support.UnwindFlagDerefCfa,
				BaseReg: support.UnwindRegFp, Param: 24, AuxBaseReg: support.UnwindRegFp},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := EncodeUnwindInfo(c.in); got != c.want {
				t.Fatalf("encoded rule: got %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestEncodeUnwindCommand(t *testing.T) {
	for _, c := range []struct{ from, to int32 }{
		{upstream.UnwindCommandInvalid, support.UnwindCommandInvalid},
		{upstream.UnwindCommandStop, support.UnwindCommandStop},
		{upstream.UnwindCommandPLT, support.UnwindCommandPLT},
		{upstream.UnwindCommandSignal, support.UnwindCommandSignal},
		{upstream.UnwindCommandFramePointer, support.UnwindCommandFramePointer},
	} {
		got, err := EncodeUnwindCommand(c.from)
		if err != nil || got != uint16(c.to)|support.DeltaCommandFlag {
			t.Fatalf("command %#x: encoded=%#x err=%v", c.from, got, err)
		}
	}
	for _, command := range []int32{-1, 5, upstream.UnwindCommandGoAsmcgocall, upstream.UnwindCommandGoMorestack} {
		if _, err := EncodeUnwindCommand(command); err == nil {
			t.Fatalf("unsupported command %#x accepted", command)
		}
	}
}
