#pragma once
#include <cstdint>
namespace kindling { namespace profiling {
int EnableNativeProfile(uint32_t pid);
int DisableNativeProfile(uint32_t pid);
}} // namespace kindling::profiling
