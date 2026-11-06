#pragma once
#include <unordered_map>
#include <vector>
#include <mutex>
#include <chrono>
#include <list>
extern "C" {
#include "bcc_syms.h"
}
#define MAX_PID_CACHES_DEFAULT 2048

#define REFRESH_COOLDOWN std::chrono::seconds(60)

struct ResolveOptions {
    bool debugSyms = true;                          // bcc_symbol_option.use_debug_file
    bool formatUnresolvedAsModulePlusOffset = true;
};

class BccPerPidSymbolizer {
public:
    BccPerPidSymbolizer();
    ~BccPerPidSymbolizer();

    int resolveProcess(uint32_t pid, const uint64_t* addrs, size_t n,
                       std::vector<std::string>& out, const ResolveOptions& opt);

    int resolveKernel(const uint64_t* addrs, size_t n, std::vector<std::string>& out);

    bool refreshPid(uint32_t pid);

    void releasePid(uint32_t pid);

    void maintainCaches(std::chrono::seconds ttl);

    void set_max_pid_caches(size_t n);

private:
    void* getOrCreateProcCache(uint32_t pid, bool debugSyms);

    bool refreshIfStale(uint32_t pid);
    static std::string formatModulePlusOffset(const bcc_symbol& sym);

    struct MapRegion {
        uint64_t start = 0;
        uint64_t end = 0;
        uint64_t pgoff = 0;

        bool isExecutable = false;
    };

    void refreshMaps(uint32_t pid);

    bool fallbackByMaps(uint32_t pid, uint64_t addr, uint64_t& out);

    struct CacheEntry {
        void* cache = nullptr;
        std::chrono::steady_clock::time_point lastUsed;
        std::chrono::steady_clock::time_point lastRefresh;
        std::vector<MapRegion> maps;
        std::list<uint32_t>::iterator lruIt;
        bool inLru = false;
        CacheEntry() : cache(nullptr), lastUsed(), lastRefresh(), maps() {}
        CacheEntry(void* c, std::chrono::steady_clock::time_point t)
                : cache(c), lastUsed(t), lastRefresh(t), maps(), inLru(false) {}
    };

    void lruTouch(uint32_t pid, CacheEntry& entry);

    size_t max_pid_caches_ = MAX_PID_CACHES_DEFAULT;

    std::mutex mu_;
    std::unordered_map<uint32_t, CacheEntry> pid2cache_;
    std::list<uint32_t> lru_;
    void* kcache_ = nullptr; // kernel cache (pid = -1)
};
