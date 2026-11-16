// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package conntracker

import "time"

type Config struct {
	Enabled                      bool
	ProcRoot                     string
	ConntrackInitTimeout         time.Duration
	ConntrackRateLimit           int
	ConntrackMaxStateSize        int
	EnableConntrackAllNamespaces bool
}

var DefaultConfig = Config{
	Enabled:                      true,
	ProcRoot:                     "/proc",
	ConntrackInitTimeout:         30 * time.Second,
	ConntrackRateLimit:           500,
	ConntrackMaxStateSize:        130000,
	EnableConntrackAllNamespaces: true,
}
