// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package analyzer

import "github.com/Kindling-project/kindling/collector/pkg/model"

type Type string

func (t Type) String() string {
	return string(t)
}

type Analyzer interface {
	// Start initializes the analyzer
	Start() error
	// ConsumeEvent gets the event from the previous component
	ConsumeEvent(event *model.KindlingEvent) error
	// Shutdown cleans all the resources used by the analyzer
	Shutdown() error
	// Type returns the type of the analyzer
	Type() Type
	// ConsumableEvents returns the events' name that this analyzer can consume
	ConsumableEvents() []string
}
