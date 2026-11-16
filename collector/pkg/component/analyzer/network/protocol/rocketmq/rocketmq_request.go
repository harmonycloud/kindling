// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package rocketmq

import (
	"encoding/json"
	"fmt"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailRocketMQRequest() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 29
	}
}

func parseRocketMQRequest() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		var (
			payloadLength int32
			headerLength  int32
			serializeType uint8
		)
		// Parsing the message content
		message.ReadInt32(0, &payloadLength)
		header := &rocketmqHeader{ExtFields: map[string]string{}}
		//When serializeType==0, the serialization type is JSON, and the json sequence is directly intercepted for deserialization
		if serializeType = message.Data[4]; serializeType == 0 {
			message.ReadInt32(4, &headerLength)
			_, headerBytes, err := message.ReadBytes(8, int(headerLength))
			if err != nil {
				return false, true
			}
			if err = json.Unmarshal(headerBytes, header); err != nil {
				return false, true
			}
		} else if serializeType == 1 {
			parseHeader(message, header)
		} else {
			return false, true
		}

		// Store the parsed attribute via AddStringAttribute() or AttIntAttribute()
		message.AddStringAttribute(constlabels.RocketMQRequestMsg, requestMsgMap[header.Code])
		message.AddIntAttribute(constlabels.RocketMQOpaque, int64(header.Opaque))

		//topicName maybe be stored in key `topic` or `b`
		if header.ExtFields["topic"] != "" {
			message.AddStringAttribute(constlabels.ContentKey, fmt.Sprintf("Topic:%v", header.ExtFields["topic"]))
		} else if header.ExtFields["b"] != "" {
			message.AddStringAttribute(constlabels.ContentKey, fmt.Sprintf("Topic:%v", header.ExtFields["b"]))
		} else {
			message.AddStringAttribute(constlabels.ContentKey, requestMsgMap[header.Code])
		}

		// Parsing succeeded
		return true, true
	}
}

// When serializeType==1, the serialization type is RocketMQ, and the fields are stored strictly in the order of [code,languagecode,version,opaque,flag,remark,extfields]
func parseHeader(message *protocol.PayloadMessage, header *rocketmqHeader) {
	var (
		remarkLen   int32
		extFieldLen int32
		offset      int
		err         error
	)
	message.ReadInt16(8, &header.Code)
	header.LanguageCode = message.Data[10]
	message.ReadInt16(11, &header.Version)
	message.ReadInt32(13, &header.Opaque)
	message.ReadInt32(17, &header.Flag)

	message.ReadInt32(21, &remarkLen)
	offset, _ = message.ReadInt32(25, &extFieldLen)
	if extFieldLen > 0 && remarkLen == 0 {
		extFieldMap := make(map[string]string)
		var (
			keyLen   int16
			valueLen int32
			key      []byte
			value    []byte
		)
		//offset starts from 29
		var extFieldBytesLen = 0
		for extFieldBytesLen < int(extFieldLen) && extFieldBytesLen+29 < len(message.Data) {
			offset, err = message.ReadInt16(offset, &keyLen)
			if err != nil {
				break
			}
			offset, key, err = message.ReadBytes(offset, int(keyLen))
			if err != nil {
				break
			}
			offset, err = message.ReadInt32(offset, &valueLen)
			if err != nil {
				break
			}
			offset, value, err = message.ReadBytes(offset, int(valueLen))
			if err != nil {
				break
			}
			extFieldMap[string(key)] = string(value)
			extFieldBytesLen = extFieldBytesLen + 2 + int(keyLen) + 4 + int(valueLen)
			if string(key) == "topic" || string(key) == "b" {
				break
			}
		}
		//Update the field `ExtFields` of the header
		header.ExtFields = extFieldMap
	}
}
