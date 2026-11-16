// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTruncateContainerId(t *testing.T) {
	testCases := []struct {
		containerId string
		expected    string
	}{
		{"docker://a1b2c3d4e5f6g7h8i9j0k1l2m3n", "a1b2c3d4e5f6"},
		{"docker://a1b2c3d4e5f6", "a1b2c3d4e5f6"},
		{"docker://a1b2c3", "a1b2c3"},
		{"containerd://a1b2c3d4e5f6g7h8i9j0k1l2m3n", "a1b2c3d4e5f6"},
		{"a1b2c3", ""},
		{"", ""},
	}
	for _, test := range testCases {
		res := TruncateContainerId(test.containerId)
		if res != test.expected {
			t.Errorf("containerId=%s, get=%s, but expected=%s", test.containerId, res, test.expected)
		}
	}
}

func TestOnAdd(t *testing.T) {
	GlobalPodInfo = &podMap{
		Info: make(map[string]map[string]*K8sPodInfo),
	}
	GlobalServiceInfo = &ServiceMap{
		ServiceMap: make(map[string]map[string]*K8sServiceInfo),
	}
	GlobalRsInfo = &ReplicaSetMap{
		Info: make(map[string]Controller),
	}
	// First add service, and then add pod
	AddService(CreateService())
	AddReplicaSet(CreateReplicaSet())
	AddPod(CreatePod(true))
	t.Log(MetaDataCache)
	// Delete podInfo must not affect serviceMap
	DeletePod(CreatePod(true))
	t.Log(MetaDataCache)
	// Empty all the metadata
	DeleteService(CreateService())
	t.Log(MetaDataCache)
}

// ISSUE https://github.com/KindlingProject/kindling/issues/229
func TestOnAddPodWhileReplicaSetUpdating(t *testing.T) {
	GlobalPodInfo = &podMap{
		Info: make(map[string]map[string]*K8sPodInfo),
	}
	GlobalServiceInfo = &ServiceMap{
		ServiceMap: make(map[string]map[string]*K8sServiceInfo),
	}
	GlobalRsInfo = &ReplicaSetMap{
		Info: make(map[string]Controller),
	}
	// Firstly deployment created and add old RS and old POD
	controller := true
	oldRs := CreateReplicaSet()
	oldRs.SetResourceVersion("old")
	newRs := CreateReplicaSet()
	newRs.SetResourceVersion("new")
	oldPOD := CreatePod(true)
	oldPOD.SetResourceVersion("old")
	oldPOD.OwnerReferences[0].Controller = &controller
	newPOD := CreatePod(true)
	newPOD.SetResourceVersion("new")
	newPOD.OwnerReferences[0].Controller = &controller
	AddReplicaSet(oldRs)
	AddPod(oldPOD)

	// Secondly POD&RS were been updated

	go func() {
		for i := 0; i < 1000; i++ {
			UpdateReplicaSet(oldRs, newRs)
		}
	}()

	for i := 0; i < 100; i++ {
		UpdatePod(oldPOD, newPOD)
		// Thirdly check the pod's workload_kind
		pod, ok := MetaDataCache.GetPodByContainerId(TruncateContainerId(newPOD.Status.ContainerStatuses[0].ContainerID))
		require.True(t, ok, "failed to get target POD")
		require.Equal(t, "deployment", pod.WorkloadKind, "failed to get the real workload_kind")
	}
}

func TestOnAddLowercaseWorkload(t *testing.T) {
	GlobalPodInfo = &podMap{
		Info: make(map[string]map[string]*K8sPodInfo),
	}
	GlobalServiceInfo = &ServiceMap{
		ServiceMap: make(map[string]map[string]*K8sServiceInfo),
	}
	GlobalRsInfo = &ReplicaSetMap{
		Info: make(map[string]Controller),
	}
	higherCase := "DaemonSet"
	lowerCase := "daemonset"
	isController := true
	AddPod(&corev1.Pod{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			OwnerReferences: []metav1.OwnerReference{{
				Kind:       higherCase,
				Controller: &isController,
			}},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "container1",
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 80,
						},
					},
				},
			}},
		Status: corev1.PodStatus{
			PodIP: "172.10.1.2",
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:        "container1",
					ContainerID: "docker://1a2b3c4d5e6f7g8h9i0j1k2",
				},
			},
		},
	})
	podInfo, ok := MetaDataCache.GetPodByContainerId("1a2b3c4d5e6f")
	if !ok || podInfo.WorkloadKind != lowerCase {
		t.Errorf("%s wanted, but get %s", higherCase, lowerCase)
	}
}

