//
// Created by zhaoxiangyu on 25-8-18.
//

#ifndef KINDLING_PROBE_KINDLING_SYMBOLIZER_H
#define KINDLING_PROBE_KINDLING_SYMBOLIZER_H

#include <string>
#include "symbolizer/bcc_per_pid_symbolizer.h"

int symbolizer_init();
void symbolizer_shutdown();

int symbolize_process_names(
        uint32_t pid,
        const uint64_t* addrs, size_t n,
        int debug_syms, int perf_map, int map_files,
        char*** out_names, uint32_t* out_n, ResolveOptions* opt);

int symbolize_process_names_with_fallback(
        uint32_t pid,
        const uint64_t *addrs, size_t n,
        int debug_syms, int perf_map, int map_files,
        char ***out_names, uint32_t *out_n);

int symbolize_kernel_names(
        const uint64_t *addrs, size_t n,
        char ***out_names, uint32_t *out_n);

void free_names(char **names, uint32_t n);

void maintain_caches(int64_t seconds);
#endif //KINDLING_PROBE_KINDLING_SYMBOLIZER_H
