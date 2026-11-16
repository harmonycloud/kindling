// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package urlclustering

type ClusteringMethod interface {
	// Clustering receives a no-host endpoint string of the HTTP request and
	// return its clustering result.
	// Some examples of the endpoint:
	// - /path/to/file/1
	// - /part1/part2/2?param=1
	// - /CloudDetective-Harmonycloud/kindling
	Clustering(endpoint string) string
}

var (
	alphabeticMethod ClusteringMethod
	noParamMethod    ClusteringMethod
)

// AlphabeticClustering is a convenient method that calls AlphabeticClusteringMethod.Clustering().
// This method is not thread-safe.
func AlphabeticClustering(endpoint string) string {
	if alphabeticMethod == nil {
		alphabeticMethod = NewAlphabeticalClusteringMethod()
	}
	return alphabeticMethod.Clustering(endpoint)
}

// NoParamClustering is a global method that calls NoParamClusteringMethod.Clustering().
// This method is not thread-safe.
func NoParamClustering(endpoint string) string {
	if noParamMethod == nil {
		noParamMethod = NewNoParamClusteringMethod()
	}
	return noParamMethod.Clustering(endpoint)
}
