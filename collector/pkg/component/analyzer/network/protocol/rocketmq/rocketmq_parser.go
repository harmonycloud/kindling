// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package rocketmq

import "github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"

func NewRocketMQParser() *protocol.ProtocolParser {
	requestParser := protocol.CreatePkgParser(fastfailRocketMQRequest(), parseRocketMQRequest())
	responseParser := protocol.CreatePkgParser(fastfailRocketMQResponse(), parseRocketMQResponse())

	return protocol.NewProtocolParser(protocol.ROCKETMQ, requestParser, responseParser, nil)
}
