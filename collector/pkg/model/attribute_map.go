// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"encoding/json"
	"strconv"
)

type AttributeValueType int

const (
	StringAttributeValueType AttributeValueType = iota
	IntAttributeValueType
	BooleanAttributeValueType
)

type AttributeMap struct {
	values map[string]AttributeValue
}

func (a AttributeMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.values)
}

func (a *AttributeMap) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &a.values)
}

func NewAttributeMap() *AttributeMap {
	values := make(map[string]AttributeValue)
	return &AttributeMap{values}
}

func NewAttributeMapWithValues(values map[string]AttributeValue) *AttributeMap {
	return &AttributeMap{values: values}
}

func NewStringValue(value string) AttributeValue {
	return &stringValue{value: value}
}

func NewIntValue(value int64) AttributeValue {
	return &intValue{value: value}
}

func NewBoolValue(value bool) AttributeValue {
	return &boolValue{value: value}
}

func (a *AttributeMap) Merge(other *AttributeMap) {
	if other == nil {
		return
	}
	for k, v := range other.values {
		a.values[k] = v
	}
}

func (a *AttributeMap) Size() int {
	return len(a.values)
}

func (a *AttributeMap) IsEmpty() bool {
	return len(a.values) == 0
}

func (a *AttributeMap) HasAttribute(key string) bool {
	_, existing := a.values[key]
	return existing
}

func (a *AttributeMap) GetStringValue(key string) string {
	value := a.values[key]
	if x, ok := value.(*stringValue); ok {
		return x.value
	}
	return ""
}

func (a *AttributeMap) AddStringValue(key string, value string) {
	a.values[key] = &stringValue{
		value: value,
	}
}

func (a *AttributeMap) UpdateAddStringValue(key string, value string) {
	if v, ok := a.values[key]; ok {
		v.(*stringValue).value = value
	} else {
		a.AddStringValue(key, value)
	}
}

func (a *AttributeMap) GetIntValue(key string) int64 {
	value := a.values[key]
	if x, ok := value.(*intValue); ok {
		return x.value
	}
	return 0
}

func (a *AttributeMap) AddIntValue(key string, value int64) {
	a.values[key] = &intValue{
		value: value,
	}
}

func (a *AttributeMap) UpdateAddIntValue(key string, value int64) {
	if v, ok := a.values[key]; ok {
		v.(*intValue).value = value
	} else {
		a.AddIntValue(key, value)
	}
}

func (a *AttributeMap) GetBoolValue(key string) bool {
	value := a.values[key]
	if x, ok := value.(*boolValue); ok {
		return x.value
	}
	return false
}

func (a *AttributeMap) AddBoolValue(key string, value bool) {
	a.values[key] = &boolValue{
		value: value,
	}
}

func (a *AttributeMap) UpdateAddBoolValue(key string, value bool) {
	if v, ok := a.values[key]; ok {
		v.(*boolValue).value = value
	} else {
		a.AddBoolValue(key, value)
	}
}

func (a *AttributeMap) RemoveAttribute(key string) {
	delete(a.values, key)
}

func (a *AttributeMap) ClearAttributes() {
	a.values = make(map[string]AttributeValue)
}

func (a *AttributeMap) ToStringMap() map[string]string {
	stringMap := make(map[string]string)
	if a == nil {
		return stringMap
	}
	for k, v := range a.values {
		stringMap[k] = v.ToString()
	}
	return stringMap
}

func (a *AttributeMap) GetValues() map[string]AttributeValue {
	if a != nil {
		return a.values
	}
	return nil
}

// ResetValues sets the default value for all elements. Used for implementing sync.Pool.
func (a *AttributeMap) ResetValues() {
	for _, v := range a.values {
		v.Reset()
	}
}

func (a *AttributeMap) String() string {
	json, _ := json.Marshal(a.ToStringMap())
	return string(json)
}

func (a *AttributeMap) Clone() *AttributeMap {
	ret := NewAttributeMap()
	for k, v := range a.values {
		switch v.Type() {
		case StringAttributeValueType:
			stringValue := v.(*stringValue).value
			ret.AddStringValue(k, stringValue)
		case IntAttributeValueType:
			intValue := v.(*intValue).value
			ret.AddIntValue(k, intValue)
		case BooleanAttributeValueType:
			boolValue := v.(*boolValue).value
			ret.AddBoolValue(k, boolValue)
		}
	}
	return ret
}

type AttributeValue interface {
	Type() AttributeValueType
	ToString() string
	Reset()
}

type stringValue struct {
	value string
}

func (v *stringValue) Type() AttributeValueType {
	return StringAttributeValueType
}

func (v stringValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *stringValue) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &v.value)
}

func (v *stringValue) ToString() string {
	return v.value
}

func (v *stringValue) Reset() {
	v.value = ""
}

type intValue struct {
	value int64
}

func (v *intValue) Type() AttributeValueType {
	return IntAttributeValueType
}

func (v *intValue) ToString() string {
	return strconv.FormatInt(v.value, 10)
}

func (v intValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *intValue) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &v.value)
}

func (v *intValue) Reset() {
	v.value = 0
}

type boolValue struct {
	value bool
}

func (v *boolValue) Type() AttributeValueType {
	return BooleanAttributeValueType
}

func (v *boolValue) ToString() string {
	return strconv.FormatBool(v.value)
}

func (v boolValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *boolValue) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &v.value)
}

func (v *boolValue) Reset() {
	v.value = false
}
