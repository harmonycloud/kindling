package framedecode

import "testing"

func TestDecodeNativeAndPython(t *testing.T) {
	// native: header(data=elfVA) + fileID
	// python: header + codeObj + lineno
	nativeHdr := (uint64(MarkerNative) << 60) | (uint64(FlagReturnAddress) << 56) |
		(uint64(2) << 52) | 0x14ecb7
	pythonHdr := (uint64(MarkerPython) << 60) | (uint64(FlagPIDSpecific) << 56) |
		(uint64(3) << 52)
	words := []uint64{
		nativeHdr, 0xabc,
		pythonHdr, 0xdeadbeef, 0x11<<32 | 0x42,
		nativeHdr, 0xdef,
	}
	frames, err := DecodeUserFrames(words, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) != 3 {
		t.Fatalf("len=%d", len(frames))
	}
	if frames[0].Kind() != MarkerNative || frames[0].Variable(0) != 0xabc || frames[0].Data() != 0x14ecb7 {
		t.Fatalf("native0=%+v", frames[0])
	}
	if frames[0].Flags()&FlagReturnAddress == 0 {
		t.Fatal("missing RETURN_ADDRESS")
	}
	if frames[1].Kind() != MarkerPython || frames[1].Variable(0) != 0xdeadbeef || frames[1].Variable(1) != 0x11<<32|0x42 {
		t.Fatalf("python=%+v", frames[1])
	}
	if frames[2].Kind() != MarkerNative || frames[2].Variable(0) != 0xdef {
		t.Fatalf("native2=%+v", frames[2])
	}
}

func TestDecodeUnknownPreservesPayloadAndAlignment(t *testing.T) {
	unknownHdr := (uint64(9) << 60) | (uint64(4) << 52) | 0x123
	nativeHdr := (uint64(MarkerNative) << 60) | (uint64(2) << 52) | 0x456
	words := []uint64{unknownHdr, 1, 2, 3, nativeHdr, 0xabc}

	frames, err := DecodeUserFrames(words, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames[0]) != 4 || frames[0].Kind() != 9 || frames[0].Variable(2) != 3 {
		t.Fatalf("unknown=%v", frames[0])
	}
	if frames[1].Kind() != MarkerNative || frames[1].Variable(0) != 0xabc {
		t.Fatalf("native=%v", frames[1])
	}

	// RawFrame is deliberately a view, not a copied interpretation.
	words[1] = 42
	if frames[0].Variable(0) != 42 {
		t.Fatal("decoded frame does not alias input")
	}
}

func TestDecodeBadLength(t *testing.T) {
	hdr := (uint64(MarkerNative) << 60) | (uint64(3) << 52) // length 3 but native expects 2
	_, err := DecodeUserFrames([]uint64{hdr, 1, 2}, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}
