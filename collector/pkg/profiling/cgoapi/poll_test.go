package cgoapi

import "testing"

// The probe reader is not initialized in these unit tests. Its no-data/error
// path must reuse the batch buffer without allocating Go memory.
func TestTracePollerNoDataAllocations(t *testing.T) {
	poller := NewTracePoller(64)
	if samples, n := poller.Poll(); n > 0 || len(samples) != 0 {
		t.Fatalf("uninitialized reader returned %d samples", n)
	}
	allocs := testing.AllocsPerRun(100, func() {
		poller.Poll()
	})
	if allocs != 0 {
		t.Fatalf("no-data poll allocated %g times; want zero", allocs)
	}
}

func TestTraceSampleOwnsFrames(t *testing.T) {
	poller := NewTracePoller(1)
	raw := &poller.buf[0]
	raw.pid = 42
	raw.frame_data_len = 2
	raw.frames[0], raw.frames[1] = 0x123, 0x456
	first := traceFromC(raw)

	// Simulate the next poll overwriting the same raw slot with a shorter trace.
	raw.pid = 43
	raw.frame_data_len = 1
	raw.frames[0] = 0x789
	second := traceFromC(raw)
	if first.PID != 42 || len(first.Frames) != 2 || first.Frames[0] != 0x123 || first.Frames[1] != 0x456 {
		t.Fatalf("previous sample changed after buffer reuse: %+v", first)
	}
	if second.PID != 43 || len(second.Frames) != 1 || second.Frames[0] != 0x789 {
		t.Fatalf("next sample includes stale data: %+v", second)
	}
}

func BenchmarkTracePollerNoData(b *testing.B) {
	b.Run("NewBufferPerPoll", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			NewTracePoller(64).Poll()
		}
	})
	b.Run("ReuseBuffer", func(b *testing.B) {
		poller := NewTracePoller(64)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			poller.Poll()
		}
	})
}

func TestTraceSampleUsesExplicitWordCountAndBoundedComm(t *testing.T) {
	raw := &NewTracePoller(1).buf[0]
	raw.num_frames = 2
	raw.frame_data_len = 0
	for i := range raw.comm {
		raw.comm[i] = 'x'
	}
	sample := traceFromC(raw)
	if len(sample.Frames) != 0 {
		t.Fatalf("invented frame words: %v", sample.Frames)
	}
	if sample.Comm != "xxxxxxxxxxxxxxxx" {
		t.Fatalf("comm exceeded its fixed buffer: %q", sample.Comm)
	}
}
