#include "profiling/profile_map_writer.h"

#include "cgo/profiling_inspector.h"
#include "sinsp.h"

#include "compat/bpf.h"

#include <cerrno>
#include <cstddef>
#include <cstdio>
#include <cstring>
#include <sys/syscall.h>
#include <unistd.h>
#include <vector>

#include "spdlog/spdlog.h"

namespace {

static_assert(sizeof(ProfileMapStackDelta) == 4, "ProfileMapStackDelta size");
static_assert(sizeof(ProfileMapUnwindInfo) == 12, "ProfileMapUnwindInfo size");
static_assert(sizeof(ProfileMapStackDeltaPageKey) == 16, "ProfileMapStackDeltaPageKey size");
static_assert(sizeof(ProfileMapStackDeltaPageInfo) == 8, "ProfileMapStackDeltaPageInfo size");
static_assert(sizeof(ProfileMapPidPage) == 16, "ProfileMapPidPage size");
static_assert(sizeof(ProfileMapPidPageMappingInfo) == 16, "ProfileMapPidPageMappingInfo size");

static_assert(sizeof(ProfileMapOffsetRange) == sizeof(ProfileOffsetRange), "Python offset range ABI");
static_assert(sizeof(ProfileMapOffsetRange) == 40, "Python offset range size");
static_assert(offsetof(ProfileMapOffsetRange, program_index) == offsetof(ProfileOffsetRange, program_index), "Python program offset");
static_assert(sizeof(ProfileMapTSDInfo) == sizeof(ProfileTSDInfo), "Python TSD ABI");
static_assert(sizeof(ProfileMapTSDInfo) == 8, "Python TSD size");
static_assert(offsetof(ProfileMapTSDInfo, key_limit) == offsetof(ProfileTSDInfo, key_limit), "Python TSD key offset");
static_assert(sizeof(ProfileMapPyProcInfo) == sizeof(ProfilePyProcInfo), "Python proc ABI");
static_assert(sizeof(ProfileMapPyProcInfo) == 48, "Python proc size");
static_assert(offsetof(ProfileMapPyProcInfo, tsd_info) == offsetof(ProfilePyProcInfo, tsd_info), "Python proc TSD offset");
static_assert(offsetof(ProfileMapPyProcInfo, thread_state_frame) == offsetof(ProfilePyProcInfo, thread_state_frame), "Python frame offset");
static_assert(offsetof(ProfileMapPyProcInfo, code_sizeof) == offsetof(ProfilePyProcInfo, code_sizeof), "Python code size offset");

int map_fd_or_err(const char* name)
{
	sinsp* inspector = profiling_inspector();
	if(inspector == nullptr)
	{
		return -ENODEV;
	}
	const int fd = inspector->get_map_fd_by_name(name);
	if(fd < 0)
	{
		return -ENOENT;
	}
	return fd;
}

sinsp* inspector_or_null()
{
	return profiling_inspector();
}

int get_map_id(uint32_t num_deltas)
{
	if(num_deltas == 0)
	{
		return -EINVAL;
	}
	const int significant = 32 - __builtin_clz(num_deltas);
	if(significant <= PROFILE_STACK_DELTA_BUCKET_SMALLEST)
	{
		return PROFILE_STACK_DELTA_BUCKET_SMALLEST;
	}
	if(significant > PROFILE_STACK_DELTA_BUCKET_LARGEST)
	{
		return -E2BIG;
	}
	return significant;
}

int outer_stack_delta_map_fd(uint16_t map_id)
{
	if(map_id < PROFILE_STACK_DELTA_BUCKET_SMALLEST ||
	   map_id > PROFILE_STACK_DELTA_BUCKET_LARGEST)
	{
		return -EINVAL;
	}
	char name[64];
	std::snprintf(name, sizeof(name), PROFILE_MAP_EXE_ID_TO_STACK_DELTAS_FMT,
		      static_cast<unsigned>(map_id));
	return map_fd_or_err(name);
}

/* LPM key: pid/page stored big-endian in the map. */
ProfileMapPidPage to_bpf_pid_page(const ProfileMapPidPage& in)
{
	ProfileMapPidPage out{};
	out.prefix_len = in.prefix_len;
#if defined(__BYTE_ORDER__) && (__BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__)
	out.pid = __builtin_bswap32(in.pid);
	out.page = __builtin_bswap64(in.page);
#else
	out.pid = in.pid;
	out.page = in.page;
#endif
	return out;
}

int update_elems(sinsp* inspector, int fd, const void* keys, uint32_t key_size,
		 const void* values, uint32_t value_size, uint32_t n, uint64_t flags)
{
	if(fd < 0)
	{
		return fd;
	}
	if(n == 0)
	{
		return 0;
	}
	if(inspector == nullptr || keys == nullptr || values == nullptr)
	{
		return -EINVAL;
	}

	int ok = 0;
	const char* key_base = static_cast<const char*>(keys);
	const char* val_base = static_cast<const char*>(values);
	for(uint32_t i = 0; i < n; ++i)
	{
		const void* key = key_base + static_cast<size_t>(i) * key_size;
		const void* value = val_base + static_cast<size_t>(i) * value_size;
		if(inspector->map_update_elem(fd, key, value, flags) == 0)
		{
			++ok;
		}
	}
	return ok;
}

int delete_elems(sinsp* inspector, int fd, const void* keys, uint32_t key_size, uint32_t n)
{
	if(fd < 0)
	{
		return fd;
	}
	if(n == 0)
	{
		return 0;
	}
	if(inspector == nullptr || keys == nullptr)
	{
		return -EINVAL;
	}

	int ok = 0;
	const char* key_base = static_cast<const char*>(keys);
	for(uint32_t i = 0; i < n; ++i)
	{
		const void* key = key_base + static_cast<size_t>(i) * key_size;
		if(inspector->map_delete_elem(fd, key) == 0)
		{
			++ok;
		}
	}
	return ok;
}

// Single-entry Python writes preserve syscall errno instead of batch counts.
int update_python_entry(const char* name, const void* key, const void* value)
{
	if(value == nullptr)
	{
		return -EINVAL;
	}
	const int fd = map_fd_or_err(name);
	if(fd < 0)
	{
		return fd;
	}
	if(inspector_or_null()->map_update_elem(fd, key, value, BPF_ANY) != 0)
	{
		return -errno;
	}
	return 0;
}

int delete_python_entry(const char* name, const void* key)
{
	const int fd = map_fd_or_err(name);
	if(fd < 0)
	{
		return fd;
	}
	if(inspector_or_null()->map_delete_elem(fd, key) != 0 && errno != ENOENT)
	{
		return -errno;
	}
	return 0;
}

} // namespace

