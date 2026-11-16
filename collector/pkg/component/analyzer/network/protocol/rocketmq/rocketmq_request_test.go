// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package rocketmq

import (
	"testing"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
)

func Test_parseHeader(t *testing.T) {
	type args struct {
		message *protocol.PayloadMessage
		header  *rocketmqHeader
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "busy-loop",
			args: args{
				message: protocol.NewRequestMessage([]byte{
					1, 23, 4, 20, 123, 213, 4, 2, 34, 12, 23, 1, 23, 4, 20, 123,
					213, 4, 2, 34, 12, 0, 0, 0, 0, 20, 123, 213, 4, 254, 34, 12,
					23, 1, 23, 4, 20, 123, 213, 4, 2, 34, 12, 23, 1, 23, 4, 20,
					123, 213, 4, 2, 34, 12, 23}),
				header: &rocketmqHeader{ExtFields: map[string]string{}},
			},
		},
	}
	finishCh := make(chan bool)
	go func() {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				parseHeader(tt.args.message, tt.args.header)
			})
		}
		finishCh <- true
	}()

	select {
	case <-time.After(5 * time.Second):
		t.Fatal("The test case didn't finish in 5 seconds.")
	case <-finishCh:
		return
	}
}
