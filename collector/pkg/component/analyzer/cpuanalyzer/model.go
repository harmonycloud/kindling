// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cpuanalyzer

import (
	"encoding/json"

	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
)

type TimedEventKind int

const (
	TimedCpuEventKind TimedEventKind = iota
	TimedJavaFutexEventKind
	TimedTransactionIdEventKind
	TimedApmSpanEventKind
	TimedInnerCallEventKind
)

const (
	CpuEventLabel           = "cpuEvents"
	JavaFutexEventLabel     = "javaFutexEvents"
	TransactionIdEventLabel = "transactionIds"
	SpanLabel               = "spans"
	InnerCallLabel          = "innerCalls"
)

type TimedEvent interface {
	StartTimestamp() uint64
	EndTimestamp() uint64
	Kind() TimedEventKind
}

type TimeSegments struct {
	Pid        uint32       `json:"pid"`
	Tid        uint32       `json:"tid"`
	ThreadName string       `json:"threadName"`
	BaseTime   uint64       `json:"baseTime"`
	Segments   *CircleQueue `json:"segments"`
}

func (t *TimeSegments) updateThreadName(threadName string) {
	t.ThreadName = threadName
}

type Segment struct {
	StartTime       uint64       `json:"startTime"`
	EndTime         uint64       `json:"endTime"`
	CpuEvents       []TimedEvent `json:"cpuEvents"`
	JavaFutexEvents []TimedEvent `json:"javaFutexEvents"`
	TransactionIds  []TimedEvent `json:"transactionIds"`
	Spans           []TimedEvent `json:"spans"`
	InnerCalls      []TimedEvent `json:"innerCalls"`
	IsSend          int
	IndexTimestamp  string `json:"indexTimestamp"`
}

func newSegment(startTime uint64, endTime uint64) *Segment {
	return &Segment{
		StartTime:       startTime,
		EndTime:         endTime,
		CpuEvents:       make([]TimedEvent, 0),
		JavaFutexEvents: make([]TimedEvent, 0),
		TransactionIds:  make([]TimedEvent, 0),
		Spans:           make([]TimedEvent, 0),
		InnerCalls:      make([]TimedEvent, 0),
		IsSend:          0,
		IndexTimestamp:  "",
	}
}

func (s *Segment) putTimedEvent(event TimedEvent) {
	switch event.Kind() {
	case TimedCpuEventKind:
		s.CpuEvents = append(s.CpuEvents, event)
	case TimedJavaFutexEventKind:
		s.JavaFutexEvents = append(s.JavaFutexEvents, event)
	case TimedTransactionIdEventKind:
		s.TransactionIds = append(s.TransactionIds, event)
	case TimedApmSpanEventKind:
		s.Spans = append(s.Spans, event)
	case TimedInnerCallEventKind:
		s.InnerCalls = append(s.InnerCalls, event)
	}
}

func (s *Segment) toDataGroup(parent *TimeSegments) *model.DataGroup {
	labels := model.NewAttributeMap()
	labels.AddIntValue(constlabels.Pid, int64(parent.Pid))
	labels.AddIntValue(constlabels.Tid, int64(parent.Tid))
	labels.AddIntValue(constlabels.IsSent, int64(s.IsSend))
	labels.AddStringValue(constlabels.ThreadName, parent.ThreadName)
	labels.AddIntValue(constlabels.StartTime, int64(s.StartTime))
	labels.AddIntValue(constlabels.EndTime, int64(s.EndTime))
	cpuEventString, err := json.Marshal(s.CpuEvents)
	if err == nil {
		labels.AddStringValue(CpuEventLabel, string(cpuEventString))
	}
	javaFutexEventString, err := json.Marshal(s.JavaFutexEvents)
	if err == nil {
		labels.AddStringValue(JavaFutexEventLabel, string(javaFutexEventString))
	}
	transactionIdEventString, err := json.Marshal(s.TransactionIds)
	if err == nil {
		labels.AddStringValue(TransactionIdEventLabel, string(transactionIdEventString))
	}
	spanEventString, err := json.Marshal(s.Spans)
	if err == nil {
		labels.AddStringValue(SpanLabel, string(spanEventString))
	}
	innerCallString, err := json.Marshal(s.InnerCalls)
	if err == nil {
		labels.AddStringValue(InnerCallLabel, string(innerCallString))
	}
	return model.NewDataGroup(constnames.CameraEventGroupName, labels, s.StartTime)
}

func (s *Segment) isNotEmpty() bool {
	return len(s.CpuEvents) > 0 || len(s.JavaFutexEvents) > 0 || len(s.TransactionIds) > 0 || len(s.Spans) > 0 || len(s.InnerCalls) > 0
}

