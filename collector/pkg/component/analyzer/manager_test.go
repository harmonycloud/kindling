// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package analyzer

import (
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	manager, err := NewManager(&testAnalyzer{}, &testConsumeAllAnalyzer{})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(manager.consumeAllEventsAnalyzers))
	assert.Equal(t, 2, len(manager.allAnalyzers))
	assert.Equal(t, 2, len(manager.eventAnalyzersMap))

	analyzers := manager.GetConsumableAnalyzers("evt1")
	assert.Equal(t, 2, len(analyzers))
	_, ok := analyzers[0].(*testAnalyzer)
	assert.True(t, ok)
	_, ok = analyzers[1].(*testConsumeAllAnalyzer)
	assert.True(t, ok)

	analyzers = manager.GetConsumableAnalyzers("evt2")
	assert.Equal(t, 2, len(analyzers))
	_, ok = analyzers[0].(*testAnalyzer)
	assert.True(t, ok)
	_, ok = analyzers[1].(*testConsumeAllAnalyzer)
	assert.True(t, ok)

	analyzers = manager.GetConsumableAnalyzers("evt3")
	assert.Equal(t, 1, len(analyzers))
	_, ok = analyzers[0].(*testConsumeAllAnalyzer)
	assert.True(t, ok)
}

type testAnalyzer struct {
}

// ConsumableEvents returns the events' name that this analyzer can consume
func (t *testAnalyzer) ConsumableEvents() []string {
	return []string{"evt1", "evt2"}
}

// Start initializes the analyzer
func (t *testAnalyzer) Start() error {
	return nil
}

// ConsumeEvent gets the event from the previous component
func (t *testAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
	return nil
}

// Shutdown cleans all the resources used by the analyzer
func (t *testAnalyzer) Shutdown() error {
	return nil
}

// Type returns the type of the analyzer
func (t *testAnalyzer) Type() Type {
	return "testanalyzer"
}

type testConsumeAllAnalyzer struct {
}

// ConsumableEvents returns the events' name that this analyzer can consume
func (t *testConsumeAllAnalyzer) ConsumableEvents() []string {
	return []string{ConsumeAllEvents}
}

// Start initializes the analyzer
func (t *testConsumeAllAnalyzer) Start() error {
	return nil
}

// ConsumeEvent gets the event from the previous component
func (t *testConsumeAllAnalyzer) ConsumeEvent(event *model.KindlingEvent) error {
	return nil
}

// Shutdown cleans all the resources used by the analyzer
func (t *testConsumeAllAnalyzer) Shutdown() error {
	return nil
}

// Type returns the type of the analyzer
func (t *testConsumeAllAnalyzer) Type() Type {
	return "testconsumeallanalyzer"
}
