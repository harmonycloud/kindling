// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"reflect"
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
)

func Test_urlMerge(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{

		{
			name: "more than two segments",
			args: args{url: "/test/123/456/test?a=1"},
			want: "/test/123/456/test",
		},
		{
			name: "two segments",
			args: args{url: "/test/123"},
			want: "/test/123",
		},
		{
			name: "two segments, but more than two /",
			args: args{url: "/test/123/"},
			want: "/test/123/",
		},
		{
			name: "one segment",
			args: args{url: "/test"},
			want: "/test",
		},
		{
			name: "one segment, but two /",
			args: args{url: "/test/"},
			want: "/test/",
		},
		{
			name: "only one /",
			args: args{url: "/"},
			want: "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getContentKey(tt.args.url); got != tt.want {
				t.Errorf("UrlMerge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHttpParser_getContentKey(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "normal", args: args{url: "/test/arg?a=12314"}, want: "/test/arg"},
		{name: "normal", args: args{url: "/test/arg/sar?a=12314"}, want: "/test/arg/sar"},
		{name: "normal", args: args{url: "/test"}, want: "/test"},
		{name: "zero", args: args{url: ""}, want: ""},
		{name: "zero", args: args{url: "/test/arg/adf/fadf/adf"}, want: "/test/arg/adf/fadf/adf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getContentKey(tt.args.url); got != tt.want {
				t.Errorf("getContentKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseHeaders(t *testing.T) {
	type args struct {
		message *protocol.PayloadMessage
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "normal case",
			args: args{
				message: protocol.NewRequestMessage([]byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\nAPM-AgentID: TTXvC3EQS6KLwxx3eIqINFjAW2olRm+cr8M+yuvwhkY=\r\nTransfer-Encoding: chunked\r\nContent-Type: application/json\r\nAPM-TransactionID: 5e480579c718a4a6498a9")),
			},
			want: map[string]string{
				"connection":        "keep-alive",
				"apm-agentid":       "TTXvC3EQS6KLwxx3eIqINFjAW2olRm+cr8M+yuvwhkY=",
				"transfer-encoding": "chunked",
				"content-type":      "application/json",
				"apm-transactionid": "5e480579c718a4a6498a9",
			},
		},
		{
			name: "no values",
			args: args{
				protocol.NewRequestMessage([]byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\nTransfer-Encoding: ")),
			},
			want: map[string]string{
				"connection": "keep-alive",
			},
		},
		{

			name: "no spaces",
			args: args{
				protocol.NewRequestMessage([]byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\nTransfer-Encoding:")),
			},
			want: map[string]string{
				"connection": "keep-alive",
			},
		},
		{
			name: "no colon",
			args: args{
				protocol.NewRequestMessage([]byte("HTTP/1.1 200 OK\r\nConnection: keep-alive\r\nTransfer-Encoding")),
			},
			want: map[string]string{
				"connection": "keep-alive",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseHeaders(tt.args.message); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseHeaders() = %v, want %v", got, tt.want)
			}
		})
	}
}
