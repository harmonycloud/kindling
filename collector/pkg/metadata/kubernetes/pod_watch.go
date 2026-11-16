// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"fmt"
	"os"
	_ "path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	_ "k8s.io/client-go/tools/clientcmd"
	_ "k8s.io/client-go/util/homedir"

	"github.com/Kindling-project/kindling/collector/pkg/compare"
	"github.com/Kindling-project/kindling/collector/pkg/model"
	"github.com/Kindling-project/kindling/collector/pkg/model/constlabels"
	"github.com/Kindling-project/kindling/collector/pkg/model/constnames"
)

type podMap struct {
	// namespace:
	//   podName: podInfo{}
	Info  map[string]map[string]*K8sPodInfo
	mutex sync.RWMutex
}

type workloadMap struct {
	Info  map[string]map[string]*workloadInfo
	mutex sync.RWMutex
}

type workloadInfo struct {
	Namespace    string
	WorkloadName string
	WorkloadKind string
}

var GlobalPodInfo = newPodMap()

// localWorkloadMap only stores the workload whose pods are in the local Node.
// The workload metadata will be sent to prometheus and used to filter metrics.
var localWorkloadMap = newWorkloadMap()

func GetWorkloadDataGroup() []*model.DataGroup {
	localWorkloadMap.mutex.RLock()
	dataGroups := make([]*model.DataGroup, 0)
	for _, workloadInfoMap := range localWorkloadMap.Info {
		for _, workloadInfo := range workloadInfoMap {
			dataGroups = append(dataGroups, &model.DataGroup{
				Name: constnames.K8sWorkloadMetricGroupName,
				Metrics: []*model.Metric{
					model.NewIntMetric(constnames.K8sWorkLoadMetricName, 1),
				},
				Labels: model.NewAttributeMapWithValues(map[string]model.AttributeValue{
					constlabels.Namespace:    model.NewStringValue(workloadInfo.Namespace),
					constlabels.WorkloadKind: model.NewStringValue(workloadInfo.WorkloadKind),
					constlabels.WorkloadName: model.NewStringValue(workloadInfo.WorkloadName),
				}),
			})
		}
	}
	localWorkloadMap.mutex.RUnlock()
	return dataGroups
}

func newPodMap() *podMap {
	return &podMap{
		Info:  make(map[string]map[string]*K8sPodInfo),
		mutex: sync.RWMutex{},
	}
}

func newWorkloadMap() *workloadMap {
	return &workloadMap{
		Info:  make(map[string]map[string]*workloadInfo),
		mutex: sync.RWMutex{},
	}
}

func (m *podMap) add(info *K8sPodInfo) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	podInfoMap, ok := m.Info[info.Namespace]
	if !ok {
		podInfoMap = make(map[string]*K8sPodInfo)
	}
	podInfoMap[info.PodName] = info
	m.Info[info.Namespace] = podInfoMap
}

func (m *podMap) delete(namespace string, name string) (*K8sPodInfo, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	podInfoMap, ok := m.Info[namespace]
	var podInfo *K8sPodInfo
	if !ok {
		return nil, false
	}
	podInfo, ok = podInfoMap[name]
	if !ok {
		return nil, false
	}
	delete(podInfoMap, name)
	return podInfo, true
}

func (m *workloadMap) add(info *workloadInfo) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	workloadInfoMap, ok := m.Info[info.Namespace]
	if !ok {
		workloadInfoMap = make(map[string]*workloadInfo)
	}
	workloadInfoMap[info.WorkloadName] = info
	m.Info[info.Namespace] = workloadInfoMap
}

func (m *workloadMap) delete(namespace string, name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	workloadInfoMap, ok := m.Info[namespace]
	if ok {
		delete(workloadInfoMap, name)
	}
}

func (m *podMap) get(namespace string, name string) (*K8sPodInfo, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	podInfoMap, ok := m.Info[namespace]
	if !ok {
		return nil, false
	}
	podInfo, ok := podInfoMap[name]
	if !ok {
		return nil, false
	}
	return podInfo, true
}

