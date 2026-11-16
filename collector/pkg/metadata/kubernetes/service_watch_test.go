// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSelectorsMatchLabels(t *testing.T) {
	testCases := []struct {
		selectors map[string]string
		labels    map[string]string
		expected  bool
	}{
		// more labels, find key, equal value
		{
			selectors: map[string]string{
				"a": "b",
				"c": "d",
			},
			labels: map[string]string{
				"a": "b",
				"c": "d",
				"e": "f",
			},
			expected: true,
		},
		// equal key, not equal value
		{
			selectors: map[string]string{
				"a": "b",
			},
			labels: map[string]string{
				"a": "c",
			},
			expected: false,
		},
		// not equal key
		{
			selectors: map[string]string{
				"a": "b",
			},
			labels: map[string]string{
				"b": "b",
			},
			expected: false,
		},
		// more selectors
		{
			selectors: map[string]string{
				"a": "b",
				"c": "d",
			},
			labels: map[string]string{
				"a": "b",
			},
			expected: false,
		},
		// empty selector value, no labels
		{
			selectors: map[string]string{
				"a": "",
			},
			labels:   map[string]string{},
			expected: false,
		},
	}

	for _, test := range testCases {
		isMatched := SelectorsMatchLabels(test.selectors, test.labels)
		if test.expected != isMatched {
			t.Errorf("selectors %s, labels %s, match %v expected, but got %v", test.selectors, test.labels, test.expected, isMatched)
		}
	}
}

func TestServiceMap_GetServiceMatchLabels(t *testing.T) {
	service1 := &K8sServiceInfo{
		Ip:          "192.168.1.1",
		ServiceName: "service1",
		Namespace:   "namespace1",
		isNodePort:  true,
		Selector: map[string]string{
			"a": "1",
			"b": "1",
		},
	}

	service2 := &K8sServiceInfo{
		Ip:          "192.168.1.2",
		ServiceName: "service2",
		Namespace:   "namespace2",
		isNodePort:  true,
		Selector: map[string]string{
			"a": "2",
			"c": "2",
		},
	}

	serviceMap := &ServiceMap{
		ServiceMap: make(map[string]map[string]*K8sServiceInfo),
	}
	serviceMap.add(service1)
	serviceMap.add(service2)
	labels := map[string]string{
		"a": "2",
		"b": "1",
		"c": "2",
	}
	serviceInfoSlice := serviceMap.GetServiceMatchLabels(service2.Namespace, labels)
	if !reflect.DeepEqual(serviceInfoSlice[0], service2) {
		t.Errorf("GetServiceMatchLabels return %v, but %v wanted", serviceInfoSlice[0], service2)
	}
}

func TestOnAddService(t *testing.T) {
	GlobalPodInfo = &podMap{
		Info: make(map[string]map[string]*K8sPodInfo),
	}
	GlobalServiceInfo = &ServiceMap{
		ServiceMap: make(map[string]map[string]*K8sServiceInfo),
	}
	GlobalRsInfo = &ReplicaSetMap{
		Info: make(map[string]Controller),
	}
	// First add pod, and then add service
	AddReplicaSet(CreateReplicaSet())
	AddPod(CreatePod(true))
	AddService(CreateService())
	t.Log(MetaDataCache)
	// Delete service must empty the serviceInfo referenced by podInfo
	DeleteService(CreateService())
	t.Log(MetaDataCache)
	// Empty all the metadata
	DeletePod(CreatePod(true))
	t.Log(MetaDataCache)
}

func TestServiceMap_Delete(t *testing.T) {
	GlobalPodInfo = &podMap{
		Info: make(map[string]map[string]*K8sPodInfo),
	}
	GlobalServiceInfo = &ServiceMap{
		ServiceMap: make(map[string]map[string]*K8sServiceInfo),
	}
	GlobalRsInfo = &ReplicaSetMap{
		Info: make(map[string]Controller),
	}
	AddService(CreateService())
	AddPod(CreatePod(true))
	containerId := "1a2b3c4d5e6f"
	podFromCache, ok := MetaDataCache.GetPodByContainerId(containerId)
	if !ok {
		t.Fatalf("Supposed to find Pod by containerId=%s, but not found", containerId)
	}
	serviceFromCache := podFromCache.ServiceInfo
	expectedService := &K8sServiceInfo{
		Ip:          "192.168.1.2",
		ServiceName: "CustomService",
		Namespace:   "CustomNamespace",
		isNodePort:  true,
		Selector: map[string]string{
			"a": "1",
			"b": "1",
		},
	}
	if !reflect.DeepEqual(serviceFromCache, expectedService) {
		t.Errorf("Before delete method is invoked %v is expected, but get %v", expectedService, serviceFromCache)
	}
	GlobalServiceInfo.delete("CustomNamespace", "CustomService")
	expectedService = &K8sServiceInfo{
		Ip:          "",
		ServiceName: "",
		Namespace:   "",
		isNodePort:  false,
		Selector:    nil,
	}
	if !reflect.DeepEqual(serviceFromCache, expectedService) {
		t.Errorf("After delete method is invoked %v is expected, but get %v", expectedService, serviceFromCache)
	}
}

func CreateService() *corev1.Service {
	var service = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "CustomService",
			Namespace: "CustomNamespace",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:     80,
					NodePort: 80,
				},
			},
			Selector: map[string]string{
				"a": "1",
				"b": "1",
			},
			ClusterIP: "192.168.1.2",
			Type:      "NodePort",
		},
	}

	return service
}
