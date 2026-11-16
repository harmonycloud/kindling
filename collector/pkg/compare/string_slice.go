// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package compare

type StringSlice struct {
	oldElements []string
	oldFlags    []bool
	newElements []string
	newFlags    []bool
}

func NewStringSlice(oldElements []string, newElements []string) StringSlice {
	return StringSlice{
		oldElements: oldElements,
		oldFlags:    make([]bool, len(oldElements)),
		newElements: newElements,
		newFlags:    make([]bool, len(newElements)),
	}
}

func (c *StringSlice) Compare() {
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

func (c *StringSlice) GetRemovedElements() []string {
	ret := make([]string, 0)
	for i, flag := range c.oldFlags {
		if !flag {
			ret = append(ret, c.oldElements[i])
		}
	}
	return ret
}

func (c *StringSlice) GetAddedElements() []string {
	ret := make([]string, 0)
	for i, flag := range c.newFlags {
		if !flag {
			ret = append(ret, c.newElements[i])
		}
	}
	return ret
}
