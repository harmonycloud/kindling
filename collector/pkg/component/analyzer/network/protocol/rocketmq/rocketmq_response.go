// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package rocketmq

import (
	"encoding/json"
	"fmt"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailRocketMQResponse() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 29
	}
}

func parseRocketMQResponse() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		var (
			payloadLength int32
			headerLength  int32
			serializeType uint8
		)

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

		if !message.HasAttribute(constlabels.RocketMQOpaque) ||
			message.GetIntAttribute(constlabels.RocketMQOpaque) != int64(header.Opaque) {

			return false, true
		}
		message.AddIntAttribute(constlabels.RocketMQErrCode, int64(header.Code))

		//add RocketMQErrMsg if responseCode > 0
		if header.Code > 0 {
			if _, ok := responseErrMsgMap[header.Code]; ok {
				message.AddStringAttribute(constlabels.RocketMQErrMsg, responseErrMsgMap[header.Code])
			} else if header.Remark != "" {
				message.AddStringAttribute(constlabels.RocketMQErrMsg, header.Remark)
			} else {
				message.AddStringAttribute(constlabels.RocketMQErrMsg, fmt.Sprintf("error:response code is %v", header.Code))
			}
			message.AddBoolAttribute(constlabels.IsError, true)
			message.AddIntAttribute(constlabels.ErrorType, int64(constlabels.ProtocolError))
		}

		return true, true
	}
}