func (s *Segment) UnmarshalJSON(data []byte) error {
	events := make(map[string]json.RawMessage)
	err := json.Unmarshal(data, &events)

	if err != nil {
		return err
	}

	for k, v := range events {
		switch k {
		case "startTime":
			var t uint64
			err := json.Unmarshal(v, &t)
			if err != nil {
				return err
			}
			s.StartTime = t
		case "endTime":
			var t uint64
			err := json.Unmarshal(v, &t)
			if err != nil {
				return err
			}
			s.EndTime = t
		case CpuEventLabel:
			var e []CpuEvent
			err := json.Unmarshal(v, &e)
			if err != nil {
				return err
			}
			for i, _ := range e {
				s.CpuEvents = append(s.CpuEvents, &e[i])
			}
		case JavaFutexEventLabel:
			var e []JavaFutexEvent
			err := json.Unmarshal(v, &e)
			if err != nil {
				return err
			}
			for _, a := range e {
				s.JavaFutexEvents = append(s.JavaFutexEvents, &a)
			}
		default:
			//return errors.New("unrecognized key")
		}
	}
	return nil
}

type CPUType uint8

const (
	CPUType_ON    CPUType = 0
	CPUType_FILE  CPUType = 1
	CPUType_NET   CPUType = 2
	CPUType_FUTEX CPUType = 3
	CPUType_IDLE  CPUType = 4
	CPUType_OTHER CPUType = 5
	CPUType_EPOLL CPUType = 6
	CPUTYPE_MAX   CPUType = 7
)

func (ct CPUType) MarshalJSON() ([]byte, error) {
	return json.Marshal(uint16(ct))
}

func (ct *CPUType) UnmarshalJSON(data []byte) error {
	var val uint16
	err := json.Unmarshal(data, &val)
	if err != nil {
		return err
	}
	*ct = CPUType(val)
	return nil
}

type CpuEvent struct {
	StartTime   uint64    `json:"startTime"`
	EndTime     uint64    `json:"endTime"`
	TypeSpecs   []uint64  `json:"typeSpecs"`
	RunqLatency []uint64  `json:"runqLatency"`
	TimeType    []CPUType `json:"timeType"`
	OnInfo      string    `json:"onInfo"`
	OffInfo     string    `json:"offInfo"`
	Log         string    `json:"log"`
	Stack       string    `json:"stack"`
}

func (c *CpuEvent) StartTimestamp() uint64 {
	return c.StartTime
}

func (c *CpuEvent) EndTimestamp() uint64 {
	return c.EndTime
}

func (c *CpuEvent) Kind() TimedEventKind {
	return TimedCpuEventKind
}

type JavaFutexEvent struct {
	StartTime uint64 `json:"startTime"`
	EndTime   uint64 `json:"endTime"`
	DataVal   string `json:"dataValue"`
}

func (j *JavaFutexEvent) StartTimestamp() uint64 {
	return j.StartTime
}

func (j *JavaFutexEvent) EndTimestamp() uint64 {
	return j.EndTime
}

func (j *JavaFutexEvent) Kind() TimedEventKind {
	return TimedJavaFutexEventKind
}

type TransactionIdEvent struct {
	Timestamp   uint64 `json:"timestamp"`
	TraceId     string `json:"traceId"`
	IsEntry     uint32 `json:"isEntry"`
	Protocol    string `json:"protocol"`
	Url         string `json:"url"`
	PidString   string `json:"pidString"`
	ContainerId string `json:"containerId"`
}

func (t *TransactionIdEvent) StartTimestamp() uint64 {
	return t.Timestamp
}

func (t *TransactionIdEvent) EndTimestamp() uint64 {
	return t.Timestamp
}

func (t *TransactionIdEvent) Kind() TimedEventKind {
	return TimedTransactionIdEventKind
}

type ApmSpanEvent struct {
	StartTime uint64 `json:"startTime"`
	EndTime   uint64 `json:"endTime"`
	TraceId   string `json:"traceId"`
	Name      string `json:"name"`
}

func (j *ApmSpanEvent) StartTimestamp() uint64 {
	return j.StartTime
}

func (j *ApmSpanEvent) EndTimestamp() uint64 {
	return j.EndTime
}

func (j *ApmSpanEvent) Kind() TimedEventKind {
	return TimedApmSpanEventKind
}

type InnerCall struct {
	StartTime uint64           `json:"startTime"`
	EndTime   uint64           `json:"endTime"`
	Trace     *model.DataGroup `json:"trace"`
}

func (c *InnerCall) StartTimestamp() uint64 {
	return c.StartTime
}

func (c *InnerCall) EndTimestamp() uint64 {
	return c.EndTime
}

func (c *InnerCall) Kind() TimedEventKind {
	return TimedInnerCallEventKind
}
