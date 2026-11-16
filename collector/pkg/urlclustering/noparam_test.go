// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package urlclustering

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func noParamTestcases() []testCase {
	return []testCase{
		{"/", "/"},
		{"/A/b/20000/", "/A/b/20000/"},
		{" /test/22?v=a", "/test/22"},
		{"/a12?a=1132", "/a12"},
		{"/abcd/1234a/efg/b&*", "/abcd/1234a/efg/b&*"},
		// Double slashes is valid but not recommended
		{"/a//b/c?d=2&e=3", "/a//b/c"},
	}
}

func TestNoParamClusteringMethod_Clustering(t *testing.T) {
	method := NewNoParamClusteringMethod()
	testCases := noParamTestcases()
	for _, c := range testCases {
		assert.Equal(t, c.want, method.Clustering(c.endpoint))
	}
}
