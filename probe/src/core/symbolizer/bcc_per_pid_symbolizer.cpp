#include "bcc_per_pid_symbolizer.h"
#include <cstring>
#include <sstream>
#include <iomanip>
#include <fstream>
#include <algorithm>
#include <cctype>
#include <cerrno>
#include <cstdlib>
#include <cstdio>
namespace {
    // 直接从 [p, q) 解析十六进制
    static inline bool parseHexU64Span(const char* p, const char* q, uint64_t& out) {
        if (!p || !q || p >= q) return false;
        uint64_t v = 0;
        while (p < q) {
            unsigned char c = static_cast<unsigned char>(*p++);
            uint64_t d;
            if (c >= '0' && c <= '9') d = c - '0';
            else if (c >= 'a' && c <= 'f') d = 10 + (c - 'a');
            else if (c >= 'A' && c <= 'F') d = 10 + (c - 'A');
            else return false;
            v = (v << 4) | d;
        }
        out = v;
        return true;
    }

    // 以小写十六进制将 value 追加到 dst 末尾（不带 0x 前缀）
    static inline void appendHexLower(std::string& dst, uint64_t value) {
        char buf[16];
        int idx = 16;
        if (value == 0) { dst.push_back('0'); return; }
        while (value != 0 && idx > 0) {
            unsigned int digit = static_cast<unsigned int>(value & 0xF);
            value >>= 4;
            buf[--idx] = static_cast<char>(digit < 10 ? ('0' + digit) : ('a' + (digit - 10)));
        }
        dst.append(buf + idx, 16 - idx);
    }
}

BccPerPidSymbolizer::BccPerPidSymbolizer() {
    kcache_ = bcc_symcache_new(-1, nullptr);
}
BccPerPidSymbolizer::~BccPerPidSymbolizer() {
    {
        for (auto& item : pid2cache_) if (item.second.cache) bcc_free_symcache(item.second.cache, item.first);
        pid2cache_.clear();
        lru_.clear();
    }
    if (kcache_) { bcc_free_symcache(kcache_, -1); kcache_ = nullptr; }
}

void BccPerPidSymbolizer::lruTouch(uint32_t pid, CacheEntry& entry) {
    if (entry.inLru) {
        lru_.splice(lru_.end(), lru_, entry.lruIt);
        return;
    }
    lru_.push_back(pid);
    entry.lruIt = --lru_.end();
    entry.inLru = true;
}

void BccPerPidSymbolizer::set_max_pid_caches(size_t n) {
    max_pid_caches_ = n == 0 ? 1 : n;
    while (pid2cache_.size() > max_pid_caches_) {
        uint32_t evictPid = lru_.front();
        lru_.pop_front();
        auto it2 = pid2cache_.find(evictPid);
        bcc_free_symcache(it2->second.cache, (int)evictPid);
        pid2cache_.erase(it2);
    }
}

void* BccPerPidSymbolizer::getOrCreateProcCache(uint32_t pid, bool debugSyms) {
    auto now = std::chrono::steady_clock::now();
    auto it = pid2cache_.find(pid);
    if (it != pid2cache_.end() && it->second.cache) {
        it->second.lastUsed = now;
        lruTouch(pid, it->second);
        return it->second.cache;
    }
    bcc_symbol_option opt{};
    opt.use_debug_file = debugSyms ? 1 : 0;
    opt.check_debug_file_crc = 1;
    opt.lazy_symbolize = 1;
    opt.use_symbol_type = 0xFFFFFFFF;
    void* c = bcc_symcache_new((int)pid, &opt);
    if(!c){
        return nullptr;
    }

    pid2cache_[pid] = CacheEntry(c, now);
    lruTouch(pid, pid2cache_[pid]);
    refreshMaps(pid);

    if (pid2cache_.size() > max_pid_caches_) {
        uint32_t evictPid = lru_.front();
        lru_.pop_front();
        auto it2 = pid2cache_.find(evictPid);
        if (it2 != pid2cache_.end()) {
            if (it2->second.cache) bcc_free_symcache(it2->second.cache, (int)evictPid);
            pid2cache_.erase(it2);
        }
    }
    return c;
}

int BccPerPidSymbolizer::resolveProcess(uint32_t pid, const uint64_t* addrs, size_t n,
                                        std::vector<std::string>& out, const ResolveOptions& opt) {
    std::unique_lock<std::mutex> lock(mu_);
    if (!addrs || n == 0) { out.clear(); return 0; }
    void* cache = getOrCreateProcCache(pid, opt.debugSyms);
    if (!cache) return -1;

    out.resize(n);
    for (size_t i = 0; i < n; i++) {
        bcc_symbol sym{};
        int rc = bcc_symcache_resolve(cache, addrs[i], &sym);
        if (rc == 0) {
            const char* nm = sym.demangle_name ? sym.demangle_name : sym.name;
            if (nm) out[i].assign(nm); else out[i].clear();
            bcc_symbol_free_demangle_name(&sym);
        } else {
            if (opt.formatUnresolvedAsModulePlusOffset) {
                uint64_t fb;
                if (fallbackByMaps(pid, addrs[i], fb)) {
                    out[i].clear();
                    if (sym.module) {
                        const char* mod = sym.module;
                        size_t ml = std::strlen(mod);
                        out[i].reserve(ml + 3 + 16);
                        out[i].append(mod, ml);
                    } else {
                        out[i].reserve(6 + 3 + 16);
                        out[i].append("<anon>", 6);
                    }
                    out[i].append("+0x", 3);
                    appendHexLower(out[i], fb);
                } else {
                    out[i].clear();
                    out.resize(i);
                    break;
                }
            } else {
                if(sym.module){
                    out[i].assign(sym.module);
                }else{
                    out[i].clear();
                    out.resize(i);
                    break;
                }
            }
        }
    }
    return 0;
}

