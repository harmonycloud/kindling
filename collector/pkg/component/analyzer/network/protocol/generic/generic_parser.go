// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package generic

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
)

func NewGenericParser() *protocol.ProtocolParser {
	requestParser := protocol.CreatePkgParser(fastfailGeneric(), parseGeneric())
	responseParser := protocol.CreatePkgParser(fastfailGeneric(), parseGeneric())

	return protocol.NewProtocolParser(protocol.NOSUPPORT, requestParser, responseParser, nil)
}

func fastfailGeneric() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return false
	}
}

func parseGeneric() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		return true, true
	}
}
