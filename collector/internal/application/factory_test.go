// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/Kindling-project/kindling/collector/pkg/component/analyzer/network"
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer/processor/k8sprocessor"
	"github.com/Kindling-project/kindling/collector/pkg/component/receiver/cgoreceiver"
	"github.com/Kindling-project/kindling/collector/pkg/metadata/kubernetes"
)

func TestConstructConfig(t *testing.T) {
	factory := NewComponentsFactory()
	factory.RegisterProcessor(k8sprocessor.K8sMetadata, k8sprocessor.NewKubernetesProcessor, &k8sprocessor.DefaultConfig)
	factory.RegisterAnalyzer(network.Network.String(), network.NewNetworkAnalyzer, network.NewDefaultConfig())
	factory.RegisterReceiver(cgoreceiver.Cgo, cgoreceiver.NewCgoReceiver, cgoreceiver.NewDefaultConfig())

	// Construct the config from the yaml file
	v := viper.New()
	v.SetConfigFile("./testdata/kindling-collector-config.yaml")
	err := v.ReadInConfig()
	assert.NoError(t, err)

	err = factory.ConstructConfig(v)
	assert.NoError(t, err)

	//// Assert the config is as expected
	k8sProcessorFactory := factory.Processors[k8sprocessor.K8sMetadata]
	k8sCfg := k8sProcessorFactory.Config.(*k8sprocessor.Config)
	// The expected config is exactly the opposite of the default config
	expectedCfg := &k8sprocessor.Config{
		Enable:                false,
		KubeAuthType:          "kubeConfig",
		KubeConfigDir:         "/opt/.kube/config",
		GraceDeletePeriod:     30,
		EnableFetchReplicaSet: true,
		MetaDataProviderConfig: &kubernetes.MetaDataProviderConfig{
			Enable:      false,
			EnableTrace: false,
			Endpoint:    "",
		},
	}
	assert.Equal(t, expectedCfg, k8sCfg)

	networkAnalyzerFactory := factory.Analyzers[network.Network.String()]
	networkConfig := networkAnalyzerFactory.Config
	expectedNetworkConfig := &network.Config{
		EventChannelSize:      10000,
		EnableTimeoutCheck:    true,
		ConnectTimeout:        100,
		FdReuseTimeout:        15,
		NoResponseThreshold:   120,
		ResponseSlowThreshold: 500,
		EnableConntrack:       true,
		ConntrackMaxStateSize: 131072,
		ConntrackRateLimit:    500,
		ProcRoot:              "/proc",
		// Case: This slice is from the default config. The config file doesn't have this field.
		ProtocolParser: []string{"http", "mysql", "dns", "redis", "kafka", "dubbo"},
		// Case: This slice is overridden by the config file. The default config is different.
		ProtocolConfigs: []network.ProtocolConfig{
			{
				Key:           "http",
				Ports:         []uint32{80, 8080},
				PayloadLength: 100,
				Threshold:     200,
			},
		},
		UrlClusteringMethod: "blank",

		IgnoreDnsRcode3Error: true,
	}
	assert.Equal(t, expectedNetworkConfig, networkConfig)

	cgoreceiverFactory := factory.Receivers[cgoreceiver.Cgo]
	cgoreceiverConfig := cgoreceiverFactory.Config
	expectedCgoreceiverConfig := &cgoreceiver.Config{
		SubscribeInfo: []cgoreceiver.SubEvent{
			{
				Name:     "syscall_exit-writev",
				Category: "net",
			},
			{
				Name:     "syscall_exit-readv",
				Category: "net",
			},
			{
				Name:     "syscall_exit-write",
				Category: "net",
			},
			{
				Name:     "syscall_exit-read",
				Category: "net",
			},
			{
				Name:     "syscall_exit-sendto",
				Category: "net",
			},
			{
				Name:     "syscall_exit-recvfrom",
				Category: "net",
			},
			{
				Name:     "syscall_exit-sendmsg",
				Category: "net",
			},
			{
				Name:     "syscall_exit-recvmsg",
				Category: "net",
			},
			{
				Name:     "syscall_exit-sendmmsg",
				Category: "net",
			},
			{
				Name: "kprobe-tcp_close",
			},
			{
				Name: "kprobe-tcp_rcv_established",
			},
			{
				Name: "kprobe-tcp_drop",
			},
			{
				Name: "kprobe-tcp_retransmit_skb",
			},
			{
				Name: "syscall_exit-connect",
			},
			{
				Name: "kretprobe-tcp_connect",
			},
			{
				Name: "kprobe-tcp_set_state",
			},
			{
				Name: "tracepoint-procexit",
			},
		},
		ProcessFilterInfo: cgoreceiver.ProcessFilter{
			Comms: []string{"kindling-collec", "containerd", "dockerd", "containerd-shim", "filebeat", "java"},
		},
	}
	assert.Equal(t, expectedCgoreceiverConfig, cgoreceiverConfig)
}
