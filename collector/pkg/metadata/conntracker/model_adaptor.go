// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package conntracker

import "net"

type IPTranslation struct {
	ReplSrcIP   net.IP
	ReplDstIP   net.IP
	ReplSrcPort uint16
	ReplDstPort uint16
}
