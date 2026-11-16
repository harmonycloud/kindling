// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	viperpackage "github.com/spf13/viper"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol/factory"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer"
	"github.com/Kindling-project/kindling/collector/pkg/model"
)

func TestHttpProtocol(t *testing.T) {
	testProtocol(t, "http/server-event.yml",
		"http/server-trace-slow.yml",
		"http/server-trace-error.yml",
		"http/server-trace-split.yml",
		"http/server-trace-normal.yml",
		"http/server-trace-continue.yml",
	)
}

func TestMySqlProtocol(t *testing.T) {
	testProtocol(t, "mysql/server-event.yml",
		"mysql/server-trace-commit.yml",
		"mysql/server-trace-query-split.yml",
		"mysql/server-trace-query.yml",
		"mysql/server-trace-oneway.yml",
		"mysql/server-trace-query-cmd.yml",
	)
}

func TestRedisProtocol(t *testing.T) {
	testProtocol(t, "redis/server-event.yml",
		"redis/server-trace-get.yml")
}

func TestDnsProtocol(t *testing.T) {
	testProtocol(t, "dns/server-event.yml",
		"dns/server-trace.yml")
	testProtocol(t, "dns/server-event.yml",
		"dns/server-trace-multi.yml")
	testProtocol(t, "dns/client-event.yml",
		"dns/client-trace-sendmmg.yml")
	testProtocol(t, "dns/client-event.yml",
		"dns/client-trace-dns3.yml")
	testProtocol(t, "dns/client-event-tcp.yml",
		"dns/client-trace-tcp.yml")
}

func TestKafkaProtocol(t *testing.T) {
	testProtocol(t, "kafka/provider-event.yml",
		"kafka/provider-trace-produce-split.yml")

	testProtocol(t, "kafka/consumer-event.yml",
		"kafka/consumer-trace-fetch-split.yml",
		"kafka/consumer-trace-fetch-multi-topics.yml")
}

func TestDubboProtocol(t *testing.T) {
	testProtocol(t, "dubbo/server-event.yml",
		"dubbo/server-trace-short.yml")
}

func TestRocketMQProtocol(t *testing.T) {
	testProtocol(t, "rocketmq/server-event.yml",
		"rocketmq/server-trace-json.yml",
		"rocketmq/server-trace-rocketmq.yml",
		"rocketmq/server-trace-error.yml")
}

func TestNoSupportProtocol(t *testing.T) {
	testProtocol(t, "nosupport/server-event.yml",
		"nosupport/server-trace-normal.yml",
	)
}

type NopProcessor struct {
}

func (n NopProcessor) Consume(dataGroup *model.DataGroup) error {
	// fmt.Printf("Consume %v\n", dataGroup)
	results = append(results, dataGroup)
	return nil
}

type NoCacheDataGroupPool struct {
}

func (p *NoCacheDataGroupPool) Get() *model.DataGroup {
	dataGroup := createDataGroup()
	return dataGroup.(*model.DataGroup)
}

func (p *NoCacheDataGroupPool) Free(_ *model.DataGroup) {
}

var na *NetworkAnalyzer
var results []*model.DataGroup

func prepareNetworkAnalyzer() *NetworkAnalyzer {
	if na == nil {
		config := &Config{}
		viper := viperpackage.New()
		viper.SetConfigFile("protocol/testdata/na-protocol-config.yaml")
		err := viper.ReadInConfig()
		if err != nil {
			fmt.Printf("Read Config File failed%v\n", err)
			return nil
		}
		_ = viper.UnmarshalKey("analyzers.networkanalyzer", config)

		na = &NetworkAnalyzer{
			cfg:           config,
			dataGroupPool: &NoCacheDataGroupPool{},
			nextConsumers: []consumer.Consumer{&NopProcessor{}},
			telemetry:     component.NewDefaultTelemetryTools(),
		}
		na.staticPortMap = map[uint32]string{}
		for _, config := range na.cfg.ProtocolConfigs {
			for _, port := range config.Ports {
				na.staticPortMap[port] = config.Key
			}
		}
		na.slowThresholdMap = map[string]int{}
		for _, config := range na.cfg.ProtocolConfigs {
			protocol.SetPayLoadLength(config.Key, config.PayloadLength)
			na.slowThresholdMap[config.Key] = config.Threshold
		}
		na.parserFactory = factory.NewParserFactory(factory.WithUrlClusteringMethod(na.cfg.UrlClusteringMethod), factory.WithIgnoreDnsRcode3Error(na.cfg.IgnoreDnsRcode3Error))
		na.snaplen = 200
		// Do not start the timeout check otherwise the test maybe fail
		na.cfg.EnableTimeoutCheck = false
		_ = na.Start()
	}
	return na
}

