// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"encoding/binary"
	"errors"
	"math"
	"net"

	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
)

const (
	LOWER32 = 0x00000000FFFFFFFF
	LOWER16 = 0x000000000000FFFF
	_       = LOWER16
)

var (
	ErrMessageNotSocket = errors.New("not a network receive/send event")
	byteOrder           = getByteOrder()
)

func getByteOrder() binary.ByteOrder {
	// Check if littleendian or bigendian
	s := int16(0x1234)
	littleVal := byte(0x34)
	if littleVal == byte(int8(s)) {
		return binary.LittleEndian
	}
	return binary.BigEndian
}

func (k *KindlingEvent) GetData() []byte {
	keyValue := k.GetUserAttribute("data")
	if keyValue != nil {
		return keyValue.GetValue()
	}
	return nil
}

func (k *KindlingEvent) GetDataLen() int {
	keyValue := k.GetUserAttribute("data")
	if keyValue != nil {
		return len(keyValue.GetValue())
	}
	return 0
}

func (k *KindlingEvent) GetResVal() int64 {
	keyValue := k.GetUserAttribute("res")
	if keyValue != nil {
		return int64(byteOrder.Uint64(keyValue.Value))
	}
	return -1
}

func (k *KindlingEvent) GetLatency() uint64 {
	return k.Latency
}

func (k *KindlingEvent) GetUintUserAttribute(key string) uint64 {
	keyValue := k.GetUserAttribute(key)
	if keyValue != nil {
		return keyValue.GetUintValue()
	}
	return 0
}

func (k *KindlingEvent) GetIntUserAttribute(key string) int64 {
	keyValue := k.GetUserAttribute(key)
	if keyValue != nil {
		return keyValue.GetIntValue()
	}
	return 0
}

func (k *KindlingEvent) GetFloatUserAttribute(key string) float32 {
	keyValue := k.GetUserAttribute(key)
	if keyValue != nil && keyValue.ValueType == ValueType_FLOAT {
		return math.Float32frombits(byteOrder.Uint32(keyValue.Value))
	}
	return 0.0
}

func (k *KindlingEvent) GetDoubleUserAttribute(key string) float64 {
	keyValue := k.GetUserAttribute(key)
	if keyValue != nil && keyValue.ValueType == ValueType_FLOAT {
		return math.Float64frombits(byteOrder.Uint64(keyValue.Value))
	}
	return 0.0
}

func (k *KindlingEvent) GetStringUserAttribute(key string) string {
	keyValue := k.GetUserAttribute(key)
	if keyValue != nil {
		return string(keyValue.GetValue())
	}
	return ""
}

func (k *KindlingEvent) GetStartTime() uint64 {
	return k.Timestamp - k.GetLatency()
}

func (k *KindlingEvent) GetUserAttribute(key string) *KeyValue {
	if k.ParamsNumber == 0 {
		return nil
	}
	for index, keyValue := range k.UserAttributes {
		if index+1 > int(k.ParamsNumber) {
			break
		}
		if keyValue.Key == key {
			return &keyValue
		}
	}
	return nil
}

func (k *KindlingEvent) SetUserAttribute(key string, value []byte) {
	if k.ParamsNumber == 0 {
		return
	}
	for index, keyValue := range k.UserAttributes {
		if index+1 > int(k.ParamsNumber) {
			break
		}
		if keyValue.Key == key {
			k.UserAttributes[index].Value = value
			break
		}
	}
}

func (k *KindlingEvent) GetPid() uint32 {
	ctx := k.GetCtx()
	if ctx == nil {
		return 0
	}
	threadInfo := ctx.GetThreadInfo()
	if threadInfo == nil {
		return 0
	}
	return threadInfo.Pid
}

func (k *KindlingEvent) GetTid() uint32 {
	ctx := k.GetCtx()
	if ctx == nil {
		return 0
	}
	threadInfo := ctx.GetThreadInfo()
	if threadInfo == nil {
		return 0
	}
	return threadInfo.Tid
}

func (k *KindlingEvent) GetComm() string {
	ctx := k.GetCtx()
	if ctx == nil {
		return ""
	}
	threadInfo := ctx.GetThreadInfo()
	if threadInfo == nil {
		return ""
	}
	return threadInfo.Comm
}

func (k *KindlingEvent) GetContainerId() string {
	ctx := k.GetCtx()
	if ctx == nil {
		return ""
	}
	threadInfo := ctx.GetThreadInfo()
	if threadInfo == nil {
		return ""
	}
	return threadInfo.ContainerId
}

