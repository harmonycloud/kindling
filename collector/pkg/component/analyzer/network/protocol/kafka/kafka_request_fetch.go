// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailRequestFetch() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.GetIntAttribute(constlabels.KafkaApi) != _apiFetch
	}
}

func parseRequestFetch() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		var (
			offset    int
			err       error
			topicNum  int32
			topicName string
		)
		version := message.GetIntAttribute(constlabels.KafkaVersion)
		compact := version >= 12

		// replica_id, max_wait_ms, min_bytes,
		offset = message.Offset + 12
		if version >= 3 {
			offset += 4 // max_bytes
		}
		if version >= 4 {
			offset += 1 // isolation_level
		}
		if version >= 7 {
			offset += 8 // session_id, session_epoch
		}

		if offset, err = message.ReadArraySize(offset, compact, &topicNum); err != nil {
			return false, true
		}
		if topicNum > 0 {
			if _, err = message.ReadString(offset, compact, &topicName); err != nil {
				return false, true
			}
			/*
				Based on following case, we	just read first topicName.

				1. The payload is substr with fixed length(1K), it's not able to get all topic names
				2. Even if we get enough length for payload, the parser will take more performance cost
				3. There is not enough cases to cover multi-topics

				Since version 13, topicName will be repalced with topicId as uuid, therefore topicName is not able to be got.
			*/
			message.AddUtf8StringAttribute(constlabels.KafkaTopic, topicName)
		}

		return true, true
	}
}
