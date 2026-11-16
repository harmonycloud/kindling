// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"strconv"
	"sync"
)

type IpPortKey struct {
	Ip   string
	Port uint32
}

type HostPortMap struct {
	HostPortInfo map[string]*K8sContainerInfo
	mutex        sync.RWMutex
}

func NewHostPortMap() *HostPortMap {
	return &HostPortMap{
		HostPortInfo: make(map[string]*K8sContainerInfo),
	}
}

func (m *HostPortMap) add(ip string, port uint32, containerInfo *K8sContainerInfo) {
	key := ip + ":" + strconv.FormatUint(uint64(port), 10)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.HostPortInfo[key] = containerInfo
}

func (m *HostPortMap) get(ip string, port uint32) (*K8sContainerInfo, bool) {
	key := ip + ":" + strconv.FormatUint(uint64(port), 10)
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	containerInfo, ok := m.HostPortInfo[key]
	if !ok {
		return nil, false
	}
	return containerInfo, true
}

func (m *HostPortMap) delete(ip string, port uint32) {
	key := ip + ":" + strconv.FormatUint(uint64(port), 10)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.HostPortInfo, key)
}
