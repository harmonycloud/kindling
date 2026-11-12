#ifndef KINDLING_CGO_PROFILING_H
#define KINDLING_CGO_PROFILING_H

#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* CPU profiling facade. Keep both probe and collector copies identical.
 * Map structures use the default ABI; the writer converts LPM keys to big endian.
 */

/* ---- CPU sampling ---- */
int initCPUProfiling(void);
void shutdownCPUProfiling(void);
int enableNativeProfile(uint32_t pid);
int disableNativeProfile(uint32_t pid);
void startCpuSampeling(void);
void startCpuSampelingAll(void);
void stopCpuSampeling(void);

/* ---- PID events / metrics ---- */
/* Drain profile_pid_events. Returns count, 0 if empty, <0 on error. */
int drainProfilePidEvents(uint32_t* out, int max_n);
int deleteProfileReportedPid(uint32_t pid);
/* Sum profile_metrics slots across CPUs. Returns slots written, <0 on error. */
int readProfileMetrics(uint64_t* out, int n);

/* ---- Symbolization (BCC) ---- */
int symbolizerInit(void);
void symbolizerShutdown(void);
/* Free process symbol caches unused for longer than seconds. */
void maintainBccCaches(int64_t seconds);
int symbolizeProcessNamesWithFallback(uint32_t pid, const uint64_t* addrs, uint32_t n,
				      int debug_syms, int perf_map, int map_files,
				      char*** out_names, uint32_t* out_n);
int symbolizeKernelNames(const uint64_t* addrs, uint32_t n, char*** out_names, uint32_t* out_n);
void freeNames(char** names, uint32_t n);

/* ---- Native-unwind MapWriter constants (mirror profile_types.h) ---- */
#define PROFILE_MAP_TRACE_EVENTS             "profile_trace_events"
#define PROFILE_MAP_UNWIND_INFO_ARRAY        "profile_unwind_info_array"
#define PROFILE_MAP_STACK_DELTA_PAGE_TO_INFO "profile_stack_delta_page_to_info"
#define PROFILE_MAP_PID_PAGE_TO_MAPPING      "profile_pid_page_to_mapping_info"
#define PROFILE_MAP_INTERPRETER_OFFSETS      "profile_interpreter_offsets"
#define PROFILE_MAP_PY_PROCS                 "profile_py_procs"
#define PROFILE_MAP_METRICS                  "profile_metrics"
/* snprintf(buf, ..., PROFILE_MAP_EXE_ID_TO_STACK_DELTAS_FMT, map_id) */
#define PROFILE_MAP_EXE_ID_TO_STACK_DELTAS_FMT "profile_exe_id_to_%u_stack_deltas"

#define PROFILE_NATIVE_MAX_FRAMES 128
/* Mirror profile_types.h: kernel PCs + variable-length user frames, power of two. */
#define PROFILE_FRAME_WORDS 512
#define PROFILE_COMM_LEN 16
#define PROFILE_UNWIND_INFO_MAX_ENTRIES 16384
#define PROFILE_STACK_DELTA_BUCKET_SMALLEST 8
#define PROFILE_STACK_DELTA_BUCKET_LARGEST 23

/* ---- Native Trace poll ABI (Go drain) ---- */
typedef struct ProfileTraceSample {
	uint32_t pid;
	uint32_t tid;
	uint64_t ktime;
	uint16_t cpu_id;
	uint16_t num_frames;       /* user frames only */
	uint32_t frame_data_len;   /* total u64 words in frames[] */
	uint16_t num_kernel_frames; /* leading frames[] entries: raw kernel PCs */
	uint16_t pad;
	char comm[PROFILE_COMM_LEN];
	/* kernel PCs first, then typed variable-length user frames */
	uint64_t frames[PROFILE_FRAME_WORDS];
} ProfileTraceSample;

/* Batch poll profile_trace_events. Returns samples written, 0 if empty, <0 on error. */
int pollProfileTraces(ProfileTraceSample* out, int max_n);

/* ---- Native-unwind MapWriter (batch) ---- */
typedef struct ProfileMapStackDelta {
	uint16_t addr_low;
	uint16_t unwind_info;
} ProfileMapStackDelta; /* sizeof == 4 */

