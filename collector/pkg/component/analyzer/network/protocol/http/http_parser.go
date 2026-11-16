// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"strings"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/urlclustering"
)

func NewHttpParser(urlClusteringMethod string) *protocol.ProtocolParser {
	method := urlclustering.NewMethod(urlClusteringMethod)
	requestParser := protocol.CreatePkgParser(fastfailHttpRequest(), parseHttpRequest(method))
	responseParser := protocol.CreatePkgParser(fastfailHttpResponse(), parseHttpResponse())

	return protocol.NewProtocolParser(protocol.HTTP, requestParser, responseParser, nil)
}

/*
Requet-Line\r\n
Key:Value\r\n
...
Key:Value\r\n
\r\n
Data
*/
func parseHeaders(message *protocol.PayloadMessage) map[string]string {
	header := make(map[string]string)

	from, data := message.ReadUntilCRLF(0)
	if data == nil {
		return header
	}
	for {
		from, data = message.ReadUntilCRLF(from)
		if data == nil {
			return header
		}
		if position := strings.Index(string(data), ":"); position > 0 && position < len(data)-2 {
			header[strings.ToLower(string(data[0:position]))] = string(data[position+2:])
			continue
		}
		return header
	}
}
