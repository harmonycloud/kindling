#include "profiling/profile_trace_reader.h"

#include <cerrno>
#include <cstddef>
#include <cstring>

#include "sinsp.h"
#include "spdlog/spdlog.h"

// ProfileTrace layout (via scap → driver/bpf/types.h).
#include "scap.h"

namespace kindling {
namespace profiling {

namespace {

bool copy_sample(const scap_evt* pe, uint16_t cpuid, ProfileTraceSample* out)
{
	if(!pe || !out)
	{
		return false;
	}
	// Reject truncated payloads before reading metadata or frame words.
	const size_t header_size = offsetof(ProfileTrace, frames);
	if(pe->len < header_size)
	{
		return false;
	}
	ProfileTrace local = {};
	std::memcpy(&local, pe, header_size);
	const uint32_t words = local.frame_data_len;
	if(words > PROFILE_FRAME_WORDS ||
	   local.num_frames > PROFILE_NATIVE_MAX_FRAMES ||
	   local.num_kernel_frames > PROFILE_NATIVE_MAX_FRAMES ||
	   local.num_kernel_frames > words ||
	   sizeof(uint64_t) * words > pe->len - header_size)
	{
		return false;
	}
	if(words > 0)
	{
		std::memcpy(local.frames, reinterpret_cast<const char*>(pe) + header_size,
		            sizeof(uint64_t) * words);
	}

	out->pid = local.pid;
	out->tid = local.tid;
	out->ktime = local.ktime;
	out->cpu_id = cpuid;
	out->num_frames = local.num_frames;
	out->num_kernel_frames = local.num_kernel_frames;
	std::memcpy(out->comm, local.comm, PROFILE_COMM_LEN);

	out->frame_data_len = words;
	if(words > 0)
	{
		std::memcpy(out->frames, local.frames, sizeof(uint64_t) * words);
	}
	return true;
}

}  // namespace

bool ProfileTraceReader::init(sinsp* inspector)
{
	if(!inspector)
	{
		return false;
	}
	const int fd = inspector->get_map_fd_by_name(PROFILE_MAP_TRACE_EVENTS);
	if(fd < 0)
	{
		spdlog::warn("profile_trace_events map not found (fd={})", fd);
		return false;
	}
	insp_ = MiniSinsp::create_mini_sinsp_with_buf_size_bytes(fd, 64ULL * 1024ULL * 1024ULL);
	if(!insp_)
	{
		spdlog::error("failed to open mini_sinsp on profile_trace_events");
		return false;
	}
	spdlog::info("ProfileTraceReader ready on profile_trace_events");
	return true;
}

int ProfileTraceReader::poll(ProfileTraceSample* out, int max_n)
{
	if(!insp_ || !out || max_n <= 0)
	{
		return -EINVAL;
	}

	int n = 0;
	while(n < max_n)
	{
		sinsp_evt* evt = nullptr;
		const int32_t res = insp_->next(&evt);
		if(res == SCAP_TIMEOUT)
		{
			break;
		}
		if(res != SCAP_SUCCESS || evt == nullptr)
		{
			if(n > 0)
			{
				break;
			}
			return res < 0 ? static_cast<int>(res) : -EIO;
		}
		scap_evt* pe = evt->get_scap_evt();
		if(!copy_sample(pe, static_cast<uint16_t>(evt->get_cpuid()), &out[n]))
		{
			continue;
		}
		++n;
	}
	return n;
}

ProfileTraceReader& profile_trace_reader()
{
	static ProfileTraceReader reader;
	return reader;
}

}  // namespace profiling
}  // namespace kindling
