// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cpuanalyzer

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/component"
)

var cnt int

func TestJavaTraceDeleteQueue(t *testing.T) {

	jt := make(map[string]*TransactionIdEvent, 100000)
	testTelemetry := component.NewTelemetryManager().GetGlobalTelemetryTools()
	mycfg := &Config{SegmentSize: 40, JavaTraceDeleteInterval: 15, JavaTraceExpirationTime: 10}
	ca = &CpuAnalyzer{javaTraces: jt, telemetry: testTelemetry, cfg: mycfg}
	ca.javaTraceExpiredQueue = newJavaTraceDeleteQueue()
	go ca.JavaTraceDelete(1*time.Second, 1*time.Second)

	for i := 0; i < 20; i++ {
		ev := new(TransactionIdEvent)
		ev.TraceId = strconv.Itoa(rand.Intn(10000))
		ev.PidString = strconv.Itoa(rand.Intn(10000))
		ev.IsEntry = 1
		key := ev.TraceId + ev.PidString
		ca.javaTraces[key] = ev
		val := new(deleteVal)
		val.key = ev.TraceId + ev.PidString
		val.enterTime = time.Now()
		ca.javaTraceExpiredQueue.Push(*val)
		t.Logf("pid=%s, tid=%s enter time=%s\n", ev.PidString, ev.TraceId, val.enterTime.Format("2006-01-02 15:04:05.000"))
		cnt++
	}
	time.Sleep(5 * time.Second)

	if len(ca.javaTraces) != 0 {
		t.Fatalf("The number of javatraces that entering and exiting the queue is not equal! "+
			"enterCount=%d , len of javatrace is : %d\n", cnt, len(ca.javaTraces))
	} else {
		t.Logf("All javatraces have cleaned normally. enterCount=%d\n", cnt)
	}
}
