// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package conntracker

type Conntracker interface {
	GetDNATTupleWithString(srcIP string, dstIP string, srcPort uint16, dstPort uint16, isUdp uint32) *IPTranslation
	GetDNATTuple(srcIP uint32, dstIP uint32, srcPort uint16, dstPort uint16, isUdp uint32) *IPTranslation
	GetStats() map[string]int64
}

type NoopConntracker struct {
	cfg   *Config
	stats map[string]int64
}

func NewNoopConntracker(cfg *Config) *NoopConntracker {
	return &NoopConntracker{
		cfg:   cfg,
		stats: make(map[string]int64),
	}
}

func (ctr *NoopConntracker) GetDNATTupleWithString(_ string, _ string, _ uint16, _ uint16, _ uint32) *IPTranslation {
	return nil
}

func (ctr *NoopConntracker) GetDNATTuple(_ uint32, _ uint32, _ uint16, _ uint16, _ uint32) *IPTranslation {
	return nil
}

func (ctr *NoopConntracker) GetStats() map[string]int64 {
	return ctr.stats
}
