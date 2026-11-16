// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cpuanalyzer

import (
	"strconv"
	"testing"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/component"
)

var (
	visit    []deleteTid
	ca       *CpuAnalyzer
	exitTid  map[uint32]int
	enterCnt int
	exitCnt  int
)

const timeDuration time.Duration = 100 * time.Millisecond

func TestDeleteQueue(t *testing.T) {

	cpupidEvents := make(map[uint32]map[uint32]*TimeSegments, 100000)
	testTelemetry := component.NewTelemetryManager().GetGlobalTelemetryTools()
	mycfg := &Config{SegmentSize: 40}
	ca = &CpuAnalyzer{cpuPidEvents: cpupidEvents, telemetry: testTelemetry, cfg: mycfg}

	ca.tidExpiredQueue = newTidDeleteQueue()

	visit = make([]deleteTid, 0)
	exitTid = make(map[uint32]int, 0)

	go ca.TidDelete(5*timeDuration, 4*timeDuration)
	go CheckQueueLoop(t)
	for i := 0; i < 20; i++ {

		ev := new(CpuEvent)
		curTime := time.Now()
		ev.EndTime = uint64(curTime.Add(timeDuration).Nanosecond())
		ev.StartTime = uint64(curTime.Nanosecond())

		//check tid which exist in queue but not in the map
		if i%4 != 0 {
			ca.PutEventToSegments(uint32(i), uint32(i)+5, "threadname"+strconv.Itoa(i+100), ev)
			time.Sleep(timeDuration)
		}

		func() {
			cacheElem := deleteTid{uint32(i), uint32(i) + 5, curTime.Add(timeDuration)}
			ca.trimExitedThread(uint32(i), uint32(i)+5)

			ca.tidExpiredQueue.queueMutex.Lock()
			defer ca.tidExpiredQueue.queueMutex.Unlock()
			visit = append(visit, cacheElem)
		}()

		t.Logf("pid=%d, tid=%d enter time=%s\n", uint32(i), uint32(i)+5, curTime.Format("2006-01-02 15:04:05.000"))
		enterCnt++
		time.Sleep(3 * timeDuration)
	}
	time.Sleep(10 * timeDuration)

	if enterCnt != exitCnt {
		t.Fatalf("The number of threads that entering and exiting the queue is not equal! enterCount=%d, exitCount=%d\n", enterCnt, exitCnt)
	} else {
		t.Logf("All threads have exited normally. enterCount=%d, exitCount=%d\n", enterCnt, exitCnt)
	}

}

func CheckQueueLoop(t *testing.T) {
	for {
		select {
		case <-time.After(timeDuration * 3):
			func() {
				ca.tidExpiredQueue.queueMutex.Lock()
				defer ca.tidExpiredQueue.queueMutex.Unlock()
				queueLen := len(ca.tidExpiredQueue.queue)
				curTime := time.Now()
				for i := 0; i < len(visit); i++ {
					tmpv := visit[i]
					var j int
					for j = 0; j < queueLen; j++ {
						tmpq := ca.tidExpiredQueue.queue[j]
						if tmpv.tid == tmpq.tid {
							if curTime.After(tmpq.exitTime.Add(12 * timeDuration)) {
								t.Errorf("there is a expired threads that is not deleted. pid=%d, tid=%d, exitTime=%s\n", tmpv.pid, tmpv.tid, tmpv.exitTime.Format("2006-01-02 15:04:05.000"))
							}
							break
						}
					}

					if _, exist := exitTid[tmpv.tid]; j >= queueLen && !exist {
						exitTid[tmpv.tid] = 1
						exitCnt++
						t.Logf("pid=%d, tid=%d exit time=%s\n", tmpv.pid, tmpv.tid, curTime.Format("2006-01-02 15:04:05.000"))
					}
				}
			}()
		}
	}
}