extern "C" int profileMapFdByName(const char* name)
{
	if(name == nullptr)
	{
		return -EINVAL;
	}
	return map_fd_or_err(name);
}

extern "C" int profileUpdateUnwindInfos(const uint32_t* indices,
					const ProfileMapUnwindInfo* infos, uint32_t n)
{
	sinsp* inspector = inspector_or_null();
	const int fd = map_fd_or_err(PROFILE_MAP_UNWIND_INFO_ARRAY);
	if(fd < 0)
	{
		return fd;
	}
	if(n == 0)
	{
		return 0;
	}
	if(inspector == nullptr || indices == nullptr || infos == nullptr)
	{
		return -EINVAL;
	}

	int ok = 0;
	for(uint32_t i = 0; i < n; ++i)
	{
		if(indices[i] >= PROFILE_UNWIND_INFO_MAX_ENTRIES)
		{
			continue;
		}
		if(inspector->map_update_elem(fd, &indices[i], &infos[i], BPF_ANY) == 0)
		{
			++ok;
		}
	}
	return ok;
}

extern "C" int profileUpdateStackDeltaPages(const ProfileMapStackDeltaPageKey* keys,
					    const ProfileMapStackDeltaPageInfo* values,
					    uint32_t n)
{
	sinsp* inspector = inspector_or_null();
	const int fd = map_fd_or_err(PROFILE_MAP_STACK_DELTA_PAGE_TO_INFO);
	return update_elems(inspector, fd, keys, sizeof(*keys), values, sizeof(*values), n,
			    BPF_ANY);
}

extern "C" int profileDeleteStackDeltaPages(const ProfileMapStackDeltaPageKey* keys, uint32_t n)
{
	sinsp* inspector = inspector_or_null();
	const int fd = map_fd_or_err(PROFILE_MAP_STACK_DELTA_PAGE_TO_INFO);
	return delete_elems(inspector, fd, keys, sizeof(*keys), n);
}

