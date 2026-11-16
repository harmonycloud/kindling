// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package processor

import (
	"github.com/Kindling-project/kindling/collector/pkg/component/consumer"
)

type Processor interface {
	consumer.Consumer
}
