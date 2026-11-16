// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailRequestProduce() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.GetIntAttribute(constlabels.KafkaApi) != _apiProduce
	}
}

func parseRequestProduce() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		var (
			offset    int
			err       error
			topicNum  int32
			topicName string
		)
		version := message.GetIntAttribute(constlabels.KafkaVersion)
		compact := version >= 9
		offset = message.Offset

		if version >= 3 {
			var transactionId string
			if offset, err = message.ReadNullableString(offset, compact, &transactionId); err != nil {
				return false, true
			}
		}
		// acks, timeout_ms
		offset += 6
		if offset, err = message.ReadArraySize(offset, compact, &topicNum); err != nil {
			return false, true
		}
		if topicNum > 0 {
			if _, err = message.ReadString(offset, compact, &topicName); err != nil {
				return false, true
			}
			// Get TopicName
			message.AddUtf8StringAttribute(constlabels.KafkaTopic, topicName)
		}
		return true, true
	}
}
