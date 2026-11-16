// Copyright (C) 2022-2024 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"encoding/json"
	"fmt"
	"sync"
)

type SetPreprocessingMetaDataCache func(cache *K8sMetaDataCache, nodeMap *NodeMap, serviceMap *ServiceMap, rsMap *ReplicaSetMap)

type K8sContainerInfo struct {
	ContainerId string
	Name        string
	HostPortMap map[int32]int32
	RefPodInfo  *K8sPodInfo
}

type K8sPodInfo struct {
	UID          string
	Ip           string
	PodName      string
	Ports        []int32
	HostPorts    []int32
	ContainerIds []string
	Labels       map[string]string
	// TODO: There may be multiple kinds of workload or services for the same pod
	WorkloadKind  string
	WorkloadName  string
	Namespace     string
	NodeName      string
	NodeAddress   string
	isHostNetwork bool
	ServiceInfo   *K8sServiceInfo
}

type K8sServiceInfo struct {
	Ip          string
	ServiceName string
	Namespace   string
	isNodePort  bool
	Selector    map[string]string
	// TODO: How to delete the workload info when it is deleted?
	WorkloadKind string
	WorkloadName string
}

func (s *K8sServiceInfo) emptySelf() {
	s.Ip = ""
	s.ServiceName = ""
	s.Namespace = ""
	s.isNodePort = false
	s.Selector = nil
	s.WorkloadKind = ""
	s.WorkloadName = ""
}

type K8sMetaDataCache struct {
	cMut            sync.RWMutex
	ContainerIdInfo map[string]*K8sContainerInfo
	//
	//    "192.168.1.14": { // podIp
	//        9093: k8sResInfo,
	//        6783: k8sResInfo
	//    },
	//    "192.168.2.15": { // podIp
	//        0: k8sResInfo,
	//        6783: k8sResInfo
	//    },
	//    "10.1.11.213": { // serviceIp
	//        8080: k8sResInfo
	//    }
	//}
	pMut            sync.RWMutex
	IpContainerInfo map[string]map[uint32]*K8sContainerInfo

	sMut          sync.RWMutex
	IpServiceInfo map[string]map[uint32]*K8sServiceInfo

	HostPortInfo *HostPortMap
}

func New() *K8sMetaDataCache {
	c := &K8sMetaDataCache{
		ContainerIdInfo: make(map[string]*K8sContainerInfo),
		IpContainerInfo: make(map[string]map[uint32]*K8sContainerInfo),
		IpServiceInfo:   make(map[string]map[uint32]*K8sServiceInfo),
		HostPortInfo:    NewHostPortMap(),
	}

	return c
}

func (c *K8sMetaDataCache) GetCacheContainerIdInfoSize() int {
	c.cMut.RLock()
	defer c.cMut.RUnlock()
	return len(c.ContainerIdInfo)
}

func (c *K8sMetaDataCache) AddByContainerId(containerId string, resource *K8sContainerInfo) {
	c.cMut.Lock()
	defer c.cMut.Unlock()
	c.ContainerIdInfo[containerId] = resource
}

func (c *K8sMetaDataCache) GetByContainerId(containerId string) (*K8sContainerInfo, bool) {
	c.cMut.RLock()
	defer c.cMut.RUnlock()
	res, ok := c.ContainerIdInfo[containerId]
	if ok {
		return res, ok
	}
	return nil, false
}

func (c *K8sMetaDataCache) GetPodByContainerId(containerId string) (*K8sPodInfo, bool) {
	c.cMut.RLock()
	defer c.cMut.RUnlock()
	containerInfo, ok := c.ContainerIdInfo[containerId]
	if ok {
		return containerInfo.RefPodInfo, ok
	}
	return nil, false
}

func (c *K8sMetaDataCache) DeleteByContainerId(containerId string) {
	c.cMut.Lock()
	defer c.cMut.Unlock()
	delete(c.ContainerIdInfo, containerId)
}

