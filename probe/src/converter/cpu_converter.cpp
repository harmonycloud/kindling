// Copyright (C) 2022-2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

//
// Created by 散养鸡 on 2022/4/29.
//

#include "cpu_converter.h"
#include <fstream>
#include <vector>
#include "iostream"

fstream debug_file;
using namespace std;

cpu_converter::cpu_converter(sinsp* inspector) : m_inspector(inspector) {}

cpu_converter::~cpu_converter() {}

int cpu_converter::convert(kindling_event_t_for_go* p_kindling_event, sinsp_evt* cpu_evt,
                           vector<QObject*> qls, bool is_profile_debug, int64_t debug_pid,
                           int64_t debug_tid) {
  // convert
  init_kindling_event(p_kindling_event, cpu_evt);
  add_threadinfo(p_kindling_event, cpu_evt);
  add_cpu_data(p_kindling_event, cpu_evt, qls, is_profile_debug, debug_pid, debug_tid);
  return 1;
}

int cpu_converter::init_kindling_event(kindling_event_t_for_go* p_kindling_event, sinsp_evt* sevt) {
  strcpy(p_kindling_event->name, "cpu_event");

  return 0;
}

int cpu_converter::add_threadinfo(kindling_event_t_for_go* p_kindling_event, sinsp_evt* evt) {
  auto threadInfo = evt->get_thread_info();
  if (!threadInfo) {
    return -1;
  }
  p_kindling_event->context.tinfo.pid = threadInfo->m_pid;
  p_kindling_event->context.tinfo.tid = threadInfo->m_tid;
  strcpy(p_kindling_event->context.tinfo.containerId, (char*)threadInfo->m_container_id.data());

  return 0;
}

