// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cpuanalyzer

import (
	"strconv"
	"sync"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/filepathhelper"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constvalues"
)

// Eager Initialization
var (
	enableProfile    bool
	triggerEventChan chan SendTriggerEvent
	traceChan        chan *model.DataGroup
	sampleMap        sync.Map
	isInstallApm     map[uint64]bool
)

// ReceiveDataGroupAsSignal receives model.DataGroup as a signal.
// Signal is used to trigger to send CPU on/off events
func ReceiveDataGroupAsSignal(data *model.DataGroup) {
	if !enableProfile {
		return
	}
	isFromApm := data.Labels.GetBoolValue("isInstallApm")
	if isFromApm {
		isInstallApm[uint64(data.Labels.GetIntValue("pid"))] = true
		if !data.Labels.GetBoolValue(constlabels.IsSlow) {
			return
		}
		// We save the trace to sampleMap to make it as the sending trigger event.
		pidString := strconv.FormatInt(data.Labels.GetIntValue("pid"), 10)
		// The data is unnecessary to be cloned as it won't be reused.
		sampleMap.LoadOrStore(data.Labels.GetStringValue(constlabels.ContentKey)+pidString, data)
	} else {
		if !data.Labels.GetBoolValue(constlabels.IsSlow) {
			return
		}
		// Clone the data for further usage. Otherwise, it will be reused and loss fields.
		trace := data.Clone()
		// CpuAnalyzer consumes all traces from the client-side to add them to TimeSegments for data enrichment.
		// Now we don't store the trace from the server-side due to the storage concern.
		if !trace.Labels.GetBoolValue(constlabels.IsServer) {
			traceChan <- trace
			// The trace sent from the client-side won't be treated as trigger event, so we just return here.
			return
		}
		// If the data is not from APM while there have been APM traces received, we don't make it as a signal.
		if !isInstallApm[uint64(data.Labels.GetIntValue("pid"))] {
			// We save the trace to sampleMap to make it as the sending trigger event.
			pidString := strconv.FormatInt(data.Labels.GetIntValue("pid"), 10)
			sampleMap.LoadOrStore(data.Labels.GetStringValue(constlabels.ContentKey)+pidString, trace)
		}
	}
}

type SendTriggerEvent struct {
	Pid          uint32           `json:"pid"`
	StartTime    uint64           `json:"startTime"`
	SpendTime    uint64           `json:"spendTime"`
	OriginalData *model.DataGroup `json:"originalData"`
}

// ReadTraceChan reads the trace channel and make cpuanalyzer consume them as general events.
func (ca *CpuAnalyzer) ReadTraceChan() {
	for {
		select {
		case trace := <-traceChan:
			ca.ConsumeTraces(trace)
		case <-ca.stopProfileChan:
			return
		}
	}
}

// ReadTriggerEventChan reads the triggerEvent channel and creates tasks to send cpuEvents.
func (ca *CpuAnalyzer) ReadTriggerEventChan() {
	for {
		select {
		case sendContent := <-triggerEventChan:
			// Only send the slow traces as the signals
			if !sendContent.OriginalData.Labels.GetBoolValue(constlabels.IsSlow) {
				continue
			}
			// Store the traces first
			for _, nexConsumer := range ca.nextConsumers {
				_ = nexConsumer.Consume(sendContent.OriginalData)
			}
			// Copy the value and then get its pointer to create a new task
			triggerEvent := sendContent
			task := &SendEventsTask{
				cpuAnalyzer:              ca,
				triggerEvent:             &triggerEvent,
				edgeEventsWindowDuration: time.Duration(ca.cfg.EdgeEventsWindowSize) * time.Second,
			}
			expiredCallback := func() {
				ca.routineSize.Dec()
			}
			// The expired duration should be windowDuration+1 because the ticker and the timer are not started together.
			NewAndStartScheduledTaskRoutine(1*time.Second, time.Duration(ca.cfg.EdgeEventsWindowSize)*time.Second+1, task, expiredCallback)
			ca.routineSize.Inc()
		case <-ca.stopProfileChan:
			return
		}
	}
}

type SendEventsTask struct {
	tickerCount              int
	cpuAnalyzer              *CpuAnalyzer
	triggerEvent             *SendTriggerEvent
	edgeEventsWindowDuration time.Duration
}

