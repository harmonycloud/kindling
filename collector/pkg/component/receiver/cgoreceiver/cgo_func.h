//
// Created by jundi zhou on 2022/6/1.
//

#ifndef SYSDIG_CGO_FUNC_H
#define SYSDIG_CGO_FUNC_H

#include <stdint.h>

/* Must match agent-libs driver/bpf/types.h layout. */
struct bpf_profile_data {
	uint32_t pid;
	uint64_t user_stack[127];
	uint64_t kernel_stack[127];
	uint64_t count;
} __attribute__((aligned(8)));
struct sample_key {
	uint32_t pid;
	int32_t kernel_stack_id;
	int32_t user_stack_id;
	char comm[16];
} __attribute__((aligned(8)));
struct sample_key_set {
	uint32_t nr_keys;
	struct sample_key *keys;
} __attribute__((aligned(8)));

struct capture_statistics_for_go {
	int evts;
	int drops;
	int drops_buffer;
	int drops_pf;
	int drops_bug;
	int preemptions;
	int suppressed;
	int tids_suppressed;
};

struct event_params_for_subscribe {
	char *name;
	char *value;
};

struct kindling_event_t_for_go {
	uint64_t timestamp;
	char *name;
	uint32_t category;
	uint16_t paramsNumber;
	uint64_t latency;
	struct KeyValue {
		char *key;
		char *value;
		uint32_t len;
		uint32_t valueType;
	} userAttributes[16];
	struct event_context {
		struct thread_info {
			uint32_t pid;
			uint32_t tid;
			uint32_t uid;
			uint32_t gid;
			char *comm;
			char *containerId;
		} tinfo;
		struct fd_info {
			int32_t num;
			uint32_t fdType;
			char *filename;
			char *directory;
			uint32_t protocol;
			uint8_t role;
			uint32_t sip[4];
			uint32_t dip[4];
			uint32_t sport;
			uint32_t dport;
			uint64_t source;
			uint64_t destination;
		} fdInfo;
	} context;
};

#ifdef __cplusplus
extern "C" {
#endif
int runForGo();
int getEventsByInterval(int interval, void **kindlingEvent, void *count);
void suppressEventsCommForGo(char *comm);
void subEventForGo(char *eventName, char *category, void *params);
int startProfile();
int stopProfile();
char *startAttachAgent(int pid);
char *stopAttachAgent(int pid);
void startProfileDebug(int pid, int tid);
void stopProfileDebug();
void getCaptureStatistics(struct capture_statistics_for_go *stats);
void catchSignalUp();

/* CPU continuous profiling */
int getProfileData(struct sample_key key, struct bpf_profile_data *sample);
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

#endif // SYSDIG_CGO_FUNC_H