extern "C" int profileUpdatePidPageMappings(const ProfileMapPidPage* keys,
					    const ProfileMapPidPageMappingInfo* values,
					    uint32_t n)
{
	sinsp* inspector = inspector_or_null();
	const int fd = map_fd_or_err(PROFILE_MAP_PID_PAGE_TO_MAPPING);
	if(fd < 0)
	{
		return fd;
	}
	if(n == 0)
	{
		return 0;
	}
	if(inspector == nullptr || keys == nullptr || values == nullptr)
	{
		return -EINVAL;
	}

	int ok = 0;
	for(uint32_t i = 0; i < n; ++i)
	{
		ProfileMapPidPage key = to_bpf_pid_page(keys[i]);
		if(inspector->map_update_elem(fd, &key, &values[i], BPF_ANY) == 0)
		{
			++ok;
		}
	}
	return ok;
}

extern "C" int profileDeletePidPageMappings(const ProfileMapPidPage* keys, uint32_t n)
{
	sinsp* inspector = inspector_or_null();
	const int fd = map_fd_or_err(PROFILE_MAP_PID_PAGE_TO_MAPPING);
	if(fd < 0)
	{
		return fd;
	}
	if(n == 0)
	{
		return 0;
	}
	if(inspector == nullptr || keys == nullptr)
	{
		return -EINVAL;
	}

	int ok = 0;
	for(uint32_t i = 0; i < n; ++i)
	{
		ProfileMapPidPage key = to_bpf_pid_page(keys[i]);
		if(inspector->map_delete_elem(fd, &key) == 0)
		{
			++ok;
		}
	}
	return ok;
}

extern "C" int profileUpdateExeIDToStackDeltas(uint64_t file_id,
					       const ProfileMapStackDelta* deltas, uint32_t n)
{
	sinsp* inspector = inspector_or_null();
	if(inspector == nullptr)
	{
		return -ENODEV;
	}
	if(n == 0 || deltas == nullptr)
	{
		return -EINVAL;
	}

	const int map_id = get_map_id(n);
	if(map_id < 0)
	{
		spdlog::error("profileUpdateExeIDToStackDeltas: {} deltas exceeds bucket {}", n,
			      PROFILE_STACK_DELTA_BUCKET_LARGEST);
		return map_id;
	}

	const int outer_fd = outer_stack_delta_map_fd(static_cast<uint16_t>(map_id));
	if(outer_fd < 0)
	{
		return outer_fd;
	}

	const int inner_fd = inspector->map_create(BPF_MAP_TYPE_ARRAY, sizeof(uint32_t),
						   sizeof(ProfileMapStackDelta), 1u << map_id, 0,
						   -1);
	if(inner_fd < 0)
	{
		const int err = errno != 0 ? -errno : -EIO;
		spdlog::error("failed to create stack delta inner map (map_id={}): {}", map_id,
			      -err);
		return err;
	}

	for(uint32_t i = 0; i < n; ++i)
	{
		if(inspector->map_update_elem(inner_fd, &i, &deltas[i], BPF_ANY) != 0)
		{
			spdlog::warn("failed to write stack delta[{}] for file_id=0x{:x}", i,
				     file_id);
		}
	}

	const uint32_t inner_as_value = static_cast<uint32_t>(inner_fd);
	const int rc = inspector->map_update_elem(outer_fd, &file_id, &inner_as_value, BPF_ANY);
	close(inner_fd);
	if(rc != 0)
	{
		spdlog::error("failed to attach inner stack deltas for file_id=0x{:x} map_id={}: {}",
			      file_id, map_id, errno);
		return -errno != 0 ? -errno : -EIO;
	}
	return map_id;
}

extern "C" int profileDeleteExeIDToStackDeltas(uint64_t file_id, uint16_t map_id)
{
	sinsp* inspector = inspector_or_null();
	const int outer_fd = outer_stack_delta_map_fd(map_id);
	if(outer_fd < 0)
	{
		return outer_fd;
	}
	if(inspector == nullptr)
	{
		return -ENODEV;
	}
	if(inspector->map_delete_elem(outer_fd, &file_id) != 0)
	{
		return -errno != 0 ? -errno : -ENOENT;
	}
	return 0;
}

static int bpf_next_key(int fd, const void* key, void* next)
{
	union bpf_attr attr;
	std::memset(&attr, 0, sizeof(attr));
	attr.map_fd = static_cast<__u32>(fd);
	attr.key = reinterpret_cast<__u64>(key);
	attr.next_key = reinterpret_cast<__u64>(next);
	if(syscall(__NR_bpf, BPF_MAP_GET_NEXT_KEY, &attr, sizeof(attr)) != 0)
	{
		return -errno;
	}
	return 0;
}

static int bpf_delete_key(int fd, const void* key)
{
	union bpf_attr attr;
	std::memset(&attr, 0, sizeof(attr));
	attr.map_fd = static_cast<__u32>(fd);
	attr.key = reinterpret_cast<__u64>(key);
	if(syscall(__NR_bpf, BPF_MAP_DELETE_ELEM, &attr, sizeof(attr)) != 0)
	{
		return -errno;
	}
	return 0;
}

