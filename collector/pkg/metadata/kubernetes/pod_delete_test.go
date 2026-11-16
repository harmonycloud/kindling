// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"testing"
	"time"
)

func TestDeleteLoop(t *testing.T) {
	pod := CreatePod(true)
	AddPod(pod)
	verifyIfPodExist(true, t)
	if len(podDeleteQueue) != 0 {
		t.Fatalf("PodDeleteQueue should be 0, but is %d", len(podDeleteQueue))
	}

	DeletePod(pod)
	verifyIfPodExist(true, t)
	if len(podDeleteQueue) != 1 {
		t.Fatalf("PodDeleteQueue should be 1, but is %d", len(podDeleteQueue))
	}

	gracePeriod := time.Millisecond * 500
	stopCh := make(chan struct{})
	go podDeleteLoop(time.Millisecond, gracePeriod, stopCh)
	go func() {
		time.Sleep(time.Millisecond * 50)
		verifyIfPodExist(true, t)
		podDeleteQueueMut.Lock()
		if len(podDeleteQueue) != 1 {
			t.Errorf("PodDeleteQueue should be 1, but is %d", len(podDeleteQueue))
		}
		podDeleteQueueMut.Unlock()

		time.Sleep(gracePeriod + (time.Millisecond * 50))
		verifyIfPodExist(false, t)
		podDeleteQueueMut.Lock()
		if len(podDeleteQueue) != 0 {
			t.Errorf("PodDeleteQueue should be 0, but is %d", len(podDeleteQueue))
		}
		podDeleteQueueMut.Unlock()
		close(stopCh)
	}()
	<-stopCh
}

func verifyIfPodExist(exist bool, t *testing.T) {
	_, ok := GlobalPodInfo.get("CustomNamespace", "deploy-1a2b3c4d-5e6f7")
	if ok != exist {
		t.Fatalf("Finding pod at globalPodInfo. Expect %v, but get %v", false, ok)
	}
	_, ok = MetaDataCache.GetByContainerId("1a2b3c4d5e6f")
	if ok != exist {
		t.Errorf("Finding container using containerid. Expect %v, but get %v", exist, ok)
	}
	_, ok = MetaDataCache.GetContainerByIpPort("172.10.1.2", 80)
	if ok != exist {
		t.Errorf("Finding container using IP:Port. Expect %v, but get %v", exist, ok)
	}
}
