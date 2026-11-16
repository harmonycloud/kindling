// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"encoding/json"
	"testing"
)

func TestMetric_Marshal(t *testing.T) {
	metrics := &Metric{
		Name: "ThisIsATestMetric",
		Data: &Int{Value: 10000},
	}

	jsonString, err := json.Marshal(metrics)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(jsonString))
}
