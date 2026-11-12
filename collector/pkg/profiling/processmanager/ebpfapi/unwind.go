package ebpfapi

import (
	"fmt"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/support"
	"go.opentelemetry.io/ebpf-profiler/nativeunwind"
	"go.opentelemetry.io/ebpf-profiler/nativeunwind/elfunwindinfo"
	sdtypes "go.opentelemetry.io/ebpf-profiler/nativeunwind/stackdeltatypes"
	upstream "go.opentelemetry.io/ebpf-profiler/support"
)

// NewStackDeltaProvider delegates all ELF rule parsing to the fork. Kindling's
// BPF has no Go stack-switch unwinder, so retain its existing PC/SP rules for
// morestack and STOP at arm64 asmcgocall.
func NewStackDeltaProvider() nativeunwind.StackDeltaProvider {
	return elfunwindinfo.NewStackDeltaProviderWithOptions(elfunwindinfo.ExtractionOptions{
		DisableGoStackSwitch: true,
	})
}

// EncodeUnwindInfo converts the fork's rule representation to Kindling's BPF
// ABI. The fork packs an x86 return-address register in BaseReg's high nibble;
// Kindling uses a flag and AuxBaseReg. That slot cannot also recover FP.
func EncodeUnwindInfo(info sdtypes.UnwindInfo) UnwindInfo {
	out := UnwindInfo(info)
	out.BaseReg = info.BaseReg & 0x0f
	if ra := info.BaseReg >> 4; ra != 0 {
		out.Flags |= support.UnwindFlagRegisterRA
		out.AuxBaseReg = ra
		out.AuxParam = 0
	}
	return out
}

// EncodeUnwindCommand accepts only commands implemented by Kindling's BPF.
// Unsupported commands must fail before any corresponding map entry is published.
func EncodeUnwindCommand(command int32) (uint16, error) {
	var encoded int32
	switch command {
	case upstream.UnwindCommandInvalid:
		encoded = support.UnwindCommandInvalid
	case upstream.UnwindCommandStop:
		encoded = support.UnwindCommandStop
	case upstream.UnwindCommandPLT:
		encoded = support.UnwindCommandPLT
	case upstream.UnwindCommandSignal:
		encoded = support.UnwindCommandSignal
	case upstream.UnwindCommandFramePointer:
		encoded = support.UnwindCommandFramePointer
	default:
		return 0, fmt.Errorf("unsupported native unwind command %#x", command)
	}
	return uint16(encoded) | support.DeltaCommandFlag, nil
}
