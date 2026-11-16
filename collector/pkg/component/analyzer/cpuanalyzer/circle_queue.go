// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cpuanalyzer

type CircleQueue struct {
	length int
	data   []interface{}
}

func NewCircleQueue(length int) *CircleQueue {
	return &CircleQueue{
		length: length,
		data:   make([]interface{}, length),
	}
}

func (s *CircleQueue) GetByIndex(index int) interface{} {
	return s.data[index]
}

func (s *CircleQueue) UpdateByIndex(index int, val interface{}) {
	s.data[index] = val
	return
}

func (s *CircleQueue) Clear() {
	s.data = make([]interface{}, s.length)
}
