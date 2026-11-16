// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"strconv"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
)

/**
*0\r\n
*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n
*3\r\n:1\r\n:2\r\n:3\r\n
 */
func fastfailRedisArray() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.Data[message.Offset] != '*'
	}
}

func parseRedisArray() protocol.ParsePkgFn {
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
