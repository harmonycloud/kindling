// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Native unwind constants adapted from the profiler's shared ABI.
package support

const (
	DeltaCommandFlag    = 0x8000
	MergeOpcodeNegative = 0x80
	StackDeltaPageBits  = 0x10
	StackDeltaPageMask  = 0xffff
)

type TSDInfo struct {
	Offset       int16
	Multiplier   uint8
	Indirect     uint8
	KeyLimit     uint16
	BlockEntries uint8
	DataOffset   uint8
}

const (
	UnwindRegInvalid uint8 = 0x0
	UnwindRegCfa     uint8 = 0x1
	UnwindRegPc      uint8 = 0x2
	UnwindRegSp      uint8 = 0x3
	UnwindRegFp      uint8 = 0x4
	UnwindRegLr      uint8 = 0x5
	UnwindRegX86RAX  uint8 = 0x6
	UnwindRegX86R9   uint8 = 0x9
	UnwindRegX86R11  uint8 = 0xa
	UnwindRegX86R13  uint8 = 0xb
	UnwindRegX86R15  uint8 = 0xc
	UnwindRegX86RDI  uint8 = 0x7
	UnwindRegX86R8   uint8 = 0x8

	UnwindFlagCommand    uint8 = 0x1
	UnwindFlagFrame      uint8 = 0x2
	UnwindFlagLeafOnly   uint8 = 0x4
	UnwindFlagDerefCfa   uint8 = 0x8
	UnwindFlagRegisterRA uint8 = 0x10

	UnwindCommandInvalid      int32 = 0x0
	UnwindCommandStop         int32 = 0x1
	UnwindCommandPLT          int32 = 0x2
	UnwindCommandSignal       int32 = 0x3
	UnwindCommandFramePointer int32 = 0x4

	UnwindDerefMask       int32 = 0x7
	UnwindDerefMultiplier int32 = 0x8
)
