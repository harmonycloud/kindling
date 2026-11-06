#include "profiling_cgo.h"
#include <exception>
#include "profiling_inspector.h"
#include "sinsp.h"
#include "profiling/pid_filter.h"
#include "profiling/profile_trace_reader.h"
#include "spdlog/spdlog.h"

int initCPUProfiling(void) {
    sinsp* inspector = profiling_inspector();
    if (!inspector) return -1;
    try {
        inspector->set_cpu_sampling(PROFILING_DISABLED);
        inspector->clear_profile_pid_config();
        return kindling::profiling::profile_trace_reader().init(inspector) ? 0 : -1;
    } catch (const std::exception& e) {
        spdlog::error("CPU profiling initialization failed: {}", e.what());
        return -1;
    }
}
void shutdownCPUProfiling(void) {
    kindling::profiling::profile_trace_reader().close();
}
int enableNativeProfile(uint32_t pid) {
    return kindling::profiling::EnableNativeProfile(pid);
}
int disableNativeProfile(uint32_t pid) {
    return kindling::profiling::DisableNativeProfile(pid);
}
int pollProfileTraces(ProfileTraceSample* out, int max_n) {
    return kindling::profiling::profile_trace_reader().poll(out, max_n);
}
