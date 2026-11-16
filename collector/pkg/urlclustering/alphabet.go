// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package urlclustering

import (
	"regexp"
	"strings"
)

// AlphabeticClusteringMethod clustering all the non-alphabetic characters to '*'
// and will trim the parameters at the end of the string.
type AlphabeticClusteringMethod struct {
	regexp *regexp.Regexp
}

func NewAlphabeticalClusteringMethod() *AlphabeticClusteringMethod {
	exp, _ := regexp.Compile("^[A-Za-z_-]+$")
	return &AlphabeticClusteringMethod{
		regexp: exp,
	}
}

// ClusteringBaseline is a more readable version of Clustering() but with a poor performance.
// Don't use this function at anytime.
func (m *AlphabeticClusteringMethod) ClusteringBaseline(endpoint string) string {
	if endpoint == "" {
		return ""
	}
	index := strings.Index(endpoint, "?")
	if index != -1 {
		endpoint = endpoint[:index]
	}

	// Split the endpoint into multiple segments
	endpoint = strings.TrimSpace(endpoint)
	segments := strings.Split(endpoint, "/")
	// Iterate over all parts and execute the regular expression.
	resultSegments := make([]string, 0, len(segments))
	for i := 0; i < len(segments); i++ {
		// If the current segment is too long, we consider it as a high-cardinality variable.
		if len(segments[i]) > 25 {
			resultSegments = append(resultSegments, "*")
			continue
		}
		if segments[i] == "" || m.regexp.MatchString(segments[i]) {
			resultSegments = append(resultSegments, segments[i])
		} else {
			// If the segment is composed of non-alphabetic characters, we replace it with a star.
			resultSegments = append(resultSegments, "*")
		}
	}
	// Re-combine all parts together
	var resultEndpoint string
	for i, seg := range resultSegments {
		resultEndpoint = resultEndpoint + seg
		if i != len(resultSegments)-1 {
			resultEndpoint += "/"
		}
	}

	return resultEndpoint
}

func (m *AlphabeticClusteringMethod) Clustering(endpoint string) string {
	if endpoint == "" {
		return ""
	}
	endpointBytes := []byte(endpoint)

	resultBytes := make([]byte, 0)

	currentSegmentIsStar := false
	currentSegment := make([]byte, 0)

	for _, b := range endpointBytes {
		if b == ' ' {
			continue
		}
		// End of the current segment.
		if b == '/' || b == '?' {
			if currentSegmentIsStar {
				resultBytes = append(resultBytes, '*')
			} else {
				// currentSegment could be empty
				resultBytes = append(resultBytes, currentSegment...)
			}
			currentSegment = make([]byte, 0)
			currentSegmentIsStar = false
			if b == '/' {
				resultBytes = append(resultBytes, '/')
				continue
			} else if b == '?' {
				break
			}
		}
		if currentSegmentIsStar {
			continue
		}
		if isAlphabetical(b) {
			currentSegment = append(currentSegment, b)
			// If the current segment is too long, we consider it as a high-cardinality variable.
			if len(currentSegment) > 25 {
				currentSegmentIsStar = true
			}
		} else {
			currentSegmentIsStar = true
		}
	}
	// Deal with the last segment.
	if currentSegmentIsStar {
		resultBytes = append(resultBytes, '*')
	} else if len(currentSegment) != 0 {
		resultBytes = append(resultBytes, currentSegment...)
	}
	return string(resultBytes)
}

func isAlphabetical(b byte) bool {
	return b == '-' || b == '_' || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