func testProtocol(t *testing.T, eventYaml string, traceYamls ...string) {
	na := prepareNetworkAnalyzer()
	if na == nil {
		return
	}

	eventCommon := getEventCommon("protocol/testdata/" + eventYaml)
	if eventCommon == nil {
		t.Errorf("Parse %v Failed", eventYaml)
		return
	}

	for _, yaml := range traceYamls {
		trace := getTrace("protocol/testdata/" + yaml)
		if trace == nil {
			t.Errorf("Parse %v Failed", yaml)
			return
		}

		t.Run(trace.Key, func(t *testing.T) {
			results = []*model.DataGroup{}
			events := trace.getSortedEvents(eventCommon)
			for _, event := range events {
				_ = na.processEvent(event)
			}
			if model.L4Proto(eventCommon.Ctx.Fd.Protocol) == model.L4Proto_TCP {
				if pairInterface, ok := na.requestMonitor.Load(getMessagePairKey(events[0])); ok {
					var oldPairs = pairInterface.(*messagePairs)
					_ = na.distributeTraceMetric(oldPairs, nil)
				}
			}
			trace.Validate(t, results)
		})
	}
}

func getEventCommon(path string) *EventCommon {
	eventCommon := &EventCommon{}
	viper := viperpackage.New()
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error%v\n", err)
		return nil
	}
	_ = viper.UnmarshalKey("eventCommon", eventCommon)
	return eventCommon
}

func getTrace(path string) *Trace {
	trace := &Trace{}
	viper := viperpackage.New()
	viper.SetConfigFile(path)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error%v\n", err)
		return nil
	}
	_ = viper.UnmarshalKey("trace", trace)
	return trace
}

type EventCommon struct {
	Source   int `mapstructure:"source"`
	Category int `mapstructure:"category"`
	Ctx      Ctx `mapstructure:"ctx"`
}

type Ctx struct {
	Thread ThreadInfo `mapstructure:"thread_info"`
	Fd     FdInfo     `mapstructure:"fd_info"`
}

type ThreadInfo struct {
	Pid         uint32 `mapstructure:"pid"`
	Tid         uint32 `mapstructure:"tid"`
	Uid         uint32 `mapstructure:"uid"`
	Gid         uint32 `mapstructure:"gid"`
	Comm        string `mapstructure:"comm"`
	ContainerId string `mapstructure:"container_id"`
}

type FdInfo struct {
	Num      int32    `mapstructure:"num"`
	TypeFd   int32    `mapstructure:"type_fd"`
	Protocol uint32   `mapstructure:"protocol"`
	Role     bool     `mapstructure:"role"`
	Sip      []uint32 `mapstructure:"sip"`
	Dip      []uint32 `mapstructure:"dip"`
	Sport    uint32   `mapstructure:"sport"`
	Dport    uint32   `mapstructure:"dport"`
}

type Trace struct {
	Key       string        `mapstructure:"key"`
	Connects  []TraceEvent  `mapstructure:"connects"`
	Requests  []TraceEvent  `mapstructure:"requests"`
	Responses []TraceEvent  `mapstructure:"responses"`
	Expects   []TraceExpect `mapstructure:"expects"`
}

