package flamepid

import "testing"

func TestContainerIDLayouts(t *testing.T) {
	id := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	short := id[:12]
	cases := []struct {
		name string
		line string
		want string
	}{
		{"docker", "0::/docker/" + id, short},
		{"docker systemd", "0::/system.slice/docker-" + id + ".scope", short},
		{"containerd", "0::/k8s.io/" + id, short},
		{"containerd systemd", "0::/kubepods.slice/cri-containerd-" + id + ".scope", short},
		{"containerd colon", "0::/default:cri-containerd:" + id, short},
		{"crio", "0::/crio-" + id, short},
		{"crio systemd", "0::/machine.slice/crio-" + id + ".scope", short},
		{"podman", "0::/machine.slice/libpod-" + id + ".scope", short},
		{"host", "0::/user.slice/user-1000.slice/session.scope", ""},
		{"short id", "0::/docker/0123456789ab", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := containerIDFromCgroupFile(tc.line)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
