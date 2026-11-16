// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package dubbo

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
)

const (
	// Zero : byte zero
	Zero = byte(0x00)

	// magic header
	MagicHigh = byte(0xda)
	MagicLow  = byte(0xbb)

	// message flag.
	FlagRequest = byte(0x80)
	FlagTwoWay  = byte(0x40)
	FlagEvent   = byte(0x20) // for heartbeat
	SerialMask  = 0x1f
)

func NewDubboParser() *protocol.ProtocolParser {
	requestParser := protocol.CreatePkgParser(fastfailDubboRequest(), parseDubboRequest())
	responseParser := protocol.CreatePkgParser(fastfailDubboResponse(), parseDubboResponse())
	return protocol.NewProtocolParser(protocol.DUBBO, requestParser, responseParser, nil)
}
