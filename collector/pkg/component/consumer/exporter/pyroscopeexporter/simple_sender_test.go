package pyroscopeexporter

import (
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"
)

func TestPrepareDataForTransmission(t *testing.T) {
	lbs := labels.Labels{
		{Name: "__name__", Value: "process_cpu"},
		{Name: "__internal__", Value: "drop-me"},
		{Name: "pid", Value: "123"},
	}
	samples := []*RawSample{{RawProfile: []byte("gzip-pprof")}}
	protoLabels, protoSamples, err := PrepareDataForTransmission(lbs, samples)
	require.NoError(t, err)
	require.Len(t, protoSamples, 1)
	require.Equal(t, []byte("gzip-pprof"), protoSamples[0].RawProfile)

	got := map[string]string{}
	for _, l := range protoLabels {
		got[l.Name] = l.Value
	}
	require.Equal(t, "process_cpu", got["__name__"])
	require.Equal(t, "123", got["pid"])
	_, hasInternal := got["__internal__"]
	require.False(t, hasInternal)
}
