// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cpuanalyzer

import (
	"sync"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/model"
)

func (ca *CpuAnalyzer) ProfileModule() (submodule string, start func() error, stop func() error) {
	return "cpuanalyzer", ca.StartProfile, ca.StopProfile
}

func (ca *CpuAnalyzer) StartProfile() error {
	// control flow changed
	// Note that these two variables belongs to the package
	triggerEventChan = make(chan SendTriggerEvent, 3e5)
	traceChan = make(chan *model.DataGroup, 1e4)
	isInstallApm = make(map[uint64]bool, 100000)
	sampleMap = sync.Map{}
	ca.stopProfileChan = make(chan struct{})
	enableProfile = true
	go ca.sampleSend()
	go ca.ReadTriggerEventChan()
	go ca.ReadTraceChan()
	go ca.TidDelete(30*time.Second, 10*time.Second)
	return nil
}

func (ca *CpuAnalyzer) StopProfile() error {
	// control flow changed
	ca.lock.Lock()
	defer ca.lock.Unlock()
	enableProfile = false
	close(ca.stopProfileChan)
	// Clear the old events even if they are not sent
	ca.cpuPidEvents = make(map[uint32]map[uint32]*TimeSegments, 100000)
	// clear the following data via gc
	triggerEventChan = make(chan SendTriggerEvent, 3e5)
	traceChan = make(chan *model.DataGroup, 1e4)
	isInstallApm = make(map[uint64]bool, 100000)
	sampleMap = sync.Map{}
	return nil
}
