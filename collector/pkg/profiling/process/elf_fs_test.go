package process

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"go.opentelemetry.io/ebpf-profiler/libpf"
	"go.opentelemetry.io/ebpf-profiler/libpf/pfelf"
)

func TestProcessELFFileSystem(t *testing.T) {
	pr := New(libpf.PID(os.Getpid()), libpf.PID(os.Getpid()))
	defer pr.Close()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ref := pfelf.NewReference(executable, pr)
	defer ref.Close()
	ef, err := ref.GetELF()
	if err != nil {
		t.Fatal(err)
	}
	if ef.Machine != pfelf.CurrentMachine {
		t.Fatalf("unexpected ELF machine: %s", ef.Machine)
	}
	for _, name := range []string{"../escape", "/absolute", "a/../escape", ""} {
		if _, err := pr.Open(name); !errors.Is(err, fs.ErrInvalid) {
			t.Fatalf("Open(%q): expected invalid fs path, got %v", name, err)
		}
	}
}

func TestCoreELFFileSystemHintsAndOwnership(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ef, err := pfelf.Open(executable)
	if err != nil {
		t.Fatal(err)
	}
	defer ef.Close()
	cd := &CoredumpProcess{File: ef, hasMusl: true, files: make(map[string]*CoredumpFile)}
	cd.files["/target"] = &CoredumpFile{parent: cd, Name: libpf.Intern("/target"), Base: 0x70000000}
	f, err := cd.Open("target")
	if err != nil {
		t.Fatal(err)
	}
	hints := f.(pfelf.LoadHinter)
	if base, musl := hints.ELFLoadHints(); base != 0x70000000 || !musl {
		t.Fatalf("core load hints lost: base=%#x musl=%v", base, musl)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := ef.GetGoBuildID(); err != nil {
		t.Fatalf("closing a borrowed core view closed its owner: %v", err)
	}
	if _, err := cd.Open("missing"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing core file: %v", err)
	}
}
