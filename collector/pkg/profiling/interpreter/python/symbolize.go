// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package python

import (
	"encoding/binary"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter"
	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.opentelemetry.io/ebpf-profiler/successfailurecounter"
	"go.opentelemetry.io/ebpf-profiler/util"
)

// reusing them. We keep that policy, omit line tables for function-level output,
// and decode PEP 393 strings instead of treating Unicode data as a C string.
type pythonCodeObject struct {
	name     string
	checksum uint32
}

var _ interpreter.InterpSymbolizer = (*pythonInstance)(nil)

type SymbolizationStats struct {
	Success, Failed, CacheHits uint64
}

func (p *pythonInstance) SymbolizationStats() SymbolizationStats {
	return SymbolizationStats{p.symbolOK.Load(), p.symbolFail.Load(), p.cacheHits.Load()}
}

// SymbolizeFrame returns an owned name, independent of the target process and
// the LRU lifetime. A zero BPF checksum always forces a fresh remote read.
func (p *pythonInstance) SymbolizeFrame(codeObj, checksum uint64) (string, bool) {
	count := successfailurecounter.New(&p.symbolOK, &p.symbolFail)
	defer count.DefaultToFailure()
	p.symbolMu.Lock()
	defer p.symbolMu.Unlock()
	if p.symbolClosed || codeObj == 0 || checksum > 0xffffffff {
		return "", false
	}
	addr := libpf.Address(codeObj)
	if checksum != 0 {
		if cached, ok := p.addrToCodeObject.Get(addr); ok && cached.checksum == uint32(checksum) {
			p.cacheHits.Add(1)
			count.ReportSuccess()
			return cached.name, true
		}
	}
	obj, err := p.getCodeObject(addr, uint32(checksum))
	if err != nil {
		return "", false
	}
	p.addrToCodeObject.Add(addr, obj)
	count.ReportSuccess()
	return obj.name, true
}

func (p *pythonInstance) getCodeObject(addr libpf.Address, checksum uint32) (pythonCodeObject, error) {
	vm := &p.d.vmStructs.PyCodeObject
	if vm.Sizeof == 0 || vm.Sizeof > 4096 || vm.Name == 0 || vm.Filename == 0 {
		return pythonCodeObject{}, fmt.Errorf("invalid code object size %d", vm.Sizeof)
	}
	buf := make([]byte, vm.Sizeof)
	if err := p.rm.Read(addr, buf); err != nil {
		return pythonCodeObject{}, err
	}
	read32 := func(offset uint) (uint32, error) {
		if len(buf) < 4 || offset > uint(len(buf)-4) {
			return 0, fmt.Errorf("code offset %d out of range", offset)
		}
		return binary.LittleEndian.Uint32(buf[offset:]), nil
	}
	var fields [4]uint32
	for i, offset := range []uint{vm.ArgCount, vm.KwOnlyArgCount, vm.Flags, vm.FirstLineno} {
		v, err := read32(offset)
		if err != nil {
			return pythonCodeObject{}, err
		}
		fields[i] = v
	}
	actual := (fields[0] << 25) + (fields[1] << 18) + (fields[2] << 10) + fields[3]
	if checksum != 0 && checksum != actual {
		return pythonCodeObject{}, fmt.Errorf("stale code object checksum")
	}
	readString := func(offset uint) (string, error) {
		if len(buf) < 8 || offset > uint(len(buf)-8) {
			return "", fmt.Errorf("code string offset %d out of range", offset)
		}
		ptr := libpf.Address(binary.LittleEndian.Uint64(buf[offset:])) - p.rm.Bias
		return p.readUnicode(ptr)
	}
	var name string
	if vm.QualName != 0 {
		name, _ = readString(vm.QualName)
	}
	if name == "" {
		var err error
		name, err = readString(vm.Name)
		if err != nil {
			return pythonCodeObject{}, err
		}
	}
	filename, err := readString(vm.Filename)
	if err != nil {
		return pythonCodeObject{}, err
	}
	if !util.IsValidString(name) || !util.IsValidString(filename) {
		return pythonCodeObject{}, fmt.Errorf("invalid Python name or filename")
	}
	if frozen, err := frozenNameToFileName(filename); err == nil {
		filename = frozen
	}
	return pythonCodeObject{name: name + " (" + filename + ")", checksum: actual}, nil
}

// readUnicode handles compact ASCII/Latin-1/UCS-2/UCS-4 and noncompact strings.
// Python 3.12 removed wstr from the Unicode header, reducing its size by 8 bytes
// (and compact non-ASCII's header by 16). Strings are copied with a fixed cap.
func (p *pythonInstance) readUnicode(addr libpf.Address) (string, error) {
	const maxCodepoints = 4096
	if addr == 0 {
		return "", fmt.Errorf("nil Python string")
	}
	headerSize := p.d.vmStructs.PyASCIIObject.Data
	if headerSize != 40 && headerSize != 48 {
		return "", fmt.Errorf("unsupported Python string header %d", headerSize)
	}
	buf := make([]byte, headerSize)
	if err := p.rm.Read(addr, buf); err != nil {
		return "", err
	}
	length := binary.LittleEndian.Uint64(buf[16:24])
	state := binary.LittleEndian.Uint32(buf[32:36])
	kind, compact, ascii := (state>>2)&7, state&(1<<5) != 0, state&(1<<6) != 0
	if length == 0 || length > maxCodepoints || (kind != 1 && kind != 2 && kind != 4) {
		return "", fmt.Errorf("invalid Python string length/kind")
	}
	if ascii && kind != 1 {
		return "", fmt.Errorf("invalid ASCII kind")
	}
	compactSize := headerSize + 24
	if p.d.version >= pythonVer(3, 12) {
		compactSize = headerSize + 16
	}
	data := addr + libpf.Address(headerSize)
	if compact && !ascii {
		data = addr + libpf.Address(compactSize)
	}
	if !compact {
		var err error
		data, err = p.rm.ReadPtr(addr + libpf.Address(compactSize))
		if err != nil {
			return "", err
		}
	}
	if data == 0 {
		return "", fmt.Errorf("nil Unicode data")
	}
	bytes := make([]byte, int(length)*int(kind))
	if err := p.rm.Read(data, bytes); err != nil {
		return "", err
	}
	var out strings.Builder
	for i := 0; i < len(bytes); i += int(kind) {
		var r rune
		switch kind {
		case 1:
			r = rune(bytes[i])
		case 2:
			r = rune(binary.LittleEndian.Uint16(bytes[i:]))
		case 4:
			r = rune(binary.LittleEndian.Uint32(bytes[i:]))
		}
		if !utf8.ValidRune(r) || (ascii && r > 127) {
			return "", fmt.Errorf("invalid Python Unicode code point")
		}
		out.WriteRune(r)
	}
	return out.String(), nil
}

func frozenNameToFileName(source string) (string, error) {
	if !strings.HasPrefix(source, "<frozen ") {
		return source, nil
	}
	if !strings.HasSuffix(source, ">") {
		return "", fmt.Errorf("missing frozen filename terminator")
	}
	start := strings.LastIndexByte(source, '.') + 1
	if start == 0 {
		start = 8
	}
	name := source[start : len(source)-1]
	if name == "" {
		return "", fmt.Errorf("empty frozen filename")
	}
	return name + ".py", nil
}
