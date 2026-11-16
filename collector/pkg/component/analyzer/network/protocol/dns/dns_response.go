// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package dns

import (
	"net"
	"strings"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

const (
	TypeA    uint16 = 1
	TypeAAAA uint16 = 28
)

func fastfailDnsResponse() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data) <= DNSHeaderSize
	}
}

func parseTcpDnsResponse(ignoreDnsRcode3Error bool) protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		message.Offset += 2
		return parseDnsResponse(message, ignoreDnsRcode3Error)
	}
}

func parseUdpDnsResponse(ignoreDnsRcode3Error bool) protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		return parseDnsResponse(message, ignoreDnsRcode3Error)
	}
}

/*
Header format

	  0  1  2  3  4  5  6  7  8  9  A  B  C  D  E  F
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	|                      ID                       |
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	|QR|   Opcode  |AA|TC|RD|RA|   Z    |   RCODE   |
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	|                    QDCOUNT                    |
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	|                    ANCOUNT                    |
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	|                    NSCOUNT                    |
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
	|                    ARCOUNT                    |
	+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+--+
*/
func parseDnsResponse(message *protocol.PayloadMessage, ignoreDnsRcode3Error bool) (bool, bool) {
	offset := message.Offset
	id, _ := message.ReadUInt16(offset)
	flags, _ := message.ReadUInt16(offset + 2)
	rcode := flags & 0xf

	numOfQuestions, _ := message.ReadUInt16(offset + 4)
	numOfAnswers, _ := message.ReadUInt16(offset + 6)

	if numOfQuestions == 0 {
		return false, true
	}

	domain, err := readQuery(message, numOfQuestions)
	if err != nil {
		return false, true
	}

	ip := readIpV4Answer(message, numOfAnswers)

	message.AddStringAttribute(constlabels.DnsDomain, domain)
	if len(ip) > 0 {
		message.AddStringAttribute(constlabels.DnsIp, ip)
	}
	message.AddIntAttribute(constlabels.DnsId, int64(id))
	message.AddIntAttribute(constlabels.DnsRcode, int64(rcode))
	if (rcode > 0 && rcode != 3) || (rcode == 3 && !ignoreDnsRcode3Error) {
		message.AddBoolAttribute(constlabels.IsError, true)
		message.AddIntAttribute(constlabels.ErrorType, int64(constlabels.ProtocolError))
	}
	return true, true
}

func readIpV4Answer(message *protocol.PayloadMessage, answerCount uint16) string {
	var (
		aType  uint16
		length uint16
		ip     net.IP
		ips    []string
		err    error
	)

	ips = make([]string, 0)
	offset := message.Offset
	for i := 0; i < int(answerCount); i++ {
		/*
			uint16 name
			uint16 type
			uint16 class
			uint32 ttl
			uint16 rdlength
			string rdata
		*/
		offset += 2
		aType, err = message.ReadUInt16(offset)
		if err != nil {
			break
		}

		offset += 8
		length, err = message.ReadUInt16(offset)
		if err != nil {
			break
		}

		offset += 2
		if aType == TypeA {
			offset, ip, err = message.ReadBytes(offset, int(length))
			if err != nil {
				break
			}
			ips = append(ips, ip.String())
		}
		offset += int(length)
	}
	message.Offset = offset
	if len(ips) == 0 {
		return ""
	}

	return strings.Join(ips, ",")
}
