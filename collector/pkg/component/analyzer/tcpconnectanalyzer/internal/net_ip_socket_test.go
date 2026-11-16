// Copyright 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"net"
	"reflect"
	"testing"
)

func Test_parseNetIPSocketLine(t *testing.T) {
	tests := []struct {
		fields  []string
		name    string
		want    *netIPSocketLine
		wantErr bool
	}{
		{
			name:   "reading valid lines, no issue should happened",
			fields: []string{"11:", "00000000:0000", "00000000:0000", "0A", "00000017:0000002A", "0:0", "0", "1000", "0", "39309"},
			want: &netIPSocketLine{
				LocalAddr: net.IP{0, 0, 0, 0},
				LocalPort: 0,
				RemAddr:   net.IP{0, 0, 0, 0},
				RemPort:   0,
				St:        "0A",
			},
		},
		{
			name:    "error case - invalid line - number of fields/columns < 10",
			fields:  []string{"1:", "00000000:0000", "00000000:0000", "07", "0:0", "0", "0"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "error case - parse local_address - not a valid hex",
			fields:  []string{"1:", "0000000O:0000", "00000000:0000", "07", "00000000:00000001", "0:0", "0", "0", "0", "39309"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "error case - parse rem_address - not a valid hex",
			fields:  []string{"1:", "00000000:0000", "0000000O:0000", "07", "00000000:00000001", "0:0", "0", "0", "0", "39309"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseNetIPSocketLine(tt.fields)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseNetIPSocketLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want == nil && got != nil {
				t.Errorf("parseNetIPSocketLine() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseNetIPSocketLine() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
