package support

import "unsafe"

const ProfileProgUnwindPython = 2

// ProfilePyProcInfo mirrors agent-libs ProfilePyProcInfo. Keep the explicit
// padding: this value crosses Go, C and BPF without field-by-field conversion.
// Unlike upstream PyProcInfo, CodeSizeof is uint16 to retain tp_basicsize > 255.
type ProfilePyProcInfo struct {
	AutoTLSKeyAddr     uint64
	NoneStructAddr     uint64
	Version            uint16
	TLSOffset          int16
	TsdInfo            TSDInfo
	ThreadStateFrame   uint8
	CFrameCurrentFrame uint8
	FrameBack          uint8
	FrameCode          uint8
	FrameLastI         uint8
	FrameEntryMember   uint8
	FrameEntryVal      uint8
	CodeArgCount       uint8
	CodeKwOnlyArgCount uint8
	CodeFlags          uint8
	CodeFirstLineno    uint8
	LastIIsCodeunit    uint8
	FrameIsCFrame      uint8
	Pad                uint8
	CodeSizeof         uint16
}

// Compile-time ABI checks, complemented by C/BPF checks in the map writer.
var (
	_ [8 - unsafe.Sizeof(TSDInfo{})]byte
	_ [unsafe.Sizeof(TSDInfo{}) - 8]byte
	_ [48 - unsafe.Sizeof(ProfilePyProcInfo{})]byte
	_ [unsafe.Sizeof(ProfilePyProcInfo{}) - 48]byte
	_ [42 - unsafe.Offsetof(ProfilePyProcInfo{}.CodeSizeof)]byte
	_ [unsafe.Offsetof(ProfilePyProcInfo{}.CodeSizeof) - 42]byte
)