func (c *K8sMetaDataCache) AddContainerByIpPort(ip string, port uint32, resource *K8sContainerInfo) {
	c.pMut.RLock()
	portContainerInfo, ok := c.IpContainerInfo[ip]
	c.pMut.RUnlock()
	if ok {
		c.pMut.Lock()
		portContainerInfo[port] = resource
		c.pMut.Unlock()
	} else {
		portContainerInfo = make(map[uint32]*K8sContainerInfo)
		portContainerInfo[port] = resource
		c.pMut.Lock()
		c.IpContainerInfo[ip] = portContainerInfo
		c.pMut.Unlock()
	}
}

func (c *K8sMetaDataCache) GetContainerByIpPort(ip string, port uint32) (*K8sContainerInfo, bool) {
	c.pMut.RLock()
	defer c.pMut.RUnlock()
	portContainerInfo, ok := c.IpContainerInfo[ip]
	if !ok {
		return nil, false
	}
	containerInfo, ok := portContainerInfo[port]
	if ok {
		return containerInfo, true
	}
	// maybe such pod has a port which is not declared explicitly
	containerInfo, ok = portContainerInfo[0]
	if !ok {
		// find the first pod whose network mode is not hostnetwork
		for _, info := range portContainerInfo {
			if !info.RefPodInfo.isHostNetwork {
				return info, true
			}
		}
		return nil, false
	} else {
		if !containerInfo.RefPodInfo.isHostNetwork {
			return containerInfo, true
		}
		return nil, false
	}
}

func (c *K8sMetaDataCache) GetPodByIpPort(ip string, port uint32) (*K8sPodInfo, bool) {
	containerInfo, ok := c.GetContainerByIpPort(ip, port)
	if !ok {
		return nil, false
	}
	return containerInfo.RefPodInfo, true
}

func (c *K8sMetaDataCache) GetPodByIp(ip string) (*K8sPodInfo, bool) {
	c.pMut.RLock()
	defer c.pMut.RUnlock()
	portContainerInfo, ok := c.IpContainerInfo[ip]
	if !ok {
		return nil, false
	}
	// find the first pod whose network mode is not hostnetwork
	for _, info := range portContainerInfo {
		if !info.RefPodInfo.isHostNetwork {
			return info.RefPodInfo, true
		}
	}
	return nil, false
}

func (c *K8sMetaDataCache) DeleteContainerByIpPort(ip string, port uint32) {
	c.pMut.RLock()
	portContainerInfo, ok := c.IpContainerInfo[ip]
	c.pMut.RUnlock()
	if !ok {
		return
	}
	c.pMut.Lock()
	delete(portContainerInfo, port)
	if len(portContainerInfo) == 0 {
		delete(c.IpContainerInfo, ip)
	}
	c.pMut.Unlock()
}

func (c *K8sMetaDataCache) AddContainerByHostIpPort(hostIp string, hostPort uint32, containerInfo *K8sContainerInfo) {
	c.HostPortInfo.add(hostIp, hostPort, containerInfo)
}

func (c *K8sMetaDataCache) GetContainerByHostIpPort(hostIp string, hostPort uint32) (*K8sContainerInfo, bool) {
	return c.HostPortInfo.get(hostIp, hostPort)
}

func (c *K8sMetaDataCache) DeleteContainerByHostIpPort(hostIp string, hostPort uint32) {
	c.HostPortInfo.delete(hostIp, hostPort)
}

func (c *K8sMetaDataCache) AddServiceByIpPort(ip string, port uint32, resource *K8sServiceInfo) {
	c.sMut.RLock()
	portServiceInfo, ok := c.IpServiceInfo[ip]
	c.sMut.RUnlock()
	if ok {
		c.sMut.Lock()
		portServiceInfo[port] = resource
		c.sMut.Unlock()
	} else {
		portServiceInfo = make(map[uint32]*K8sServiceInfo)
		portServiceInfo[port] = resource
		c.sMut.Lock()
		c.IpServiceInfo[ip] = portServiceInfo
		c.sMut.Unlock()
	}
}

