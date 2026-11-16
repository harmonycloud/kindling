// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package controller

/*
#cgo LDFLAGS: -L ./ -lkindling  -lstdc++ -ldl
#cgo CFLAGS: -I .
#include <stdlib.h>
#include <stdint.h>
#include <stdio.h>
#include "../receiver/cgoreceiver/cgo_func.h"
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"github.com/Kindling-project/kindling/collector/pkg/component"
)

const (
	NoError        = 1
	StartWithError = 2
	StopWithError  = 3
	NoOperation    = 4
)

const ProfileModule = "profile"

type Profile struct {
	Module
}

type ProfileOption struct {
	// seconds
	Duration time.Duration

	// bytes
	FileWatch *FileWatchOption
}

type FileWatchOption struct {
	File     string
	DataSize int
}

type ExportSubModule func() (name string, start func() error, stop func() error)

func NewProfileController(tools *component.TelemetryTools) *Profile {
	profile := NewModule(ProfileModule, tools, Stopped)
	return &Profile{
		Module: profile,
	}
}

func (p *Profile) RegistSubModules(subModules ...ExportSubModule) {
	for _, subModule := range subModules {
		p.Module.RegisterSubModule(subModule())
	}
}

func (o *ProfileOption) UnmarshalJSON(b []byte) error {
	var v struct {
		FileWatch *FileWatchOption
		Duration  interface{}
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	o.FileWatch = v.FileWatch
	switch value := v.Duration.(type) {
	case float64:
		o.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		o.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

func (p *Profile) GetModuleKey() string {
	return p.Name()
}

func startAttachAgent(pid int) string {
	result := C.startAttachAgent(C.int(pid))
	errorMsg := C.GoString(result)
	C.free(unsafe.Pointer(result))
	if len(errorMsg) > 0 {
		return errorMsg
	}
	return ""
}

func stopAttachAgent(pid int) string {
	result := C.stopAttachAgent(C.int(pid))
	errorMsg := C.GoString(result)
	C.free(unsafe.Pointer(result))
	if len(errorMsg) > 0 {
		return errorMsg
	}
	return ""
}

func startDebug(pid int, tid int) error {
	C.startProfileDebug(C.int(pid), C.int(tid))
	return nil
}

func stopDebug() error {
	C.stopProfileDebug()
	return nil
}

func (p *Profile) HandRequest(req *ControlRequest) *ControlResponse {
	switch req.Operation {
	case "start":
		opts := p.GetOptions(req.Options)
		if err := p.Start(opts...); err != nil {
			return &ControlResponse{
				Code: StartWithError,
				Msg:  err.Error(),
			}
		}
		return &ControlResponse{
			Code: NoError,
			Msg:  "start success",
		}
	case "stop":
		if err := p.Stop("module stop by manual"); err != nil {
			return &ControlResponse{
				Code: StopWithError,
				Msg:  err.Error(),
			}
		}
		return &ControlResponse{
			Code: NoError,
			Msg:  "stop success",
		}
	case "start_attach_agent":
		if errMsg := startAttachAgent(req.Pid); errMsg != "" {
			return &ControlResponse{
				Code: StartWithError,
				Msg:  errMsg,
			}
		}
		return &ControlResponse{
			Code: NoError,
			Msg:  "start success",
		}
	case "stop_attach_agent":
		if errMsg := stopAttachAgent(req.Pid); errMsg != "" {
			return &ControlResponse{
				Code: StopWithError,
				Msg:  errMsg,
			}
		}
		return &ControlResponse{
			Code: NoError,
			Msg:  "stop success",
		}
	case "status":
		var status string
		switch p.Status() {
		case Started:
			status = "running"
		case Stopped:
			status = "stopped"
		}
		return &ControlResponse{
			Code: NoError,
			Msg:  status,
		}
	case "start_debug":
		if err := startDebug(req.Pid, req.Tid); err != nil {
			return &ControlResponse{
				Code: StopWithError,
				Msg:  err.Error(),
			}
		}
		return &ControlResponse{
			Code: NoError,
			Msg:  "start debug success",
		}
	case "stop_debug":
		stopDebug()
		return &ControlResponse{
			Code: NoError,
			Msg:  "stop debug success",
		}

	default:
		return &ControlResponse{
			Code: NoOperation,
			Msg:  fmt.Sprintf("unexpected operation:%s", req.Operation),
		}
	}
}

func (p *Profile) GetOptions(raw_opts *json.RawMessage) []Option {
	if raw_opts == nil {
		return nil
	}
	var pOption ProfileOption
	var opts []Option
	json.Unmarshal(*raw_opts, &pOption)
	if pOption.Duration > 0 {
		opts = append(opts, WithStopInterval(pOption.Duration))
	}
	if pOption.FileWatch != nil {
		opts = append(opts, WithStopSignal(fileSizeWatch(pOption.FileWatch)))
	}
	return opts
}

func fileSizeWatch(filewatch *FileWatchOption) <-chan struct{} {
	// TDDO
	panic("file size watch is not implemented yet")
}