func CreatePod(hasPort bool) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			UID:       "0ae5c03d-5fb3-4eb9-9de8-2bd4b51606ba",
			Name:      "deploy-1a2b3c4d-5e6f7",
			Namespace: "CustomNamespace",
			Labels: map[string]string{
				"a": "1",
				"b": "1",
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: ReplicaSetKind,
					Name: "deploy-1a2b3c4d",
				},
			},
		},
		Spec: corev1.PodSpec{
			NodeName:    "node1",
			HostNetwork: false,
		},
		Status: corev1.PodStatus{
			PodIP: "172.10.1.2",
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name:        "container1",
					ContainerID: "docker://1a2b3c4d5e6f7g8h9i0j1k2",
				},
			},
		},
	}
	if hasPort {
		pod.Spec.Containers = []corev1.Container{
			{
				Name: "container1",
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 80,
					},
				},
			},
		}
	} else {
		pod.Spec.Containers = []corev1.Container{
			{
				Name: "container1",
			},
		}
	}
	return pod
}

func TestUpdateAndDelayDelete(t *testing.T) {
	addObjJson := "{\"metadata\":{\"name\":\"testdemo2-5c86748464-26crb\",\"namespace\":\"test-ns\",\"uid\":\"0ae5c03d-5fb3-4eb9-9de8-2bd4b51606ba\",\"resourceVersion\":\"44895976\"},\"spec\":{\"containers\":[{\"name\":\"testdemo2\",\"ports\":[{\"containerPort\":9001,\"protocol\":\"TCP\",\"hostPort\":9001}]}]},\"status\":{\"phase\":\"Running\",\"podIP\":\"192.168.136.210\",\"hostIP\":\"10.10.10.101\",\"containerStatuses\":[{\"name\":\"testdemo2\",\"state\":{\"running\":{\"startedAt\":\"2022-05-25T08:55:36Z\"}},\"lastState\":{},\"ready\":true,\"restartCount\":5,\"image\":\"\",\"imageID\":\"docker-pullable://10.10.102.213:8443/cloudnevro-test/test-netserver@sha256:6720f648b74ed590f36094a1c7a58b01b6881396409784c17f471ecfe445e3fd\",\"containerID\":\"docker://d505f50edb4e204cf31840e3cb8d26d33f212d4ebef994d0c3fc151d57e17413\",\"started\":true}]}}"
	updateObjJson := "{\"metadata\":{\"name\":\"testdemo2-5c86748464-26crb\",\"namespace\":\"test-ns\",\"uid\":\"0ae5c03d-5fb3-4eb9-9de8-2bd4b51606ba\",\"resourceVersion\":\"44895977\"},\"spec\":{\"containers\":[{\"name\":\"testdemo2\",\"ports\":[{\"containerPort\":9001,\"protocol\":\"TCP\",\"hostPort\":9001}]}]},\"status\":{\"phase\":\"Running\",\"podIP\":\"192.168.136.212\",\"hostIP\":\"10.10.10.102\",\"containerStatuses\":[{\"name\":\"testdemo2\",\"state\":{\"running\":{\"startedAt\":\"2022-05-25T08:55:36Z\"}},\"lastState\":{},\"ready\":true,\"restartCount\":5,\"image\":\"\",\"imageID\":\"docker-pullable://10.10.102.213:8443/cloudnevro-test/test-netserver@sha256:6720f648b74ed590f36094a1c7a58b01b6881396409784c17f471ecfe445e3fd\",\"containerID\":\"docker://d000f50edb4e204cf31840e3cb8d26d33f212d4ebef994d0c3fc151d57e17413\",\"started\":true}]}}"
	addObj := new(corev1.Pod)
	err := json.Unmarshal([]byte(addObjJson), addObj)
	if err != nil {
		t.Errorf("error unmarshalling %v", err)
	}
	updateObj := new(corev1.Pod)
	err = json.Unmarshal([]byte(updateObjJson), updateObj)
	if err != nil {
		t.Fatalf("error unmarshalling %v", err)
	}
	podIp := addObj.Status.PodIP
	port := addObj.Spec.Containers[0].Ports[0].ContainerPort
	AddPod(addObj)
	_, ok := MetaDataCache.GetContainerByIpPort(podIp, uint32(port))
	if !ok {
		t.Fatalf("Not found container [%s:%d]", podIp, port)
	}
	stopCh := make(chan struct{})
	go podDeleteLoop(100*time.Millisecond, 500*time.Millisecond, stopCh)
	UpdatePod(addObj, updateObj)

	// Check if the new container can be found
	assertFindPod(t, updateObj)
	// Wait for the deleting
	time.Sleep(1000 * time.Millisecond)
	// Double check for the new container
	assertFindPod(t, updateObj)

	// Check if the old container has been deleted
	_, find := MetaDataCache.GetByContainerId(TruncateContainerId(addObj.Status.ContainerStatuses[0].ContainerID))
	assert.False(t, find, "OldContainerId should be deletedin MetaDataCache")
	_, find = MetaDataCache.GetContainerByIpPort(addObj.Status.PodIP, uint32(port))
	assert.False(t, find, "OldContainer IP should be deleted in MetaDataCache")
	_, find = MetaDataCache.GetContainerByHostIpPort(addObj.Status.HostIP, uint32(port))
	assert.False(t, find, "OldHostIp Port should be deleted in MetaDataCache")

	stopCh <- struct{}{}
}

