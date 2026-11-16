// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cpuanalyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPointer(t *testing.T) {
	queue := make(chan MyTask, 10)
	for i := 0; i < 5; i++ {
		queue <- MyTask{i}
	}
	close(queue)
	array := make([]Task, 0)
	for v := range queue {
		// Copy the value and the get its pointer
		tmp := v
		array = append(array, &tmp)
	}
	for i, v := range array {
		assert.Equal(t, i, v.getAge())
	}
}

type MyTask struct {
	age int
}

func (t *MyTask) getAge() int {
	return t.age
}

type Task interface {
	getAge() int
}