func (c *K8sMetaDataCache) GetServiceByIpPort(ip string, port uint32) (*K8sServiceInfo, bool) {
	c.sMut.RLock()
	portServiceInfo, ok := c.IpServiceInfo[ip]
	defer c.sMut.RUnlock()
	if !ok {
		return nil, false
	}
	serviceInfo, ok := portServiceInfo[port]
	if ok {
		return serviceInfo, true
	}
	return nil, false
}

func (c *K8sMetaDataCache) DeleteServiceByIpPort(ip string, port uint32) {
	c.sMut.RLock()
	portServiceInfo, ok := c.IpServiceInfo[ip]
	c.sMut.RUnlock()
	if !ok {
		return
	}
	c.sMut.Lock()
	delete(portServiceInfo, port)
	if len(portServiceInfo) == 0 {
		delete(c.IpServiceInfo, ip)
	}
	c.sMut.Unlock()
}

func (c *K8sMetaDataCache) ClearAll() {
	c.pMut.Lock()
	c.IpContainerInfo = make(map[string]map[uint32]*K8sContainerInfo)
	c.pMut.Unlock()

	c.sMut.Lock()
	c.IpServiceInfo = make(map[string]map[uint32]*K8sServiceInfo)
	c.sMut.Unlock()

	c.cMut.Lock()
	c.ContainerIdInfo = make(map[string]*K8sContainerInfo)
	c.cMut.Unlock()
}

func (c *K8sMetaDataCache) String() string {
	containerIdPodJson, _ := json.Marshal(c.ContainerIdInfo)
	ipContainerJson, _ := json.Marshal(c.IpContainerInfo)
	ipServiceJson, _ := json.Marshal(c.IpServiceInfo)
	return fmt.Sprintf("{\"containerIdPodInfo\": %s, \"ipContainerInfo\": %s, \"ipServiceInfo\": %s}",
		string(containerIdPodJson), string(ipContainerJson), string(ipServiceJson))
}

func (c *K8sMetaDataCache) GetNodeNameByIp(ip string) (string, bool) {
	return GlobalNodeInfo.getNodeName(ip)
}

func SetupCache(cache *K8sMetaDataCache, nodeMap *NodeMap, serviceMap *ServiceMap, rsMap *ReplicaSetMap) {
	LockMetadataCache()
	defer UnlockMetadataCache()
	if cache != nil {
		if cache.ContainerIdInfo != nil {
			// Recalculate local cacheMap
			tmpLocalWorkloadMap := newWorkloadMap()
			tmpGlobalPodInfo := newPodMap()
			for _, containersInfo := range cache.ContainerIdInfo {
				refPodInfo := containersInfo.RefPodInfo
				tmpGlobalPodInfo.add(refPodInfo)
				if refPodInfo.NodeName == MyNodeName {
					// Recalculate local workloadMap
					tmpLocalWorkloadMap.add(&workloadInfo{
						Namespace:    refPodInfo.Namespace,
						WorkloadName: refPodInfo.WorkloadName,
						WorkloadKind: refPodInfo.WorkloadKind,
					})
				}
			}
			GlobalPodInfo.Info = tmpGlobalPodInfo.Info
			localWorkloadMap.Info = tmpLocalWorkloadMap.Info
			MetaDataCache.ContainerIdInfo = cache.ContainerIdInfo
		}
		if cache.HostPortInfo != nil {
			MetaDataCache.HostPortInfo.HostPortInfo = cache.HostPortInfo.HostPortInfo
		}
		if cache.IpContainerInfo != nil {
			MetaDataCache.IpContainerInfo = cache.IpContainerInfo
		}
		if cache.IpServiceInfo != nil {
			MetaDataCache.IpServiceInfo = cache.IpServiceInfo
		}
	}
	if nodeMap != nil {
		GlobalNodeInfo.Info = nodeMap.Info
	}
	if serviceMap != nil {
		GlobalServiceInfo.ServiceMap = serviceMap.ServiceMap
	}
	if rsMap != nil {
		GlobalRsInfo.Info = rsMap.Info
	}
	podDeleteQueue = make([]deleteRequest, 0)
}
