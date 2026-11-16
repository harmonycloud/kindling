// Copyright (C) 2026 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

#pragma once
#include <cstdint>
namespace kindling { namespace profiling {
int EnableNativeProfile(uint32_t pid);
int DisableNativeProfile(uint32_t pid);
}} // namespace kindling::profiling
