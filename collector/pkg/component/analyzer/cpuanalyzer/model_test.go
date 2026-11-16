// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cpuanalyzer

import (
	"encoding/json"
	"testing"
)

func TestMarshalSomeFields(t *testing.T) {
	raw := "{\"cpuEvents\": [{\"stack\": \"\",\"log\": \"\",\"typeSpecs\": \"56651993,4709969826,2611411,1315845777,\",\"onInfo\": \"net@write@10.10.103.148:55738->10.10.103.148:9999@1660790509723744474@77660@1@296#net@write@10.10.103.148:55738->10.10.103.148:9999@1660790509724168721@52734@1@5|net@write@10.10.103.148:55744->10.10.103.148:9999@1660790514436177779@90940@1@297#net@write@10.10.103.148:55744->10.10.103.148:9999@1660790514436641680@71964@1@5|\",\"timeType\": \"0,3,0,3,\",\"startTime\": 1660790509667848205,\"endTime\": 1660790515752931068,\"runqLatency\": \"0,0,\",\"offInfo\": \"futex@addr140553945667444@1660790509724485534@4710000334|futex@addr140553945667444@1660790514437063928@1315884467|\"}],\"transactionIds\": [{\"traceId\": \"88786150748bd45008f7510696bce47c4^389\",\"isEntry\": 1,\"timestamp\": 1660790514434929615},{\"traceId\": \"88786150748bd45008f7510696bce47c4^389\",\"isEntry\": 0,\"timestamp\": 1660790514436590940}],\"pid\": 13403,\"startTime\": 1660790514000000000,\"endTime\": 1660790515000000000,\"javaFutexEvents\": [{\"dataValue\": \"kd-jf@1660790509724364808!1660790514434516975!13600!f501cc48!UnsafePark!XNIO-1 task-3!4710152167!13529!park.Lsun/misc/Unsafe;parkNanos.Ljava/util/concurrent/locks/LockSupport;park.Lorg/jboss/threads/EnhancedQueueExecutor$PoolThreadNode;run.Lorg/jboss/threads/EnhancedQueueExecutor$ThreadBody;run.Ljava/lang/Thread;!\\n\",\"startTime\": 1660790509724364808,\"endTime\": 1660790514434516975},{\"dataValue\": \"kd-jf@1660790514436874909!1660790515752985916!13600!f501cc48!UnsafePark!XNIO-1 task-3!1316111007!-1!park.Lsun/misc/Unsafe;parkNanos.Ljava/util/concurrent/locks/LockSupport;park.Lorg/jboss/threads/EnhancedQueueExecutor$PoolThreadNode;run.Lorg/jboss/threads/EnhancedQueueExecutor$ThreadBody;run.Ljava/lang/Thread;!\\n\",\"startTime\": 1660790514436874909,\"endTime\": 1660790515752985916}],\"tid\": 13600,\"threadName\": \"XNIO-1 task-3\",\"IsSend\": 1,\"indexTimestamp\": \"2022-08-18 10:41:56.338762903 +0800 CST m=+511.580582242\"}"
	segment := &Segment{}
	err := json.Unmarshal([]byte(raw), segment)
	if err != nil {
		t.Failed()
	}
	cpuEvents, err := json.Marshal(segment.CpuEvents)
	if err != nil {
		t.Failed()
	}
	t.Log(string(cpuEvents))
}

func TestCPUType_MarshalJSON(t *testing.T) {
	ct := CPUType_FILE
	data, err := ct.MarshalJSON()
	if err != nil {
		t.Failed()
	}
	t.Log(string(data))
}

func TestCPUType_UnmarshalJSON(t *testing.T) {
	data := []byte{49}

	var ct CPUType
	err := ct.UnmarshalJSON(data)
	if err != nil {
		t.Failed()
	}
	t.Log(ct)
}