func TestUpdateAndDelayDeleteWhenOnlyPodIpChanged(t *testing.T) {
	addObjJson := "{\"metadata\":{\"name\":\"testdemo2-5c86748464-26crb\",\"namespace\":\"test-ns\",\"uid\":\"0ae5c03d-5fb3-4eb9-9de8-2bd4b51606ba\",\"resourceVersion\":\"44895976\"},\"spec\":{\"containers\":[{\"name\":\"testdemo2\",\"ports\":[{\"containerPort\":9001,\"protocol\":\"TCP\",\"hostPort\":9001}]}]},\"status\":{\"phase\":\"Running\",\"podIP\":\"192.168.136.210\",\"hostIP\":\"10.10.10.101\",\"containerStatuses\":[{\"name\":\"testdemo2\",\"state\":{\"running\":{\"startedAt\":\"2022-05-25T08:55:36Z\"}},\"lastState\":{},\"ready\":true,\"restartCount\":5,\"image\":\"\",\"imageID\":\"docker-pullable://10.10.102.213:8443/cloudnevro-test/test-netserver@sha256:6720f648b74ed590f36094a1c7a58b01b6881396409784c17f471ecfe445e3fd\",\"containerID\":\"docker://d505f50edb4e204cf31840e3cb8d26d33f212d4ebef994d0c3fc151d57e17413\",\"started\":true}]}}"
	updateObjJson := "{\"metadata\":{\"name\":\"testdemo2-5c86748464-26crb\",\"namespace\":\"test-ns\",\"uid\":\"0ae5c03d-5fb3-4eb9-9de8-2bd4b51606ba\",\"resourceVersion\":\"44895977\"},\"spec\":{\"containers\":[{\"name\":\"testdemo2\",\"ports\":[{\"containerPort\":9001,\"protocol\":\"TCP\",\"hostPort\":9001}]}]},\"status\":{\"phase\":\"Running\",\"podIP\":\"192.168.136.212\",\"hostIP\":\"10.10.10.101\",\"containerStatuses\":[{\"name\":\"testdemo2\",\"state\":{\"running\":{\"startedAt\":\"2022-05-25T08:55:36Z\"}},\"lastState\":{},\"ready\":true,\"restartCount\":5,\"image\":\"\",\"imageID\":\"docker-pullable://10.10.102.213:8443/cloudnevro-test/test-netserver@sha256:6720f648b74ed590f36094a1c7a58b01b6881396409784c17f471ecfe445e3fd\",\"containerID\":\"docker://d000f50edb4e204cf31840e3cb8d26d33f212d4ebef994d0c3fc151d57e17413\",\"started\":true}]}}"
	addObj := new(corev1.Pod)
	_ = json.Unmarshal([]byte(addObjJson), addObj)
	updateObj := new(corev1.Pod)
	_ = json.Unmarshal([]byte(updateObjJson), updateObj)

	AddPod(addObj)
	stopCh := make(chan struct{})
	go podDeleteLoop(100*time.Millisecond, 500*time.Millisecond, stopCh)
	UpdatePod(addObj, updateObj)

	// Check if the new container can be found
	assertFindPod(t, updateObj)
	// Wait for the deleting
	time.Sleep(1000 * time.Millisecond)
	// Double check for the new container
	assertFindPod(t, updateObj)

	// Check if the old container has been deleted
	port := addObj.Spec.Containers[0].Ports[0].ContainerPort
	_, find := MetaDataCache.GetByContainerId(TruncateContainerId(addObj.Status.ContainerStatuses[0].ContainerID))
	assert.False(t, find, "OldContainerId should be deleted in MetaDataCache")
	_, find = MetaDataCache.GetContainerByIpPort(addObj.Status.PodIP, uint32(port))
	assert.False(t, find, "OldContainer IP should be deleted in MetaDataCache")

	stopCh <- struct{}{}
}

