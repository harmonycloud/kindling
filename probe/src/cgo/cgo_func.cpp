//
// Created by jundi zhou on 2022/6/1.
//

#include "cgo_func.h"
#include "kindling.h"
#include "catch_sig.h"
#include <cstdlib>

int runForGo() { return init_probe(); }

int getEventsByInterval(int interval, void** kindlingEvent, void* count) { 
    return get_events_by_interval((uint64_t)interval, kindlingEvent, count); 
}

int startProfile() { return start_profile(); }
int stopProfile() { return stop_profile(); }

char* startAttachAgent(int pid) { return start_attach_agent(pid); }
char* stopAttachAgent(int pid) { return stop_attach_agent(pid); }

void suppressEventsCommForGo(char *comm) { suppress_events_comm(string(comm)); }
void subEventForGo(char* eventName, char* category, void *params) { sub_event(eventName, category, (event_params_for_subscribe *)params); }

void startProfileDebug(int pid, int tid) { start_profile_debug(pid, tid); }
void stopProfileDebug() { stop_profile_debug(); }

void getCaptureStatistics(struct capture_statistics_for_go* stats) { get_capture_statistics(stats); }
void catchSignalUp() { sig_set_up(); }

int getProfileData(struct sample_key key, struct bpf_profile_data* sample) {
    return get_profile_data(key, sample);
}

int getProfileKeys(struct sample_key_set *set) {
    return get_profile_keys(set);
}

void freeProfileKeys(struct sample_key_set *set) {
    if (set && set->keys) {
        free(set->keys);
        set->keys = NULL;
    }
    if (set) set->nr_keys = 0;
}

void clearProfileMap() { clear_profile_map(); }
void clearStacksMap() { clear_stacks_map(); }

/* Keep ENABLED for API parity; open-source default path uses All. */
void startCpuSampeling() { start_cpu_sampeling(); }
void startCpuSampelingAll() { start_cpu_sampeling_all(); }
void stopCpuSampeling() { stop_cpu_sampeling(); }

int setProfilePids(const uint32_t *pids, uint32_t n) {
    return set_profile_pids(pids, n);
}

int symbolizerInit(void) { return symbolizer_init(); }
void symbolizerShutdown(void) { symbolizer_shutdown(); }

int symbolizeProcessNamesWithFallback(
        uint32_t pid,
        const uint64_t *addrs, uint32_t n,
        int debug_syms, int perf_map, int map_files,
        char ***out_names, uint32_t *out_n) {
    return symbolize_process_names_with_fallback(
            pid, addrs, n, debug_syms, perf_map, map_files, out_names, out_n);
}

int symbolizeKernelNames(
        const uint64_t *addrs, uint32_t n,
        char ***out_names, uint32_t *out_n) {
    return symbolize_kernel_names(addrs, n, out_names, out_n);
}

void freeNames(char **names, uint32_t n) { free_names(names, n); }
void maintainBccCaches(int64_t seconds) { maintain_caches(seconds); }
