// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package esclient

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewEsClient(t *testing.T) {
	_, err := NewEsClient("http://10.10.100.100:9000")
	assert.Error(t, err)
	aliveEsHost := "http://10.10.103.96:9200"
	esClient, err := NewEsClient(aliveEsHost)
	if err != nil {
		t.Logf("esClient cannot connect to %s, %v", aliveEsHost, err)
		return
	}
	for i := 0; i < 1000; i++ {
		esClient.AddIndexRequestWithParams("test_index", "{\"test\": \"test\"}")
	}
	_ = esClient.bulkProcessor.Flush()
}
