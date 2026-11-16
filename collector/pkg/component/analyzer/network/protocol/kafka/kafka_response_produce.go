// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailResponseProduce() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.GetIntAttribute(constlabels.KafkaApi) != _apiProduce
	}
}

func parseResponseProduce() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		var (
			offset       int
			err          error
			topicNum     int32
			topicName    string
			partitionNum int32
			errorCode    int16
		)
		version := message.GetIntAttribute(constlabels.KafkaVersion)
		compact := version >= 9
		offset = message.Offset
		if offset, err = message.ReadArraySize(offset, compact, &topicNum); err != nil {
			return false, true
		}
		if topicNum > 0 {
			if offset, err = message.ReadString(offset, compact, &topicName); err != nil {
				return false, true
			}
			if offset, err = message.ReadArraySize(offset, compact, &partitionNum); err != nil {
				return false, true
			}
			if partitionNum > 0 {
				offset += 4
				// Read ErrorCode in First Partition
				if _, err = message.ReadInt16(offset, &errorCode); err != nil {
					return false, true
				}
			}
			// Get topicName
			message.AddUtf8StringAttribute(constlabels.KafkaTopic, topicName)
		}
		message.AddIntAttribute(constlabels.KafkaErrorCode, int64(errorCode))
		return true, true
	}
}
