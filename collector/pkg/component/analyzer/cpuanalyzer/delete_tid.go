// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cpuanalyzer

import (
	"sync"
	"time"
)

type tidDeleteQueue struct {
	queueMutex sync.Mutex
	queue      []deleteTid
}

type deleteTid struct {
	pid      uint32
	tid      uint32
	exitTime time.Time
}

func newTidDeleteQueue() *tidDeleteQueue {
	return &tidDeleteQueue{queue: make([]deleteTid, 0)}
}

func (dq *tidDeleteQueue) GetFront() *deleteTid {
	if len(dq.queue) > 0 {
		return &dq.queue[0]
	}
	return nil
}

func (dq *tidDeleteQueue) Push(elem deleteTid) {
	dq.queue = append(dq.queue, elem)
}

func (dq *tidDeleteQueue) Pop() {
	if len(dq.queue) > 0 {
		dq.queue = dq.queue[1:len(dq.queue)]
	}
}

func (ca *CpuAnalyzer) TidDelete(interval time.Duration, expiredDuration time.Duration) {
	for {
		select {
		case <-ca.stopProfileChan:
			return
		case <-time.After(interval):
			now := time.Now()
			func() {
				ca.tidExpiredQueue.queueMutex.Lock()
				defer ca.tidExpiredQueue.queueMutex.Unlock()
				for {
					elem := ca.tidExpiredQueue.GetFront()
					if elem == nil {
						break
					}
					if elem.exitTime.Add(expiredDuration).After(now) {
						break
					}
					//Delete expired threads (current_time >= thread_exit_time + interval_time).
					func() {
						ca.lock.Lock()
						defer ca.lock.Unlock()
						tidEventsMap := ca.cpuPidEvents[elem.pid]
						if tidEventsMap == nil {
							ca.tidExpiredQueue.Pop()
						} else {
							ca.telemetry.Logger.Debugf("Delete expired thread... pid=%d, tid=%d", elem.pid, elem.tid)
							//fmt.Printf("Go Test: Delete expired thread... pid=%d, tid=%d\n", elem.pid, elem.tid)
							delete(tidEventsMap, elem.tid)
							ca.tidExpiredQueue.Pop()
						}
					}()
				}
			}()
		}
	}
}
