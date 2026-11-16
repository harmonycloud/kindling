// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package service

import "github.com/Kindling-project/kindling/collector/pkg/metadata/kubernetes"

type Config struct {
	KubeAuthType  kubernetes.AuthType
	KubeConfigDir string
	// EnableFetchReplicaSet controls whether to fetch ReplicaSet information.
	// The default value is false. It should be enabled if the ReplicaSet
	// is used to control pods in the third-party CRD except for Deployment.
	EnableFetchReplicaSet bool
	LogInterval           int
}
