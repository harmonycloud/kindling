// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/Kindling-project/kindling/collector/pkg/component"
	"github.com/hashicorp/go-multierror"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

type Module interface {
	RegisterSubModule(name string, start func() error, stop func() error) error
	Name() string
	Start(...Option) error
	Stop(msg string) error
	Status() ModuleStatus
	ClearSignal() <-chan struct{}
}

type ModuleStatus int

const (
	Started ModuleStatus = iota
	Stopped
)

type StartModule func() error

type StopModule func() error

type DefaultModule struct {
	name string

	subStarts map[string]StartModule
	subStops  map[string]StopModule

	clearSignal chan struct{}
	status      ModuleStatus

	tools *component.TelemetryTools

	check sync.Mutex
}

func NewModule(name string, tools *component.TelemetryTools, status ModuleStatus) *DefaultModule {
	return &DefaultModule{
		name:      name,
		subStarts: make(map[string]StartModule),
		subStops:  make(map[string]StopModule),
		tools:     tools,
		status:    status,
	}
}

func (m *DefaultModule) Name() string {
	return m.name
}

func (m *DefaultModule) RegisterSubModule(subModule string, start func() error, stop func() error) error {
	m.subStarts[subModule] = start
	m.subStops[subModule] = stop
	return nil
}

func (m *DefaultModule) Start(opts ...Option) error {
	m.check.Lock()
	defer m.check.Unlock()
	if m.status != Stopped {
		return fmt.Errorf("module is running now")
	}
	m.clearSignal = make(chan struct{})
	var err error
	for _, start := range m.subStarts {
		if perr := start(); perr != nil {
			err = multierr.Append(err, perr)
		}
	}
	m.status = Started
	for _, opt := range opts {
		opt.apply(m)
	}
	m.tools.Logger.Info("module start", zap.String("module", m.name), zap.Time("startTime", time.Now()))
	return err
}

func (m *DefaultModule) Stop(msg string) error {
	m.check.Lock()
	defer m.check.Unlock()
	if m.status == Stopped {
		m.tools.Logger.Infof("received a stop signal, while the module already stopped")
		return fmt.Errorf("module is stopped now")
	}
	defer close(m.clearSignal)
	var err error
	for _, stop := range m.subStops {
		if perr := stop(); perr != nil {
			err = multierror.Append(err, perr)
		}
	}
	m.status = Stopped
	m.tools.Logger.Info(msg, zap.String("module", m.name), zap.Time("stopTime", time.Now()))
	return err
}

func (m *DefaultModule) Status() ModuleStatus {
	m.check.Lock()
	defer m.check.Unlock()
	return m.status
}

func (m *DefaultModule) ClearSignal() <-chan struct{} {
	return m.clearSignal
}

type Option interface {
	apply(Module)
}

type optionFunc func(Module)

func (f optionFunc) apply(c Module) {
	f(c)
}

func WithStopInterval(duration time.Duration) Option {
	return optionFunc(func(c Module) {
		timer := time.NewTimer(duration)
		go func() {
			select {
			case <-timer.C:
				c.Stop(fmt.Sprintf("module stoped after running %v seconds", duration.Seconds()))
			case <-c.ClearSignal():
				return
			}
		}()
	})
}

func WithStopSignal(stopch <-chan struct{}) Option {
	return optionFunc(func(c Module) {
		go func() {
			select {
			case <-stopch:
				c.Stop("module stoped by manual signal")
			case <-c.ClearSignal():
				return
			}
		}()
	})
}
