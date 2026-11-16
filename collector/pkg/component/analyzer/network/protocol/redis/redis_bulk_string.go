// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"strconv"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

/*
$3\r\nbar\r\n
*/
func fastfailRedisBulkString() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.Data[message.Offset] != '$'
	}
}

func parseRedisBulkString() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		offset, data := message.ReadUntilCRLF(message.Offset + 1)
		if data == nil {
			return false, true
		}

		size, err := strconv.Atoi(string(data))
		if err != nil {
			return false, true
		}

		// $-1\r\n
		if size == -1 {
			message.Offset = offset
			return true, message.IsComplete()
		}

		offset, data = message.ReadUntilCRLF(offset)
		if data == nil {
			return false, true
		}

		/**
		$0\r\n\r\n
		$6\r\nfoobar\r\n
		*/
		if len(data) != size {
			return false, true
		}

		command := string(data)
		if !message.HasAttribute(command) && IsRedisCommand(data) {
			message.AddUtf8StringAttribute(constlabels.ContentKey, command)
			message.AddUtf8StringAttribute(constlabels.RedisCommand, command)
		}

		message.Offset = offset
		return true, message.IsComplete()
	}
}
