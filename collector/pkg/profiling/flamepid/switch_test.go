package flamepid

import (
	"os"
	"path/filepath"
	"testing"
)

func TestApplyCPUContainerJavaAndFilter(t *testing.T) {
	root := t.TempDir()
	writeProc(t, root, 10, "0::/docker/"+hex64("a"), "java\x00", "app")
	writeProc(t, root, 11, "0::/docker/"+hex64("a"), "cat\x00", "nginx")
	writeProc(t, root, 20, "0::/docker/"+hex64("b"), "server\x00", "cat")
	writeProc(t, root, 21, "0::/docker/"+hex64("b"), "server\x00", "app")

	var nativeOn, nativeOff, dropped []uint32
	sw := newSwitch(Hooks{
		EnableNative:  func(pid uint32) { nativeOn = append(nativeOn, pid) },
		DisableNative: func(pid uint32) { nativeOff = append(nativeOff, pid) },

		DropProcess: func(pid uint32) { dropped = append(dropped, pid) },
	}, root)

	sw.ApplyCPU(Targets{ContainerIDs: []string{hex64("a")[:12], hex64("b")}})
	if len(nativeOn) != 1 || nativeOn[0] != 21 {
		t.Fatalf("native on %v", nativeOn)
	}

	sw.ApplyCPU(Targets{})
	if len(nativeOff) != 1 || nativeOff[0] != 21 {
		t.Fatalf("native off %v", nativeOff)
	}
	if len(dropped) != 1 || dropped[0] != 21 {
		t.Fatalf("drop %v", dropped)
	}
}

func TestApplyCPUSkipsWhenTableMissing(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")
	called := false
	sw := newSwitch(Hooks{
		DisableNative: func(uint32) { called = true },
	}, root)
	sw.native[7] = struct{}{}
	sw.ApplyCPU(Targets{ContainerIDs: []string{"0123456789ab"}})
	if called {
		t.Fatal("missing table turned native off")
	}
	if _, on := sw.native[7]; !on {
		t.Fatal("roster changed")
	}
}

func TestApplyCPUPIDList(t *testing.T) {
	root := t.TempDir()
	writeProc(t, root, 5, "0::/user.slice", "java\x00", "java")
	writeProc(t, root, 6, "0::/user.slice", "app\x00", "cat")
	var nativeOn []uint32
	sw := newSwitch(Hooks{
		EnableNative: func(pid uint32) { nativeOn = append(nativeOn, pid) },
	}, root)
	sw.ApplyCPU(Targets{PIDs: []uint32{5, 6}})
	if len(nativeOn) != 1 || nativeOn[0] != 6 {
		t.Fatalf("native %v, cat on the pid list stays", nativeOn)
	}
}

func writeProc(t *testing.T, root string, pid uint32, cgroup, cmdline, comm string) {
	t.Helper()
	dir := procPath(root, pid, "")
	dir = dir[:len(dir)-1]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(dir, "cgroup"), cgroup+"\n")
	mustWrite(t, filepath.Join(dir, "cmdline"), cmdline)
	mustWrite(t, filepath.Join(dir, "comm"), comm+"\n")
	mustWrite(t, filepath.Join(dir, "maps"), "libfoo.so\n")
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func hex64(prefix string) string {
	const hex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	return prefix + hex[:64-len(prefix)]
}

func sameSet(got, want []uint32) bool {
	if len(got) != len(want) {
		return false
	}
	seen := map[uint32]int{}
	for _, p := range got {
		seen[p]++
	}
	for _, p := range want {
		seen[p]--
	}
	for _, n := range seen {
		if n != 0 {
			return false
		}
	}
	return true
}
