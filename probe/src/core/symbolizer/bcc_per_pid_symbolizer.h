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

struct ResolveOptions {
    bool debugSyms = true;                          // bcc_symbol_option.use_debug_file
    bool formatUnresolvedAsModulePlusOffset = true; // 轻回退：若有 module 则拼 "module+0xoff"
};

class BccPerPidSymbolizer {
public:
    BccPerPidSymbolizer();
    ~BccPerPidSymbolizer();

    // 进程符号解析
    int resolveProcess(uint32_t pid, const uint64_t* addrs, size_t n,
                       std::vector<std::string>& out, const ResolveOptions& opt);

    // 内核符号解析
    int resolveKernel(const uint64_t* addrs, size_t n, std::vector<std::string>& out);

    // 显式刷新某 pid（dlopen/perf-map 变化时调用）
    bool refreshPid(uint32_t pid);

    // 释放某 pid 的缓存（进程退出/长时间不用）
    void releasePid(uint32_t pid);

    // 定期清理：超过 ttl 未使用的缓存回收
    void maintainCaches(std::chrono::seconds ttl);

    // 设置/获取 LRU 容量（单位：pid 数）
    void set_max_pid_caches(size_t n);

private:
    void* getOrCreateProcCache(uint32_t pid, bool debugSyms);
    static std::string formatModulePlusOffset(const bcc_symbol& sym);

    // -------------------- /proc/<pid>/maps 回退缓存 --------------------
    struct MapRegion {
        uint64_t start = 0;
        uint64_t end = 0;
        uint64_t pgoff = 0;              // 文件内偏移（十六进制列）
        //std::string path;               // 用bcc module代替, 映射文件路径（可能为空/匿名）
        bool isExecutable = false;       // perms 是否包含 'x'
    };

    // 解析并刷新指定 pid 的 maps 缓存
    void refreshMaps(uint32_t pid);

    // 使用 maps 回退：将 addr 映射为 "basename+0xoffset"，成功返回 true
    bool fallbackByMaps(uint32_t pid, uint64_t addr, uint64_t& out);

    struct CacheEntry {
        void* cache = nullptr;
        std::chrono::steady_clock::time_point lastUsed;
        std::vector<MapRegion> maps;
        std::list<uint32_t>::iterator lruIt; // 指向 lru_ 中该 pid 的位置
        bool inLru = false;
        CacheEntry() : cache(nullptr), lastUsed(), maps() {}
        CacheEntry(void* c, std::chrono::steady_clock::time_point t)
                : cache(c), lastUsed(t), maps(), inLru(false) {}
    };

    // LRU 触达：将 pid 移动到末尾，必要时插入
    void lruTouch(uint32_t pid, CacheEntry& entry);

    // 当前容量（可运行时修改）
    size_t max_pid_caches_ = MAX_PID_CACHES_DEFAULT;

    std::mutex mu_;
    std::unordered_map<uint32_t, CacheEntry> pid2cache_;
    std::list<uint32_t> lru_; // LRU 次序：front 最久未用，back 最近使用
    void* kcache_ = nullptr; // kernel cache (pid = -1)
};
