//
// Created by jundi zhou on 2022/6/1.
//

#ifndef SYSDIG_CGO_FUNC_H
#define SYSDIG_CGO_FUNC_H

#include <stdint.h>
#include "kindling.h"
#include "core/symbolizer/kindling_symbolizer.h"

#ifdef __cplusplus
extern "C" {
#endif
int runForGo();
int getEventsByInterval(int interval, void** kindlingEvent, void* count);
void suppressEventsCommForGo(char *comm);
void subEventForGo(char* eventName, char* category, void* params);
int startProfile();
int stopProfile();
char* startAttachAgent(int pid);
char* stopAttachAgent(int pid);
void startProfileDebug(int pid, int tid);
void stopProfileDebug();
void getCaptureStatistics(struct capture_statistics_for_go* stats);
void catchSignalUp();

/* CPU continuous profiling CGO surface (default sampling: CAPTURE_ALL). */
int getProfileData(struct sample_key key, struct bpf_profile_data* sample);
int getProfileKeys(struct sample_key_set *set);
void freeProfileKeys(struct sample_key_set *set);
void clearProfileMap();
void clearStacksMap();
void startCpuSampeling();
void startCpuSampelingAll();
void stopCpuSampeling();
int setProfilePids(const uint32_t *pids, uint32_t n);

int symbolizerInit(void);
void symbolizerShutdown(void);
int symbolizeProcessNamesWithFallback(
        uint32_t pid,
        const uint64_t *addrs, uint32_t n,
        int debug_syms, int perf_map, int map_files,
        char ***out_names, uint32_t *out_n);
int symbolizeKernelNames(
        const uint64_t *addrs, uint32_t n,
        char ***out_names, uint32_t *out_n);
void freeNames(char **names, uint32_t n);
void maintainBccCaches(int64_t seconds);

#ifdef __cplusplus
}
#endif

#endif  // SYSDIG_CGO_FUNC_H
