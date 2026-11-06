package python

import "fmt"

// validateOffsets rejects unsupported layouts before narrowing offsets into
// BPF's compact cookbook. Frame/code reads use fixed 128/192-byte scratch areas.
func (d *pythonData) validateOffsets() error {
	vm := &d.vmStructs
	if d.staticTLSOffset < -32768 || d.staticTLSOffset > 32767 {
		return fmt.Errorf("Python TLS offset %d exceeds int16", d.staticTLSOffset)
	}
	if vm.PyCodeObject.Sizeof == 0 || vm.PyCodeObject.Sizeof > 4096 {
		return fmt.Errorf("unsupported PyCodeObject size %d", vm.PyCodeObject.Sizeof)
	}
	if vm.PyThreadState.Frame == 0 || vm.PyFrameObject.Back == 0 || vm.PyFrameObject.LastI == 0 ||
		vm.PyCodeObject.ArgCount == 0 || vm.PyCodeObject.KwOnlyArgCount == 0 ||
		vm.PyCodeObject.Flags == 0 || vm.PyCodeObject.FirstLineno == 0 {
		return fmt.Errorf("missing required Python frame/code offsets")
	}
	checks := []struct {
		name   string
		offset uint
		max    uint
	}{
		{"thread_state.frame", vm.PyThreadState.Frame, 255},
		{"cframe.current_frame", vm.PyCFrame.CurrentFrame, 255},
		{"frame.back", vm.PyFrameObject.Back, 120},
		{"frame.code", vm.PyFrameObject.Code, 120},
		{"frame.lasti", vm.PyFrameObject.LastI, 120},
		{"frame.entry_member", vm.PyFrameObject.EntryMember, 127},
		{"frame.entry_val", vm.PyFrameObject.EntryVal, 255},
		{"code.argcount", vm.PyCodeObject.ArgCount, 188},
		{"code.kwonlyargcount", vm.PyCodeObject.KwOnlyArgCount, 188},
		{"code.flags", vm.PyCodeObject.Flags, 188},
		{"code.firstlineno", vm.PyCodeObject.FirstLineno, 188},
	}
	for _, c := range checks {
		if c.offset > c.max {
			return fmt.Errorf("unsupported Python offset %s=%d (max %d)", c.name, c.offset, c.max)
		}
	}
	return nil
}
