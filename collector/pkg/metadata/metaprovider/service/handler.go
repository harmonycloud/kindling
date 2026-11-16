// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/Kindling-project/kindling/collector/pkg/metadata/metaprovider/api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

type boardcast func(data api.MetaDataRsyncResponse)

type K8sResourceHandler struct {
	resType string

	add    api.AddObj
	update api.UpdateObj
	delete api.DeleteObj

	boardcast
}

func (m *K8sResourceHandler) AddObj(obj interface{}) {
	m.add(obj)
	m.boardcast(api.MetaDataRsyncResponse{
		Type:      m.resType,
		Operation: string(api.Add),
		NewObj:    obj,
	})
}

func (m *K8sResourceHandler) UpdateObj(objOld interface{}, objNew interface{}) {
	m.update(objOld, objNew)
	m.boardcast(api.MetaDataRsyncResponse{
		Type:      m.resType,
		Operation: string(api.Update),
		OldObj:    objOld,
		NewObj:    objNew,
	})
}

func (m *K8sResourceHandler) DeleteObj(obj interface{}) {
	m.delete(obj)
	m.boardcast(api.MetaDataRsyncResponse{
		Type:      m.resType,
		Operation: string(api.Delete),
		OldObj:    obj,
	})
}

type PodResourceHandler struct {
	K8sResourceHandler
}

func (prh *PodResourceHandler) AddPod(obj interface{}) {
	decreasePodInfo(obj)
	prh.K8sResourceHandler.AddObj(obj)
}

func (prh *PodResourceHandler) UpdatePod(objOld interface{}, objNew interface{}) {
	decreasePodInfo(objOld)
	decreasePodInfo(objNew)
	prh.K8sResourceHandler.UpdateObj(objNew, objOld)
}

func (prh *PodResourceHandler) DeleteObj(obj interface{}) {
	decreasePodInfo(obj)
	prh.K8sResourceHandler.DeleteObj(obj)
}

func NewHandler(typeName string, add api.AddObj, update api.UpdateObj, delete api.DeleteObj, boardcast boardcast) cache.ResourceEventHandlerFuncs {
	handler := K8sResourceHandler{
		resType:   typeName,
		add:       add,
		update:    update,
		delete:    delete,
		boardcast: boardcast,
	}

	if typeName == "pod" {
		prh := PodResourceHandler{handler}

		return cache.ResourceEventHandlerFuncs{
			AddFunc:    prh.AddObj,
			UpdateFunc: prh.UpdateObj,
			DeleteFunc: prh.DeleteObj,
		}
	}

	return cache.ResourceEventHandlerFuncs{
		AddFunc:    handler.AddObj,
		UpdateFunc: handler.UpdateObj,
		DeleteFunc: handler.DeleteObj,
	}
}

func decreasePodInfo(obj interface{}) {
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

	if pod != nil {
		pod.ManagedFields = nil
		pod.Spec.Volumes = nil
		pod.Status.Conditions = nil
	}
}
