// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cgoreceiver

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	otelprocessor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"

	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
)

func runTest(counter eventCounter, workerNum int, loopNum int) {
	wg := sync.WaitGroup{}
	for i := 0; i < workerNum; i++ {
		wg.Add(1)
		go func() {
			runRecordCounter(loopNum, counter)
			wg.Done()
		}()
	}
	wg.Wait()
}

var eventLists = []string{constnames.ReadEvent, constnames.WriteEvent, constnames.ReadvEvent, constnames.WritevEvent,
	constnames.SendToEvent, constnames.RecvFromEvent, constnames.SendMsgEvent, constnames.RecvMsgEvent,
	constnames.GrpcUprobeEvent, constnames.TcpCloseEvent, constnames.TcpRcvEstablishedEvent, constnames.TcpDropEvent,
	constnames.TcpRetransmitSkbEvent, constnames.ConnectEvent, constnames.TcpConnectEvent,
	constnames.TcpSetStateEvent, "another_event"}

func runRecordCounter(loopNum int, counter eventCounter) {
	for i := 0; i < loopNum; i++ {
		for _, name := range eventLists {
			counter.add(name, 1)
		}
	}
}

func assertTest(t *testing.T, counter eventCounter, workerNum int, loopNum int) {
	runTest(counter, workerNum, loopNum)
	expectedNum := workerNum * loopNum
	for key, value := range counter.getStats() {
		if value != int64(expectedNum) {
			t.Errorf("The count of [%s] is expected to be %d, but got %d", key, expectedNum, value)
		}
	}
}

func TestCounterMutexMap(t *testing.T) {
	counter := &mutexMap{m: make(map[string]int64)}
	assertTest(t, counter, 5, 100000)
}

func TestCounterRwMutexMap(t *testing.T) {
	counter := &rwMutexMap{m: make(map[string]int64)}
	assertTest(t, counter, 5, 100000)
}

func TestCounterIntCombination(t *testing.T) {
	counter := &intCombinationCounter{}
	assertTest(t, counter, 5, 100000)
}

func TestCounterRwAtomicMap(t *testing.T) {
	counter := newDynamicStats([]SubEvent{
		{Category: "net", Name: "syscall_exit-writev"},
		{Category: "net", Name: "syscall_exit-readv"},
		{Category: "net", Name: "syscall_exit-write"},
		{Category: "net", Name: "syscall_exit-read"},
		{Category: "net", Name: "syscall_exit-sendto"},
		{Category: "net", Name: "syscall_exit-recvfrom"},
		{Category: "net", Name: "syscall_exit-sendmsg"},
		{Category: "net", Name: "syscall_exit-recvmsg"},
		{Category: "net", Name: "grpc_uprobe"},
		{Name: "kprobe-tcp_close"},
		{Name: "kprobe-tcp_rcv_established"},
		{Name: "kprobe-tcp_drop"},
		{Name: "kprobe-tcp_retransmit_skb"},
		{Name: "syscall_exit-connect"},
		{Name: "kretprobe-tcp_connect"},
		{Name: "kprobe-tcp_set_state"},
	})
	assertTest(t, counter, 5, 100000)
}

func BenchmarkCounterMutexMap(b *testing.B) {
	counter := &mutexMap{m: make(map[string]int64)}
	initOtelCounterObserver(counter)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runTest(counter, 5, 1000)
	}
}

func BenchmarkCounterRwMutexMap(b *testing.B) {
	counter := &rwMutexMap{m: make(map[string]int64)}
	initOtelCounterObserver(counter)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runTest(counter, 5, 1000)
	}
}

func BenchmarkCounterIntCombination(b *testing.B) {
	counter := &intCombinationCounter{}
	initOtelCounterObserver(counter)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runTest(counter, 5, 1000)
	}
}

func BenchmarkCounterOtelCounter(b *testing.B) {
	counter := newOtelRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runTest(counter, 5, 1000)
	}
}

func BenchmarkCounterRwAtomicMap(b *testing.B) {
	counter := newDynamicStats([]SubEvent{
		{Category: "net", Name: "syscall_exit-writev"},
		{Category: "net", Name: "syscall_exit-readv"},
		{Category: "net", Name: "syscall_exit-write"},
		{Category: "net", Name: "syscall_exit-read"},
		{Category: "net", Name: "syscall_exit-sendto"},
		{Category: "net", Name: "syscall_exit-recvfrom"},
		{Category: "net", Name: "syscall_exit-sendmsg"},
		{Category: "net", Name: "syscall_exit-recvmsg"},
		{Name: "kprobe-tcp_close"},
		{Name: "kprobe-tcp_rcv_established"},
		{Name: "kprobe-tcp_drop"},
		{Name: "kprobe-tcp_retransmit_skb"},
		{Name: "syscall_exit-connect"},
		{Name: "kretprobe-tcp_connect"},
		{Name: "kprobe-tcp_set_state"},
	})
	initOtelCounterObserver(counter)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runTest(counter, 5, 1000)
	}
}

// It's not practical to implement with sync.Map,
// because you still need to lock the value when increasing it.
// type syncMap struct {
// 	m sync.Map
// 	mutex sync.Mutex
// }

type mutexMap struct {
	m     map[string]int64
	mutex sync.Mutex
}

func (m *mutexMap) add(name string, value int64) {
	m.mutex.Lock()
	v := m.m[name]
	m.m[name] = v + value
	m.mutex.Unlock()
}
func (m *mutexMap) getStats() map[string]int64 {
	m.mutex.Lock()
	ret := make(map[string]int64, len(m.m))
	for k, v := range m.m {
		ret[k] = v
	}
	m.mutex.Unlock()
	return ret
}

