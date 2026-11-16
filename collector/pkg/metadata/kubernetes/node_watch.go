// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"fmt"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type NodeInfo struct {
	Ip     string
	Name   string
	Labels map[string]string
}

type NodeMap struct {
	Info  map[string]*NodeInfo
	mutex sync.RWMutex
}

func newNodeMap() *NodeMap {
	return &NodeMap{
		Info: make(map[string]*NodeInfo),
	}
}

func (n *NodeMap) add(info *NodeInfo) {
	if info == nil {
		return
	}
	n.mutex.Lock()
	n.Info[info.Ip] = info
	n.mutex.Unlock()
}

func (n *NodeMap) getNodeName(ip string) (string, bool) {
	n.mutex.RLock()
	ret, ok := n.Info[ip]
	n.mutex.RUnlock()
	if !ok {
		return "", false
	}
	return ret.Name, true
}

func (n *NodeMap) getAllNodeAddresses() []string {
	ret := make([]string, 0)
	n.mutex.RLock()
	for address := range n.Info {
		ret = append(ret, address)
	}
	n.mutex.RUnlock()
	return ret
}

func (n *NodeMap) delete(name string) {
	n.mutex.Lock()
	delete(n.Info, name)
	n.mutex.Unlock()
}

var GlobalNodeInfo = newNodeMap()

func NodeWatch(clientSet *kubernetes.Clientset, handler cache.ResourceEventHandler) {
	stopper := make(chan struct{})
	defer close(stopper)

	factory := informers.NewSharedInformerFactory(clientSet, 0)
	nodeInformer := factory.Core().V1().Nodes()
	informer := nodeInformer.Informer()
	defer runtime.HandleCrash()

	go factory.Start(stopper)

	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	if handler != nil {
		informer.AddEventHandler(handler)
	} else {
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    AddNode,
			UpdateFunc: UpdateNode,
			DeleteFunc: DeleteNode,
		})
	}

	// TODO: use workqueue to avoid blocking
	<-stopper
}

func AddNode(obj interface{}) {
	node := obj.(*corev1.Node)
	nI := &NodeInfo{
		Ip:     "",
		Name:   node.Name,
		Labels: node.Labels,
	}

	for _, nodeAddress := range node.Status.Addresses {
		if nodeAddress.Type == "InternalIP" {
			nI.Ip = nodeAddress.Address
		}
	}
	GlobalNodeInfo.add(nI)
}

func UpdateNode(objOld interface{}, objNew interface{}) {
	DeleteNode(objOld)
	AddNode(objNew)
}

func DeleteNode(obj interface{}) {
	// Maybe get DeletedFinalStateUnknown instead of *corev1.Pod.
	// Fix https://github.com/KindlingProject/kindling/issues/445
	node, ok := obj.(*corev1.Node)
	if !ok {
		deletedState, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return
		}
		node, ok = deletedState.Obj.(*corev1.Node)
		if !ok {
			return
		}
	}
	GlobalNodeInfo.delete(node.Name)
}
