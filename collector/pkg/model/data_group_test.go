// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"reflect"
	"testing"
)

func TestDataGroup_RemoveMetric(t *testing.T) {
	type fields struct {
		Values []*Metric
	}
	type args struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fields
	}{
		{
			name: "normal",
			fields: fields{
				Values: []*Metric{
					NewIntMetric("a", 1),
					NewIntMetric("b", 2),
					NewIntMetric("c", 3),
				},
			},
			args: args{name: "b"},
			want: fields{
				Values: []*Metric{
					NewIntMetric("a", 1),
					NewIntMetric("c", 3),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &DataGroup{
				Metrics: tt.fields.Values,
			}
			g.RemoveMetric(tt.args.name)
			// For output string
			wantG := &DataGroup{Metrics: tt.want.Values}
			if !reflect.DeepEqual(g.Metrics, tt.want.Values) {
				t.Errorf("expected %s, got %s", wantG, g)
			}
		})
	}
}
