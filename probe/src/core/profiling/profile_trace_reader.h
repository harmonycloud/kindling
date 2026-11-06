#pragma once

#include <memory>

#include "cgo/profiling_cgo.h"
#include "mini_sinsp.h"

class sinsp;

namespace kindling {
namespace profiling {

// Thin MiniSinsp reader for profile_trace_events. Copies payload into
// ProfileTraceSample; does not parse sinsp event params.
class ProfileTraceReader {
 public:
	bool init(sinsp* inspector);
	void close() { insp_.reset(); }
	// Fills up to max_n samples. Returns count written, 0 if empty/timeout,
	// negative errno-style on hard failure.
	int poll(ProfileTraceSample* out, int max_n);

 private:
	std::shared_ptr<MiniSinsp::mini_sinsp> insp_;
};

ProfileTraceReader& profile_trace_reader();

}  // namespace profiling
}  // namespace kindling
