// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package urlclustering

func NewMethod(urlClusteringMethod string) ClusteringMethod {
	switch urlClusteringMethod {
	case "alphabet":
		return NewAlphabeticalClusteringMethod()
	case "noparam":
		return NewNoParamClusteringMethod()
	case "blank":
		return NewBlankClusteringMethod()
	default:
		return NewAlphabeticalClusteringMethod()
	}
}
