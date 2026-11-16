// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"reflect"
	"testing"
)

func TestK8sMetaDataCache_AddPodByIpPort(t *testing.T) {
	containerInfo := &K8sContainerInfo{
		ContainerId: "123abc456def",
		Name:        "containername",
		RefPodInfo: &K8sPodInfo{
			Ip:            "10.1.11.2",
			PodName:       "pod1-xxx-123",
			WorkloadKind:  "Deployment",
			WorkloadName:  "pod1",
			Namespace:     "default",
			NodeName:      "node1",
			isHostNetwork: false,
			ServiceInfo:   nil,
		},
	}
	var port uint32 = 80
	portContainerInfo := make(map[uint32]*K8sContainerInfo)
	portContainerInfo[port] = containerInfo
	cacheManual := New()
	cacheManual.IpContainerInfo[containerInfo.RefPodInfo.Ip] = portContainerInfo

	cacheFunc := New()
	cacheFunc.AddContainerByIpPort(containerInfo.RefPodInfo.Ip, port, containerInfo)
	if !reflect.DeepEqual(cacheManual, cacheFunc) {
		t.Errorf("Expected %s,\n but got %s", cacheManual, cacheFunc)
	}

	containerInfoDiff := &K8sContainerInfo{
		ContainerId: "123abc456def",
		Name:        "containername",
		RefPodInfo: &K8sPodInfo{
			Ip:            "10.1.11.2",
			PodName:       "pod1-xxx-456",
			WorkloadKind:  "Deployment",
			WorkloadName:  "pod1",
			Namespace:     "default",
			NodeName:      "node1",
			isHostNetwork: false,
			ServiceInfo:   nil,
		},
	}
	cacheFunc.AddContainerByIpPort(containerInfoDiff.RefPodInfo.Ip, port, containerInfoDiff)
	if reflect.DeepEqual(cacheManual, cacheFunc) {
		t.Errorf("Expected %s,\n but got %s", cacheManual, cacheFunc)
	}
}

func TestK8sServiceInfo_emptySelf(t *testing.T) {
	serviceInfo := &K8sServiceInfo{
		Ip:          "192.168.1.2",
		ServiceName: "Service1",
		Namespace:   "Namespace1",
		isNodePort:  true,
		Selector: map[string]string{
			"s1": "v1",
			"s2": "v2",
		},
		WorkloadKind: "deployment",
		WorkloadName: "deploy",
	}
	emptyServiceInfo := &K8sServiceInfo{}
	serviceInfo.emptySelf()
	if !reflect.DeepEqual(serviceInfo, emptyServiceInfo) {
		t.Errorf("serviceInfo should be empty, but get %v", serviceInfo)
	}
}

func TestK8sMetaDataCache_DeleteServiceByIpPort(t *testing.T) {
	serviceInfo := &K8sServiceInfo{
		Ip:          "192.168.2.1",
		ServiceName: "service",
		Namespace:   "custom-namespace",
		isNodePort:  false,
		Selector:    nil,
	}
	MetaDataCache.AddServiceByIpPort("192.168.2.1", 80, serviceInfo)
	if len(MetaDataCache.IpServiceInfo) != 1 &&
		len(MetaDataCache.IpServiceInfo["192.168.2.1"]) != 1 {
		t.Fatalf("no service was added")
	}
	MetaDataCache.DeleteServiceByIpPort("192.168.2.1", 80)
	if len(MetaDataCache.IpServiceInfo) != 0 {
		t.Fatalf("cache is not empty after deleting service")
	}
}
