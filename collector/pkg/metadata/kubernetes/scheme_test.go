// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import "testing"

func TestNewScheme(t *testing.T) {
	scheme := NewKnownScheme()
	t.Log(scheme.groupVersions)
}

func TestBuiltInScheme_IsBuiltInGV(t *testing.T) {
	scheme := NewKnownScheme()
	testCases := []struct {
		in       string
		expected bool
	}{
		{"apps/v1", true},
		{"test", false},
		{"", false},
	}
	for _, testCase := range testCases {
		ret := scheme.IsBuiltInGV(testCase.in)
		if testCase.expected != ret {
			t.Errorf("expected %v, but get %v", testCase.expected, ret)
		}
	}
}