// |________________|______________|_________________|
// 0  (edgeWindow)  1  (duration)  2  (edgeWindow)   3
// 0: The start time of the windows where the events we need are.
// 1: The start time of the "trace".
// 2: The end time of the "trace". This is nearly equal to the creating time of the task.
// 3: The end time of the windows where the events we need are.
func (t *SendEventsTask) run() {
	currentWindowsStartTime := uint64(t.tickerCount)*uint64(time.Second) + t.triggerEvent.StartTime - uint64(t.edgeEventsWindowDuration)
	currentWindowsEndTime := uint64(t.tickerCount)*uint64(time.Second) + t.triggerEvent.StartTime + t.triggerEvent.SpendTime
	t.tickerCount++
	// keyElements are used to correlate the cpuEvents with the trace.
	keyElements := filepathhelper.GetFilePathElements(t.triggerEvent.OriginalData, t.triggerEvent.StartTime)
	t.cpuAnalyzer.sendEvents(keyElements.ToAttributes(), t.triggerEvent.Pid, currentWindowsStartTime, currentWindowsEndTime)
}

func (ca *CpuAnalyzer) sampleSend() {
	timer := time.NewTicker(time.Duration(ca.cfg.SamplingInterval) * time.Second)
	for {
		select {
		case <-timer.C:
			sampleMap.Range(func(k, v interface{}) bool {
				data := v.(*model.DataGroup)
				duration, ok := data.GetMetric(constvalues.RequestTotalTime)
				if !ok {
					sampleMap.Delete(k)
					return true
				}
				event := SendTriggerEvent{
					Pid:          uint32(data.Labels.GetIntValue("pid")),
					StartTime:    data.Timestamp,
					SpendTime:    uint64(duration.GetInt().Value),
					OriginalData: data,
				}
				triggerEventChan <- event
				sampleMap.Delete(k)
				return true
			})
		case <-ca.stopProfileChan:
			return
		}
	}
}

func (ca *CpuAnalyzer) sendEvents(keyElements *model.AttributeMap, pid uint32, startTime uint64, endTime uint64) {
	ca.lock.RLock()
	defer ca.lock.RUnlock()
	if !enableProfile {
		ca.telemetry.Logger.Infof("The profiling is disabled, so won't send the cpu events. "+
			"pid=%d, startTime=%d, endTime=%d", pid, startTime, endTime)
		return
	}
	maxSegmentSize := ca.cfg.SegmentSize
	tidCpuEvents, exist := ca.cpuPidEvents[pid]
	if !exist {
		ca.telemetry.Logger.Infof("Not found the cpu events with the pid=%d, startTime=%d, endTime=%d",
			pid, startTime, endTime)
		return
	}
	startTimeSecond := startTime / nanoToSeconds
	endTimeSecond := endTime / nanoToSeconds

	for _, timeSegments := range tidCpuEvents {
		if endTimeSecond < timeSegments.BaseTime || startTimeSecond > timeSegments.BaseTime+uint64(maxSegmentSize) {
			ca.telemetry.Logger.Infof("pid=%d tid=%d events are beyond the time windows. BaseTimeSecond=%d, "+
				"startTimeSecond=%d, endTimeSecond=%d", pid, timeSegments.Tid, timeSegments.BaseTime, startTimeSecond, endTimeSecond)
			continue
		}
		startIndex := int(startTimeSecond - timeSegments.BaseTime)
		if startIndex < 0 {
			startIndex = 0
		}
		endIndex := endTimeSecond - timeSegments.BaseTime
		if endIndex > timeSegments.BaseTime+uint64(maxSegmentSize) {
			endIndex = timeSegments.BaseTime + uint64(maxSegmentSize)
		}
		ca.telemetry.Logger.Infof("pid=%d tid=%d sends events. startSecond=%d, endSecond=%d",
			pid, timeSegments.Tid, startTimeSecond, endTimeSecond)
		for i := startIndex; i <= int(endIndex) && i < maxSegmentSize; i++ {
			val := timeSegments.Segments.GetByIndex(i)
			if val == nil {
				continue
			}
			segment := val.(*Segment)
			if segment.isNotEmpty() {
				// Don't remove the duplicated one
				segment.IndexTimestamp = time.Now().String()
				dataGroup := segment.toDataGroup(timeSegments)
				dataGroup.Labels.Merge(keyElements)
				for _, nexConsumer := range ca.nextConsumers {
					_ = nexConsumer.Consume(dataGroup)
				}
				segment.IsSend = 1
			}
		}
	}
}
