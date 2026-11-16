// Copyright 2020 OpenTelemetry Authors
// Source: https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/processor/k8sattributesprocessor/internal/kube/client.go
// Modification: Use deletedPodInfo as deleted elements and delete our cache map as needed.

package kubernetes

import (
	"log"
	"sync"
	"time"
)

var (
	podDeleteQueueMut sync.Mutex
	podDeleteQueue    []deleteRequest
)

type deleteRequest struct {
	podInfo *deletedPodInfo
	ts      time.Time
}

type deletedPodInfo struct {
	uid          string
	name         string
	namespace    string
	containerIds []string
	ip           string
	ports        []int32
	hostIp       string
	hostPorts    []int32
}

// deleteLoop deletes pods from cache periodically.
func podDeleteLoop(interval time.Duration, gracePeriod time.Duration, stopCh chan struct{}) {
	// This loop runs after N seconds and deletes pods from cache.
	// It iterates over the delete queue and deletes all that aren't
	// in the grace period anymore.
	for {
		select {
		case <-time.After(interval):
			var cutoff int
			now := time.Now()
			podDeleteQueueMut.Lock()
			for i, d := range podDeleteQueue {
				if d.ts.Add(gracePeriod).After(now) {
					break
				}
				cutoff = i + 1
			}
			toDelete := podDeleteQueue[:cutoff]
			podDeleteQueue = podDeleteQueue[cutoff:]
			podDeleteQueueMut.Unlock()
			for _, d := range toDelete {
				deletePodInfo(d.podInfo)
			}

		case <-stopCh:
			return
		}
	}
}

func deletePodInfo(podInfo *deletedPodInfo) {
	if podInfo.name != "" {
		deletePodInfo, ok := GlobalPodInfo.delete(podInfo.namespace, podInfo.name)
		if ok && deletePodInfo != nil && localWorkloadMap != nil {
			localWorkloadMap.delete(deletePodInfo.Namespace, deletePodInfo.WorkloadName)
		} else if ok && (deletePodInfo == nil || localWorkloadMap == nil) {
			// don't know why this happen , print more message to console
			log.Println("unexpected error happened when delete PodInfo from cache, which could lead to a segmentation violation")
			log.Println("detailed info print below:")
			log.Printf("\tdeletedPodInfo: %+v", podInfo)
			if deletePodInfo != nil {
				log.Printf("\tmatchedPodInCache: %+v", *deletePodInfo)
			}
			if localWorkloadMap != nil {
				log.Printf("\tlocalWorkloadMap is not nullptr")
			} else {
				log.Printf("\tlocalWorkloadMap is nullptr")
			}
		}
	}
	if len(podInfo.containerIds) != 0 {
		for i := 0; i < len(podInfo.containerIds); i++ {
			// Assume that container id can't be reused in a few seconds
			MetaDataCache.DeleteByContainerId(podInfo.containerIds[i])
		}
	}
	if podInfo.ip != "" && len(podInfo.ports) != 0 {
		for _, port := range podInfo.ports {
			containerInfo, ok := MetaDataCache.GetContainerByIpPort(podInfo.ip, uint32(port))
			if !ok {
				continue
			}
			// PodIP:Port can be reused in a few seconds, so we check its UID
			if containerInfo.RefPodInfo.UID == podInfo.uid {
				MetaDataCache.DeleteContainerByIpPort(podInfo.ip, uint32(port))
			}
		}
	}
	if podInfo.hostIp != "" && len(podInfo.hostPorts) != 0 {
		for _, port := range podInfo.hostPorts {
			containerInfo, ok := MetaDataCache.GetContainerByHostIpPort(podInfo.hostIp, uint32(port))
			if !ok {
				continue
			}
			// PodIP:Port can be reused in a few seconds, so we check its UID
			if containerInfo.RefPodInfo.UID == podInfo.uid {
				MetaDataCache.DeleteContainerByHostIpPort(podInfo.hostIp, uint32(port))
			}
		}
	}
}
