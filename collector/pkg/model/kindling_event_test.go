// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"encoding/base64"
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalEvent(t *testing.T) {
	v, _ := base64.StdEncoding.DecodeString("sd4btBwgHxc=")
	att1 := &KeyValue{
		Key:       "start_time",
		ValueType: 8,
		Value:     v,
	}
	assert.Equal(t, uint64(1666085694803271345), att1.GetUintValue())
	v, _ = base64.StdEncoding.DecodeString("19uyLHcgHxc=")
	att2 := &KeyValue{
		Key:       "end_time",
		ValueType: 8,
		Value:     v,
	}
	assert.Equal(t, uint64(1666086083373489111), att2.GetUintValue())
}

func TestTextEvent(t *testing.T) {

	v, _ := base64.StdEncoding.DecodeString("sd4btBwgHxc=")

	evt := &KindlingEvent{
		Source:       Source_TRACEPOINT,
		Timestamp:    160000,
		Name:         constnames.ReadEvent,
		Category:     Category_CAT_NET,
		ParamsNumber: 3,
		Latency:      100,
		UserAttributes: [16]KeyValue{{Key: "starttime", ValueType: ValueType_UINT64, Value: v}, {Key: "data", ValueType: ValueType_BYTEBUF, Value: []byte(`POST /test?sleep=800 HTTP/1.1
		Host: 10.10.103.148:9000
		Content-Type: application/x-www-form-urlencoded 
		xxxxjkajdnaksndandlkajelkqjweklqwjeklqjlwejklasdas
		qwelkqnmwlkenqlwenqwnekqkwelkqmnwlekqlwnelknqewlal
		naknslkdnalksndlkanlskndlaksdasdasdasdasdasdasdasd
		`)}},
		Ctx: Context{ThreadInfo: Thread{Pid: 22592, Tid: 22592, Uid: 0, Gid: 0, Comm: "node", ContainerId: "123123", ContainerName: "test"}, FdInfo: Fd{Num: 19, TypeFd: 3, Filename: "tmp", Directory: "/root", Protocol: 1, Role: true, Sip: []uint32{16777343, 16777343}, Dip: []uint32{16777343}, Sport: 39620, Dport: 37093, Source: 0, Destination: 0}},
	}
	assert.Equal(t,
		`{Name:read Category:3 ParamsNumber:3 UserAttributes:[starttime:1666085694803271345 data:<POST /test?sleep=800 HTTP/1.1...Host: 10.10.103.148:9000...Content-Type: application/x-www-form-urlencoded ...xxxxjkajdnaksndandlkajelkqjweklqwjeklqjlwejklasdas...qwelkqnmwlkenqlwenqwnekqkwelkqmnwlekq(69 bytes more)>] Ctx:{ThreadInfo:{Pid:22592 Tid:22592 Uid:0 Gid:0 Comm:node ContainerId:123123 ContainerName:test} FdInfo:{Num:19 TypeFd:3 Filename:tmp Directory:/root Protocol:1 Role:true Sip:[127.0.0.1,127.0.0.1] Dip:[127.0.0.1] Sport:39620 Dport:37093 Source:0 Destination:0}} Latency:100 Timestamp:160000}`,
		TextKindlingEvent(evt),
		"Unexpected log output format from KindlingEvent",
	)
}
