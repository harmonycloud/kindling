// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package metadataclient

import (
	"fmt"
	"testing"

	"github.com/Kindling-project/kindling/collector/pkg/metadata/kubernetes"
)

// func TestClients_ListAndWatch(t *testing.T) {
// 	for i := 0; i < 1000; i++ {
// 		go func() {
// 			cli := NewMetaDataWrapperClient("http://localhost:9504", true)
// 			err := cli.ListAndWatch(kubernetes.SetupCache)
// 			if err != nil {
// 				fmt.Println(err)
// 			}
// 		}()
// 		fmt.Printf("init: %d\n", i)
// 	}

// 	select {}
// }

func TestClient_ListAndWatch(t *testing.T) {
	cli := NewMetaDataWrapperClient("http://localhost:9504", true)
	err := cli.ListAndWatch(kubernetes.SetupCache)
	if err != nil {
		fmt.Println(err)
	}
}
