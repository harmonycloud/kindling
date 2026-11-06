// Package symbolizer combines retained interpreter names and native address symbols.
//
// Each frame is routed on its own: kernel PCs go to the kernel symbolizer,
// frames whose file has per-file data (Go pclntab) are named on the spot, the
// rest go to the symbol table symbolizer by process address. Kernel and symtab
// frames are resolved in one batch each and written back by slot, so the
// original order survives any interleaving (Go -> C -> Go callbacks under cgo).
package symbolizer

import (
	"fmt"
	"sync/atomic"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/traceagg"
	"go.opentelemetry.io/ebpf-profiler/host"
	"go.opentelemetry.io/ebpf-profiler/libpf"
)

// UnknownName fills a symtab slot that got no name back.
const UnknownName = "[unknown]"

// FileSymbolizer names a frame from per-file data. Implemented by ProcessManager.
type FileSymbolizer interface {
	SymbolizeAddr(fileID host.FileID, elfVA libpf.Address) (string, bool)
}

// ProcPCResolver maps a frame back to a process address. Implemented by ProcessManager.
type ProcPCResolver interface {
	ResolveProcPC(pid libpf.PID, fileID host.FileID, elfVA libpf.Address) (uint64, bool)
}

// SymtabSymbolizer names addresses from symbol tables: kallsyms for kernel
// frames, ELF .symtab/.dynsym and debuginfo files for process frames. Both
// calls return names in input order; ProcessNames may return fewer names than
// addresses. Implemented by cgoapi.SymtabSymbolizer (BCC).
type SymtabSymbolizer interface {
	KernelNames(addrs []uint64) ([]string, error)
	ProcessNames(pid uint32, procPCs []uint64) ([]string, error)
}

// Stats are cumulative frame counters.
type Stats struct {
	// File counts frames named by FileSymbolizer (Go pclntab).
	File uint64
	// Symtab counts frames sent to SymtabSymbolizer.ProcessNames.
	Symtab uint64
	// SymtabShort counts stacks where ProcessNames returned fewer names than asked.
	SymtabShort uint64
	// NoMapping counts frames dropped because no mapping covers them.
	NoMapping uint64
}

type Symbolizer struct {
	files  FileSymbolizer
	procs  ProcPCResolver
	symtab SymtabSymbolizer

	file, symtabFrames, symtabShort, noMapping atomic.Uint64
}

func New(files FileSymbolizer, procs ProcPCResolver, symtab SymtabSymbolizer) *Symbolizer {
	return &Symbolizer{files: files, procs: procs, symtab: symtab}
}

func (s *Symbolizer) Stats() Stats {
	return Stats{
		File:        s.file.Load(),
		Symtab:      s.symtabFrames.Load(),
		SymtabShort: s.symtabShort.Load(),
		NoMapping:   s.noMapping.Load(),
	}
}

// IsKernelPC reports whether addr is a kernel address.
func IsKernelPC(addr uint64) bool {
	return (addr & 0xFF00000000000000) != 0
}

type slotKind uint8

const (
	slotKernel slotKind = iota
	slotFile
	slotInterp
	slotSymtab
)

type slot struct {
	kind slotKind
	name string // set for slotFile / slotInterp
}

// Symbolize returns names leaf first with comm appended as the root.
// It returns nil when no frame could be named.
func (s *Symbolizer) Symbolize(pid uint32, entry *traceagg.Entry) ([]string, error) {
	if entry == nil || len(entry.Frames) == 0 {
		return nil, fmt.Errorf("empty entry")
	}
	frames, comm := entry.Frames, entry.Comm
	slots := make([]slot, 0, len(frames))
	var kernelAddrs, symtabAddrs []uint64
	hasInterpreter := false
	for i, fr := range frames {
		if pc, ok := fr.KernelPC(); ok && IsKernelPC(pc) {
			kernelAddrs = append(kernelAddrs, pc)
			slots = append(slots, slot{kind: slotKernel})
			continue
		}
		// Python names were copied at first aggregation. Never dereference a
		// target code object during the delayed flush (the process may be gone).
		if fr.Kind() == traceagg.KindPython {
			hasInterpreter = true
			slots = append(slots, slot{kind: slotInterp, name: nameAt(entry.InterpNames, i)})
			continue
		}
		native, ok := fr.Native()
		if !ok {
			continue
		}
		// Resolve with the raw elfVA. A return address points at the
		// instruction after the call, which stays in the same function and
		// the same mapping as the call itself, so no -1 adjustment is
		// needed here. The -1 that unwind uses lives in BPF only
		// (profile_get_stack_delta) and never changes the stored address.
		fid, va := host.FileID(native.FileID), libpf.Address(native.ElfVA)
		if name, ok := s.files.SymbolizeAddr(fid, va); ok {
			slots = append(slots, slot{kind: slotFile, name: name})
			continue
		}
		procPC, ok := s.procs.ResolveProcPC(libpf.PID(pid), fid, va)
		if !ok {
			s.noMapping.Add(1)
			continue
		}
		symtabAddrs = append(symtabAddrs, procPC)
		slots = append(slots, slot{kind: slotSymtab})
	}
	if len(slots) == 0 {
		return nil, nil
	}

	var kernelNames, symtabNames []string
	var err error
	if len(kernelAddrs) > 0 {
		if kernelNames, err = s.symtab.KernelNames(kernelAddrs); err != nil {
			if !hasInterpreter {
				return nil, err
			}
			kernelNames = nil
		}
	}
	if len(symtabAddrs) > 0 {
		if symtabNames, err = s.symtab.ProcessNames(pid, symtabAddrs); err != nil {
			// Target exit must not discard Python names already retained at
			// ingestion. Keep unresolved native slots as [unknown].
			if !hasInterpreter {
				return nil, err
			}
			symtabNames = nil
		}
		s.symtabFrames.Add(uint64(len(symtabAddrs)))
		if len(symtabNames) < len(symtabAddrs) {
			s.symtabShort.Add(1)
		}
	}

	stack := make([]string, 0, len(slots)+1)
	ki, si := 0, 0
	for _, sl := range slots {
		switch sl.kind {
		case slotKernel:
			stack = append(stack, nameAt(kernelNames, ki))
			ki++
		case slotSymtab:
			stack = append(stack, nameAt(symtabNames, si))
			si++
		case slotInterp:
			stack = append(stack, sl.name)
		case slotFile:
			stack = append(stack, sl.name)
			s.file.Add(1)
		}
	}
	return append(stack, comm), nil
}

func nameAt(names []string, i int) string {
	if i < len(names) && names[i] != "" {
		return names[i]
	}
	return UnknownName
}