func TestUpdateAndDelayDeleteWhenOnlyPortChanged(t *testing.T) {
	addObjJson := "{\"metadata\":{\"name\":\"testdemo2-5c86748464-26crb\",\"namespace\":\"test-ns\",\"uid\":\"0ae5c03d-5fb3-4eb9-9de8-2bd4b51606ba\",\"resourceVersion\":\"44895976\"},\"spec\":{\"containers\":[{\"name\":\"testdemo2\",\"ports\":[{\"containerPort\":9001,\"protocol\":\"TCP\",\"hostPort\":9001}]}]},\"status\":{\"phase\":\"Running\",\"podIP\":\"192.168.136.210\",\"hostIP\":\"10.10.10.101\",\"containerStatuses\":[{\"name\":\"testdemo2\",\"state\":{\"running\":{\"startedAt\":\"2022-05-25T08:55:36Z\"}},\"lastState\":{},\"ready\":true,\"restartCount\":5,\"image\":\"\",\"imageID\":\"docker-pullable://10.10.102.213:8443/cloudnevro-test/test-netserver@sha256:6720f648b74ed590f36094a1c7a58b01b6881396409784c17f471ecfe445e3fd\",\"containerID\":\"docker://d505f50edb4e204cf31840e3cb8d26d33f212d4ebef994d0c3fc151d57e17413\",\"started\":true}]}}"
	updateObjJson := "{\"metadata\":{\"name\":\"testdemo2-5c86748464-26crb\",\"namespace\":\"test-ns\",\"uid\":\"0ae5c03d-5fb3-4eb9-9de8-2bd4b51606ba\",\"resourceVersion\":\"44895977\"},\"spec\":{\"containers\":[{\"name\":\"testdemo2\",\"ports\":[{\"containerPort\":9002,\"protocol\":\"TCP\",\"hostPort\":9002}]}]},\"status\":{\"phase\":\"Running\",\"podIP\":\"192.168.136.210\",\"hostIP\":\"10.10.10.101\",\"containerStatuses\":[{\"name\":\"testdemo2\",\"state\":{\"running\":{\"startedAt\":\"2022-05-25T08:55:36Z\"}},\"lastState\":{},\"ready\":true,\"restartCount\":5,\"image\":\"\",\"imageID\":\"docker-pullable://10.10.102.213:8443/cloudnevro-test/test-netserver@sha256:6720f648b74ed590f36094a1c7a58b01b6881396409784c17f471ecfe445e3fd\",\"containerID\":\"docker://d000f50edb4e204cf31840e3cb8d26d33f212d4ebef994d0c3fc151d57e17413\",\"started\":true}]}}"
	addObj := new(corev1.Pod)
	_ = json.Unmarshal([]byte(addObjJson), addObj)
	updateObj := new(corev1.Pod)
	_ = json.Unmarshal([]byte(updateObjJson), updateObj)

	AddPod(addObj)
	stopCh := make(chan struct{})
	go podDeleteLoop(100*time.Millisecond, 500*time.Millisecond, stopCh)
	UpdatePod(addObj, updateObj)

	// Check if new container can be found
	assertFindPod(t, updateObj)
	// Wait for the deleting
	time.Sleep(1000 * time.Millisecond)
	// Double check for the new container
	assertFindPod(t, updateObj)

	// Check the old Container has been deleted
	port := addObj.Spec.Containers[0].Ports[0].ContainerPort
	_, find := MetaDataCache.GetByContainerId(TruncateContainerId(addObj.Status.ContainerStatuses[0].ContainerID))
	assert.False(t, find, "OldContainerId should be deleted in MetaDataCache")
	_, find = MetaDataCache.GetContainerByIpPort(addObj.Status.PodIP, uint32(port))
	assert.True(t, find, "If podIp is not changed, Old IP can still be found in MetaDataCache")
	_, find = MetaDataCache.GetContainerByHostIpPort(addObj.Status.HostIP, uint32(port))
	assert.False(t, find, "OldHostIp Port should be deleted in MetaDataCache")

	stopCh <- struct{}{}
}

