//
// Created by jundi zhou on 2022/6/1.
//

#ifndef SYSDIG_CGO_FUNC_H
#define SYSDIG_CGO_FUNC_H

#include <stdint.h>

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

#ifdef __cplusplus
}
#endif

#endif // SYSDIG_CGO_FUNC_H
