// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package k8sprocessor

import (
	"github.com/Kindling-project/kindling/collector/pkg/metadata/kubernetes"
)

type Config struct {
	KubeAuthType  kubernetes.AuthType `mapstructure:"kube_auth_type"`
	KubeConfigDir string              `mapstructure:"kube_config_dir"`
	// GraceDeletePeriod controls the delay interval after receiving delete event.
	// The unit is seconds, and the default value is 60 seconds.
	// Should not be lower than 30 seconds.
	GraceDeletePeriod int `mapstructure:"grace_delete_period"`
	// EnableFetchReplicaSet controls whether to fetch ReplicaSet information.
	// The default value is false. It should be enabled if the ReplicaSet
	// is used to control pods in the third-party CRD except for Deployment.
	EnableFetchReplicaSet bool `mapstructure:"enable_fetch_replicaset"`
	// Set "Enable" false if you want to run the agent in the non-Kubernetes environment.
	// Otherwise, the agent will panic if it can't connect to the API-server.
	Enable bool `mapstructure:"enable"`

	// MetaDataProviderConfig is optional config to use another source of K8sMetadata named metadata-provider
	// Used to reduce the stress caused by agent directly on APIServer
	// Set "metadata_provider_config.enable" true and "metadata_provider_config.endpoint" as target service to enable it
	MetaDataProviderConfig *kubernetes.MetaDataProviderConfig `mapstructure:"metadata_provider_config"`
}

var DefaultConfig Config = Config{
	KubeAuthType:           "serviceAccount",
	KubeConfigDir:          "/root/.kube/config",
	GraceDeletePeriod:      60,
	Enable:                 true,
	MetaDataProviderConfig: &kubernetes.MetaDataProviderConfig{Enable: false, EnableTrace: false, Endpoint: ""},
}
