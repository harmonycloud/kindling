// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"fmt"
	"strings"

	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/Kindling-project/kindling/collector/pkg/metadata/kubernetes"
)

type AddObj func(obj interface{})

type UpdateObj func(oldObj interface{}, newObj interface{})

type DeleteObj func(obj interface{})

type Operation string

const (
	Add    Operation = "add"
	Update Operation = "update"
	Delete Operation = "delete"
)

type MetaDataService interface {
	ListAndWatch()
}

type MetaDataRsyncResponse struct {
	Type      string // pod,rs,node,service
	Operation string // add,update,delete
	NewObj    interface{}
	OldObj    interface{}
}

type MetaDataVO struct {
	Type      string // pod,rs,node,service
	Operation string // add,update,delete
	NewObj    json.RawMessage
	OldObj    json.RawMessage
}

type ListVO struct {
	Cache             *kubernetes.K8sMetaDataCache
	GlobalNodeInfo    *kubernetes.NodeMap
	GlobalRsInfo      *kubernetes.ReplicaSetMap
	GlobalServiceInfo *kubernetes.ServiceMap
}

func MetaDataVO2String(resp *MetaDataVO) string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("Operation: [%s], ResType: [%s]", resp.Operation, resp.Type))

	switch resp.Type {
	case "pod":
		if resp.NewObj != nil {
			obj := corev1.Pod{}
			err := json.Unmarshal(resp.NewObj, &obj)
			if err == nil {
				str.WriteString(fmt.Sprintf(", newObj: [%s/%s]", obj.Namespace, obj.Name))
			}
		}
		if resp.OldObj != nil {
			obj := corev1.Pod{}
			err := json.Unmarshal(resp.OldObj, &obj)
			if err == nil {
				str.WriteString(fmt.Sprintf(", oldObj: [%s/%s]", obj.Namespace, obj.Name))
			}
		}
	case "rs":
		if resp.NewObj != nil {
			obj := appv1.ReplicaSet{}
			err := json.Unmarshal(resp.NewObj, &obj)
			if err == nil {
				str.WriteString(fmt.Sprintf(", newObj: [%s/%s]", obj.Namespace, obj.Name))
			}
		}
		if resp.OldObj != nil {
			obj := appv1.ReplicaSet{}
			err := json.Unmarshal(resp.OldObj, &obj)
			if err == nil {
				str.WriteString(fmt.Sprintf(", oldObj: [%s/%s]", obj.Namespace, obj.Name))
			}
		}
	case "service":
		if resp.NewObj != nil {
			obj := corev1.Service{}
			err := json.Unmarshal(resp.NewObj, &obj)
			if err == nil {
				str.WriteString(fmt.Sprintf(", newObj: [%s/%s]", obj.Namespace, obj.Name))
			}
		}
		if resp.OldObj != nil {
			obj := corev1.Service{}
			err := json.Unmarshal(resp.OldObj, &obj)
			if err == nil {
				str.WriteString(fmt.Sprintf(", oldObj: [%s/%s]", obj.Namespace, obj.Name))
			}
		}
	case "node":
		if resp.NewObj != nil {
			obj := corev1.Node{}
			err := json.Unmarshal(resp.NewObj, &obj)
			if err == nil {
				str.WriteString(fmt.Sprintf(", newObj: [%s]", obj.Name))
			}
		}
		if resp.OldObj != nil {
			obj := corev1.Node{}
			err := json.Unmarshal(resp.OldObj, &obj)
			if err == nil {
				str.WriteString(fmt.Sprintf(", oldObj: [%s]", obj.Name))
			}
		}
	}
	return str.String()
}