int BccPerPidSymbolizer::resolveKernel(const uint64_t* addrs, size_t n,
                                       std::vector<std::string>& out) {
    if (!addrs || n == 0) { out.clear(); return 0; }
    if (!kcache_) return -1;
    out.resize(n);
    for (size_t i = 0; i < n; i++) {
        bcc_symbol sym{};
        int rc = bcc_symcache_resolve(kcache_, addrs[i], &sym);
        if (rc == 0) {
            const char* nm = sym.demangle_name ? sym.demangle_name : sym.name;
            out[i] = nm ? nm : "";
            bcc_symbol_free_demangle_name(&sym);
        } else {
            out[i].clear();
            out.resize(i);
            break;
        }
    }
    return 0;
}

bool BccPerPidSymbolizer::refreshPid(uint32_t pid) {
    auto it = pid2cache_.find(pid);
    if (it != pid2cache_.end() && it->second.cache) bcc_symcache_refresh(it->second.cache);
    else return false;
    // 同步刷新 maps
    it->second.lastUsed = std::chrono::steady_clock::now();
    lruTouch(pid, it->second);
    refreshMaps(pid);
    return true;
}

void BccPerPidSymbolizer::releasePid(uint32_t pid) {
    auto it = pid2cache_.find(pid);
    if (it != pid2cache_.end()) {
        if (it->second.cache) {
            bcc_free_symcache(it->second.cache, (int)pid);
            lru_.erase(it->second.lruIt);
        }
        pid2cache_.erase(it);
    }
}

void BccPerPidSymbolizer::maintainCaches(std::chrono::seconds ttl) {
    std::unique_lock<std::mutex> lock(mu_);
    auto now = std::chrono::steady_clock::now();
    // 因为 LRU 顺序与 lastUsed 一致（每次 touch 都挪到尾），可以从头部按需清理
    auto it = lru_.begin();
    while (it != lru_.end()) {
        uint32_t pid = *it;
        auto mit = pid2cache_.find(pid);
        if (mit == pid2cache_.end()) { it = lru_.erase(it); continue; }
        if ((now - mit->second.lastUsed) > ttl) {
            if (mit->second.cache) bcc_free_symcache(mit->second.cache, (int)pid);
            it = lru_.erase(it);
            pid2cache_.erase(mit);
        } else {
            break;
        }
    }

    // 将剩余（活跃）pid 收集后再统一刷新，避免遍历中移动 LRU 迭代器, 刷新失败会删除。
    std::vector<uint32_t> toRefresh;
    toRefresh.reserve(lru_.size());
    for (auto it2 = it; it2 != lru_.end(); ++it2) toRefresh.push_back(*it2);

    for (uint32_t rpid : toRefresh) {
        if(!refreshPid(rpid))releasePid(rpid);
    }

    if (kcache_)bcc_symcache_refresh(kcache_);
    else kcache_ =  bcc_symcache_new(-1, nullptr);//todo: 报错原因未分类，如果因为权限原因 应该不再重新创建。
}

void BccPerPidSymbolizer::refreshMaps(uint32_t pid) {
    std::ostringstream oss;
    oss << "/proc/" << pid << "/maps";
    std::ifstream ifs(oss.str());
    if (!ifs) return;

    auto it = pid2cache_.find(pid);
    if (it == pid2cache_.end()) {
        fprintf(stderr, "pid = %u, bcc cache is NULL!\n", pid);
        return;
    }

    auto& cache = it->second;
    std::vector<MapRegion> regions;
    std::string line;
    while (std::getline(ifs, line)) {
        const char* s = line.c_str();
        const char* e = s + line.size();
        // start-end
        const char* dash = (const char*)::memchr(s, '-', (size_t)(e - s));
        if (!dash) continue;
        const char* p = dash + 1;
        while (p < e && *p == ' ') ++p; // should not be spaces, safeguard
        const char* sp1 = p; while (p < e && *p != ' ') ++p; const char* sp1_end = p; // end hex
        while (p < e && *p == ' ') ++p;
        const char* perms_beg = p; while (p < e && *p != ' ') ++p; const char* perms_end = p;
        while (p < e && *p == ' ') ++p;
        const char* off_beg = p; while (p < e && *p != ' ') ++p; const char* off_end = p;

        uint64_t start = 0, endv = 0, pgoff = 0;
        if (!parseHexU64Span(s, dash, start)) continue;
        if (!parseHexU64Span(sp1, sp1_end, endv)) continue;
        if (!parseHexU64Span(off_beg, off_end, pgoff)) continue;

        bool x = false;
        for (const char* t = perms_beg; t < perms_end; ++t) { if (*t == 'x') { x = true; break; } }

        MapRegion r;
        r.start = start;
        r.end = endv;
        r.pgoff = pgoff;
        r.isExecutable = x;
        regions.push_back(std::move(r));
    }
    cache.maps = std::move(regions);
}

bool BccPerPidSymbolizer::fallbackByMaps(uint32_t pid, uint64_t addr, uint64_t& off) {
    auto it = pid2cache_.find(pid);
    if (it == pid2cache_.end() || it->second.maps.empty()) {
        return false;
    }

    const auto& regions = it->second.maps;
    // 二分查找：找到最后一个 start <= addr 的区间，并检查 addr < end
    size_t lo = 0, hi = regions.size();
    while (lo < hi) {
        size_t mid = lo + (hi - lo) / 2;
        if (regions[mid].start <= addr) {
            lo = mid + 1;
        } else {
            hi = mid;
        }
    }
    if (lo == 0) return false;
    const auto& r = regions[lo - 1];
    if (!(addr < r.end)) return false;

    off = (addr - r.start) + r.pgoff;
    return true;
}
