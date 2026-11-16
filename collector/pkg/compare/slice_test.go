// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package compare

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt32Slice(t *testing.T) {
	oldElements := []int32{0, 0, 1, 2, 3}
	newElements := []int32{1, 3, 4, 6}
	compareSlice := NewInt32Slice(oldElements, newElements)
	compareSlice.Compare()
	removedElements := compareSlice.GetRemovedElements()
	assert.ElementsMatch(t, removedElements, []int32{0, 0, 2})
	addedElements := compareSlice.GetAddedElements()
	assert.ElementsMatch(t, addedElements, []int32{4, 6})
}

func TestStringSlice(t *testing.T) {
	oldElements := []string{"a", "b", "c"}
	newElements := []string{"d", "e", "f"}
	compareSlice := NewStringSlice(oldElements, newElements)
	compareSlice.Compare()
	removedElements := compareSlice.GetRemovedElements()
	assert.ElementsMatch(t, removedElements, []string{"a", "b", "c"})
	addedElements := compareSlice.GetAddedElements()
	assert.ElementsMatch(t, addedElements, []string{"d", "e", "f"})
}