func (trace *Trace) prepareMessagePairs(common *EventCommon) *messagePairs {
	mps := &messagePairs{
		connects:  nil,
		requests:  nil,
		responses: nil,
		mutex:     sync.RWMutex{},
	}
	if trace.Connects != nil {
		for _, connect := range trace.Connects {
			mps.mergeConnect(connect.exchange(common))
		}
	}
	if trace.Requests != nil {
		for _, request := range trace.Requests {
			mps.mergeRequest(request.exchange(common))
		}
	}
	if trace.Responses != nil {
		for _, response := range trace.Responses {
			mps.mergeResponse(response.exchange(common))
		}
	}
	return mps
}

func (trace *Trace) Validate(t *testing.T, results []*model.DataGroup) {
	checkSize(t, "Expect Size", len(trace.Expects), len(results))

	for i, result := range results {
		expect := trace.Expects[i]
		checkUint64Equal(t, "Timestamp", expect.Timestamp, result.Timestamp)

		// Validate Metrics Metrics
		checkSize(t, "Metrics Size", len(expect.Values), len(result.Metrics))
		for _, value := range result.Metrics {
			expectValue, ok := expect.Values[value.Name]
			if !ok {
				t.Errorf("[Miss %s] want=nil, got=%d", value.Name, value.GetInt().Value)
			} else {
				checkInt64Equal(t, value.Name, expectValue, value.GetInt().Value)
			}
		}

		// Validate Metrics Attributes
		checkSize(t, "Labels Size", len(expect.Labels), result.Labels.Size())
		for labelKey, labelValue := range expect.Labels {
			if reflect.TypeOf(labelValue).Name() == "int" {
				gotValue := result.Labels.GetIntValue(labelKey)
				checkInt64Equal(t, labelKey, int64(labelValue.(int)), gotValue)
			} else if reflect.TypeOf(labelValue).Name() == "bool" {
				gotValue := result.Labels.GetBoolValue(labelKey)
				checkBoolEqual(t, labelKey, labelValue.(bool), gotValue)
			} else {
				gotValue := result.Labels.GetStringValue(labelKey)
				checkStringEqual(t, labelKey, labelValue.(string), gotValue)
			}
		}

		for labelKey, labelValue := range result.Labels.ToStringMap() {
			if _, ok := expect.Labels[labelKey]; !ok {
				t.Errorf("[Miss %s] want=nil, got=%s", labelKey, labelValue)
			}
		}
	}
}

func checkBoolEqual(t *testing.T, key string, expect bool, got bool) {
	if expect != got {
		t.Errorf("[Check %s] want=%t, got=%t", key, expect, got)
	}
}

func checkStringEqual(t *testing.T, key string, expect string, got string) {
	if expect != got {
		t.Errorf("[Check %s] want=%s, got=%s", key, expect, got)
	}
}

func checkUint64Equal(t *testing.T, key string, expect uint64, got uint64) {
	if expect != got {
		t.Errorf("[Check %s] want=%d, got=%d", key, expect, got)
	}
}

func checkInt64Equal(t *testing.T, key string, expect int64, got int64) {
	if expect != got {
		t.Errorf("[Check %s] want=%d, got=%d", key, expect, got)
	}
}

func checkSize(t *testing.T, key string, expect int, got int) {
	if expect != got {
		t.Errorf("[Check %s] want=%d, got=%d", key, expect, got)
	}
}

func (trace *Trace) getSortedEvents(common *EventCommon) []*model.KindlingEvent {
	events := make([]*model.KindlingEvent, 0)
	if trace.Connects != nil {
		for _, connect := range trace.Connects {
			events = append(events, connect.exchange(common))
		}
	}
	if trace.Requests != nil {
		for _, request := range trace.Requests {
			events = append(events, request.exchange(common))
		}
	}
	if trace.Responses != nil {
		for _, response := range trace.Responses {
			events = append(events, response.exchange(common))
		}
	}
	// Sort By Event Timestamp.
	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Timestamp < events[j].Timestamp
	})
	return events
}

type TraceEvent struct {
	Name           string         `mapstructure:"name"`
	Timestamp      uint64         `mapstructure:"timestamp"`
	UserAttributes UserAttributes `mapstructure:"user_attributes"`
}

