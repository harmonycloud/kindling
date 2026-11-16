// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

//go:build !linux

package conntracker

func NewConntracker(config *Config) (Conntracker, error) {
	return NewNoopConntracker(config), nil
}
