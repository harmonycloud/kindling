// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"path"
	"strconv"
)

func NewPidTcpStat(hostProc string, pid int) (NetSocketStateMap, error) {
	tcpFilePath := path.Join(hostProc, strconv.Itoa(pid), "net/tcp")
	return newNetIPSocket(tcpFilePath)
}
