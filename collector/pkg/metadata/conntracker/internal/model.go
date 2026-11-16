package internal

import (
	"net"
)

// ConnectionType will be either TCP or UDP
type ConnectionType uint8

const (
	// TCP connection type
	TCP ConnectionType = 0

	// UDP connection type
	UDP ConnectionType = 1
)

func (c ConnectionType) String() string {
	if c == TCP {
		return "TCP"
	}
	return "UDP"
}

type IPTranslation struct {
	ReplSrcIP   net.IP
	ReplDstIP   net.IP
	ReplSrcPort uint16
	ReplDstPort uint16
}

// ConnectionStats stores statistics for a single connection.  Field order in the struct should be 8-byte aligned
type ConnectionStats struct {
	Source net.IP
	Dest   net.IP

	SPort uint16
	DPort uint16
	Type  ConnectionType
}
