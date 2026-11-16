// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package aggregator

import (
	"sort"
	"strconv"

	"github.com/Kindling-project/kindling/collector/pkg/model"
)

type vType string

const (
	BooleanType vType = "boolean"
	StringType  vType = "string"
	IntType     vType = "int"
)

type LabelSelectors struct {
	// Name: VType
	selectors []LabelSelector
}
type LabelSelector struct {
	Name  string
	VType vType
}

func NewLabelSelectors(selectors ...LabelSelector) *LabelSelectors {
	return &LabelSelectors{selectors: selectors}
}

func (s *LabelSelectors) GetLabelKeys(labels *model.AttributeMap) *LabelKeys {
	keys := &LabelKeys{}
	for i := 0; i < maxLabelKeySize && i < len(s.selectors); i++ {
		selector := s.selectors[i]
		keys.keys[i].Name = selector.Name
		keys.keys[i].VType = selector.VType
		switch selector.VType {
		case BooleanType:
			keys.keys[i].Value = strconv.FormatBool(labels.GetBoolValue(selector.Name))
		case StringType:
			keys.keys[i].Value = labels.GetStringValue(selector.Name)
		case IntType:
			keys.keys[i].Value = strconv.FormatInt(labels.GetIntValue(selector.Name), 10)
		}
	}
	return keys
}

func (s *LabelSelectors) AppendSelectors(selectors ...LabelSelector) {
	s.selectors = append(s.selectors, selectors...)
}

const maxLabelKeySize = 37

type LabelKeys struct {
	// LabelKeys will be used as key of map, so it is must be an array instead of a slice.
	// Now 34 is enough for all cases. If there is more than 34 labels, must increase this value.
	keys [maxLabelKeySize]LabelKey
}

func (k *LabelKeys) Len() int {
	return len(k.keys)
}

func (k *LabelKeys) Swap(i, j int) {
	k.keys[i], k.keys[j] = k.keys[j], k.keys[i]
}

func (k *LabelKeys) Less(i, j int) bool {
	return k.keys[i].Name < k.keys[j].Name
}

type LabelKey struct {
	Name  string
	Value string
	VType vType
	sort.Interface
}

func NewLabelKeys(keys ...LabelKey) *LabelKeys {
	ret := &LabelKeys{}
	length := len(keys)
	// only the first maxLabelKeySize number of keys are valid
	for i := 0; i < maxLabelKeySize && i < length; i++ {
		ret.keys[i] = keys[i]
	}
	return ret
}

func (k *LabelKeys) GetLabels() *model.AttributeMap {
	ret := model.NewAttributeMap()
	for _, key := range k.keys {
		switch key.VType {
		case BooleanType:
			value, _ := strconv.ParseBool(key.Value)
			ret.AddBoolValue(key.Name, value)
		case StringType:
			ret.AddStringValue(key.Name, key.Value)
		case IntType:
			value, _ := strconv.ParseInt(key.Value, 10, 64)
			ret.AddIntValue(key.Name, value)
		}
	}
	return ret
}

// GetLabelsKeys TODO Sort before you real need to use this key
func GetLabelsKeys(attributeMap *model.AttributeMap) *LabelKeys {
	keys := &LabelKeys{}
	labels := attributeMap.GetValues()
	index := 0
	for k, v := range labels {
		keys.keys[index].Name = k
		keys.keys[index].Value = v.ToString()
		switch v.Type() {
		case model.StringAttributeValueType:
			keys.keys[index].VType = StringType
		case model.IntAttributeValueType:
			keys.keys[index].VType = IntType
		case model.BooleanAttributeValueType:
			keys.keys[index].VType = BooleanType
		}
		index++
	}
	sort.Sort(keys)
	return keys
}
