// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

#ifndef CPU_CONVERTER_H
#define CPU_CONVERTER_H
#include <map>
#include <string>
#include "cgo/kindling.h"
#include "sinsp.h"

class cpu_converter {
 public:
  cpu_converter(sinsp* inspector);
  ~cpu_converter();

  int convert(kindling_event_t_for_go* p_kindling_event, sinsp_evt* cpu_evt, vector<QObject*> qls,
              bool is_profile_debug, int64_t debug_pid, int64_t debug_tid);

  bool Cache(sinsp_evt* evt);

 private:
  int init_kindling_event(kindling_event_t_for_go* p_kindling_event, sinsp_evt* sevt);

  int add_threadinfo(kindling_event_t_for_go* p_kindling_event, sinsp_evt* sevt);

  int add_cpu_data(kindling_event_t_for_go* p_kindling_event, sinsp_evt* sevt, vector<QObject*> qls,
                   bool is_profile_debug, int64_t debug_pid, int64_t debug_tid);

  int32_t set_boot_time(uint64_t* boot_time);

  sinsp* m_inspector;

  uint64_t sample_interval;
};

#endif  // CPU_CONVERTER_H
