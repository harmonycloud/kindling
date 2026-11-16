// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model"
)

type benchCase struct {
	protocol   string
	commonFile string
	dataFile   string
}

const (
	BENCH_CASE_HTTP           = "http"
	BENCH_CASE_MYSQL          = "mysql"
	BENCH_CASE_REDIS          = "redis"
	BENCH_CASE_DNS            = "dns"
	BENCH_CASE_KAFKA_PRODUCER = "kafka_producer"
	BENCH_CASE_KAFKA_FETCHER  = "kafka_fetcher"
	BENCH_CASE_DUBBO          = "dubbo"
	BENCH_CASE_ROCKETMQ       = "rocketmq"
)

var benchCaseMap = map[string]benchCase{
	BENCH_CASE_HTTP:           {protocol.HTTP, "http/server-event.yml", "http/1k-trace.yml"},
	BENCH_CASE_MYSQL:          {protocol.MYSQL, "mysql/server-event.yml", "mysql/1k-trace.yml"},
	BENCH_CASE_REDIS:          {protocol.REDIS, "redis/server-event.yml", "redis/1k-trace.yml"},
	BENCH_CASE_DNS:            {protocol.DNS, "dns/server-event.yml", "dns/1k-trace.yml"},
	BENCH_CASE_KAFKA_PRODUCER: {protocol.KAFKA, "kafka/provider-event.yml", "kafka/1k-provider-trace.yml"},
	BENCH_CASE_KAFKA_FETCHER:  {protocol.KAFKA, "kafka/consumer-event.yml", "kafka/1k-consumer-trace.yml"},
	BENCH_CASE_DUBBO:          {protocol.DUBBO, "dubbo/server-event.yml", "dubbo/1k-trace.yml"},
	BENCH_CASE_ROCKETMQ:       {protocol.ROCKETMQ, "rocketmq/server-event.yml", "rocketmq/1k-trace.yml"},
}

const (
	SIZE_MESSAGE_PAIR int = 500
)

func BenchmarkHttp(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_HTTP)
}

func BenchmarkMySql(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_MYSQL)
}

func BenchmarkRedis(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_REDIS)
}

func BenchmarkDns(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_DNS)
}

func BenchmarkKafkaProducer(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_KAFKA_PRODUCER)
}

func BenchmarkKafkaFetcher(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_KAFKA_FETCHER)
}

func BenchmarkDubo(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_DUBBO)
}

func BenchmarkRocketmq(b *testing.B) {
	testProtocolBench(b, b.N, SIZE_MESSAGE_PAIR, BENCH_CASE_ROCKETMQ)
}

func testProtocolBench(b *testing.B, tps int, mpSize int, caseKey string) {
	na := prepareNetworkAnalyzer()
	if na == nil {
		return
	}

	benchCase := benchCaseMap[caseKey]
	eventCommon := getEventCommon("protocol/testdata/" + benchCase.commonFile)
	if eventCommon == nil {
		b.Errorf("Parse %v Failed", benchCase.commonFile)
		return
	}

	trace := getTrace("protocol/testdata/" + benchCase.dataFile)
	if trace == nil {
		b.Errorf("Parse %v Failed", benchCase.dataFile)
		return
	}

	// Prepare Base Events
	evts := prepareEvents(mpSize, eventCommon, trace)
	// Produce Event
	for i := 0; i < tps; i++ {
		na.ConsumeEvent(evts[i%mpSize])
	}
}

func prepareEvents(mpSize int, eventCommon *EventCommon, trace *Trace) []*model.KindlingEvent {
	baseEvents := make([]*model.KindlingEvent, 0)
	for _, request := range trace.Requests {
		baseEvents = append(baseEvents, request.exchange(eventCommon))

	}
	for _, response := range trace.Responses {
		baseEvents = append(baseEvents, response.exchange(eventCommon))
	}

	size := len(baseEvents)
	events := make([]*model.KindlingEvent, size*mpSize)
	for i := 0; i < mpSize; i++ {
		for j, basebaseEvent := range baseEvents {
			events[i*size+j] = prepareEvent(basebaseEvent, i+1)
		}
	}
	return events
}

func prepareEvent(evt *model.KindlingEvent, fdNum int) *model.KindlingEvent {
	newEvt := &model.KindlingEvent{
		Source:         evt.Source,
		Timestamp:      evt.Timestamp,
		Name:           evt.Name,
		Category:       evt.Category,
		UserAttributes: evt.UserAttributes,
		Ctx: model.Context{
			ThreadInfo: evt.Ctx.ThreadInfo,
			FdInfo: model.Fd{
				Num:      int32(fdNum),
				TypeFd:   evt.Ctx.FdInfo.TypeFd,
				Protocol: evt.Ctx.FdInfo.Protocol,
				Role:     evt.Ctx.FdInfo.Role,
				Sip:      evt.Ctx.FdInfo.Sip,
				Dip:      evt.Ctx.FdInfo.Dip,
				Sport:    evt.Ctx.FdInfo.Sport,
				Dport:    evt.Ctx.FdInfo.Dport,
			},
		},
	}
	return newEvt
}
