// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package urlclustering

import (
	"strings"
)

// NoParamClusteringMethod removes the parameters that are end of the URL string.
type NoParamClusteringMethod struct {
}

func NewNoParamClusteringMethod() ClusteringMethod {
	return &NoParamClusteringMethod{}
}

func (m *NoParamClusteringMethod) Clustering(endpoint string) string {
	if endpoint == "" {
		return ""
	}
	endpoint = strings.TrimSpace(endpoint)
	// Remove the parameters first
	index := strings.Index(endpoint, "?")
	if index != -1 {
		endpoint = endpoint[:index]
	}

	return endpoint
}
