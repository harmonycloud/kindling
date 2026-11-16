// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package factory

import (
	"sync"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol/rocketmq"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol/dns"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol/dubbo"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol/generic"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol/http"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol/kafka"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol/mysql"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol/redis"
)

type ParserFactory struct {
	cachePortParsersMap map[uint32][]*protocol.ProtocolParser
	mutex               sync.Mutex
	protocolParsers     map[string]*protocol.ProtocolParser
	udpDnsParser        *protocol.ProtocolParser

	config *config
}

func NewParserFactory(options ...Option) *ParserFactory {
	factory := &ParserFactory{
		cachePortParsersMap: make(map[uint32][]*protocol.ProtocolParser),
		protocolParsers:     make(map[string]*protocol.ProtocolParser),
		config:              newDefaultConfig(),
	}
	for _, option := range options {
		option(factory.config)
	}
	factory.protocolParsers[protocol.HTTP] = http.NewHttpParser(factory.config.urlClusteringMethod)
	factory.protocolParsers[protocol.KAFKA] = kafka.NewKafkaParser()
	factory.protocolParsers[protocol.MYSQL] = mysql.NewMysqlParser()
	factory.protocolParsers[protocol.REDIS] = redis.NewRedisParser()
	factory.protocolParsers[protocol.DUBBO] = dubbo.NewDubboParser()
	factory.protocolParsers[protocol.DNS] = dns.NewTcpDnsParser(factory.config.ignoreDnsRcode3Error)
	factory.protocolParsers[protocol.ROCKETMQ] = rocketmq.NewRocketMQParser()
	factory.protocolParsers[protocol.NOSUPPORT] = generic.NewGenericParser()

	factory.udpDnsParser = dns.NewUdpDnsParser(factory.config.ignoreDnsRcode3Error)
	return factory
}

func (f *ParserFactory) GetUdpDnsParser() *protocol.ProtocolParser {
	return f.udpDnsParser
}

func (f *ParserFactory) GetParser(key string) *protocol.ProtocolParser {
	return f.protocolParsers[key]
}

func (f *ParserFactory) GetGenericParser() *protocol.ProtocolParser {
	return f.protocolParsers[protocol.NOSUPPORT]
}

func (f *ParserFactory) GetCachedParsersByPort(port uint32) ([]*protocol.ProtocolParser, bool) {
	f.mutex.Lock()
	parser, ok := f.cachePortParsersMap[port]
	f.mutex.Unlock()

	return parser, ok
}

func (f *ParserFactory) AddCachedParser(port uint32, parser *protocol.ProtocolParser) {
	f.mutex.Lock()
	if val := f.cachePortParsersMap[port]; val == nil {
		parsers := make([]*protocol.ProtocolParser, 0)
		parsers = append(parsers, parser)
		f.cachePortParsersMap[port] = parsers
	} else {
		exist := false
		for _, value := range val {
			if value == parser {
				exist = true
				break
			}
		}
		genericParser := f.GetGenericParser()
		if !exist {
			// Make sure Generic is last
			if len(val) > 0 && val[len(val)-1] == genericParser {
				parsers := append(val[0:len(val)-1], parser)
				parsers = append(parsers, genericParser)
				f.cachePortParsersMap[port] = parsers
			} else {
				parsers := append(val, parser)
				f.cachePortParsersMap[port] = parsers
			}
		}
	}
	f.mutex.Unlock()
}

func (f *ParserFactory) RemoveCachedParser(port uint32, parser *protocol.ProtocolParser) {
	f.mutex.Lock()
	if val, ok := f.cachePortParsersMap[port]; ok {
		for i, value := range val {
			if value == parser {
				val = append(val[:i], val[i+1:]...)
				f.cachePortParsersMap[port] = val
			}
		}
	}
	f.mutex.Unlock()
}