typedef struct ProfileMapUnwindInfo {
	uint8_t flags;
	uint8_t base_reg;
	uint8_t aux_base_reg;
	uint8_t merge_opcode;
	int32_t param;
	int32_t aux_param;
} ProfileMapUnwindInfo; /* sizeof == 12 */

typedef struct ProfileMapStackDeltaPageKey {
	uint64_t file_id;
	uint64_t page;
} ProfileMapStackDeltaPageKey; /* sizeof == 16 */

typedef struct ProfileMapStackDeltaPageInfo {
	uint32_t first_delta;
	uint16_t num_deltas;
	uint16_t map_id;
} ProfileMapStackDeltaPageInfo; /* sizeof == 8 */

typedef struct ProfileMapPidPage {
	uint32_t prefix_len; /* 32 + page prefix bits */
	uint32_t pid;
	uint64_t page;
} ProfileMapPidPage; /* sizeof == 16 */

typedef struct ProfileMapPidPageMappingInfo {
	uint64_t file_id;
	uint64_t bias_and_unwind_program;
} ProfileMapPidPageMappingInfo; /* sizeof == 16 */

/* Python preprocessing MapWriter ABI (host endian, default alignment). */
typedef struct ProfileMapOffsetRange {
	uint64_t lower_offset1;
	uint64_t upper_offset1; /* inclusive; converted from Go's exclusive End */
	uint64_t lower_offset2;
	uint64_t upper_offset2;
	uint16_t program_index;
	uint16_t pad[3];
} ProfileMapOffsetRange; /* sizeof == 40 */

typedef struct ProfileMapTSDInfo {
	int16_t offset;
	uint8_t multiplier;
	uint8_t indirect;
	uint16_t key_limit;
	uint8_t block_entries;
	uint8_t data_offset;
} ProfileMapTSDInfo; /* sizeof == 8 */

typedef struct ProfileMapPyProcInfo {
	uint64_t auto_tls_key_addr;
	uint64_t none_struct_addr;
	uint16_t version;
	int16_t tls_offset;
	ProfileMapTSDInfo tsd_info;
	uint8_t thread_state_frame;
	uint8_t cframe_current_frame;
	uint8_t frame_back;
	uint8_t frame_code;
	uint8_t frame_lasti;
	uint8_t frame_entry_member;
	uint8_t frame_entry_val;
	uint8_t code_argcount;
	uint8_t code_kwonlyargcount;
	uint8_t code_flags;
	uint8_t code_firstlineno;
	uint8_t lasti_is_codeunit;
	uint8_t frame_is_cframe;
	uint8_t pad;
	uint16_t code_sizeof;
} ProfileMapPyProcInfo; /* sizeof == 48 */

/* Single-entry writes return 0 or -errno. Deleting a missing key succeeds. */
int profileUpdateInterpreterOffsets(uint64_t file_id, const ProfileMapOffsetRange* range);
int profileDeleteInterpreterOffsets(uint64_t file_id);
int profileUpdatePyProcInfo(uint32_t pid, const ProfileMapPyProcInfo* info);
int profileDeletePyProcInfo(uint32_t pid);

int profileMapFdByName(const char* name);

int profileUpdateUnwindInfos(const uint32_t* indices, const ProfileMapUnwindInfo* infos,
			     uint32_t n);

int profileUpdateStackDeltaPages(const ProfileMapStackDeltaPageKey* keys,
				 const ProfileMapStackDeltaPageInfo* values, uint32_t n);

int profileDeleteStackDeltaPages(const ProfileMapStackDeltaPageKey* keys, uint32_t n);

int profileUpdatePidPageMappings(const ProfileMapPidPage* keys,
				 const ProfileMapPidPageMappingInfo* values, uint32_t n);

int profileDeletePidPageMappings(const ProfileMapPidPage* keys, uint32_t n);

/* Returns map_id (8..23) or -errno. */
int profileUpdateExeIDToStackDeltas(uint64_t file_id, const ProfileMapStackDelta* deltas,
				    uint32_t n);

int profileDeleteExeIDToStackDeltas(uint64_t file_id, uint16_t map_id);

#ifdef __cplusplus
}
#endif

#endif /* KINDLING_CGO_PROFILING_H */
