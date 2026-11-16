package internal

import "time"

type Config struct {
	Enabled                      bool
	ProcRoot                     string
	ConntrackInitTimeout         time.Duration
	ConntrackRateLimit           int
	ConntrackMaxStateSize        int
	EnableConntrackAllNamespaces bool
}
