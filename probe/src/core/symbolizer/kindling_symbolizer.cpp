#include "kindling_symbolizer.h"
#include <memory>
#include <vector>
#include <cstring>
#include <cstdlib>
#include <chrono>

static std::unique_ptr<BccPerPidSymbolizer> g_sym;

static size_t trim_trailing_zeros(const uint64_t* addrs, size_t n) {
    while (n > 0 && addrs[n - 1] == 0) n--;
    return n;
}

static int copy_out(const std::vector<std::string>& names, char*** out_names, uint32_t* out_n) {
    const size_t cnt = names.size();
    char** arr = cnt ? (char**)std::calloc(cnt , sizeof(char*)) : nullptr;
    if (!arr && cnt) return -1;
    for (size_t i = 0; i < cnt; i++) {
        const std::string& s = names[i];
        arr[i] = (char*)std::malloc(s.size() + 1);
        if (!arr[i]) { for (size_t j = 0; j < i; j++) std::free(arr[j]); std::free(arr); return -1; }
        std::memcpy(arr[i], s.c_str(), s.size());
        arr[i][s.size()] = '\0';
    }
    *out_names = arr;
    *out_n = (uint32_t)cnt;
    return 0;
}

int symbolizer_init() {
    if (!g_sym) g_sym.reset(new BccPerPidSymbolizer());
    return g_sym ? 0 : -1;
}

void symbolizer_shutdown() {
    g_sym.reset();
}

int symbolize_process_names(
        uint32_t pid,
        const uint64_t* addrs, size_t n,
        int debug_syms, int /*perf_map*/, int /*map_files*/,
        char*** out_names, uint32_t* out_n, ResolveOptions* opt) {

    if (!addrs || n == 0) { *out_names = nullptr; *out_n = 0; return 0; }
    if (!g_sym && symbolizer_init() != 0) return -1;

    n = trim_trailing_zeros(addrs, n);
    if (n == 0) { *out_names = nullptr; *out_n = 0; return 0; }

    std::vector<std::string> names;
    int rc = g_sym->resolveProcess(pid, addrs, n, names, *opt);
    if (rc != 0) return rc;
    return copy_out(names, out_names, out_n);
}

int symbolize_process_names_with_fallback(
        uint32_t pid,
        const uint64_t* addrs, size_t n,
        int debug_syms, int /*perf_map*/, int /*map_files*/,
        char*** out_names, uint32_t* out_n) {
    ResolveOptions opt;
    opt.debugSyms = (debug_syms != 0);
    opt.formatUnresolvedAsModulePlusOffset = true;
    return symbolize_process_names(pid, addrs, n, debug_syms, 0, 0, out_names, out_n, &opt);
}

int symbolize_kernel_names(
        const uint64_t* addrs, size_t n,
        char*** out_names, uint32_t* out_n) {

    if (!addrs || n == 0) { *out_names = nullptr; *out_n = 0; return 0; }
    if (!g_sym && symbolizer_init() != 0) return -1;

    n = trim_trailing_zeros(addrs, n);
    if (n == 0) { *out_names = nullptr; *out_n = 0; return 0; }

    std::vector<std::string> names;
    int rc = g_sym->resolveKernel(addrs, n, names);
    if (rc != 0) return rc;
    return copy_out(names, out_names, out_n);
}

void free_names(char** names, uint32_t n) {
    if (!names) return;
    for (uint32_t i = 0; i < n; i++) std::free(names[i]);
    std::free(names);
}

void maintain_caches(int64_t seconds){
    if (!g_sym) return;
    std::chrono::seconds ttl{seconds};
    g_sym->maintainCaches(ttl);
}