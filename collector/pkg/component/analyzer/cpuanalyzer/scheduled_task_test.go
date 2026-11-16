// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cpuanalyzer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestScheduledTask(t *testing.T) {
	// test case 1: Normal expired exit
	task1 := &testIncrementTask{0}
	routine1 := NewAndStartScheduledTaskRoutine(1*time.Millisecond, 5*time.Millisecond, task1, nil)
	_ = routine1.Start()
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, false, routine1.isRunning.Load())
	assert.Equal(t, 5, task1.count)

	// Case 2: Double start or double stop
	task3 := &testIncrementTask{0}
	routine3 := NewAndStartScheduledTaskRoutine(1*time.Millisecond, 5*time.Millisecond, task3, nil)
	err := routine3.Start()
	assert.Error(t, err)
	err = routine3.Stop()
	assert.NoError(t, err)
	err = routine3.Stop()
	assert.Error(t, err)
}

type testIncrementTask struct {
	count int
}

func (t *testIncrementTask) run() {
	t.count++
}