extern "C" int drainProfilePidEvents(uint32_t* out, int max_n)
{
	if(out == nullptr || max_n <= 0)
	{
		return -EINVAL;
	}
	const int fd = map_fd_or_err("profile_pid_events");
	if(fd < 0)
	{
		return fd;
	}
	uint32_t cur = 0;
	uint32_t next = 0;
	int rc = bpf_next_key(fd, nullptr, &next);
	int n = 0;
	while(rc == 0 && n < max_n)
	{
		out[n++] = next;
		cur = next;
		rc = bpf_next_key(fd, &cur, &next);
	}
	for(int i = 0; i < n; ++i)
	{
		bpf_delete_key(fd, &out[i]);
	}
	if(rc != 0 && rc != -ENOENT)
	{
		return n > 0 ? n : rc;
	}
	return n;
}

extern "C" int deleteProfileReportedPid(uint32_t pid)
{
	if(pid == 0)
	{
		return 0;
	}
	const int fd = map_fd_or_err("profile_reported_pids");
	if(fd < 0)
	{
		return fd;
	}
	const int rc = bpf_delete_key(fd, &pid);
	if(rc == -ENOENT)
	{
		return 0;
	}
	return rc;
}

/* PERCPU lookups return one value per possible CPU, not per online CPU. */
static int possible_cpus()
{
	static int cached = 0;
	if(cached > 0)
	{
		return cached;
	}
	FILE* f = std::fopen("/sys/devices/system/cpu/possible", "r");
	if(f == nullptr)
	{
		return -errno;
	}
	char buf[256] = {};
	const bool ok = std::fgets(buf, sizeof(buf), f) != nullptr;
	std::fclose(f);
	if(!ok)
	{
		return -EIO;
	}
	int max_cpu = -1;
	for(char* tok = std::strtok(buf, ",\n"); tok != nullptr; tok = std::strtok(nullptr, ",\n"))
	{
		int lo = 0;
		int hi = 0;
		const int got = std::sscanf(tok, "%d-%d", &lo, &hi);
		if(got == 1)
		{
			hi = lo;
		}
		else if(got != 2)
		{
			continue;
		}
		if(hi > max_cpu)
		{
			max_cpu = hi;
		}
	}
	if(max_cpu < 0)
	{
		return -EIO;
	}
	cached = max_cpu + 1;
	return cached;
}

extern "C" int readProfileMetrics(uint64_t* out, int n)
{
	if(out == nullptr || n <= 0)
	{
		return -EINVAL;
	}
	const int fd = map_fd_or_err(PROFILE_MAP_METRICS);
	if(fd < 0)
	{
		return fd;
	}
	const int ncpu = possible_cpus();
	if(ncpu < 0)
	{
		return ncpu;
	}
	std::vector<uint64_t> per_cpu(static_cast<size_t>(ncpu));
	for(int i = 0; i < n; ++i)
	{
		uint32_t key = static_cast<uint32_t>(i);
		union bpf_attr attr;
		std::memset(&attr, 0, sizeof(attr));
		attr.map_fd = static_cast<__u32>(fd);
		attr.key = reinterpret_cast<__u64>(&key);
		attr.value = reinterpret_cast<__u64>(per_cpu.data());
		if(syscall(__NR_bpf, BPF_MAP_LOOKUP_ELEM, &attr, sizeof(attr)) != 0)
		{
			return -errno;
		}
		uint64_t sum = 0;
		for(uint64_t v : per_cpu)
		{
			sum += v;
		}
		out[i] = sum;
	}
	return n;
}

extern "C" int profileUpdateInterpreterOffsets(uint64_t file_id, const ProfileMapOffsetRange* range)
{
	return update_python_entry(PROFILE_MAP_INTERPRETER_OFFSETS, &file_id, range);
}

extern "C" int profileDeleteInterpreterOffsets(uint64_t file_id)
{
	return delete_python_entry(PROFILE_MAP_INTERPRETER_OFFSETS, &file_id);
}

extern "C" int profileUpdatePyProcInfo(uint32_t pid, const ProfileMapPyProcInfo* info)
{
	return update_python_entry(PROFILE_MAP_PY_PROCS, &pid, info);
}

extern "C" int profileDeletePyProcInfo(uint32_t pid)
{
	return delete_python_entry(PROFILE_MAP_PY_PROCS, &pid);
}