// getPodsMatchSelectors gets K8sPodInfo(s) whose labels match with selectors in such namespace.
// Return empty slice if not found. Note there may be multiple match.
func (m *podMap) getPodsMatchSelectors(namespace string, selectors map[string]string) []*K8sPodInfo {
	retPodInfoSlice := make([]*K8sPodInfo, 0)
	if len(selectors) == 0 {
		return retPodInfoSlice
	}
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	podInfoMap, ok := m.Info[namespace]
	if !ok {
		return retPodInfoSlice
	}
	for _, podInfo := range podInfoMap {
		if SelectorsMatchLabels(selectors, podInfo.Labels) {
			retPodInfoSlice = append(retPodInfoSlice, podInfo)
		}
	}
	return retPodInfoSlice
}

func PodWatch(clientSet *kubernetes.Clientset, graceDeletePeriod time.Duration, handler cache.ResourceEventHandler) {
	stopper := make(chan struct{})
	defer close(stopper)

	factory := informers.NewSharedInformerFactory(clientSet, 0)
	podInformer := factory.Core().V1().Pods()
	informer := podInformer.Informer()
	defer runtime.HandleCrash()

	// Start informer, list & watch
	go factory.Start(stopper)

	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	go podDeleteLoop(10*time.Second, graceDeletePeriod, stopper)

	if handler == nil {
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    AddPod,
			UpdateFunc: UpdatePod,
			DeleteFunc: DeletePod,
		})
	} else {
		informer.AddEventHandler(handler)
	}
	// TODO: use workqueue to avoid blocking
	<-stopper
}

func AddPod(obj interface{}) {
	pod := obj.(*corev1.Pod)

	// Find the controller workload of the pod
	rsUpdateMutex.RLock()
	workloadTypeTmp, workloadNameTmp := getControllerKindName(pod)
	rsUpdateMutex.RUnlock()

	// Find one of the services of the pod
	serviceInfoSlice := GlobalServiceInfo.GetServiceMatchLabels(pod.Namespace, pod.Labels)
	var serviceInfo *K8sServiceInfo
	if len(serviceInfoSlice) == 0 {
		serviceInfo = nil
	} else {
		// When span target is a kind of service, workload should also be filled in case to display
		// the real topology in model level. Service is considered as abstract level, instead.
		// So here the information of workload is assigned to serviceInfo.
		for _, service := range serviceInfoSlice {
			service.WorkloadKind = workloadTypeTmp
			service.WorkloadName = workloadNameTmp
		}
		// Only one of the matched services is cared, here we get the first one
		serviceInfo = serviceInfoSlice[0]
	}

	var cachePodInfo = &K8sPodInfo{
		UID:           string(pod.UID),
		Ip:            pod.Status.PodIP,
		Namespace:     pod.Namespace,
		PodName:       pod.Name,
		Ports:         make([]int32, 0),
		HostPorts:     make([]int32, 0),
		ContainerIds:  make([]string, 0, 2),
		Labels:        pod.Labels,
		WorkloadKind:  workloadTypeTmp,
		WorkloadName:  workloadNameTmp,
		NodeName:      pod.Spec.NodeName,
		NodeAddress:   pod.Status.HostIP,
		isHostNetwork: pod.Spec.HostNetwork,
		ServiceInfo:   serviceInfo,
	}

	// Add containerId map
	for _, containerStatus := range pod.Status.ContainerStatuses {
		shortenContainerId := TruncateContainerId(containerStatus.ContainerID)
		if shortenContainerId == "" {
			continue
		}
		cachePodInfo.ContainerIds = append(cachePodInfo.ContainerIds, shortenContainerId)
		containerInfo := &K8sContainerInfo{
			ContainerId: shortenContainerId,
			Name:        containerStatus.Name,
			RefPodInfo:  cachePodInfo,
		}
		MetaDataCache.AddByContainerId(shortenContainerId, containerInfo)
	}

	// Add pod IP and port map
	if len(pod.Status.PodIP) > 0 {
		for _, tmpContainer := range pod.Spec.Containers {
			containerInfo := &K8sContainerInfo{
				Name:        tmpContainer.Name,
				HostPortMap: make(map[int32]int32),
				RefPodInfo:  cachePodInfo,
			}
			// Not specifying a port DOES NOT prevent that port from being exposed.
			// So Ports could be empty, if so we only record its IP address.
			if len(tmpContainer.Ports) == 0 {
				// If there is more than one container that doesn't specify a port,
				// we would rather get an empty name than get an incorrect one.
				if len(pod.Spec.Containers) > 1 {
					containerInfo.Name = ""
				}
				// When there are many containers in one pod and only part of them have ports,
				// the containers at the back will overwrite the ones at the front here.
				MetaDataCache.AddContainerByIpPort(pod.Status.PodIP, 0, containerInfo)
				cachePodInfo.Ports = append(cachePodInfo.Ports, 0)
				continue
			}
			for _, port := range tmpContainer.Ports {
				cachePodInfo.Ports = append(cachePodInfo.Ports, port.ContainerPort)
				// If hostPort is specified, add the container using HostIP and HostPort
				if port.HostPort != 0 {
					containerInfo.HostPortMap[port.HostPort] = port.ContainerPort
					cachePodInfo.HostPorts = append(cachePodInfo.HostPorts, port.HostPort)
					MetaDataCache.AddContainerByHostIpPort(pod.Status.HostIP, uint32(port.HostPort), containerInfo)
				}
				MetaDataCache.AddContainerByIpPort(pod.Status.PodIP, uint32(port.ContainerPort), containerInfo)
			}
		}
	}
	GlobalPodInfo.add(cachePodInfo)
	nodeName, _ := os.LookupEnv("MY_NODE_NAME")
	//workloadMap only restore the workload in this machine
	if cachePodInfo.NodeName == nodeName {
		localWorkloadMap.add(&workloadInfo{
			Namespace:    cachePodInfo.Namespace,
			WorkloadName: cachePodInfo.WorkloadName,
			WorkloadKind: cachePodInfo.WorkloadKind,
		})
	}
}

