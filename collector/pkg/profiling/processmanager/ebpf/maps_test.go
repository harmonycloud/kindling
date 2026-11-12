package ebpf

import (
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/cgoapi"
	"go.opentelemetry.io/ebpf-profiler/lpm"
)

// Helpers themselves are pure Go. CGO_ENABLED=1 still links cgoapi
// (libkindling.so via cgoapi LDFLAGS).

func TestEncodeBiasAndUnwindProgram(t *testing.T) {
	v, err := EncodeBiasAndUnwindProgram(0x1234, ProgUnwindNative)
	if err != nil {
		t.Fatal(err)
	}
	bias, prog := DecodeBiasAndUnwindProgram(v)
	if bias != 0x1234 || prog != ProgUnwindNative {
		t.Fatalf("got bias=0x%x prog=%d", bias, prog)
	}
}

func TestEncodeBiasTooLarge(t *testing.T) {
	if _, err := EncodeBiasAndUnwindProgram(1<<56, ProgUnwindNative); err == nil {
		t.Fatal("expected error")
	}
}

func TestMapIDForNumDeltas(t *testing.T) {
	cases := []struct {
		n    uint32
		want uint16
	}{
		{1, 8},
		{255, 8},
		{256, 9},
		{512, 10},
		{1 << 22, 23},
	}
	for _, tc := range cases {
		got, err := MapIDForNumDeltas(tc.n)
		if err != nil || got != tc.want {
			t.Fatalf("n=%d: got %d err=%v want %d", tc.n, got, err, tc.want)
		}
	}
	if _, err := MapIDForNumDeltas(1 << 23); err == nil {
		t.Fatal("expected overflow for 1<<23")
	}
}

func TestBuildStackDeltaPageEntries(t *testing.T) {
	keys, values := buildStackDeltaPageEntries(0xabc, []uint16{40, 35, 28}, 9, 0x10000)
	if len(keys) != 3 || len(values) != 3 {
		t.Fatalf("len keys=%d values=%d", len(keys), len(values))
	}
	if keys[0].Page != 0x10000 || keys[1].Page != 0x20000 || keys[2].Page != 0x30000 {
		t.Fatalf("pages: %+v", keys)
	}
	if values[0].FirstDelta != 0 || values[0].NumDeltas != 40 || values[0].MapID != 9 {
		t.Fatalf("page0: %+v", values[0])
	}
	if values[1].FirstDelta != 40 || values[2].FirstDelta != 75 {
		t.Fatalf("firstDelta chain: %d %d", values[1].FirstDelta, values[2].FirstDelta)
	}
}

func TestPidPagePrefixLen(t *testing.T) {
	prefix := lpm.Prefix{Key: 0x7f0000000000, Length: 40}
	key := cgoapi.PidPage{
		PrefixLen: BitWidthPID + prefix.Length,
		PID:       42,
		Page:      prefix.Key,
	}
	if key.PrefixLen != 72 || key.PID != 42 || key.Page != prefix.Key {
		t.Fatalf("%+v", key)
	}
}
