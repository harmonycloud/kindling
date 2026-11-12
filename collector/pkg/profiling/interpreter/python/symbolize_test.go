package python

import (
	"encoding/binary"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"go.opentelemetry.io/ebpf-profiler/remotememory"
)

type objectMemory struct {
	data  []byte
	reads atomic.Uint64
	fail  atomic.Bool
}

func (m *objectMemory) ReadAt(out []byte, offset int64) (int, error) {
	m.reads.Add(1)
	if m.fail.Load() || offset < 0 || offset >= int64(len(m.data)) {
		return 0, io.EOF
	}
	n := copy(out, m.data[int(offset):])
	if n != len(out) {
		return n, io.EOF
	}
	return n, nil
}

func putUnicode(m *objectMemory, addr int, value string, kind uint32, compact bool, version uint16) {
	header, extra := 48, 24
	if version >= pythonVer(3, 12) {
		header, extra = 40, 16
	}
	runes := []rune(value)
	binary.LittleEndian.PutUint64(m.data[addr+16:], uint64(len(runes)))
	ascii := true
	for _, r := range runes {
		if r > 127 {
			ascii = false
		}
	}
	state := kind<<2 | 1<<7
	if compact {
		state |= 1 << 5
	}
	if ascii {
		state |= 1 << 6
	}
	binary.LittleEndian.PutUint32(m.data[addr+32:], state)
	data := addr + header
	if !ascii {
		data += extra
	}
	if !compact {
		data = addr + 256
		binary.LittleEndian.PutUint64(m.data[addr+header+extra:], uint64(data))
	}
	for i, r := range runes {
		switch kind {
		case 1:
			m.data[data+i] = byte(r)
		case 2:
			binary.LittleEndian.PutUint16(m.data[data+i*2:], uint16(r))
		case 4:
			binary.LittleEndian.PutUint32(m.data[data+i*4:], uint32(r))
		}
	}
}

func symbolFixture(t *testing.T, version uint16) (*pythonInstance, *objectMemory, uint64) {
	t.Helper()
	d := testPythonData(version)
	d.vmStructs.PyCodeObject.Sizeof = 176
	d.vmStructs.PyCodeObject.Name = 64
	d.vmStructs.PyCodeObject.Filename = 56
	d.vmStructs.PyASCIIObject.Data = 48
	if version >= pythonVer(3, 12) {
		d.vmStructs.PyASCIIObject.Data = 40
	}
	m := &objectMemory{data: make([]byte, 1<<16)}
	binary.LittleEndian.PutUint32(m.data[0x1000+16:], 2)
	binary.LittleEndian.PutUint32(m.data[0x1000+24:], 1)
	binary.LittleEndian.PutUint32(m.data[0x1000+36:], 3)
	binary.LittleEndian.PutUint32(m.data[0x1000+40:], 42)
	binary.LittleEndian.PutUint64(m.data[0x1000+64:], 0x2000)
	binary.LittleEndian.PutUint64(m.data[0x1000+56:], 0x3000)
	putUnicode(m, 0x2000, "worker", 1, true, version)
	putUnicode(m, 0x3000, "app.py", 1, true, version)
	inst, err := d.Attach(&recordingEBPF{}, 42, 0, remotememory.RemoteMemory{ReaderAt: m})
	if err != nil {
		t.Fatal(err)
	}
	return inst.(*pythonInstance), m, (2 << 25) + (1 << 18) + (3 << 10) + 42
}

func TestSymbolizeObjectCacheAndReuse(t *testing.T) {
	p, m, checksum := symbolFixture(t, pythonVer(3, 10))
	name, ok := p.SymbolizeFrame(0x1000, checksum)
	if !ok || name != "worker (app.py)" {
		t.Fatalf("%q %v", name, ok)
	}
	reads := m.reads.Load()
	// Cached names remain usable after target memory disappears.
	m.fail.Store(true)
	if name, ok := p.SymbolizeFrame(0x1000, checksum); !ok || name != "worker (app.py)" {
		t.Fatal("cache not used")
	}
	if m.reads.Load() != reads {
		t.Fatal("cache hit read target memory")
	}
	if _, ok := p.SymbolizeFrame(0x1000, 0); ok {
		t.Fatal("zero checksum incorrectly used cache")
	}
	m.fail.Store(false)
	binary.LittleEndian.PutUint32(m.data[0x1000+40:], 43)
	putUnicode(m, 0x2000, "replacement", 1, true, p.d.version)
	if _, ok := p.SymbolizeFrame(0x1000, checksum+2); ok {
		t.Fatal("stale checksum accepted")
	}
	if name, ok := p.SymbolizeFrame(0x1000, checksum+1); !ok || name != "replacement (app.py)" {
		t.Fatalf("address reuse: %q %v", name, ok)
	}
	if st := p.SymbolizationStats(); st.Success != 3 || st.Failed != 2 || st.CacheHits != 1 {
		t.Fatalf("%+v", st)
	}
}

