// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package kubernetes

const DeploymentKind = "deployment"

// CompleteGVK returns the complete string of the workload kind.
// If apiVersion is not one of the built-in groupVersion(see scheme.go), return {apiVersion}-{kind};
// return {kind}, otherwise.
func CompleteGVK(apiVersion string, kind string) string {
	if apiVersion == "" {
		return kind
	}
	// builtInScheme is a package-scope variable
	if builtInScheme.IsBuiltInGV(apiVersion) {
		return kind
	} else {
		return apiVersion + "/" + kind
	}
}

func mapKey(namespace string, name string) string {
	return namespace + "/" + name
}
