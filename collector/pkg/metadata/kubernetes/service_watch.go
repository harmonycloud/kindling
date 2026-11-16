// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"fmt"
	_ "path/filepath"
	"sync"

	corev1 "k8s.io/api/core/v1"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	_ "k8s.io/client-go/tools/clientcmd"
	_ "k8s.io/client-go/util/homedir"
)

type ServiceMap struct {
	// Service name could be duplicated in different namespace, so here
	// service name must not be the key of map. Therefore, a map with the
	// following structure is constructed.
	//
	// namespace1:
	//   servicename1: ServiceInfo{}
	//   servicename2: ServiceInfo{}
	// namespace2:
	//   servicename1: ServiceInfo{}
	ServiceMap map[string]map[string]*K8sServiceInfo
	mut        sync.RWMutex
}

var GlobalServiceInfo = newServiceMap()
var serviceUpdatedMutex sync.Mutex

func newServiceMap() *ServiceMap {
	return &ServiceMap{
		ServiceMap: make(map[string]map[string]*K8sServiceInfo),
		mut:        sync.RWMutex{},
	}
}

// GetServiceMatchLabels gets K8sServiceInfos which match labels in such namespace.
// Return empty slice if not found. Note there may be multiple matches.
func (s *ServiceMap) GetServiceMatchLabels(namespace string, labels map[string]string) []*K8sServiceInfo {
	s.mut.RLock()
	defer s.mut.RUnlock()
	retServiceInfoSlice := make([]*K8sServiceInfo, 0)
	serviceNameMap, ok := s.ServiceMap[namespace]
	if !ok {
		return retServiceInfoSlice
	}
	for _, serviceInfo := range serviceNameMap {
		if len(serviceInfo.Selector) == 0 {
			continue
		}
		if SelectorsMatchLabels(serviceInfo.Selector, labels) {
			retServiceInfoSlice = append(retServiceInfoSlice, serviceInfo)
		}
	}
	return retServiceInfoSlice
}

func (s *ServiceMap) add(info *K8sServiceInfo) {
	s.mut.Lock()
	defer s.mut.Unlock()
	serviceNameMap, ok := s.ServiceMap[info.Namespace]
	if !ok {
		serviceNameMap = make(map[string]*K8sServiceInfo)
	}
	serviceNameMap[info.ServiceName] = info
	s.ServiceMap[info.Namespace] = serviceNameMap
}

func (s *ServiceMap) delete(namespace string, serviceName string) {
	s.mut.Lock()
	defer s.mut.Unlock()
	serviceNameMap, ok := s.ServiceMap[namespace]
	if ok {
		serviceInfo, ok := serviceNameMap[serviceName]
		if ok {
			// Set the value empty via its pointer, in which way all serviceInfo related to
			// K8sPodInfo.K8sServiceInfo will be set to empty.
			// The following data will be affected:
			// - K8sMetaDataCache.containerIdInfo
			// - K8sMetaDataCache.ipContainerInfo
			// - K8sMetaDataCache.ipServiceInfo
			serviceInfo.emptySelf()
		}
	}
}

func ServiceWatch(clientSet *kubernetes.Clientset, handler cache.ResourceEventHandler) {
	stopper := make(chan struct{})
	defer close(stopper)

	factory := informers.NewSharedInformerFactory(clientSet, 0)
	serviceInformer := factory.Core().V1().Services()
	informer := serviceInformer.Informer()
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
			AddFunc:    AddService,
			UpdateFunc: UpdateService,
			DeleteFunc: DeleteService,
		})
	}
	// TODO: use workqueue to avoid blocking
	<-stopper
}

func AddService(obj interface{}) {
	service := obj.(*corev1.Service)
	sI := &K8sServiceInfo{
		Ip:          service.Spec.ClusterIP,
		ServiceName: service.Name,
		Namespace:   service.Namespace,
		isNodePort:  service.Spec.Type == "NodePort",
		Selector:    service.Spec.Selector,
	}
	GlobalServiceInfo.add(sI)
	// When new service is added, podInfo should be updated
	podInfoSlice := GlobalPodInfo.getPodsMatchSelectors(sI.Namespace, sI.Selector)
	for _, podInfo := range podInfoSlice {
		for _, containerId := range podInfo.ContainerIds {
			if podInfo, ok := MetaDataCache.GetPodByContainerId(containerId); ok {
				sI.WorkloadName = podInfo.WorkloadName
				sI.WorkloadKind = podInfo.WorkloadKind
				podInfo.ServiceInfo = sI
			}
		}
		// update Ip-Pod Map
		for _, port := range podInfo.Ports {
			if podInfo, ok := MetaDataCache.GetPodByIpPort(podInfo.Ip, uint32(port)); ok {
				sI.WorkloadName = podInfo.WorkloadName
				sI.WorkloadKind = podInfo.WorkloadKind
				podInfo.ServiceInfo = sI
			}
		}
	}

	if sI.Ip == "" || sI.Ip == "None" {
		return
	}
	for _, port := range service.Spec.Ports {
		MetaDataCache.AddServiceByIpPort(service.Spec.ClusterIP, uint32(port.Port), sI)
		if sI.isNodePort {
			nodeAddresses := GlobalNodeInfo.getAllNodeAddresses()
			for _, nodeAddress := range nodeAddresses {
				MetaDataCache.AddServiceByIpPort(nodeAddress, uint32(port.NodePort), sI)
			}
		}
	}
}

func UpdateService(objOld interface{}, objNew interface{}) {
	oldSvc := objOld.(*corev1.Service)
	newSvc := objNew.(*corev1.Service)
	if oldSvc.ResourceVersion == newSvc.ResourceVersion {
		return
	}
	serviceUpdatedMutex.Lock()
	defer serviceUpdatedMutex.Unlock()
	// TODO: re-implement the updated logic to reduce computation
	DeleteService(objOld)
	AddService(objNew)
}

func DeleteService(obj interface{}) {
	// Maybe get DeletedFinalStateUnknown instead of *corev1.Pod.
	// Fix https://github.com/KindlingProject/kindling/issues/445
	service, ok := obj.(*corev1.Service)
	if !ok {
		deletedState, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return
		}
		service, ok = deletedState.Obj.(*corev1.Service)
		if !ok {
			return
		}
	}
	// 'delete' will delete all such service in MetaDataCache
	GlobalServiceInfo.delete(service.Namespace, service.Name)
	ip := service.Spec.ClusterIP
	if ip == "" || ip == "None" {
		return
	}
	for _, port := range service.Spec.Ports {
		MetaDataCache.DeleteServiceByIpPort(ip, uint32(port.Port))
		if service.Spec.Type == "NodePort" {
			nodeAddresses := GlobalNodeInfo.getAllNodeAddresses()
			for _, nodeAddress := range nodeAddresses {
				MetaDataCache.DeleteServiceByIpPort(nodeAddress, uint32(port.NodePort))
			}
		}
	}
}

// SelectorsMatchLabels return true only if labels match all [keys:values] with selectors
func SelectorsMatchLabels(selectors map[string]string, labels map[string]string) bool {
	for key, value := range selectors {
		if labelValue, ok := labels[key]; !ok || labelValue != value {
			return false
		}
	}
	return true
}
