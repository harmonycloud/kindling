// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

/**
-Error message\r\n
*/
func fastfailRedisError() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.Data[message.Offset] != '-'
	}
}

func parseRedisError() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		offset, data := message.ReadUntilCRLF(message.Offset + 1)
		if data == nil {
			return false, true
		}

		message.Offset = offset
		if len(data) > 0 && !message.HasAttribute(constlabels.RedisErrMsg) {
			message.AddByteArrayUtf8Attribute(constlabels.RedisErrMsg, data)
			message.AddBoolAttribute(constlabels.IsError, true)
			message.AddIntAttribute(constlabels.ErrorType, int64(constlabels.ProtocolError))
		}
		return true, message.IsComplete()
	}
}
