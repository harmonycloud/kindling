// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"encoding/base64"
	"strings"
)

func ParseTraceHeader(headers map[string]string) (string, string) {
	if skywalking, ok := headers["sw8"]; ok {
		traceId := parseSkyWalkingTraceId(skywalking)
		return "skywalking", traceId
	}

	if harmonycloud, ok := headers["apm-transactionid"]; ok {
		return "harmonycloud", harmonycloud
	}

	if zipkin, ok := headers["x-b3-traceid"]; ok {
		return "zipkin", zipkin
	}

	if jaeger, ok := headers["uber-trace-id"]; ok {
		if pos := strings.Index(jaeger, ":"); pos > 0 {
			return "jaeger", jaeger[0:pos]
		}
		return "jaeger", jaeger
	}

	if w3c, ok := headers["traceparent"]; ok && len(w3c) >= 35 {
		return "w3c", w3c[3:35]
	}
	if w3c, ok := headers["traceresponse"]; ok && len(w3c) >= 35 {
		return "w3c", w3c[3:35]
	}

	return "", ""
}

// See the doc
// https://github.com/apache/skywalking/blob/master/docs/en/protocols/Skywalking-Cross-Process-Propagation-Headers-Protocol-v3.md
func parseSkyWalkingTraceId(value string) string {
	if len(value) < 3 {
		return ""
	}
	valueBytes := []byte(value)
	traceIdBase64 := make([]byte, 0)
	// We traverse the string from the first '-'.
	for i := 2; i < len(valueBytes); i++ {
		if valueBytes[i] == '-' {
			break
		}
		traceIdBase64 = append(traceIdBase64, valueBytes[i])
	}
	traceId, err := base64.StdEncoding.DecodeString(string(traceIdBase64))
	if err != nil {
		return ""
	}
	return string(traceId)
}
