// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kafka

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailResponseOther() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return message.GetIntAttribute(constlabels.KafkaApi) <= _apiFetch
	}
}

func parseResponseOther() protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		return true, true
	}
}
