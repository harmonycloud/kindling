// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"sync"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network/protocol"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
)

type DnsUdpCache struct {
	count        int
	requestCache sync.Map
}

func newDnsUdpCache() *DnsUdpCache {
	return &DnsUdpCache{}
}

func (cache *DnsUdpCache) addRequest(request *udpRequest) {
	cache.requestCache.Store(request.id, request)
	cache.count += 1
}

func (cache *DnsUdpCache) getMatchRequest(attributes *model.AttributeMap) (*model.KindlingEvent, int) {
	if request, exist := cache.requestCache.LoadAndDelete(attributes.GetIntValue(constlabels.DnsId)); exist {
		cache.count -= 1
		return request.(*udpRequest).event, cache.count
	}
	return nil, cache.count
}

func (cache *DnsUdpCache) deleteRequest(key interface{}) {
	cache.requestCache.Delete(key)
	cache.count -= 1
}

func (cache *DnsUdpCache) isEmpty() bool {
	return cache.count == 0
}

func parseDnsUdpRequest(parser *protocol.ProtocolParser, event *model.KindlingEvent) (parsedRequest *udpRequest, success bool) {
	message := protocol.NewRequestMessage(event.GetData())
	parser.ParseRequest(message)
	success = message.HasAttribute(constlabels.DnsId)
	if !success {
		return
	}

	parsedRequest = newUdpRequest(event, message.GetAttributes())
	return
}

func parseDnsUdpResponse(parser *protocol.ProtocolParser, event *model.KindlingEvent) (attributes *model.AttributeMap, success bool) {
	message := protocol.NewResponseMessage(event.GetData(), model.NewAttributeMap())
	parser.ParseResponse(message)
	success = message.HasAttribute(constlabels.DnsId)
	if !success {
		return
	}

	attributes = message.GetAttributes()
	return
}

type udpRequest struct {
	id         int64
	event      *model.KindlingEvent
	attritutes *model.AttributeMap
}

func newUdpRequest(event *model.KindlingEvent, attributes *model.AttributeMap) *udpRequest {
	return &udpRequest{
		id:         attributes.GetIntValue(constlabels.DnsId),
		event:      event,
		attritutes: attributes,
	}
}

// DNS will send different ip and port data with sharing fd and pid socket.
type udpKey struct {
	pid   uint32
	fd    int32
	sip   string
	dip   string
	sport uint32
	dport uint32
}

func getUdpKey(evt *model.KindlingEvent) udpKey {
	return udpKey{
		pid:   evt.GetPid(),
		fd:    evt.GetFd(),
		sip:   evt.GetSip(),
		dip:   evt.GetDip(),
		sport: evt.GetSport(),
		dport: evt.GetDport(),
	}
}
