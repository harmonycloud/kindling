// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

package receiver

type Receiver interface {
	// Start initializes the receiver and start to receive events.
	Start() error
	// Shutdown closes the receiver and stops receiving events.
	// Note receiver should not shutdown other components though it holds a reference
	Shutdown() error
}