func getControllerKindName(pod *corev1.Pod) (workloadKind string, workloadName string) {
	for _, owner := range pod.OwnerReferences {
		// only care about the controller
		if owner.Controller == nil || *owner.Controller != true {
			continue
		}
		if owner.Kind == ReplicaSetKind {
			// The owner of Pod is ReplicaSet, and it is Workload such as Deployment for ReplicaSet.
			// Therefore, find ReplicaSet's name in 'globalRsInfo' to find which kind of workload
			// the Pod belongs to.
			if workload, ok := GlobalRsInfo.GetOwnerReference(mapKey(pod.Namespace, owner.Name)); ok {
				workloadKind = CompleteGVK(workload.APIVersion, strings.ToLower(workload.Kind))
				workloadName = workload.Name
			} else {
				// If not found in 'globalRsInfo', just use 'Deployment' as the workload kind.
				workloadKind = DeploymentKind
				workloadName = extractDeploymentName(owner.Name)
			}
			return
		}
		// If the owner of pod is not ReplicaSet
		workloadKind = CompleteGVK(owner.APIVersion, strings.ToLower(owner.Kind))
		workloadName = owner.Name
		return
	}
	return
}

var rRegex = regexp.MustCompile(`^(.*)-[0-9a-z]+$`)

func extractDeploymentName(replicaSetName string) string {
	parts := rRegex.FindStringSubmatch(replicaSetName)
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

func UpdatePod(objOld interface{}, objNew interface{}) {
	oldPod := objOld.(*corev1.Pod)
	newPod := objNew.(*corev1.Pod)
	if oldPod.ResourceVersion == newPod.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same pod will always have different RVs.
		return
	}

	oldCachePod, ok := GlobalPodInfo.get(oldPod.Namespace, oldPod.Name)
	if !ok {
		AddPod(objNew)
		return
	}
	// Always override the old pod in the cache
	AddPod(objNew)

	// Delay delete the pod using the difference between the old pod and the new one
	deletedPodInfo := &deletedPodInfo{
		uid:          string(oldPod.UID),
		name:         "",
		namespace:    oldPod.Namespace,
		containerIds: nil,
		ip:           oldPod.Status.PodIP,
		ports:        nil,
		hostIp:       oldPod.Status.HostIP,
		hostPorts:    nil,
	}

	if oldPod.Name != newPod.Name {
		deletedPodInfo.name = oldPod.Name
	}

	// Check the containers' ID
	newContainerIds := make([]string, 0)
	for _, containerStatus := range newPod.Status.ContainerStatuses {
		shortenContainerId := TruncateContainerId(containerStatus.ContainerID)
		if shortenContainerId == "" {
			continue
		}
		newContainerIds = append(newContainerIds, shortenContainerId)
	}
	containerIdCompare := compare.NewStringSlice(oldCachePod.ContainerIds, newContainerIds)
	containerIdCompare.Compare()
	deletedPodInfo.containerIds = containerIdCompare.GetRemovedElements()

	// Check the ports specified.
	newPorts := make([]int32, 0)
	newHostPorts := make([]int32, 0)
	for _, container := range newPod.Spec.Containers {
		if len(container.Ports) == 0 {
			newPorts = append(newPorts, 0)
			continue
		}
		for _, port := range container.Ports {
			newPorts = append(newPorts, port.ContainerPort)
			// If hostPort is specified, add the container using HostIP and HostPort
			if port.HostPort != 0 {
				newHostPorts = append(newHostPorts, port.HostPort)
			}
		}
	}

	if oldPod.Status.PodIP != newPod.Status.PodIP {
		deletedPodInfo.ports = oldCachePod.Ports
	} else {
		portsCompare := compare.NewInt32Slice(oldCachePod.Ports, newPorts)
		portsCompare.Compare()
		deletedPodInfo.ports = portsCompare.GetRemovedElements()
	}

	if oldPod.Status.HostIP != newPod.Status.HostIP {
		deletedPodInfo.hostPorts = oldCachePod.HostPorts
	} else {
		hostPortsCompare := compare.NewInt32Slice(oldCachePod.HostPorts, newHostPorts)
		hostPortsCompare.Compare()
		deletedPodInfo.hostPorts = hostPortsCompare.GetRemovedElements()
	}

	// Wait for a few seconds to remove the cache data
	podDeleteQueueMut.Lock()
	podDeleteQueue = append(podDeleteQueue, deleteRequest{
		podInfo: deletedPodInfo,
		ts:      time.Now(),
	})
	podDeleteQueueMut.Unlock()
}

