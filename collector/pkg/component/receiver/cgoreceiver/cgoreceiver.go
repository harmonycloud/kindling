package cgoreceiver

/*
#cgo LDFLAGS: -L ./ -lkindling  -lstdc++ -ldl
#cgo CFLAGS: -I .
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>
#include "cgo_func.h"
*/
import "C"
import (
	"fmt"
	"sync"
	"time"
	"unsafe"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	analyzerpackage "github.com/Kindling-project/kindling/collector/pkg/component/analyzer"
	"github.com/Kindling-project/kindling/collector/pkg/component/receiver"
	"github.com/Kindling-project/kindling/collector/pkg/model"
)

const (
	Cgo = "cgoreceiver"
)

type CKindlingEventForGo C.struct_kindling_event_t_for_go

type CEventParamsForSubscribe C.struct_event_params_for_subscribe

type CgoReceiver struct {
	cfg               *Config
	analyzerManager   *analyzerpackage.Manager
	shutdownWG        sync.WaitGroup
	telemetry         *component.TelemetryTools
	eventChannel      chan *model.KindlingEvent
	stopCh            chan interface{}
	stats             eventCounter
	probeCounter      *probeCounter
	probeCounterMutex sync.RWMutex

	// eBPF CPU profiling / Pyroscope
	symOnce   sync.Once
	symInited bool
	symOpts   SymOptions
}

func NewCgoReceiver(config interface{}, telemetry *component.TelemetryTools, analyzerManager *analyzerpackage.Manager) receiver.Receiver {
	cfg, ok := config.(*Config)
	if !ok {
		telemetry.Logger.Panicf("Cannot convert [%s] config", Cgo)
	}
	cgoReceiver := &CgoReceiver{
		cfg:             cfg,
		analyzerManager: analyzerManager,
		telemetry:       telemetry,
		eventChannel:    make(chan *model.KindlingEvent, 3e5),
		stopCh:          make(chan interface{}, 1),
		probeCounter:    &probeCounter{},
	}
	cgoReceiver.stats = newDynamicStats(cfg.SubscribeInfo)
	newSelfMetrics(telemetry.MeterProvider, cgoReceiver)
	return cgoReceiver
}

func (r *CgoReceiver) Start() error {
	r.telemetry.Logger.Info("Start CgoReceiver")
	res := int(C.runForGo())
	if res == 1 {
		return fmt.Errorf("fail to init probe")
	}
	go r.catchSignalUp()
	time.Sleep(2 * time.Second)
	r.suppressEventsComm()
	_ = r.subEvent()
	// Wait for the C routine running
	time.Sleep(2 * time.Second)
	go r.consumeEvents()
	go r.startGetEvents()
	go r.getCaptureStatisticsByInterval(15 * time.Second)
	r.EnablePeriodicFlamegraphSampling(15 * time.Second)

	return nil
}

func (r *CgoReceiver) startGetEvents() {
	var count = 0
	var pKindlingEvents [1000]unsafe.Pointer
	r.shutdownWG.Add(1)

	for {
		select {
		case <-r.stopCh:
			r.shutdownWG.Done()
			return
		default:
			evt_count := int(C.getEventsByInterval(C.int(100000000), &pKindlingEvents[0], (unsafe.Pointer)(&count)))
			for i := 0; i < evt_count; i++ {
				event := convertEvent((*CKindlingEventForGo)(pKindlingEvents[i]))
				r.eventChannel <- event
				r.stats.add(event.Name, 1)
			}
		}
	}
}

func (r *CgoReceiver) consumeEvents() {
	r.shutdownWG.Add(1)
	for {
		select {
		case <-r.stopCh:
			r.shutdownWG.Done()
			return
		case ev := <-r.eventChannel:
			err := r.sendToNextConsumer(ev)
			if err != nil {
				r.telemetry.Logger.Info("Failed to send KindlingEvent: ", zap.Error(err))
			}
		}
	}
}

func (r *CgoReceiver) Shutdown() error {
	// TODO stop the C routine
	C.stopProfile()
	close(r.stopCh)
	r.shutdownWG.Wait()
	return nil
}

func convertEvent(cgoEvent *CKindlingEventForGo) *model.KindlingEvent {
	ev := new(model.KindlingEvent)
	ev.Timestamp = uint64(cgoEvent.timestamp)
	ev.Name = C.GoString(cgoEvent.name)
	ev.Category = model.Category(cgoEvent.category)
	ev.Ctx.ThreadInfo.Pid = uint32(cgoEvent.context.tinfo.pid)
	ev.Ctx.ThreadInfo.Tid = uint32(cgoEvent.context.tinfo.tid)
	ev.Ctx.ThreadInfo.Uid = uint32(cgoEvent.context.tinfo.uid)
	ev.Ctx.ThreadInfo.Gid = uint32(cgoEvent.context.tinfo.gid)
	ev.Ctx.ThreadInfo.Comm = C.GoString(cgoEvent.context.tinfo.comm)
	ev.Ctx.ThreadInfo.ContainerId = C.GoString(cgoEvent.context.tinfo.containerId)
	ev.Ctx.FdInfo.Protocol = model.L4Proto(cgoEvent.context.fdInfo.protocol)
	ev.Ctx.FdInfo.Num = int32(cgoEvent.context.fdInfo.num)
	ev.Ctx.FdInfo.TypeFd = model.FDType(cgoEvent.context.fdInfo.fdType)
	ev.Ctx.FdInfo.Filename = C.GoString(cgoEvent.context.fdInfo.filename)
	ev.Ctx.FdInfo.Directory = C.GoString(cgoEvent.context.fdInfo.directory)
	ev.Ctx.FdInfo.Role = If(cgoEvent.context.fdInfo.role != 0, true, false).(bool)
	ev.Ctx.FdInfo.Sip = make([]uint32, 4)
	ev.Ctx.FdInfo.Dip = make([]uint32, 4)
	for i := 0; i < 4; i++ {
		ev.Ctx.FdInfo.Sip[i] = uint32(cgoEvent.context.fdInfo.sip[i])
		ev.Ctx.FdInfo.Dip[i] = uint32(cgoEvent.context.fdInfo.dip[i])
	}
	ev.Ctx.FdInfo.Sport = uint32(cgoEvent.context.fdInfo.sport)
	ev.Ctx.FdInfo.Dport = uint32(cgoEvent.context.fdInfo.dport)
	ev.Ctx.FdInfo.Source = uint64(cgoEvent.context.fdInfo.source)
	ev.Ctx.FdInfo.Destination = uint64(cgoEvent.context.fdInfo.destination)

	ev.ParamsNumber = uint16(cgoEvent.paramsNumber)
	ev.Latency = uint64(cgoEvent.latency)
	for i := 0; i < int(ev.ParamsNumber); i++ {
		ev.UserAttributes[i].Key = C.GoString(cgoEvent.userAttributes[i].key)
		userAttributesLen := cgoEvent.userAttributes[i].len
		ev.UserAttributes[i].Value = C.GoBytes(unsafe.Pointer(cgoEvent.userAttributes[i].value), C.int(userAttributesLen))
		ev.UserAttributes[i].ValueType = model.ValueType(cgoEvent.userAttributes[i].valueType)
	}
	return ev
}