func TestPythonUnicodeLayouts(t *testing.T) {
	cases := []struct {
		name, value string
		kind        uint32
		compact     bool
	}{
		{"ascii", "main", 1, true}, {"latin1", "café", 1, true},
		{"ucs2", "处理数据", 2, true}, {"ucs4", "模型🙂", 4, true},
		{"noncompact", "处理🙂", 4, false},
	}
	for _, version := range []uint16{pythonVer(3, 10), pythonVer(3, 12), pythonVer(3, 14)} {
		for _, tc := range cases {
			t.Run(pythonVersionLabel(version)+"/"+tc.name, func(t *testing.T) {
				p, m, _ := symbolFixture(t, version)
				putUnicode(m, 0x4000, tc.value, tc.kind, tc.compact, version)
				got, err := p.readUnicode(0x4000)
				if err != nil || got != tc.value {
					t.Fatalf("%q %v", got, err)
				}
			})
		}
	}
}
func pythonVersionLabel(v uint16) string { return (&pythonData{version: v}).String() }

func TestSymbolizeQualnameAndUnicodeFilename(t *testing.T) {
	p, m, checksum := symbolFixture(t, pythonVer(3, 12))
	p.d.vmStructs.PyCodeObject.QualName = 72
	binary.LittleEndian.PutUint64(m.data[0x1000+72:], 0x4000)
	putUnicode(m, 0x4000, "Handler.处理", 2, true, p.d.version)
	putUnicode(m, 0x3000, "模型/训练🙂.py", 4, true, p.d.version)
	name, ok := p.SymbolizeFrame(0x1000, checksum)
	if !ok || name != "Handler.处理 (模型/训练🙂.py)" {
		t.Fatalf("%q %v", name, ok)
	}
}

func TestRejectInvalidPythonObjects(t *testing.T) {
	p, m, checksum := symbolFixture(t, pythonVer(3, 10))
	if _, ok := p.SymbolizeFrame(0, checksum); ok {
		t.Fatal("nil object")
	}
	if _, ok := p.SymbolizeFrame(0x1000, 1<<32); ok {
		t.Fatal("oversized checksum")
	}
	binary.LittleEndian.PutUint64(m.data[0x2000+16:], 1<<40)
	if _, ok := p.SymbolizeFrame(0x1000, checksum); ok {
		t.Fatal("oversized Unicode string")
	}
	putUnicode(m, 0x2000, "x", 2, true, p.d.version)
	binary.LittleEndian.PutUint16(m.data[0x2000+72:], 0xd800)
	binary.LittleEndian.PutUint32(m.data[0x2000+32:], 2<<2|1<<5|1<<7)
	if _, ok := p.SymbolizeFrame(0x1000, checksum); ok {
		t.Fatal("surrogate accepted")
	}
	p.d.vmStructs.PyCodeObject.Sizeof = 1 << 30
	if _, ok := p.SymbolizeFrame(0x1000, checksum); ok {
		t.Fatal("oversized code object")
	}
}

func TestConcurrentSymbolizeAndDetach(t *testing.T) {
	p, _, checksum := symbolFixture(t, pythonVer(3, 10))
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				p.SymbolizeFrame(0x1000, checksum)
			}
		}()
	}
	if err := p.Detach(&recordingEBPF{}, 42); err != nil {
		t.Fatal(err)
	}
	wg.Wait()
	if _, ok := p.SymbolizeFrame(0x1000, checksum); ok {
		t.Fatal("detached cache still accessible")
	}
}

func TestFrozenNames(t *testing.T) {
	for input, want := range map[string]string{"app.py": "app.py", "<frozen importlib._bootstrap>": "_bootstrap.py"} {
		got, err := frozenNameToFileName(input)
		if err != nil || got != want {
			t.Fatalf("%q %v", got, err)
		}
	}
	if _, err := frozenNameToFileName("<frozen >"); err == nil {
		t.Fatal("empty frozen name")
	}
	if _, err := frozenNameToFileName("<frozen broken"); err == nil {
		t.Fatal("unterminated frozen name")
	}
	p, m, checksum := symbolFixture(t, pythonVer(3, 10))
	putUnicode(m, 0x3000, "<frozen importlib._bootstrap>", 1, true, p.d.version)
	if name, ok := p.SymbolizeFrame(0x1000, checksum); !ok || !strings.HasSuffix(name, "(_bootstrap.py)") {
		t.Fatalf("%q %v", name, ok)
	}
}

var _ io.ReaderAt = (*objectMemory)(nil)
