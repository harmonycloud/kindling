// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"strconv"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
)

/*
:1\r\n
*/
func fastfailRedisInteger() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.Data[message.Offset] != ':'
	}
}

func parseRedisInteger() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		offset, data := message.ReadUntilCRLF(message.Offset + 1)
		if data == nil {
			return false, true
		}

		_, err := strconv.Atoi(string(data))
		if err != nil {
			return false, true
		}
		message.Offset = offset
		return true, message.IsComplete()
	}
}