func DeletePod(obj interface{}) {
	// Maybe get DeletedFinalStateUnknown instead of *corev1.Pod.
	// Fix https://github.com/KindlingProject/kindling/issues/445
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		deletedState, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return
		}
		pod, ok = deletedState.Obj.(*corev1.Pod)
		if !ok {
			return
		}
	}
	podInfo := &deletedPodInfo{
		uid:          string(pod.UID),
		name:         pod.Name,
		namespace:    pod.Namespace,
		containerIds: make([]string, 0),
		ip:           pod.Status.PodIP,
		ports:        make([]int32, 0),
		hostIp:       pod.Status.HostIP,
		hostPorts:    make([]int32, 0),
	}

	for _, containerStatus := range pod.Status.ContainerStatuses {
		shortenContainerId := TruncateContainerId(containerStatus.ContainerID)
		if shortenContainerId == "" {
			continue
		}
		podInfo.containerIds = append(podInfo.containerIds, shortenContainerId)
	}

	for _, container := range pod.Spec.Containers {
		if len(container.Ports) == 0 {
			podInfo.ports = append(podInfo.ports, 0)
			continue
		}
		for _, port := range container.Ports {
			podInfo.ports = append(podInfo.ports, port.ContainerPort)
			// If hostPort is specified, add the container using HostIP and HostPort
			if port.HostPort != 0 {
				podInfo.hostPorts = append(podInfo.hostPorts, port.HostPort)
			}
		}
	}
	// Wait for a few seconds to remove the cache data
	podDeleteQueueMut.Lock()
	podDeleteQueue = append(podDeleteQueue, deleteRequest{
		podInfo: podInfo,
		ts:      time.Now(),
	})
	podDeleteQueueMut.Unlock()
}

// TruncateContainerId slices the input containerId into two parts separated by "://",
// and return the first 12 bytes at most of the second part.
//
// If no second part found, return empty string.
func TruncateContainerId(containerId string) string {
	sep := "://"
	separated := strings.SplitN(containerId, sep, 2)
	if len(separated) < 2 {
		return ""
	}
	secondString := separated[1]
	l := len(secondString)
	if l > 12 {
		l = 12
	}
	return secondString[0:l]
}
