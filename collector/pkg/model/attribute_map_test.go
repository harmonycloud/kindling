// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttributeMap(t *testing.T) {
	attributeMap := NewAttributeMap()

	attributeMap.AddStringValue("aa", "000")
	attributeMap.AddStringValue("bb", "111")
	attributeMap.AddStringValue("cc", "222")
	attributeMap.AddIntValue("dd", 444)

	assert.EqualValues(t, "000", attributeMap.GetStringValue("aa"))
	assert.EqualValues(t, "111", attributeMap.GetStringValue("bb"))
	assert.EqualValues(t, "222", attributeMap.GetStringValue("cc"))
	assert.EqualValues(t, 444, attributeMap.GetIntValue("dd"))

	attributeMap.RemoveAttribute("ee")
	attributeMap.RemoveAttribute("aa")

	assert.EqualValues(t, 3, attributeMap.Size())

	attributeMap.ClearAttributes()
	assert.EqualValues(t, 0, attributeMap.Size())

	attributeMap.AddIntValue("ee", 555)
	assert.EqualValues(t, 555, attributeMap.GetIntValue("ee"))
	assert.EqualValues(t, 1, attributeMap.Size())
}
