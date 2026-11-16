// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package dubbo

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailDubboRequest() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 16 || message.Data[0] != MagicHigh || message.Data[1] != MagicLow
	}
}

func parseDubboRequest() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		contentKey := getContentKey(message.Data)
		if contentKey == "" {
			return false, true
		}

		message.AddStringAttribute(constlabels.ContentKey, contentKey)
		return true, true
	}
}

func getContentKey(requestData []byte) string {
	serialID := requestData[2] & SerialMask
	if serialID == Zero {
		return ""
	}
	if (requestData[2] & FlagEvent) != Zero {
		return "Heartbeat"
	}
	if (requestData[2] & FlagRequest) == Zero {
		// Invalid Data
		return ""
	}
	if (requestData[2] & FlagTwoWay) == Zero {
		// Ignore Oneway Data
		return "Oneway"
	}

	serializer := GetSerializer(serialID)
	if serializer == serialUnsupport {
		// Unsupport Serial. only support hessian and fastjson.
		return "UnSupportSerialFormat"
	}

	var (
		service string
		method  string
	)
	// version
	offset := serializer.eatString(requestData, 16)

	// service name
	offset, service = serializer.getStringValue(requestData, offset)

	// service version
	offset = serializer.eatString(requestData, offset)

	// method name
	_, method = serializer.getStringValue(requestData, offset)

	return service + "#" + method
}
