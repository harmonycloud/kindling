// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package dubbo

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailDubboResponse() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) < 16 || message.Data[0] != MagicHigh || message.Data[1] != MagicLow
	}
}

func parseDubboResponse() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		errorCode := getErrorCode(message.Data)
		if errorCode == -1 {
			return false, true
		}

		message.AddIntAttribute(constlabels.DubboErrorCode, errorCode)
		if errorCode > 20 {
			message.AddBoolAttribute(constlabels.IsError, true)
			message.AddIntAttribute(constlabels.ErrorType, int64(constlabels.ProtocolError))
		}
		return true, true
	}
}

func getErrorCode(responseData []byte) int64 {
	SerialID := responseData[2] & SerialMask
	if SerialID == Zero {
		return -1
	}
	if (responseData[2] & FlagEvent) != Zero {
		return 20
	}
	if (responseData[2] & FlagRequest) != Zero {
		// Invalid Data
		return -1
	}

	return int64(responseData[3])
}
