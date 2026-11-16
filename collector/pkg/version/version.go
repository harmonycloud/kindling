// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

// version.go

package version

import (
	"fmt"
	"runtime"
)

var (
	CodeVersion = ""
)

func RuntimeVersion() string {
	return fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func CodeBaseVersion() string {
	return CodeVersion
}

func Version() string {
	return fmt.Sprintf("%s; %s", CodeBaseVersion(), RuntimeVersion())
}