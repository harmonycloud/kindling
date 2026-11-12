// Package framedecode parses typed, variable-length ProfileTrace user frames.

package framedecode

import "fmt"

const (
	MarkerUnknown = 0x0
	MarkerPython  = 0x1
	MarkerNative  = 0x3

	FlagError         = 1 << 0
	FlagReturnAddress = 1 << 1
	FlagPIDSpecific   = 1 << 2
)

const dataMask = uint64(1<<52 - 1)

// RawFrame is a zero-copy view of one complete user frame in the BPF wire
// format. It is only valid while its backing TraceSample.Frames is valid.
// Runtime-specific code should use Data and Variable instead of assigning
// generic names such as file ID or virtual address to the payload words.
type RawFrame []uint64

func (f RawFrame) Kind() uint8 {
	return uint8(f[0] >> 60)
}

func (f RawFrame) Flags() uint8 {
	return uint8((f[0] >> 56) & 0xf)
}

func (f RawFrame) Length() uint8 {
	return uint8((f[0] >> 52) & 0xf)
}

func (f RawFrame) Data() uint64 {
	return f[0] & dataMask
}

// Variable returns a runtime-specific payload word. DecodeUserFrames has
// already validated the frame boundary, so callers must only request indices
// defined by the selected runtime's ABI.
func (f RawFrame) Variable(index int) uint64 {
	return f[index+1]
}

// DecodeHeader splits one header word.
func DecodeHeader(h uint64) (kind, flags, length uint8, data uint64) {
	f := RawFrame{h}
	return f.Kind(), f.Flags(), f.Length(), f.Data()
}

// WalkUserFrames parses numFrames user frames starting at words[0] and calls
// visit for each one. words must be exactly the user-frame region (after
// kernel PCs). Each RawFrame aliases words. Walking avoids an intermediate
// per-sample frame allocation on the profiling hot path.
func WalkUserFrames(words []uint64, numFrames int, visit func(RawFrame)) error {
	if numFrames < 0 {
		return fmt.Errorf("negative num_frames")
	}
	i := 0
	for n := 0; n < numFrames; n++ {
		if i >= len(words) {
			return fmt.Errorf("frame %d: header past end (i=%d len=%d)", n, i, len(words))
		}
		length := int(RawFrame(words[i:]).Length())
		if length < 1 {
			return fmt.Errorf("frame %d: length %d", n, length)
		}
		if i+length > len(words) {
			return fmt.Errorf("frame %d: length %d overruns buffer", n, length)
		}
		frame := RawFrame(words[i : i+length])
		switch frame.Kind() {
		case MarkerNative:
			if length != 2 {
				return fmt.Errorf("native frame %d: length %d want 2", n, length)
			}
		case MarkerPython:
			if length != 3 {
				return fmt.Errorf("python frame %d: length %d want 3", n, length)
			}
		}
		visit(frame)
		i += length
	}
	if i != len(words) {
		return fmt.Errorf("trailing %d words after %d frames", len(words)-i, numFrames)
	}
	return nil
}

// DecodeUserFrames is a convenience wrapper for callers that need to retain
// the RawFrame views. The profiling hot path uses WalkUserFrames directly.
func DecodeUserFrames(words []uint64, numFrames int) ([]RawFrame, error) {
	if numFrames < 0 {
		return nil, fmt.Errorf("negative num_frames")
	}
	out := make([]RawFrame, 0, numFrames)
	err := WalkUserFrames(words, numFrames, func(frame RawFrame) {
		out = append(out, frame)
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
