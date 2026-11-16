// Copyright (C) 2026 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

#include "pid_filter.h"

#include "sinsp.h"
#include "cgo/profiling_inspector.h"

namespace kindling {
namespace profiling {

int EnableNativeProfile(uint32_t pid) {
    if (!profiling_inspector() || pid == 0) return -1;
    pid_config config = {PROFILING_TYPE_FRAMEPOINTERS, 1, 0};
    return profiling_inspector()->set_profile_pid_config(pid, &config);
}

int DisableNativeProfile(uint32_t pid) {
    if (!profiling_inspector() || pid == 0) return -1;
    return profiling_inspector()->unset_profile_pid_config(pid);
}

} // namespace profiling
} // namespace kindling
