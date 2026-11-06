package traceagg

import "testing"

func TestAggregatorMergeSameStack(t *testing.T) {
	a := New()
	frames := []Frame{NewNativeFrame(1, 0x100, 0), NewNativeFrame(2, 0x200, 0)}
	a.Add(42, frames, "foo")
	a.Add(42, frames, "foo")
	a.Add(42, []Frame{NewNativeFrame(1, 0x100, 0), NewNativeFrame(2, 0x201, 0)}, "foo")

	if a.Len() != 2 {
		t.Fatalf("Len=%d want 2", a.Len())
	}
	got := a.Swap()
	if a.Len() != 0 {
		t.Fatalf("after Swap Len=%d want 0", a.Len())
	}
	var merged uint64
	for _, e := range got {
		if e.Count == 2 {
			merged = e.Count
			if len(e.Frames) != 2 {
				t.Fatalf("frames len=%d", len(e.Frames))
			}
		}
	}
	if merged != 2 {
		t.Fatalf("expected one entry with count=2, got map size=%d", len(got))
	}
}

func TestAggregatorDifferentPID(t *testing.T) {
	a := New()
	frames := []Frame{NewNativeFrame(1, 0x10, 0)}
	a.Add(1, frames, "a")
	a.Add(2, frames, "b")
	if a.Len() != 2 {
		t.Fatalf("Len=%d want 2", a.Len())
	}
}

func TestAggregatorCap(t *testing.T) {
	a := New()
	a.max = 2
	f := func(va uint64) []Frame { return []Frame{NewNativeFrame(1, va, 0)} }
	a.Add(1, f(0x10), "a")
	a.Add(1, f(0x20), "a")
	a.Add(1, f(0x30), "a") // new stack, table full
	a.Add(1, f(0x30), "a")
	a.Add(1, f(0x10), "a") // known stack still counts

	if a.Len() != 2 || a.Dropped() != 2 {
		t.Fatalf("Len=%d Dropped=%d, want 2 and 2", a.Len(), a.Dropped())
	}
	got := a.Swap()
	var total uint64
	for _, e := range got {
		total += e.Count
	}
	if total != 3 {
		t.Fatalf("counted %d samples, want 3", total)
	}
	a.Add(1, f(0x30), "a")
	if a.Len() != 1 {
		t.Fatalf("after Swap Len=%d want 1", a.Len())
	}
}

func TestAggregatorTakePeak(t *testing.T) {
	a := New()
	add := func(n int) {
		for i := 0; i < n; i++ {
			a.Add(1, []Frame{NewNativeFrame(1, uint64(i), 0)}, "a")
		}
		a.Swap()
	}
	add(3)
	add(5)
	add(2)
	if p := a.TakePeak(); p != 5 {
		t.Fatalf("peak=%d want 5", p)
	}
	add(1)
	if p := a.TakePeak(); p != 1 {
		t.Fatalf("peak after reset=%d want 1", p)
	}
}

func TestAggregatorEmptyIgnored(t *testing.T) {
	a := New()
	a.Add(1, nil, "x")
	if a.Len() != 0 {
		t.Fatalf("empty frames should be ignored")
	}
}

func TestPythonAggregationIgnoresLastI(t *testing.T) {
	a := New()
	a.Add(1, []Frame{NewPythonFrame(0x1000, 0x11, 0x42, 4)}, "python")
	a.Add(1, []Frame{NewPythonFrame(0x1000, 0x22, 0x42, 4)}, "python")
	a.Add(1, []Frame{NewPythonFrame(0x1000, 0x22, 0x43, 4)}, "python")

	if a.Len() != 2 {
		t.Fatalf("Len=%d want 2", a.Len())
	}
	var merged *Entry
	for _, entry := range a.Swap() {
		if entry.Count == 2 {
			merged = entry
		}
	}
	if merged == nil {
		t.Fatal("different lasti values did not merge")
	}
	py, ok := merged.Frames[0].Python()
	if !ok || py.CodeObject != 0x1000 || py.LastI != 0x11 || py.Checksum != 0x42 {
		t.Fatalf("python=%+v ok=%v", py, ok)
	}
}

func TestInterpreterNamesAreOwnedAndOnlyReadOnce(t *testing.T) {
	a := New()
	frames := []Frame{NewPythonFrame(0x1000, 1, 99, 4), NewNativeFrame(7, 123, 0)}
	names := []string{"处理 (app.py)", ""}
	calls := 0
	resolve := func() []string { calls++; return names }
	a.AddWithNames(1, frames, "python", resolve)
	a.AddWithNames(1, []Frame{NewPythonFrame(0x1000, 2, 99, 4), NewNativeFrame(7, 123, 0)}, "python", resolve)
	names[0] = "overwritten"
	frames[0] = NewPythonFrame(0x2000, 3, 88, 4)
	for _, entry := range a.Swap() {
		if entry.Count != 2 || entry.InterpNames[0] != "处理 (app.py)" {
			t.Fatalf("%+v", entry)
		}
		if py, _ := entry.Frames[0].Python(); py.CodeObject != 0x1000 {
			t.Fatal("retained caller frame slice")
		}
	}
	if calls != 1 {
		t.Fatalf("name reads=%d", calls)
	}
}

func TestSwapDuringInterpreterReadKeepsNames(t *testing.T) {
	a := New()
	a.Add(9, []Frame{NewNativeFrame(1, 1, 0)}, "native")
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	go func() {
		a.AddWithNames(1, []Frame{NewPythonFrame(0x1000, 1, 99, 4)}, "python", func() []string {
			close(entered)
			<-release
			return []string{"leaf (app.py)"}
		})
		close(done)
	}()
	<-entered
	// Would deadlock if the remote read held the table lock.
	if old := a.Swap(); len(old) != 1 {
		t.Fatalf("old entries=%d", len(old))
	}
	close(release)
	<-done
	for _, entry := range a.Swap() {
		if entry.Count != 1 || len(entry.InterpNames) != 1 || entry.InterpNames[0] != "leaf (app.py)" {
			t.Fatalf("%+v", entry)
		}
	}
}

func TestFullTableSkipsInterpreterRead(t *testing.T) {
	a := New()
	a.max = 1
	a.Add(1, []Frame{NewNativeFrame(1, 1, 0)}, "native")
	a.AddWithNames(1, []Frame{NewPythonFrame(0x1000, 1, 99, 4)}, "python", func() []string { t.Fatal("read names for rejected entry"); return nil })
	if a.Dropped() != 1 {
		t.Fatal("missing dropped sample")
	}
}
