package flamepid

import "strings"

const containerIDLen = 64
const reportedIDLen = 12

// layout matches libsinsp runc cgroup_layout. First hit wins.
// Order follows the engines that produce the 12-char id: podman, docker, cri.
type layout struct {
	prefix string
	suffix string
}

var cgroupLayouts = []layout{
	{prefix: "/libpod-", suffix: ".scope"},
	{prefix: "/docker-", suffix: ".scope"},
	{prefix: "/", suffix: ""},
	{prefix: "/crio-", suffix: ""},
	{prefix: "/cri-containerd-", suffix: ".scope"},
	{prefix: "/crio-", suffix: ".scope"},
	{prefix: ":cri-containerd:", suffix: ""},
}

// containerIDFromCgroupFile returns the 12-char id, or "" when the process is on the host.
func containerIDFromCgroupFile(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		path, ok := cgroupPath(line)
		if !ok {
			continue
		}
		if id := matchContainerID(path); id != "" {
			return id
		}
	}
	return ""
}

// cgroupPath keeps colons that belong to the path. The file line is hierarchy:controllers:path.
func cgroupPath(line string) (string, bool) {
	parts := strings.SplitN(line, ":", 3)
	if len(parts) < 3 || parts[2] == "" {
		return "", false
	}
	return parts[2], true
}

func matchContainerID(path string) string {
	for _, lay := range cgroupLayouts {
		if id, ok := matchOne(path, lay.prefix, lay.suffix); ok {
			return id
		}
	}
	return ""
}

func matchOne(cgroup, prefix, suffix string) (string, bool) {
	start := strings.LastIndex(cgroup, prefix)
	if start < 0 {
		return "", false
	}
	start += len(prefix)
	end := strings.LastIndex(cgroup, suffix)
	if end < 0 || end < start {
		return "", false
	}
	if end-start != containerIDLen {
		return "", false
	}
	raw := cgroup[start:end]
	if !isHex(raw) {
		return "", false
	}
	return raw[:reportedIDLen], true
}

func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}
