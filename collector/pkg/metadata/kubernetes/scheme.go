// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import "k8s.io/client-go/kubernetes/scheme"

var builtInScheme = NewKnownScheme()

type BuiltInScheme struct {
	groupVersions map[string]bool
}

func NewKnownScheme() *BuiltInScheme {
	groupVersions := scheme.Scheme.PrioritizedVersionsAllGroups()
	ret := make(map[string]bool)
	for _, groupVersion := range groupVersions {
		ret[groupVersion.String()] = true
	}
	return &BuiltInScheme{
		groupVersions: ret,
	}
}

func (s *BuiltInScheme) IsBuiltInGV(groupVersion string) bool {
	return s.groupVersions[groupVersion]
}
