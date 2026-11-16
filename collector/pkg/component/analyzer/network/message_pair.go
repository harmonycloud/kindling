// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package network

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/metadata/conntracker"
	"github.com/Kindling-project/kindling/collector/pkg/model"
)

type mergableEvent struct {
	events []*model.KindlingEvent // Keep No more than 10 events.

	latency uint64
	resVal  int64
	ts      uint64
	data    []byte
}

type events struct {
	event            *model.KindlingEvent
	mergable         *mergableEvent
	maxPayloadLength int
}

func newEvents(evt *model.KindlingEvent, maxPayloadSize int) *events {
	return &events{
		event:            evt,
		maxPayloadLength: maxPayloadSize,
	}
}

func (evts *events) size() int {
	if evts.mergable == nil {
		return 1
	}
	return len(evts.mergable.events)
}

func (evts *events) getEvent(index int) *model.KindlingEvent {
	if index == 0 {
		return evts.event
	}
	if index > 0 && evts.mergable != nil && index < len(evts.mergable.events) {
		return evts.mergable.events[index]
	}
	return nil
}

func (evts *events) putEventBack(originEvts *events) {
	newEvt := evts.event
	evts.event = originEvts.event
	evts.mergable = originEvts.mergable
	evts.mergeEvent(newEvt)
}

func (evts *events) mergeEvent(evt *model.KindlingEvent) {
	if evts.mergable == nil {
		firstEvt := evts.event
		evts.mergable = &mergableEvent{
			events:  []*model.KindlingEvent{firstEvt},
			latency: firstEvt.GetLatency(),
			resVal:  firstEvt.GetResVal(),
			ts:      firstEvt.Timestamp,
			data:    firstEvt.GetData(),
		}
	}

	if len(evts.mergable.events) < 10 {
		// persistent connect
		evts.mergable.events = append(evts.mergable.events, evt)
	}
	evts.mergable.latency = evt.Timestamp - evts.mergable.ts + evts.mergable.latency
	evts.mergable.resVal += evt.GetResVal()
	evts.mergable.ts = evt.Timestamp

	// We have a constraint on the payload size. The merged data can accommodate a maximum payload size
	// as same as SANPLEN that is also the maximum size of the syscall data.
	// If the previous data size is smaller than maxPayloadLength, the later one would fill that gap.
	appendLength := evts.getAppendLength(len(evt.GetData()))
	if appendLength == 0 {
		return
	}
	evts.mergable.data = append(evts.mergable.data, evt.GetData()[0:appendLength]...)
}

// getAppendLength returns the length to accommodate the new event according to the remaining size and
// the new event's size.
func (evts *events) getAppendLength(newEventLength int) int {
	remainingSize := evts.maxPayloadLength - len(evts.mergable.data)
	// If the merged data is full
	if remainingSize <= 0 {
		return 0
	}
	// If the merged data is not full, return the smaller size
	if remainingSize > newEventLength {
		return newEventLength
	} else {
		return remainingSize
	}
}

func (evts *events) getData() []byte {
	if evts.mergable == nil {
		return evts.event.GetData()
	}
	return evts.mergable.data
}

func (evts *events) getFirstTimestamp() uint64 {
	return evts.event.Timestamp
}

func (evts *events) getLastTimestamp() uint64 {
	if evts.mergable == nil {
		return evts.event.Timestamp
	}
	return evts.mergable.ts
}

func (evts *events) IsTimeout(newEvt *model.KindlingEvent, timeout int) bool {
	// old event is stale if timeout or sPort mismatched
	// be careful about overflow
	firstEvt := evts.event
	firstStartTime := firstEvt.Timestamp - firstEvt.GetLatency()
	if newEvt.Timestamp < firstStartTime {
		return false
	}
	if newEvt.Timestamp-firstStartTime > uint64(timeout)*uint64(time.Second) || newEvt.GetSport() != firstEvt.GetSport() {
		return true
	}
	return false
}

func (evts *events) IsSportChanged(newEvt *model.KindlingEvent) bool {
	return newEvt.GetSport() != evts.event.GetSport()
}

func (evts *events) getDuration() uint64 {
	if evts == nil {
		return 0
	}

	if evts.mergable == nil {
		return evts.event.GetLatency()
	}

	return evts.mergable.latency
}

type messagePairs struct {
	connects         *events
	requests         *events
	responses        *events
	natTuple         *conntracker.IPTranslation
	isSend           int32
	mutex            sync.RWMutex // only for update latency and resval now
	maxPayloadLength int
}

func (mps *messagePairs) checkSend() bool {
	// Check Send Once.
	return atomic.AddInt32(&mps.isSend, 1) == 1
}