type rwMutexMap struct {
	m     map[string]int64
	mutex sync.RWMutex
}

func (m *rwMutexMap) add(name string, value int64) {
	m.mutex.Lock()
	v := m.m[name]
	m.m[name] = v + value
	m.mutex.Unlock()
}
func (m *rwMutexMap) getStats() map[string]int64 {
	m.mutex.RLock()
	ret := make(map[string]int64, len(m.m))
	for k, v := range m.m {
		ret[k] = v
	}
	m.mutex.RUnlock()
	return ret
}

type otelRecorder struct {
	otelCounter metric.Int64Counter
}

func newOtelRecorder() *otelRecorder {
	meter := initOpentelemetry()
	return &otelRecorder{otelCounter: meter.NewInt64Counter("event_counter_total")}
}
func (m *otelRecorder) add(name string, value int64) {
	m.otelCounter.Add(context.Background(), value, attribute.String("name", name))
}
func (m *otelRecorder) getStats() map[string]int64 {
	return nil
}

func initOpentelemetry() metric.MeterMust {
	devNullWriter, _ := os.Open(os.DevNull)
	exp, _ := stdoutmetric.New(stdoutmetric.WithWriter(devNullWriter))

	cont := controller.New(
		otelprocessor.NewFactory(selector.NewWithInexpensiveDistribution(), exp),
		controller.WithExporter(exp),
		controller.WithCollectPeriod(100*time.Millisecond),
	)
	_ = cont.Start(context.Background())
	return metric.Must(cont.Meter("kindling"))
}

type intCombinationCounter struct {
	read              int64
	write             int64
	readv             int64
	writev            int64
	sendTo            int64
	recvFrom          int64
	sendMsg           int64
	recvMsg           int64
	connect           int64
	grpcUprobe        int64
	tcpClose          int64
	tcpRcvEstablished int64
	tcpDrop           int64
	tcpRetransmitSkb  int64
	tcpConnect        int64
	tcpSetState       int64
	other             int64
}

func (i *intCombinationCounter) add(name string, value int64) {
	switch name {
	case constnames.ReadEvent:
		atomic.AddInt64(&i.read, value)
	case constnames.WriteEvent:
		atomic.AddInt64(&i.write, value)
	case constnames.ReadvEvent:
		atomic.AddInt64(&i.readv, value)
	case constnames.WritevEvent:
		atomic.AddInt64(&i.writev, value)
	case constnames.SendToEvent:
		atomic.AddInt64(&i.sendTo, value)
	case constnames.RecvFromEvent:
		atomic.AddInt64(&i.recvFrom, value)
	case constnames.SendMsgEvent:
		atomic.AddInt64(&i.sendMsg, value)
	case constnames.RecvMsgEvent:
		atomic.AddInt64(&i.recvMsg, value)
	case constnames.GrpcUprobeEvent:
		atomic.AddInt64(&i.grpcUprobe, value)
	case constnames.TcpCloseEvent:
		atomic.AddInt64(&i.tcpClose, value)
	case constnames.TcpRcvEstablishedEvent:
		atomic.AddInt64(&i.tcpRcvEstablished, value)
	case constnames.TcpDropEvent:
		atomic.AddInt64(&i.tcpDrop, value)
	case constnames.TcpRetransmitSkbEvent:
		atomic.AddInt64(&i.tcpRetransmitSkb, value)
	case constnames.ConnectEvent:
		atomic.AddInt64(&i.connect, value)
	case constnames.TcpConnectEvent:
		atomic.AddInt64(&i.tcpConnect, value)
	case constnames.TcpSetStateEvent:
		atomic.AddInt64(&i.tcpSetState, value)
	default:
		atomic.AddInt64(&i.other, value)
	}
}

func (i *intCombinationCounter) getStats() map[string]int64 {
	ret := make(map[string]int64)
	ret[constnames.ReadEvent] = atomic.LoadInt64(&i.read)
	ret[constnames.WriteEvent] = atomic.LoadInt64(&i.write)
	ret[constnames.ReadvEvent] = atomic.LoadInt64(&i.readv)
	ret[constnames.WritevEvent] = atomic.LoadInt64(&i.writev)
	ret[constnames.SendToEvent] = atomic.LoadInt64(&i.sendTo)
	ret[constnames.RecvFromEvent] = atomic.LoadInt64(&i.recvFrom)
	ret[constnames.SendMsgEvent] = atomic.LoadInt64(&i.sendMsg)
	ret[constnames.RecvMsgEvent] = atomic.LoadInt64(&i.recvMsg)
	ret[constnames.GrpcUprobeEvent] = atomic.LoadInt64(&i.grpcUprobe)
	ret[constnames.TcpCloseEvent] = atomic.LoadInt64(&i.tcpClose)
	ret[constnames.TcpRcvEstablishedEvent] = atomic.LoadInt64(&i.tcpRcvEstablished)
	ret[constnames.TcpRetransmitSkbEvent] = atomic.LoadInt64(&i.tcpRetransmitSkb)
	ret[constnames.ConnectEvent] = atomic.LoadInt64(&i.connect)
	ret[constnames.TcpConnectEvent] = atomic.LoadInt64(&i.tcpConnect)
	ret[constnames.TcpSetStateEvent] = atomic.LoadInt64(&i.tcpSetState)
	ret[constnames.OtherEvent] = atomic.LoadInt64(&i.other)
	return ret
}

func initOtelCounterObserver(counter eventCounter) {
	meter := initOpentelemetry()
	meter.NewInt64CounterObserver("event_counter_total",
		func(ctx context.Context, result metric.Int64ObserverResult) {
			metrics := counter.getStats()
			for name, value := range metrics {
				result.Observe(value, attribute.String("name", name))
			}
		})
}