func (k *KindlingEvent) GetFd() int32 {
	ctx := k.GetCtx()
	if ctx == nil {
		return 0
	}
	fdInfo := ctx.GetFdInfo()
	if fdInfo == nil {
		return 0
	}
	return fdInfo.Num
}

func (k *KindlingEvent) GetSip() string {
	ctx := k.GetCtx()
	if ctx == nil {
		return ""
	}
	fdInfo := ctx.GetFdInfo()
	if fdInfo == nil {
		return ""
	}
	// TODO Only support IPV4, may support IPV6 in future.
	return IPLong2String(fdInfo.Sip[0])
}

func (k *KindlingEvent) GetDip() string {
	ctx := k.GetCtx()
	if ctx == nil {
		return ""
	}
	fdInfo := ctx.GetFdInfo()
	if fdInfo == nil {
		return ""
	}
	// TODO Only support IPV4, may support IPV6 in future.
	return IPLong2String(fdInfo.Dip[0])
}

func IPLong2String(i uint32) string {
	if i > math.MaxUint32 {
		return ""
	}

	ip := make(net.IP, net.IPv4len)
	ip[3] = byte(i >> 24)
	ip[2] = byte(i >> 16)
	ip[1] = byte(i >> 8)
	ip[0] = byte(i)

	return ip.String()
}

func (k *KindlingEvent) GetSport() uint32 {
	ctx := k.GetCtx()
	if ctx == nil {
		return 0
	}
	fdInfo := ctx.GetFdInfo()
	if fdInfo == nil {
		return 0
	}
	return fdInfo.Sport
}

func (k *KindlingEvent) GetDport() uint32 {
	ctx := k.GetCtx()
	if ctx == nil {
		return 0
	}
	fdInfo := ctx.GetFdInfo()
	if fdInfo == nil {
		return 0
	}
	return fdInfo.Dport
}

func (k *KindlingEvent) IsUdp() uint32 {
	if k.GetCtx().GetFdInfo().GetProtocol() == L4Proto_UDP {
		return 1
	}
	return 0
}

func (k *KindlingEvent) IsTcp() bool {
	context := k.GetCtx()
	if context == nil {
		return false
	}
	fd := context.GetFdInfo()
	if fd == nil {
		return false
	}
	return fd.GetProtocol() == L4Proto_TCP
}

func (k *KindlingEvent) IsConnect() bool {
	return k.Name == "connect"
}

func (k *KindlingEvent) IsRequest() (bool, error) {
	if k.Category == Category_CAT_NET {
		switch k.Name {
		case constnames.ReadEvent, constnames.RecvFromEvent, constnames.RecvMsgEvent, constnames.ReadvEvent:
			fallthrough
		case constnames.PReadEvent, constnames.PReadvEvent:
			return k.isRequest(true)
		case constnames.WriteEvent, constnames.SendToEvent, constnames.SendMsgEvent, constnames.WritevEvent:
			fallthrough
		case constnames.SendMMsgEvent, constnames.PWriteEvent, constnames.PWritevEvent:
			return k.isRequest(false)
		default:
			break
		}
	}
	return false, ErrMessageNotSocket
}

func (k *KindlingEvent) isRequest(in bool) (bool, error) {
	ctx := k.GetCtx()
	if ctx == nil {
		return false, ErrMessageNotSocket
	}
	fdInfo := ctx.GetFdInfo()
	if fdInfo == nil {
		return false, ErrMessageNotSocket
	}
	return fdInfo.Role == in, nil
}

func (k *KindlingEvent) GetSocketKey() uint64 {
	return uint64(int64(k.Ctx.ThreadInfo.Pid)<<32) | uint64(k.Ctx.FdInfo.Num)&LOWER32
}

func (kv *KeyValue) GetUintValue() uint64 {
	switch kv.ValueType {
	case ValueType_UINT8:
		return uint64(kv.Value[0])
	case ValueType_UINT16:
		return uint64(byteOrder.Uint16(kv.Value))
	case ValueType_UINT32:
		return uint64(byteOrder.Uint32(kv.Value))
	case ValueType_UINT64:
		return byteOrder.Uint64(kv.Value)
	}
	return 0
}

func (kv *KeyValue) GetIntValue() int64 {
	switch kv.ValueType {
	case ValueType_INT8:
		return int64(int8(kv.Value[0]))
	case ValueType_INT16:
		return int64(int16(byteOrder.Uint16(kv.Value)))
	case ValueType_INT32:
		return int64(int32(byteOrder.Uint32(kv.Value)))
	case ValueType_INT64:
		return int64(byteOrder.Uint64(kv.Value))
	}
	return 0
}