func (mps *messagePairs) getKey() messagePairKey {
	if mps.connects != nil {
		return getMessagePairKey(mps.connects.event)
	} else if mps.requests != nil {
		return getMessagePairKey(mps.requests.event)
	} else if mps.responses != nil {
		return getMessagePairKey(mps.responses.event)
	}
	return messagePairKey{}
}

func (mps *messagePairs) mergeConnect(evt *model.KindlingEvent) {
	mps.mutex.Lock()
	if mps.requests == nil {
		mps.connects = newEvents(evt, mps.maxPayloadLength)
	} else {
		mps.connects.mergeEvent(evt)
	}
	mps.mutex.Unlock()
}

func (mps *messagePairs) putRequestBack(evts *events) {
	mps.mutex.Lock()
	if mps.requests == nil {
		mps.requests = evts
	} else {
		mps.requests.putEventBack(evts)
	}
	mps.mutex.Unlock()
}

func (mps *messagePairs) mergeRequest(evt *model.KindlingEvent) {
	mps.mutex.Lock()
	if mps.requests == nil {
		mps.requests = newEvents(evt, mps.maxPayloadLength)
	} else {
		mps.requests.mergeEvent(evt)
	}
	mps.mutex.Unlock()
}

func (mps *messagePairs) mergeResponse(evt *model.KindlingEvent) {
	mps.mutex.Lock()
	if mps.responses == nil {
		mps.responses = newEvents(evt, mps.maxPayloadLength)
	} else {
		mps.responses.mergeEvent(evt)
	}
	mps.mutex.Unlock()
}

func (mps *messagePairs) getPort() uint32 {
	if mps.requests != nil {
		return mps.requests.event.GetDport()
	}
	if mps.responses != nil {
		return mps.responses.event.GetDport()
	}
	return 0
}

func (mps *messagePairs) getTimeoutTs() uint64 {
	if mps.responses != nil {
		return mps.responses.getLastTimestamp()
	}
	if mps.requests != nil {
		return mps.requests.getLastTimestamp()
	}
	if mps.connects != nil {
		return mps.connects.getLastTimestamp()
	}
	return 0
}

func (mps *messagePairs) getConnectDuration() uint64 {
	if mps.connects == nil {
		return 0
	}

	return mps.connects.getDuration()
}

func (mps *messagePairs) getSentTime() int64 {
	if mps.requests == nil {
		return -1
	}

	return int64(mps.requests.getDuration())
}

func (mps *messagePairs) getWaitingTime() int64 {
	if mps.responses == nil {
		return -1
	}

	return int64(mps.responses.getFirstTimestamp() - mps.responses.event.GetLatency() - mps.requests.getLastTimestamp())
}

func (mps *messagePairs) getDownloadTime() int64 {
	if mps.responses == nil {
		return -1
	}

	return int64(mps.responses.getDuration())
}

func (mps *messagePairs) getDuration() uint64 {
	if mps.responses == nil {
		return 0
	}

	return mps.responses.getLastTimestamp() - mps.requests.event.GetStartTime()
}

func (mps *messagePairs) getRquestSize() uint64 {
	if mps.requests == nil {
		return 0
	}

	if mps.requests.mergable == nil {
		return uint64(mps.requests.event.GetResVal())
	}
	return uint64(mps.requests.mergable.resVal)
}

func (mps *messagePairs) getResponseSize() uint64 {
	if mps.responses == nil {
		return 0
	}

	if mps.responses.mergable == nil {
		return uint64(mps.responses.event.GetResVal())
	}
	return uint64(mps.responses.mergable.resVal)
}

type messagePair struct {
	request  *model.KindlingEvent
	response *model.KindlingEvent
	natTuple *conntracker.IPTranslation
}

func (mp *messagePair) getSentTime() int64 {
	if mp.request == nil {
		return -1
	}

	return int64(mp.request.GetLatency())
}

func (mp *messagePair) getWaitingTime() int64 {
	if mp.response == nil {
		return -1
	}

	return int64(mp.response.Timestamp - mp.response.GetLatency() - mp.request.Timestamp)
}

func (mp *messagePair) getDownloadTime() int64 {
	if mp.response == nil {
		return -1
	}

	return int64(mp.response.GetLatency())
}

func (mp *messagePair) getRquestSize() uint64 {
	if mp.request == nil {
		return 0
	}
	return uint64(mp.request.GetResVal())
}

func (mp *messagePair) getResponseSize() uint64 {
	if mp.response == nil {
		return 0
	}
	return uint64(mp.response.GetResVal())
}

func (mp *messagePair) getDuration() uint64 {
	if mp.response == nil {
		return 0
	}

	return mp.response.Timestamp + mp.request.GetLatency() - mp.request.Timestamp
}

type messagePairKey struct {
	pid uint32
	fd  int32
}

func getMessagePairKey(evt *model.KindlingEvent) messagePairKey {
	return messagePairKey{
		pid: evt.GetPid(),
		fd:  evt.GetFd(),
	}
}
