// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/metadata/kubernetes"
)

func TestMetaDataWrapper_ListAndWatch(t *testing.T) {
	authType := flag.String("authType", "kubeConfig", "AuthType describes the type of authentication to use for the K8s API, support 'kubeConfig' or 'serviceAccount'. ")
	kubeConfigPath := flag.String("kubeConfig", "/root/.kube/config", "kubeConfig describe the filePath to your kubeConfig,only used when authType is 'kubeConfig'")
	httpPort := flag.Int("http-port", 9504, "port describe which port will be used to expose data")
	enableFetchReplicaset := flag.Bool("enableFetchReplicaset", false, "controls whether to fetch ReplicaSet information. The default value is false. It should be enabled if the ReplicaSet is used to control pods in the third-party CRD except for Deployment.")
	logInterval := flag.Int("logInterval", 10, "Interval(Second) to show how many event mp received, default 120s")

	flag.Parse()

	config := &Config{
		KubeAuthType:          kubernetes.AuthType(*authType),
		KubeConfigDir:         *kubeConfigPath,
		EnableFetchReplicaSet: *enableFetchReplicaset,
		LogInterval:           *logInterval,
	}

	if mdw, err := NewMetaDataWrapper(config); err != nil {
		log.Fatalf("create MetaData Wrapper failed, err: %v", err)
	} else {
		http.HandleFunc("/listAndWatch", mdw.ListAndWatch)
		log.Printf("[http] service start at port: %d", *httpPort)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *httpPort), nil))
	}
}
