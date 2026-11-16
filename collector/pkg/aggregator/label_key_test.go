// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package aggregator

import (
	"reflect"
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/model"
)

func TestLabelSelectors_GetLabelKeys(t *testing.T) {
	type fields struct {
		selectors []LabelSelector
	}
	type args struct {
		labels *model.AttributeMap
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *LabelKeys
	}{
		{
			fields: fields{selectors: []LabelSelector{
				{Name: "stringKey", VType: StringType},
				{Name: "booleanKey", VType: BooleanType},
				{Name: "intKey", VType: IntType},
			}},
			args: args{labels: createAttributeMap()},
			want: createLabelKeys(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LabelSelectors{
				selectors: tt.fields.selectors,
			}
			if got := s.GetLabelKeys(tt.args.labels); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLabelKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLabelKeys_GetLabels(t *testing.T) {
	type fields struct {
		keys [maxLabelKeySize]LabelKey
	}
	tests := []struct {
		name   string
		fields fields
		want   *model.AttributeMap
	}{
		{
			fields: fields{createLabelKeys().keys},
			want:   createAttributeMap(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &LabelKeys{
				keys: tt.fields.keys,
			}
			if got := k.GetLabels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createAttributeMap() *model.AttributeMap {
	ret := model.NewAttributeMap()
	ret.AddIntValue("intKey", 100)
	ret.AddStringValue("stringKey", "stringValue")
	ret.AddBoolValue("booleanKey", true)
	return ret
}

func createLabelKeys() *LabelKeys {
	ret := NewLabelKeys([]LabelKey{
		{
			Name:  "stringKey",
			Value: "stringValue",
			VType: StringType,
		},
		{
			Name:  "booleanKey",
			Value: "true",
			VType: BooleanType,
		},
		{
			Name:  "intKey",
			Value: "100",
			VType: IntType,
		},
	}...)
	return ret
}

func TestLabelKeysMap(t *testing.T) {
	store := make(map[LabelKeys]int)
	labelKeys := createLabelKeys()
	store[*labelKeys] = 10

	labelKeys2 := createLabelKeys()
	value := store[*labelKeys2]
	if value != 10 {
		t.Errorf("Expected 10, but got %v", value)
	}
}
