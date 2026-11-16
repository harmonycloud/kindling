// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package cpuanalyzer

import (
	"errors"
	"time"

	"go.uber.org/atomic"
)

type ScheduledTask interface {
	run()
}

type ScheduledTaskRoutine struct {
	expiredDuration time.Duration
	ticker          *time.Ticker
	timer           *time.Timer
	stopCh          chan struct{}

	task      ScheduledTask
	isRunning *atomic.Bool

	expiredCallback func()
}

// NewAndStartScheduledTaskRoutine creates a new routine and start it immediately.
func NewAndStartScheduledTaskRoutine(
	tickerDuration time.Duration,
	expiredDuration time.Duration,
	task ScheduledTask,
	expiredCallback func()) *ScheduledTaskRoutine {
	ret := &ScheduledTaskRoutine{
		expiredDuration: expiredDuration,
		ticker:          time.NewTicker(tickerDuration),
		timer:           time.NewTimer(expiredDuration),
		task:            task,
		isRunning:       atomic.NewBool(false),
		stopCh:          make(chan struct{}),
		expiredCallback: expiredCallback,
	}
	// Start the routine once it is created.
	_ = ret.Start()
	return ret
}

func (s *ScheduledTaskRoutine) Start() error {
	swapped := s.isRunning.CAS(false, true)
	if !swapped {
		return errors.New("the routine has been started")
	}
	go func() {
		if s.expiredCallback != nil {
			defer s.expiredCallback()
		}
		for {
			select {
			case <-s.ticker.C:
				// do some work
				s.task.run()
			case <-s.timer.C:
				// The current task is expired.
				s.isRunning.CAS(true, false)
				s.ticker.Stop()
				return
			case <-s.stopCh:
				s.timer.Stop()
				s.ticker.Stop()
				return
			}
		}
	}()
	return nil
}

func (s *ScheduledTaskRoutine) Stop() error {
	swapped := s.isRunning.CAS(true, false)
	if !swapped {
		return errors.New("the routine is not running")
	}
	s.stopCh <- struct{}{}
	return nil
}
