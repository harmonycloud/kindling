package processmanager

import (
	"context"
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/profiling/interpreter/interpreterconfig"
)

func TestNewPreservesInterpreterConfig(t *testing.T) {
	pythonOnly := interpreterconfig.NoInterpreters()
	pythonOnly.Python.Disabled = false
	for _, tc := range []struct {
		name string
		cfg  interpreterconfig.Config
		want int
	}{
		{"default loaders", interpreterconfig.Config{}, 2},
		{"explicitly disabled", interpreterconfig.NoInterpreters(), 0},
		{"python only", pythonOnly, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			pm, err := New(ctx, Config{InterpretersConfig: tc.cfg})
			if err != nil {
				t.Fatal(err)
			}
			defer pm.Close()
			if got := pm.eim.NumInterpreterLoaders(); got != tc.want {
				t.Fatalf("registered loaders=%d, want %d", got, tc.want)
			}
			if pm.interpreterTracerEnabled != (tc.want > 0) {
				t.Fatalf("interpreterTracerEnabled=%v, want %v", pm.interpreterTracerEnabled, tc.want > 0)
			}
		})
	}
}
