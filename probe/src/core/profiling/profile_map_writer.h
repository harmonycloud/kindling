// Copyright (C) 2026 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

#pragma once

#include "cgo/profiling_cgo.h"

/* MapWriter: writers for profile_* native-unwind and Python preprocessing maps.
 * C API declared in cgo/profiling_cgo.h; implemented here.
 * Go callers use processmanager/ebpf (ebpfapi.Handler), not these C symbols directly.
 */
