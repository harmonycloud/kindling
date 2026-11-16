// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import "testing"

func TestCompleteGVK(t *testing.T) {
	testCases := []struct {
		APIVersion string
		kind       string
		expected   string
	}{
		{
			APIVersion: "apps/v1",
			kind:       "StatefulSet",
			expected:   "StatefulSet",
		},
		{
			APIVersion: "apps.kruise.io/v1beta1",
			kind:       "StatefulSet",
			expected:   "apps.kruise.io/v1beta1/StatefulSet",
		},
		{
			APIVersion: "",
			kind:       "DaemonSet",
			expected:   "DaemonSet",
		},
	}

	for _, testCase := range testCases {
		ret := CompleteGVK(testCase.APIVersion, testCase.kind)
		if ret != testCase.expected {
			t.Errorf("expect %v, but get %v", testCase.expected, ret)
		}
	}
}