int cpu_converter::add_cpu_data(kindling_event_t_for_go* p_kindling_event, sinsp_evt* sevt,
                                vector<QObject*> qls, bool is_profile_debug, int64_t debug_pid,
                                int64_t debug_tid) {
  uint64_t start_time = *reinterpret_cast<uint64_t*>(sevt->get_param_value_raw("start_ts")->m_val);
  uint64_t end_time = *reinterpret_cast<uint64_t*>(sevt->get_param_value_raw("end_ts")->m_val);
  uint32_t cnt = *reinterpret_cast<uint32_t*>(sevt->get_param_value_raw("cnt")->m_val);
  uint64_t* time_specs =
      reinterpret_cast<uint64_t*>(sevt->get_param_value_raw("time_specs")->m_val);
  uint64_t* runq_latency =
      reinterpret_cast<uint64_t*>(sevt->get_param_value_raw("runq_latency")->m_val);
  uint8_t* time_type = reinterpret_cast<uint8_t*>(sevt->get_param_value_raw("time_type")->m_val);

  vector<pair<uint64_t, uint64_t>> on_time, off_time;
  vector<uint8_t> off_type;
  uint64_t start = start_time;

  for (int i = 0; i < cnt; i++) {
    if (time_type[i] == 0) {
      on_time.emplace_back(start, start + time_specs[i]);
    } else {
      off_time.emplace_back(start, start + time_specs[i]);
      off_type.emplace_back(time_type[i]);
    }
    start = start + time_specs[i];
  }

  uint16_t userAttNumber = 0;
  // start_time
  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "start_time");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, &start_time, 8);
  p_kindling_event->userAttributes[userAttNumber].valueType = UINT64;
  p_kindling_event->userAttributes[userAttNumber].len = 8;
  userAttNumber++;

  // end_time
  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "end_time");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, &end_time, 8);
  p_kindling_event->userAttributes[userAttNumber].valueType = UINT64;
  p_kindling_event->userAttributes[userAttNumber].len = 8;
  userAttNumber++;

  // time_specs
  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "time_specs");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, time_specs, 8 * cnt);
  p_kindling_event->userAttributes[userAttNumber].valueType = BYTEBUF;
  p_kindling_event->userAttributes[userAttNumber].len = 8 * cnt;
  userAttNumber++;

  // runq_latency
  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "runq_latency");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, runq_latency, cnt / 2 * 8);
  p_kindling_event->userAttributes[userAttNumber].valueType = BYTEBUF;
  p_kindling_event->userAttributes[userAttNumber].len =  cnt / 2 * 8;
  userAttNumber++;

  // time_type
  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "time_type");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, time_type, cnt);
  p_kindling_event->userAttributes[userAttNumber].valueType = BYTEBUF;
  p_kindling_event->userAttributes[userAttNumber].len = cnt;
  userAttNumber++;

  // on_stack
  auto s_tinfo = sevt->get_thread_info();
  string data = "";
  for (auto it = qls.begin(); it != qls.end(); it++) {
    KindlingInterface* plugin = qobject_cast<KindlingInterface*>(*it);
    if (plugin) {
      data.append(plugin->getCache(s_tinfo->m_tid, on_time, off_type, false, false, true));
    }
  }
  if (data != "") {
    strcpy(p_kindling_event->userAttributes[userAttNumber].key, "stack");
    memcpy(p_kindling_event->userAttributes[userAttNumber].value, data.data(), data.length());
    p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
    p_kindling_event->userAttributes[userAttNumber].len = data.length();
    userAttNumber++;
  }

  string log_msg = "";
  for (auto it = qls.begin(); it != qls.end(); it++) {
    KindlingInterface* plugin = qobject_cast<KindlingInterface*>(*it);
    if (plugin) {
      log_msg.append(plugin->getCache(s_tinfo->m_tid, on_time, off_type, false, true, false));
    }
  }
  if (log_msg != "") {
    strcpy(p_kindling_event->userAttributes[userAttNumber].key, "log");
    memcpy(p_kindling_event->userAttributes[userAttNumber].value, log_msg.data(), log_msg.length());
    p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
    p_kindling_event->userAttributes[userAttNumber].len = log_msg.length();
    userAttNumber++;
  }

  string on_info = "";
  for (auto it = qls.begin(); it != qls.end(); it++) {
    KindlingInterface* plugin = qobject_cast<KindlingInterface*>(*it);
    if (plugin) {
      on_info.append(plugin->getCache(s_tinfo->m_tid, on_time, off_type, false, false, false));
    }
  }

  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "on_info");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, on_info.data(), on_info.length());
  p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
  p_kindling_event->userAttributes[userAttNumber].len = on_info.length();
  userAttNumber++;

  string info = "";
  for (auto it = qls.begin(); it != qls.end(); it++) {
    KindlingInterface* plugin = qobject_cast<KindlingInterface*>(*it);
    if (plugin) {
      info.append(plugin->getCache(s_tinfo->m_tid, off_time, off_type, true, false, false));
    }
  }
  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "off_info");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, info.data(), info.length());
  p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
  p_kindling_event->userAttributes[userAttNumber].len = info.length();
  userAttNumber++;
  p_kindling_event->paramsNumber = userAttNumber;

  if (is_profile_debug && s_tinfo->m_tid == debug_tid && s_tinfo->m_pid == debug_pid) {
    if (!debug_file.is_open()) {
      debug_file.open("profile_debug_cpu.log", ios::app | ios::out);
    } else {
      for (int i = 0; i < on_time.size(); i++) {
        string debug_info = to_string(on_time[i].first) + "  -  " + to_string(on_time[i].second) +
                            "on, " + " onnumber: " + to_string(i + 1) + "oninfo:" + on_info;
        debug_file << debug_info << "\n";
      }
      for (int i = 0; i < off_time.size(); i++) {
        string debug_info = to_string(off_time[i].first) + "  -  " + to_string(off_time[i].second) +
                            "off, " + " offnumber: " + to_string(i + 1) + "offinfo:" + info;
        debug_file << debug_info << "\n";
      }
    }
  }

  return 0;
}