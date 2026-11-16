// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package consumer

import "github.com/Kindling-project/kindling/collector/pkg/model"

type Consumer interface {
	Consume(dataGroup *model.DataGroup) error
}
