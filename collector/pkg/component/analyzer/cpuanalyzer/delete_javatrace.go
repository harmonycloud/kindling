// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cpuanalyzer

import (
	"sync"
	"time"
)

type javaTraceDeleteQueue struct {
	queueMutex sync.Mutex
	queue      []deleteVal
}

type deleteVal struct {
	key       string
	enterTime time.Time
}

func newJavaTraceDeleteQueue() *javaTraceDeleteQueue {
	return &javaTraceDeleteQueue{queue: make([]deleteVal, 0)}
}

func (dq *javaTraceDeleteQueue) GetFront() *deleteVal {
	if len(dq.queue) > 0 {
		return &dq.queue[0]
	}
	return nil
}

func (dq *javaTraceDeleteQueue) Push(elem deleteVal) {
	dq.queue = append(dq.queue, elem)
}

func (dq *javaTraceDeleteQueue) Pop() {
	if len(dq.queue) > 0 {
		dq.queue = dq.queue[1:len(dq.queue)]
	}
}

func (ca *CpuAnalyzer) JavaTraceDelete(interval time.Duration, expiredDuration time.Duration) {
	for {
		select {
		case <-ca.stopProfileChan:
			return
		case <-time.After(interval):
			ca.telemetry.Logger.Debug("Start regular cleaning of javatrace...")
			now := time.Now()
			func() {
				ca.javaTraceExpiredQueue.queueMutex.Lock()
				defer ca.javaTraceExpiredQueue.queueMutex.Unlock()
				for {
					val := ca.javaTraceExpiredQueue.GetFront()
					if val == nil {
						break
					}
					if val.enterTime.Add(expiredDuration).After(now) {
						break
					}

					func() {
						ca.jtlock.Lock()
						defer ca.jtlock.Unlock()
						event := ca.javaTraces[val.key]
						if event == nil {
							ca.javaTraceExpiredQueue.Pop()
						} else {
							ca.telemetry.Logger.Debugf("Delete expired javatrace... pid=%s, tid=%s", event.PidString, event.TraceId)
							delete(ca.javaTraces, val.key)
							ca.javaTraceExpiredQueue.Pop()
						}
					}()
				}
			}()
		}
	}
}
