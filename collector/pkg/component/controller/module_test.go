// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"testing"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/component"
)

func TestModule(t *testing.T) {
	sample := NewModule("profile", component.NewTelemetryManager().GetGlobalTelemetryTools(), Stopped)

	stopSignal := make(chan struct{})
	go sample.Start(
		WithStopInterval(5*time.Second),
		WithStopSignal(stopSignal),
	)

	timer := time.NewTimer(6 * time.Second)
	<-timer.C
	close(stopSignal)

	<-time.After(10 * time.Second)
}

func TestReOpenModule(t *testing.T) {
	sample := NewModule("profile", component.NewTelemetryManager().GetGlobalTelemetryTools(), Stopped)

	stopSignal := make(chan struct{})
	go sample.Start(
		WithStopInterval(2*time.Second),
		WithStopSignal(stopSignal),
	)

	timer := time.NewTimer(3 * time.Second)
	<-timer.C
	close(stopSignal)

	stopSignal = make(chan struct{})
	go sample.Start(
		WithStopInterval(3*time.Second),
		WithStopSignal(stopSignal),
	)
	timer = time.NewTimer(2 * time.Second)
	<-timer.C
	close(stopSignal)

	<-time.After(5 * time.Second)
}
