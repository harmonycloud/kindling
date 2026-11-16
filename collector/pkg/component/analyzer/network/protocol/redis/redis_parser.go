// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
)

func NewRedisParser() *protocol.ProtocolParser {
	requestParser := protocol.CreatePkgParser(fastfailRedisRequest(), parseRedisRequest())
	requestParser.Add(fastfailRedisArray(), parseRedisArray())
	requestParser.Add(fastfailRedisBulkString(), parseRedisBulkString())
	requestParser.Add(fastfailRedisInteger(), parseRedisInteger())

	responseParser := protocol.CreatePkgParser(fastfailResponse(), parseResponse())
	responseParser.Add(fastfailRedisArray(), parseRedisArray())
	responseParser.Add(fastfailRedisBulkString(), parseRedisBulkString())
	responseParser.Add(fastfailRedisInteger(), parseRedisInteger())
	responseParser.Add(fastfailRedisSimpleString(), parseRedisSimpleString())
	responseParser.Add(fastfailRedisError(), parseRedisError())

	redisParser := protocol.NewProtocolParser(protocol.REDIS, requestParser, responseParser, nil)
	redisParser.EnableMultiFrame()
	return redisParser
}
