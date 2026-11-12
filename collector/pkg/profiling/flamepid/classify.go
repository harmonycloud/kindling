package flamepid

import (
	"os"
	"strconv"
	"strings"
)

func procPath(root string, pid uint32, name string) string {
	return root + "/" + strconv.FormatUint(uint64(pid), 10) + "/" + name
}

var dropComm = map[string]struct{}{
	"cat": {}, "tail": {}, "head": {}, "sleep": {}, "yes": {}, "tee": {},
	"grep": {}, "awk": {}, "sed": {}, "dd": {}, "sh": {}, "bash": {},
	"zsh": {}, "dash": {}, "curl": {}, "wget": {},
}

func isJavaProcess(procRoot string, pid uint32) bool {
	if fileContains(procPath(procRoot, pid, "maps"), "libjvm.so") {
		return true
	}
	return fileContains(procPath(procRoot, pid, "cmdline"), "java")
}

func shouldDropCmd(procRoot string, pid uint32) bool {
	raw, err := os.ReadFile(procPath(procRoot, pid, "comm"))
	if err != nil {
		return false
	}
	comm := strings.TrimRight(string(raw), "\n")
	_, ok := dropComm[comm]
	return ok
}

func fileContains(path, needle string) bool {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(raw), needle)
}
