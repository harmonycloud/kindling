package profiling

import (
	"testing"
)

func TestLocalPIDs(t *testing.T) {
	cfg := Config{Pids: []uint32{42, 42, 43}}
	got := cfg.localPIDs()
	if len(got) != 2 {
		t.Fatalf("len=%d got=%v", len(got), got)
	}
	if got[0] != 42 || got[1] != 43 {
		t.Fatalf("got=%v want [42 43]", got)
	}
	if n := (Config{}).localPIDs(); len(n) != 0 {
		t.Fatalf("empty config pids=%v", n)
	}
}

func TestPIDSelection(t *testing.T) {
	for _, tc := range []struct {
		name  string
		pids  []uint32
		all   bool
		valid bool
	}{
		{name: "omitted", valid: true},
		{name: "empty", pids: []uint32{}, valid: true},
		{name: "all", pids: []uint32{0}, all: true, valid: true},
		{name: "whitelist", pids: []uint32{1234, 5678}, valid: true},
		{name: "mixed", pids: []uint32{0, 1234}},
		{name: "trailing marker", pids: []uint32{1234, 0}},
		{name: "repeated marker", pids: []uint32{0, 0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Config{Pids: tc.pids}
			if got := cfg.captureAll(); got != tc.all {
				t.Fatalf("captureAll = %v, want %v", got, tc.all)
			}
			if err := cfg.validateTargets(); (err == nil) != tc.valid {
				t.Fatalf("validateTargets = %v, valid = %v", err, tc.valid)
			}
		})
	}
}
