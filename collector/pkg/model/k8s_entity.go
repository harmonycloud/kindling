// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package model

// KEntitySet links related kubernetes entities around a container as a whole set named KEntitySet.
type KEntitySet struct {
	// name of namespace related
	// optional, not empty
	// Could be a name of "cloud instance" or a domain name while other fields are empty.
	Namespace string `json:"namespace"`
	// name of service related
	// optional
	SvcName string `json:"svcName,omitempty"`
	// kind of workload related
	// optional
	WorkloadKind string `json:"workloadKind,omitempty"`
	// name of workload related
	// optional
	WorkloadName string `json:"workloadName,omitempty"`
	// name of pod related
	// optional
	PodName string `json:"podName,omitempty"`
	// id of container
	// optional
	ContainerId string `json:"containerId,omitempty"`
	// name of container
	// optional
	ContainerName string `json:"containerName,omitempty"`
	// ip of container
	IP string `json:"ip"`
	// port of container
	// optional
	Port uint32 `json:"port,omitempty"`
	// name of node related
	// optional
	NodeName string `json:"nodeName,omitempty"`
}
