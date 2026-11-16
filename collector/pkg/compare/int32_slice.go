// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package compare

type Int32Slice struct {
	oldElements []int32
	oldFlags    []bool
	newElements []int32
	newFlags    []bool
}

func NewInt32Slice(oldElements []int32, newElements []int32) Int32Slice {
	return Int32Slice{
		oldElements: oldElements,
		oldFlags:    make([]bool, len(oldElements)),
		newElements: newElements,
		newFlags:    make([]bool, len(newElements)),
	}
}

func (c *Int32Slice) Compare() {
	// The elements could be repeated, so we must iterate all the elements.
	for i, newElement := range c.newElements {
		for j, oldElement := range c.oldElements {
			if oldElement == newElement {
				c.newFlags[i] = true
				c.oldFlags[j] = true
			}
		}
	}
}

func (c *Int32Slice) GetRemovedElements() []int32 {
	ret := make([]int32, 0)
	for i, flag := range c.oldFlags {
		if !flag {
			ret = append(ret, c.oldElements[i])
		}
	}
	return ret
}

func (c *Int32Slice) GetAddedElements() []int32 {
	ret := make([]int32, 0)
	for i, flag := range c.newFlags {
		if !flag {
			ret = append(ret, c.newElements[i])
		}
	}
	return ret
}
