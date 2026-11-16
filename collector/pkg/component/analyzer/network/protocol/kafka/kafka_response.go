// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailResponse() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 8
	}
}

func parseResponse() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		var (
			payloadLength int32
			correlationId int32
		)
		message.ReadInt32(0, &payloadLength)
		if payloadLength <= 4 {
			return false, true
		}

		message.ReadInt32(4, &correlationId)
		if !message.HasAttribute(constlabels.KafkaCorrelationId) ||
			message.GetIntAttribute(constlabels.KafkaCorrelationId) != int64(correlationId) {

			return false, true
		}
		message.Offset = 8
		return true, false
	}
}
