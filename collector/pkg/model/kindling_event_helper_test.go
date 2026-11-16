// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUintUserAttribute(t *testing.T) {
	tests := []struct {
		key       string
		valueType ValueType
		value     []byte
		expect    uint64
	}{
		{"uint8_large", ValueType_UINT8, []byte{0xff}, 255},
		{"uint8_small", ValueType_UINT8, []byte{0x7f}, 127},
		{"uint16_large", ValueType_UINT16, []byte{0xff, 0xff}, 65535},
		{"uint16_small", ValueType_UINT16, []byte{0xff, 0x7f}, 32767},
		{"uint32_large", ValueType_UINT32, []byte{0xff, 0xff, 0xff, 0xff}, 4294967295},
		{"uint32_small", ValueType_UINT32, []byte{0xff, 0xff, 0xff, 0x7f}, 2147483647},
		{"uint64_large", ValueType_UINT32, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, 4294967295},
		{"uint64_small", ValueType_UINT32, []byte{0xff, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x00, 0x00}, 2147483647},
	}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			event := &KindlingEvent{
				ParamsNumber: 1,
				UserAttributes: [16]KeyValue{
					{Key: test.key, ValueType: test.valueType, Value: test.value},
				},
			}
			assert.Equal(t, test.expect, event.GetUintUserAttribute(test.key))
		})
	}
}

func TestGetIntUserAttribute(t *testing.T) {
	tests := []struct {
		key       string
		valueType ValueType
		value     []byte
		expect    int64
	}{
		{"int8_negative", ValueType_INT8, []byte{0xff}, -1},
		{"int8_postive", ValueType_INT8, []byte{0x7f}, 127},
		{"int16_negative", ValueType_INT16, []byte{0xff, 0xff}, -1},
		{"int16_postive", ValueType_INT16, []byte{0xff, 0x7f}, 32767},
		{"int32_negative", ValueType_INT32, []byte{0xff, 0xff, 0xff, 0xff}, -1},
		{"int32_postive", ValueType_INT32, []byte{0xff, 0xff, 0xff, 0x7f}, 2147483647},
		{"int64_negative", ValueType_INT32, []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, -1},
		{"int64_postive", ValueType_INT32, []byte{0xff, 0xff, 0xff, 0x7f, 0x00, 0x00, 0x00, 0x00}, 2147483647},
	}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			event := &KindlingEvent{
				ParamsNumber: 1,
				UserAttributes: [16]KeyValue{
					{Key: test.key, ValueType: test.valueType, Value: test.value},
				},
			}
			assert.Equal(t, test.expect, event.GetIntUserAttribute(test.key))
		})
	}
}

func TestGetFloatUserAttribute(t *testing.T) {
	tests := []struct {
		key    string
		value  []byte
		expect float32
	}{
		{"float_negative", []byte{0x66, 0xe6, 0xf6, 0xc2}, -123.45},
		{"float_zero", []byte{0x00, 0x00, 0x00, 0x00}, 0.0},
		{"float_postive", []byte{0x66, 0xe6, 0xf6, 0x42}, 123.45},
	}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			event := &KindlingEvent{
				ParamsNumber: 1,
				UserAttributes: [16]KeyValue{
					{Key: test.key, ValueType: ValueType_FLOAT, Value: test.value},
				},
			}
			assert.Equal(t, test.expect, event.GetFloatUserAttribute(test.key))
		})
	}
}

func TestGetDoubleUserAttribute(t *testing.T) {
	tests := []struct {
		key    string
		value  []byte
		expect float64
	}{
		{"double_negative", []byte{0xcd, 0xcc, 0xcc, 0xcc, 0xcc, 0xdc, 0x5e, 0xc0}, -123.45},
		{"double_zero", []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 0.0},
		{"double_postive", []byte{0xcd, 0xcc, 0xcc, 0xcc, 0xcc, 0xdc, 0x5e, 0x40}, 123.45},
	}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			event := &KindlingEvent{
				ParamsNumber: 1,
				UserAttributes: [16]KeyValue{
					{Key: test.key, ValueType: ValueType_FLOAT, Value: test.value},
				},
			}
			assert.Equal(t, test.expect, event.GetDoubleUserAttribute(test.key))
		})
	}
}