func If(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

func (r *CgoReceiver) sendToNextConsumer(evt *model.KindlingEvent) error {
	if ce := r.telemetry.Logger.Check(zapcore.DebugLevel, ""); ce != nil {
		r.telemetry.Logger.Debug(fmt.Sprintf("Event Output: %+v", model.TextKindlingEvent(evt)))
	}
	analyzers := r.analyzerManager.GetConsumableAnalyzers(evt.Name)
	if analyzers == nil || len(analyzers) == 0 {
		//r.telemetry.Logger.Info("analyzer not found for event ", zap.String("eventName", evt.Name))
		return nil
	}
	for _, analyzer := range analyzers {
		err := analyzer.ConsumeEvent(evt)
		if err != nil {
			r.telemetry.Logger.Warn("Error sending event to next consumer: ", zap.Error(err))
		}
	}
	return nil
}

func (r *CgoReceiver) suppressEventsComm() {
	comms := r.cfg.ProcessFilterInfo.Comms
	if len(comms) > 0 {
		r.telemetry.Logger.Infof("Filter out process with command: %v", comms)
	}
	for _, comm := range comms {
		csComm := C.CString(comm)
		C.suppressEventsCommForGo(csComm)
		C.free(unsafe.Pointer(csComm))
	}
}

func (r *CgoReceiver) subEvent() error {
	if len(r.cfg.SubscribeInfo) == 0 {
		r.telemetry.Logger.Warn("No events are subscribed by cgoreceiver. Please check your configuration.")
	} else {
		r.telemetry.Logger.Infof("The subscribed events are: %v", r.cfg.SubscribeInfo)
	}
	for _, value := range r.cfg.SubscribeInfo {
		//to do. analyze params filed in the value
		paramsList := make([]CEventParamsForSubscribe, 0)
		var temp CEventParamsForSubscribe
		temp.name = C.CString("terminator")
		paramsList = append(paramsList, temp)
		csName := C.CString(value.Name)
		csCategory := C.CString(value.Category)
		C.subEventForGo(csName, csCategory, (unsafe.Pointer)(&paramsList[0]))
		C.free(unsafe.Pointer(csName))
		C.free(unsafe.Pointer(csCategory))
	}
	return nil
}

func (r *CgoReceiver) StartProfile() error {
	r.telemetry.Logger.Infof("Start Profile")
	if C.startProfile() == 0 {
		return nil
	} else {
		return fmt.Errorf("start Profile failed")
	}
}

func (r *CgoReceiver) StopProfile() error {
	r.telemetry.Logger.Infof("Stop Profile")
	if C.stopProfile() == 0 {
		return nil
	} else {
		return fmt.Errorf("stop Profile failed")
	}
}

func (r *CgoReceiver) ProfileModule() (submodule string, start func() error, stop func() error) {
	return "cgoreceiver", r.StartProfile, r.StopProfile
}

func (r *CgoReceiver) getCaptureStatisticsByInterval(interval time.Duration) {
	timer := time.NewTicker(interval)
	for {
		select {
		case <-timer.C:
			r.getCaptureStatistics()
		case <-r.stopCh:
			return
		}
	}
}

func (r *CgoReceiver) getCaptureStatistics() {
	var captureStatistics C.struct_capture_statistics_for_go
	C.getCaptureStatistics((*C.struct_capture_statistics_for_go)(&captureStatistics))
	r.probeCounterMutex.Lock()
	defer r.probeCounterMutex.Unlock()
	r.probeCounter.evts = int64(captureStatistics.evts)
	r.probeCounter.drops = int64(captureStatistics.drops)
	r.probeCounter.dropsBuffer = int64(captureStatistics.drops_buffer)
	r.probeCounter.dropsPf = int64(captureStatistics.drops_pf)
	r.probeCounter.dropsBug = int64(captureStatistics.drops_bug)
	r.probeCounter.preemptions = int64(captureStatistics.preemptions)
	r.probeCounter.suppressed = int64(captureStatistics.suppressed)
	r.probeCounter.tidsSuppressed = int64(captureStatistics.tids_suppressed)
}

func (r *CgoReceiver) catchSignalUp() {
	C.catchSignalUp()
}
