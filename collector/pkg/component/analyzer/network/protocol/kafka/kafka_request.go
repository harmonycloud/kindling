// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailRequest() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 12
	}
}

func parseRequest() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		var (
			payloadLength  int32
			apiKey         int16
			apiVersion     int16
			correlationId  int32
			clientIdLength int16
		)
		message.ReadInt32(0, &payloadLength)
		if payloadLength <= 8 {
			return false, true
		}

		message.ReadInt16(4, &apiKey)
		message.ReadInt16(6, &apiVersion)
		if !IsValidVersion(int(apiKey), int(apiVersion)) {
			return false, true
		}

		message.ReadInt32(8, &correlationId)
		message.ReadInt16(12, &clientIdLength)
		if correlationId < 0 || clientIdLength < 0 {
			return false, true
		}
		var offset = int(clientIdLength) + 14
		if len(message.Data) < offset {
			return false, true
		}
		message.Offset = offset
		message.AddIntAttribute(constlabels.KafkaApi, int64(apiKey))
		message.AddIntAttribute(constlabels.KafkaVersion, int64(apiVersion))
		message.AddIntAttribute(constlabels.KafkaCorrelationId, int64(correlationId))
		return true, false
	}
}