func TestDelayDeleteThenAddWithSameIP(t *testing.T) {
	addObjJson := "{\"metadata\":{\"name\":\"testdemo2-0\",\"namespace\":\"test-ns\",\"uid\":\"0ae5c03d-5fb3-4eb9-9de8-2bd4b51606ba\",\"resourceVersion\":\"44895976\"},\"spec\":{\"containers\":[{\"name\":\"testdemo2\",\"ports\":[{\"containerPort\":9001,\"protocol\":\"TCP\",\"hostPort\":9001}]}]},\"status\":{\"phase\":\"Running\",\"podIP\":\"192.168.136.210\",\"hostIP\":\"10.10.10.101\",\"containerStatuses\":[{\"name\":\"testdemo2\",\"state\":{\"running\":{\"startedAt\":\"2022-05-25T08:55:36Z\"}},\"lastState\":{},\"ready\":true,\"restartCount\":5,\"image\":\"\",\"imageID\":\"docker-pullable://10.10.102.213:8443/cloudnevro-test/test-netserver@sha256:6720f648b74ed590f36094a1c7a58b01b6881396409784c17f471ecfe445e3fd\",\"containerID\":\"docker://d505f50edb4e204cf31840e3cb8d26d33f212d4ebef994d0c3fc151d57e17413\",\"started\":true}]}}"
	addObj := new(corev1.Pod)
	_ = json.Unmarshal([]byte(addObjJson), addObj)

	deletedObjJson := "{\"metadata\":{\"name\":\"testdemo2-0\",\"namespace\":\"test-ns\",\"uid\":\"0ae5c03d-5fb3-4eb9-9de8-2bd4b51606ba\",\"resourceVersion\":\"44895977\"},\"spec\":{\"containers\":[{\"name\":\"testdemo2\",\"ports\":[{\"containerPort\":9001,\"protocol\":\"TCP\",\"hostPort\":9001}]}]},\"status\":{\"phase\":\"Running\",\"podIP\":\"192.168.136.210\",\"hostIP\":\"10.10.10.101\",\"containerStatuses\":[{\"name\":\"testdemo2\",\"state\":{\"running\":{\"startedAt\":\"2022-05-25T08:55:37Z\"}},\"lastState\":{},\"ready\":true,\"restartCount\":5,\"image\":\"\",\"imageID\":\"docker-pullable://10.10.102.213:8443/cloudnevro-test/test-netserver@sha256:6720f648b74ed590f36094a1c7a58b01b6881396409784c17f471ecfe445e3fd\",\"containerID\":\"docker://d505f50edb4e204cf31840e3cb8d26d33f212d4ebef994d0c3fc151d57e17413\",\"started\":true}]}}"
	deletedObj := new(corev1.Pod)
	_ = json.Unmarshal([]byte(deletedObjJson), deletedObj)

	// Note that uid is different
	newAddObjJson := "{\"metadata\":{\"name\":\"testdemo2-0\",\"namespace\":\"test-ns\",\"uid\":\"00000000-5fb3-4eb9-9de8-2bd4b51606ba\",\"resourceVersion\":\"44895978\"},\"spec\":{\"containers\":[{\"name\":\"testdemo2\",\"ports\":[{\"containerPort\":9001,\"protocol\":\"TCP\",\"hostPort\":9001}]}]},\"status\":{\"phase\":\"Running\",\"podIP\":\"192.168.136.210\",\"hostIP\":\"10.10.10.101\",\"containerStatuses\":[{\"name\":\"testdemo2\",\"state\":{\"running\":{\"startedAt\":\"2022-05-25T08:55:38Z\"}},\"lastState\":{},\"ready\":true,\"restartCount\":5,\"image\":\"\",\"imageID\":\"docker-pullable://10.10.102.213:8443/cloudnevro-test/test-netserver@sha256:6720f648b74ed590f36094a1c7a58b01b6881396409784c17f471ecfe445e3fd\",\"containerID\":\"docker://d000f50edb4e204cf31840e3cb8d26d33f212d4ebef994d0c3fc151d57e17413\",\"started\":true}]}}"
	newAddObj := new(corev1.Pod)
	_ = json.Unmarshal([]byte(newAddObjJson), newAddObj)

	stopCh := make(chan struct{})
	go podDeleteLoop(20*time.Millisecond, 100*time.Millisecond, stopCh)

	AddPod(addObj)
	// Check if the container can be found
	assertFindPod(t, addObj)

	DeletePod(deletedObj)
	AddPod(newAddObj)
	time.Sleep(200 * time.Millisecond)

	// Check if the new container can be found
	assertFindPod(t, newAddObj)
}

