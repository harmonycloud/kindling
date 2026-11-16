// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import "testing"

func TestParseTraceHeader(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		traceType string
		traceId   string
	}{
		{name: "skywalking", key: "sw8", value: "1-OWZhM2VkOGRhZDhmNGExNWFkYjIzNDgyNWJmYzcxMzUuNTQuMTY2MDE4OTQxNzk2MzAwMDE=-OWZhM2V", traceType: "skywalking", traceId: "9fa3ed8dad8f4a15adb234825bfc7135.54.16601894179630001"},
		{name: "zipkin", key: "x-b3-traceid", value: "223f3b00a283c75c", traceType: "zipkin", traceId: "223f3b00a283c75c"},
		{name: "jaeger", key: "uber-trace-id", value: "3997ed0a6a71f050:cf49be2de63d86e7:e02475aab05fd358:1", traceType: "jaeger", traceId: "3997ed0a6a71f050"},
		{name: "w3c-request", key: "traceparent", value: "00-4bf92f3577b34da6a3ce929d0e0e4736-d75597dee50b0cac-00", traceType: "w3c", traceId: "4bf92f3577b34da6a3ce929d0e0e4736"},
		{name: "w3c-response", key: "traceresponse", value: "00-4bf92f3577b34da6a3ce929d0e0e4736-828c5d0d435ba505-01", traceType: "w3c", traceId: "4bf92f3577b34da6a3ce929d0e0e4736"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				tt.key: tt.value,
			}

			traceType, traceId := ParseTraceHeader(headers)
			if traceType != tt.traceType {
				t.Errorf("Fail to check traceType, got = %s, want %s", traceType, tt.traceType)
			}

			if traceId != tt.traceId {
				t.Errorf("Fail to check traceId, got = %s, want %s", traceId, tt.traceId)
			}
		})
	}
}

func Test_parseSkyWalkingTraceId(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Complete TraceID",
			args: args{value: "1-OWZhM2VkOGRhZDhmNGExNWFkYjIzNDgyNWJmYzcxMzUuNTQuMTY2MDE4OTQxNzk2MzAwMDE=-OWZhM2VkOGRhZDhmNGExNWFkYjIzNDgyNWJmYzcxMzUuNTQuMTY2MDE4OTQxNzk2MzAwMDA=-1-ZHViYm8tY29uc3VtZXI=-Y2RlMTQ0MWI5NjMwNGY1M2I0NzJjZ"},
			want: "9fa3ed8dad8f4a15adb234825bfc7135.54.16601894179630001",
		},
		{
			name: "Half TraceID",
			args: args{value: "1-OWZhM2VkOGRhZDhmNGExNWFkYjIz"},
			want: "9fa3ed8dad8f4a15adb23",
		},
		{
			name: "No TraceID",
			args: args{value: "1-"},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseSkyWalkingTraceId(tt.args.value); got != tt.want {
				t.Errorf("parseSkyWalkingTraceId() = %v, want %v", got, tt.want)
			}
		})
	}
}
