// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package conntracker

import (
	"net"
	"reflect"
	"testing"
)

func Test_int32ToIp(t *testing.T) {
	type args struct {
		i uint32
	}
	tests := []struct {
		name string
		args args
		want net.IP
	}{
		{"normal", args{67305985}, net.IP{1, 2, 3, 4}},
		{"zero", args{0}, net.IP{0, 0, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := int32ToIp(tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("int32ToIp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPToUInt32(t *testing.T) {
	type args struct {
		ip net.IP
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{name: "", args: args{ip: net.IPv4(1, 2, 3, 4)}, want: 67305985},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IPToUInt32(tt.args.ip); got != tt.want {
				t.Errorf("IPToUInt32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStringToUint32(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		{name: "normal", args: args{ip: "1.2.3.4"}, want: 67305985},
		{name: "illegal", args: args{ip: "1.2"}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringToUint32(tt.args.ip); got != tt.want {
				t.Errorf("StringToUint32() = %v, want %v", got, tt.want)
			}
		})
	}
}