func assertFindPod(t *testing.T, pod *corev1.Pod) {
	_, find := MetaDataCache.GetByContainerId(TruncateContainerId(pod.Status.ContainerStatuses[0].ContainerID))
	assert.True(t, find, "Didn't find the new container ID in MetaDataCache")
	_, find = MetaDataCache.GetContainerByIpPort(pod.Status.PodIP, uint32(pod.Spec.Containers[0].Ports[0].ContainerPort))
	assert.True(t, find, "Didn't find the new container IP Port in MetaDataCache")
	_, find = MetaDataCache.GetContainerByHostIpPort(pod.Status.HostIP, uint32(pod.Spec.Containers[0].Ports[0].HostPort))
	assert.True(t, find, "Didn't find the new HostIP Port in MetaDataCache")
}

func Test_extractDeploymentName(t *testing.T) {
	type args struct {
		replicaSetName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{args: args{"nginx-deployment-75675f5897"}, want: "nginx-deployment"},
		{args: args{"nginx-75675f5897"}, want: "nginx"},
		{args: args{"nginx75675f5897"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, extractDeploymentName(tt.args.replicaSetName), "extractDeploymentName(%v)", tt.args.replicaSetName)
		})
	}
}

func TestDelete(t *testing.T) {
	pm := newPodMap()

	info := &K8sPodInfo{
		Namespace: "test-namespace",
		PodName:   "test-podname",
	}
	pm.add(info)

	deletedPodInfo, ok := pm.delete("test-namespace", "test-podname")
	assert.True(t, ok, "expected pod to be deleted successfully")
	assert.Equal(t, info, deletedPodInfo, "deleted pod info should match the added one")

	_, ok = pm.delete("test-namespace", "nonexistent-pod")
	assert.False(t, ok, "expected pod deletion to fail for nonexistent pod")
}