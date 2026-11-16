// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cameraexporter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDateString(t *testing.T) {
	nanoseconds := 1672887314101975422
	dateString := getDateString(int64(nanoseconds))
	assert.Equal(t, "20230105105514.101975422", dateString)
}
