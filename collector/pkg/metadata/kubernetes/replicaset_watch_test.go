// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"testing"

	appv1 "k8s.io/api/apps/v1"
	apimachinery "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InitGlobalRsInfo() {
	GlobalRsInfo = &ReplicaSetMap{
		Info: make(map[string]Controller),
	}
}

func TestOnAddReplicaSet(t *testing.T) {
	InitGlobalRsInfo()
	AddReplicaSet(CreateReplicaSet())
	owner, ok := GlobalRsInfo.GetOwnerReference(mapKey("CustomNamespace", "deploy-1a2b3c4d"))
	if !ok || owner.Kind != "Deployment" || owner.APIVersion != "apps/v1" {
		t.Errorf("Error")
	}
}

func CreateReplicaSet() *appv1.ReplicaSet {
	isController := true
	return &appv1.ReplicaSet{
		ObjectMeta: apimachinery.ObjectMeta{
			Name:      "deploy-1a2b3c4d",
			Namespace: "CustomNamespace",
			OwnerReferences: []apimachinery.OwnerReference{
				{
					Kind:       "Custom",
					Name:       "deploy",
					APIVersion: "my.apps.io/v1",
				},
				{
					Kind:       "Deployment",
					Name:       "deploy",
					APIVersion: "apps/v1",
					Controller: &isController,
				},
			},
		},
	}
}
