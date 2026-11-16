// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"strings"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/tools"
	"github.com/Kindling-project/kindling/collector/pkg/urlclustering"

	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

func fastfailHttpRequest() protocol.FastFailFn {
	return func(message *protocol.PayloadMessage) bool {
		return len(message.Data[message.Offset:]) < 14
	}
}

/*
Request line

	    Method [GET/POST/PUT/DELETE/HEAD/TRACE/OPTIONS/CONNECT]
		Blank
		Request-URI [eg. /xxx/yyy?parm0=aaa&param1=bbb]
		Blank
		HTTP-Version [HTTP/1.0 | HTTP/1.2]
		\r\n

Request header
Request body
*/
func parseHttpRequest(urlClusteringMethod urlclustering.ClusteringMethod) protocol.ParsePkgFn {
	return func(message *protocol.PayloadMessage) (bool, bool) {
		offset, method := message.ReadUntilBlankWithLength(message.Offset, 8)

		if !httpMethodsList[string(method)] {
			if message.Data[offset-1] != ' ' || message.Data[offset] != '/' {
				return false, true
			}
			// FIX ET /xxx Data with split payload.
			if replaceMethod, ok := splitMethodsList[string(method)]; ok {
				method = replaceMethod
			} else {
				return false, true
			}
		}

		_, url := message.ReadUntilBlank(offset)

		headers := parseHeaders(message)
		traceType, traceId := tools.ParseTraceHeader(headers)
		if len(traceType) > 0 && len(traceId) > 0 {
			message.AddStringAttribute(constlabels.HttpApmTraceType, traceType)
			message.AddStringAttribute(constlabels.HttpApmTraceId, traceId)
		}

		message.AddStringAttribute(constlabels.HttpMethod, string(method))
		message.AddByteArrayUtf8Attribute(constlabels.HttpUrl, url)

		contentKey := urlClusteringMethod.Clustering(string(url))
		if len(contentKey) == 0 {
			contentKey = "*"
		}
		message.AddUtf8StringAttribute(constlabels.ContentKey, contentKey)
		return true, true
	}
}

func getContentKey(url string) string {
	if url == "" {
		return ""
	}
	index := strings.Index(url, "?")
	if index != -1 {
		url = url[:index]
	}
	return url
}

var httpMethodsList = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"HEAD":    true,
	"TRACE":   true,
	"OPTIONS": true,
	"CONNECT": true,
}

var splitMethodsList = map[string][]byte{
	"ET": {'G', 'E', 'T'},
}