func (evt *TraceEvent) exchange(common *EventCommon) *model.KindlingEvent {
	byteData, err := getData(evt.UserAttributes.Data)
	if err != nil {
		fmt.Printf("%s\n", err)
		return nil
	}

	modelEvt := &model.KindlingEvent{
		Source:       model.Source(common.Source),
		Timestamp:    evt.Timestamp,
		Latency:      uint64(evt.UserAttributes.Latency),
		Name:         evt.Name,
		Category:     model.Category(common.Category),
		ParamsNumber: 3,
		UserAttributes: [16]model.KeyValue{
			{Key: "res", ValueType: model.ValueType_INT64, Value: Int64ToBytes(evt.UserAttributes.Res)},
			{Key: "data", ValueType: model.ValueType_BYTEBUF, Value: byteData},
		},
		Ctx: model.Context{
			ThreadInfo: model.Thread{
				Pid:  common.Ctx.Thread.Pid,
				Tid:  common.Ctx.Thread.Tid,
				Uid:  common.Ctx.Thread.Uid,
				Gid:  common.Ctx.Thread.Gid,
				Comm: common.Ctx.Thread.Comm,
			},
			FdInfo: model.Fd{
				Num:      common.Ctx.Fd.Num,
				TypeFd:   model.FDType(common.Ctx.Fd.TypeFd),
				Protocol: model.L4Proto(common.Ctx.Fd.Protocol),
				Role:     common.Ctx.Fd.Role,
				Sip:      common.Ctx.Fd.Sip,
				Dip:      common.Ctx.Fd.Dip,
				Sport:    common.Ctx.Fd.Sport,
				Dport:    common.Ctx.Fd.Dport,
			},
		},
	}
	return modelEvt
}

// getData converts the following format to byte array.
//
// There are the following formats supported:
//  1. {hex number}|{string}
//     The first part is a number in hexadecimal which is part of the original data.
//     It holds different meanings in different protocols.
//  2. (hex)|{hex value}
//     The first part is the constant "hex" and the second part is its value
//  3. (string)|{string value}
//     The first part is the constant "string" and the second part is its value
//  4. {string value}
//     If there are no the separator "|" existing, the data is considered as a string
//
// See the files under the "testdata" directory for how to write your data.
func getData(datas []string) ([]byte, error) {
	dataBytes := make([]byte, 0)
	for _, data := range datas {
		if len(data) <= 0 {
			continue
		}
		splitIndex := getSplitIndex(data)
		// If no separator exists, the data is a string
		if splitIndex == 0 {
			byteArray := []byte(data)
			dataBytes = append(dataBytes, byteArray...)
			continue
		}
		// If there is a separator.
		prefix := strings.TrimSpace(data[0:splitIndex])
		suffix := strings.TrimSpace(data[splitIndex+1:])
		switch prefix {
		case "hex":
			hexArray, err := hex.DecodeString(suffix)
			if err != nil {
				return []byte{}, fmt.Errorf("the second part is not a hexadecimal number: %w", err)
			}
			dataBytes = append(dataBytes, hexArray...)
		case "string":
			byteArray := []byte(suffix)
			dataBytes = append(dataBytes, byteArray...)
		default:
			// The first part should be a hexadecimal number
			hexArray, err := hex.DecodeString(prefix)
			if err != nil {
				return []byte{}, fmt.Errorf("the first part of data is not correct: %w", err)
			}
			dataBytes = append(dataBytes, hexArray...)
			dataBytes = append(dataBytes, suffix...)
		}
	}
	return dataBytes, nil
}

func getSplitIndex(data string) int {
	index := strings.Index(data, "|")
	if index == -1 {
		return 0
	}
	return index
}

type UserAttributes struct {
	Latency int64    `mapstructure:"latency"`
	Res     int64    `mapstructure:"res"`
	Data    []string `mapstructure:"data"`
}

func Int64ToBytes(value int64) []byte {
	var buf = make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	return buf
}

type TraceExpect struct {
	Timestamp uint64                 `mapstructure:"Timestamp"`
	Values    map[string]int64       `mapstructure:"Values"`
	Labels    map[string]interface{} `mapstructure:"Labels"`
}
