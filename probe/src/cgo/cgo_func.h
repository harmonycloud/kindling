//
// Created by jundi zhou on 2022/6/1.
//

#ifndef SYSDIG_CGO_FUNC_H
#define SYSDIG_CGO_FUNC_H

#include <stdint.h>
#include "profiling_cgo.h"
#include "kindling.h"
#include "core/symbolizer/kindling_symbolizer.h"

#ifdef __cplusplus
extern "C" {
#endif
int runForGo();
int getEventsByInterval(int interval, void** kindlingEvent, void* count);
void suppressEventsCommForGo(char *comm);
void subEventForGo(char* eventName, char* category, void* params);
int startProfile();
int stopProfile();
char* startAttachAgent(int pid);
char* stopAttachAgent(int pid);
void startProfileDebug(int pid, int tid);
void stopProfileDebug();
void getCaptureStatistics(struct capture_statistics_for_go* stats);
void catchSignalUp();

#ifdef __cplusplus
}
#endif

#endif  // SYSDIG_CGO_FUNC_H
