// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
)

/*
For Simple Strings the first byte of the reply is "+"
For Errors the first byte of the reply is "-"
For Integers the first byte of the reply is ":"
For Bulk Strings the first byte of the reply is "$"
For Arrays the first byte of the reply is "*"
*/
func fastfailResponse() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		keyword := message.Data[message.Offset]
		return keyword != '+' &&
			keyword != '-' &&
			keyword != '*' &&
			keyword != '$' &&
			keyword != ':'
	}
}

func parseResponse() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		return true, false
	}
}
