// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package protocol

const (
	HTTP      = "http"
	DNS       = "dns"
	KAFKA     = "kafka"
	MYSQL     = "mysql"
	REDIS     = "redis"
	DUBBO     = "dubbo"
	ROCKETMQ  = "rocketmq"
	NOSUPPORT = "NOSUPPORT"
)

var payloadLength map[string]int = map[string]int{}

func SetPayLoadLength(protocol string, length int) {
	if length > 0 {
		payloadLength[protocol] = length
	} else {
		payloadLength[protocol] = 200
	}
}

func GetPayLoadLength(protocol string) int {
	if length, ok := payloadLength[protocol]; ok {
		return length
	}
	return 200
}
