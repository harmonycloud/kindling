package process

import (
	"strings"
	"testing"

	"go.opentelemetry.io/ebpf-profiler/libpf"
)

func TestIterateMappingsInternedPathsSurviveBufferReuse(t *testing.T) {
	const maps = "1000-2000 r-xp 00000000 08:01 123 /maps-buffer-test/first.so\n" +
		"3000-4000 r-xp 00000000 08:01 124 /maps-buffer-test/other.so\n"
	want := []string{"/maps-buffer-test/first.so", "/maps-buffer-test/other.so"}
	var paths []libpf.String
	parseErrors, err := iterateMappings(strings.NewReader(maps), func(m RawMapping) bool {
		paths = append(paths, libpf.Intern(m.Path))
		return true
	})
	if err != nil || parseErrors != 0 || len(paths) != len(want) {
		t.Fatalf("iterateMappings: err=%v parseErrors=%d paths=%d", err, parseErrors, len(paths))
	}
	// iterateMappings has now zeroed its buffer and returned it to the pool.
	for i, path := range paths {
		if path.String() != want[i] || path != libpf.Intern(want[i]) {
			t.Fatalf("path %d lost its value or canonical identity after pool return", i)
		}
	}
}
