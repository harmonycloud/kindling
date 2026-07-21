//
// Created by jundi zhou on 2022/6/1.
//

#include "kindling.h"
#include <cstdlib>
#include <fstream>
#include <iostream>
#include <thread>
#include "converter/cpu_converter.h"
#include "scap_open_exception.h"
#include "sinsp_capture_interrupt_exception.h"
#include "utils.h"

cpu_converter* cpuConverter;
fstream debug_file_log;
map<uint64_t, char*> ptid_comm;
static sinsp* inspector = nullptr;
sinsp_evt_formatter* formatter = nullptr;
bool printEvent = false;
int cnt = 0;
int MAX_USERATTR_NUM = 8;
map<string, ppm_event_type> m_events;
map<string, Category> m_categories;
vector<QObject*> qls;

bool is_start_profile = false;
bool all_attach = true;
bool is_profile_debug = false;
int64_t debug_pid = 0;
int64_t debug_tid = 0;
char* traceId = new char[128];
char* isEnter = new char[16];
char* protocol = new char[32];
char* url = new char[256];
char* start_time_char = new char[32];
char* end_time_char = new char[32];
char* tid_char = new char[32];
char* time_char = new char[32];
char* depth_char = new char[8];
char* finish_char = new char[4];
char* kd_stack = new char[1024];
char* duration_char = new char[32];
char* span_char = new char[1024];

int16_t event_filters[1024][16];
uint64_t receiver_ts = 0;

void init_sub_label() {
  for (auto e : kindling_to_sysdig) {
    m_events[e.event_name] = e.event_type;
  }
  for (auto c : category_map) {
    m_categories[c.cateogry_name] = c.category_value;
  }
  for (int i = 0; i < 1024; i++) {
    for (int j = 0; j < 16; j++) {
      event_filters[i][j] = 0;
    }
  }
}

void sub_event(char* eventName, char* category, event_params_for_subscribe params[]) {
  auto it_type = m_events.find(eventName);
  if (it_type == m_events.end()) {
    cout << "failed to find event " << eventName << endl;
    return;
  }
  if (category == nullptr || category[0] == '\0') {
    for (int j = 0; j < 16; j++) {
      event_filters[it_type->second][j] = 1;
    }
    inspector->set_eventmask(it_type->second);
    if (PPME_IS_EXIT(it_type->second)) {
      inspector->set_eventmask(it_type->second - 1);
    }
    cout << "sub event name: " << eventName << endl;
  } else {
    auto it_category = m_categories.find(category);
    if (it_category == m_categories.end()) {
      cout << "failed to find category " << category << endl;
      return;
    }
    event_filters[it_type->second][it_category->second] = 1;
    inspector->set_eventmask(it_type->second);
    if (PPME_IS_EXIT(it_type->second)) {
      inspector->set_eventmask(it_type->second - 1);
    }
    cout << "sub event name: " << eventName << "  &&  category:" << category << endl;
  }
}

#define MAX_EVENTS 998
int get_events_by_interval(uint64_t interval, void** kindlingEvents, void* count) {
  uint64_t tmp_ts = receiver_ts;
  int i = 0;
    while (true) {
      if (getEvent(interval, &kindlingEvents[i], (int*)count) == 1){
        if(i >= MAX_EVENTS) {
          break;
        }
        i++;
      }
      if (tmp_ts != receiver_ts) {
        break;
      }
    }
  return i;  
}

void suppress_events_comm(string comm) {
  printCurrentTime();
  cout << "suppress_events for process " << comm << endl;
  inspector->suppress_events_comm(comm);
}

void set_eventmask(sinsp* inspector) {
  inspector->clear_eventmask();
  const enum ppm_event_type enables[] = {
      // add event type (no sub_event trigger) here to collect
  };
  for (auto event : enables) {
    inspector->set_eventmask(event);
  }
}

#define KINDLING_DEFAULT_SNAPLEN 1000
void set_snaplen(sinsp* inspector) {
  uint32_t snaplen = KINDLING_DEFAULT_SNAPLEN;

  char* env_snaplen = getenv("SNAPLEN");
  if (env_snaplen != nullptr) {
    snaplen = atol(env_snaplen);
    if (snaplen == 0 || snaplen > RW_MAX_SNAPLEN) {
      snaplen = KINDLING_DEFAULT_SNAPLEN;
      cout << "Invalid snaplen value, reset to default " << KINDLING_DEFAULT_SNAPLEN << endl;
    }
  }

  cout << "Set snaplen to value: " << snaplen << endl;
  inspector->set_snaplen(snaplen);
}

int init_probe() {
  int argc = 1;
  QCoreApplication app(argc, 0);

  // w.show();
  QObject* object;
  app.addLibraryPath(QString("../KindlingPlugin"));  // 加入库路径
  // 载入插件，取得实例
  QPluginLoader l(QString("kindling-plugin"));
  object = l.instance();
  if (object != NULL) {
    qls.push_back(object);
  }

  char* isPrintEvent = getenv("IS_PRINT_EVENT");
  if (isPrintEvent != nullptr && strncmp("true", isPrintEvent, sizeof(isPrintEvent)) == 0) {
    printEvent = true;
  }
  init_sub_label();
  string output_format =
      "*%evt.num %evt.outputtime %evt.cpu %container.name (%container.id) %proc.name "
      "(%thread.tid:%thread.vtid) %evt.dir %evt.type %evt.info";

  try {
    inspector = new sinsp();
    formatter = new sinsp_evt_formatter(inspector, output_format);
    inspector->set_hostname_and_port_resolution_mode(false);
    set_snaplen(inspector);
    inspector->open("");
    set_eventmask(inspector);

    cpuConverter = new cpu_converter(inspector);
  } catch (const exception& e) {
    fprintf(stderr, "kindling probe init err: %s", e.what());
    return 1;
  }
  return 0;
}

int getEvent(uint64_t interval, void** pp_kindling_event, int* event_count) {
  int32_t res;
  sinsp_evt* ev;
  res = inspector->next(&ev);

  ppm_event_category category;
  int result = is_normal_event(res, ev, &category);
  if (result == -1) {
    return -1;
  }
  auto threadInfo = ev->get_thread_info();
  if (is_start_profile &&
      (ev->get_type() == PPME_SYSCALL_EXECVE_8_X || ev->get_type() == PPME_SYSCALL_EXECVE_13_X ||
       ev->get_type() == PPME_SYSCALL_EXECVE_15_X || ev->get_type() == PPME_SYSCALL_EXECVE_16_X ||
       ev->get_type() == PPME_SYSCALL_EXECVE_17_X || ev->get_type() == PPME_SYSCALL_EXECVE_18_X ||
       ev->get_type() == PPME_SYSCALL_EXECVE_19_X || ev->get_type() == PPME_SYSCALL_CLONE_11_X ||
       ev->get_type() == PPME_SYSCALL_CLONE_16_X || ev->get_type() == PPME_SYSCALL_CLONE_17_X ||
       ev->get_type() == PPME_SYSCALL_CLONE_20_X || ev->get_type() == PPME_SYSCALL_FORK_X ||
       ev->get_type() == PPME_SYSCALL_FORK_17_X || ev->get_type() == PPME_SYSCALL_FORK_20_X ||
       ev->get_type() == PPME_SYSCALL_VFORK_X || ev->get_type() == PPME_SYSCALL_VFORK_17_X ||
       ev->get_type() == PPME_SYSCALL_VFORK_20_X) &&
      threadInfo->is_main_thread()) {
    if (strstr(threadInfo->m_comm.c_str(), "java") != NULL) {
      string pid_str = std::to_string(threadInfo->m_pid);
      char* temp_char = (char*)pid_str.data();
      thread attach(attach_pid, temp_char, true, true, false, false);
      attach.join();
    }
  }
  uint16_t kindling_category = get_kindling_category(ev);
  uint16_t ev_type = ev->get_type();
  int evtcnt = *event_count;

  print_event(ev);
  if (ev_type != PPME_CPU_ANALYSIS_E && is_profile_debug && threadInfo->m_tid == debug_tid &&
      threadInfo->m_pid == debug_pid) {
    print_profile_debug_info(ev);
  }
  kindling_event_t_for_go* p_kindling_event;
  init_kindling_event(p_kindling_event, pp_kindling_event);

  sinsp_fdinfo_t* fdInfo = ev->get_fd_info();
  p_kindling_event = (kindling_event_t_for_go*)*pp_kindling_event;
  uint16_t userAttNumber = 0;
  uint16_t source = get_kindling_source(ev->get_type());
  if (is_start_profile) {
    for (auto it = qls.begin(); it != qls.end(); it++) {
      KindlingInterface* plugin = qobject_cast<KindlingInterface*>(*it);
      if (plugin) {
        plugin->addCache(ev, inspector);
      }
    }
  }

  if (is_start_profile && ev->get_type() == PPME_SYSCALL_WRITE_X && fdInfo != nullptr &&
      fdInfo->is_file()) {
    auto data_param = ev->get_param_value_raw("data");
    if (data_param != nullptr) {
      char* data_val = data_param->m_val;
      if (data_param->m_len > 6 && memcmp(data_val, "kd-jf@", 6) == 0) {
        parse_jf(data_val, *data_param, p_kindling_event, threadInfo, userAttNumber);
        return 1;
      }
      if (data_param->m_len > 8 && memcmp(data_val, "kd-txid@", 8) == 0) {
        parse_xtid(ev, data_val, *data_param, p_kindling_event, threadInfo, userAttNumber);
        return 1;
      }
      if (data_param->m_len > 8 && memcmp(data_val, "kd-span@", 8) == 0) {
        parse_span(ev, data_val, *data_param, p_kindling_event, threadInfo, userAttNumber);
        return 1;
      }
      if (data_param->m_len > 6 && memcmp(data_val, "kd-tm@", 6) == 0) {
        parse_tm(data_val, *data_param, threadInfo);
        return -1;
      }
    }
  }

  if (is_start_profile && ev_type == PPME_CPU_ANALYSIS_E) {
    char* tmp_comm;

    map<uint64_t, char*>::iterator key =
        ptid_comm.find(threadInfo->m_pid << 32 | (threadInfo->m_tid & 0xFFFFFFFF));
    if (key != ptid_comm.end()) {
      tmp_comm = key->second;
    } else {
      tmp_comm = (char*)threadInfo->m_comm.data();
    }

    strcpy(p_kindling_event->context.tinfo.comm, tmp_comm);
    return cpuConverter->convert(p_kindling_event, ev, qls, is_profile_debug, debug_pid, debug_tid);
  }

  if (event_filters[ev_type][kindling_category] == 0) {
    return -1;
  }

  if (source == SYSCALL_EXIT) {
    p_kindling_event->latency = threadInfo->m_latency;
  }
  p_kindling_event->timestamp = ev->get_ts();
  p_kindling_event->category = kindling_category;
  p_kindling_event->context.tinfo.pid = threadInfo->m_pid;
  p_kindling_event->context.tinfo.tid = threadInfo->m_tid;
  p_kindling_event->context.tinfo.uid = threadInfo->m_uid;
  p_kindling_event->context.tinfo.gid = threadInfo->m_gid;
  p_kindling_event->context.fdInfo.num = ev->get_fd_num();
  if (nullptr != fdInfo) {
    p_kindling_event->context.fdInfo.fdType = fdInfo->m_type;

    switch (fdInfo->m_type) {
      case SCAP_FD_FILE:
      case SCAP_FD_FILE_V2: {
        string name = fdInfo->m_name;
        size_t pos = name.rfind('/');
        if (pos != string::npos) {
          if (pos < name.size() - 1) {
            string fileName = name.substr(pos + 1, string::npos);
            memcpy(p_kindling_event->context.fdInfo.filename, fileName.data(), fileName.length());
            if (pos != 0) {
              name.resize(pos);

              strcpy(p_kindling_event->context.fdInfo.directory, (char*)name.data());
            } else {
              strcpy(p_kindling_event->context.fdInfo.directory, "/");
            }
          }
        }
        break;
      }
      case SCAP_FD_IPV4_SOCK:
        p_kindling_event->context.fdInfo.protocol = get_protocol(fdInfo->get_l4proto());
        p_kindling_event->context.fdInfo.role = fdInfo->is_role_server();
        p_kindling_event->context.fdInfo.sip[0] = fdInfo->m_sockinfo.m_ipv4info.m_fields.m_sip;
        p_kindling_event->context.fdInfo.dip[0] = fdInfo->m_sockinfo.m_ipv4info.m_fields.m_dip;
        p_kindling_event->context.fdInfo.sport = fdInfo->m_sockinfo.m_ipv4info.m_fields.m_sport;
        p_kindling_event->context.fdInfo.dport = fdInfo->m_sockinfo.m_ipv4info.m_fields.m_dport;
        break;
      case SCAP_FD_IPV4_SERVSOCK:
        p_kindling_event->context.fdInfo.protocol = get_protocol(fdInfo->get_l4proto());
        p_kindling_event->context.fdInfo.role = fdInfo->is_role_server();
        p_kindling_event->context.fdInfo.dip[0] = fdInfo->m_sockinfo.m_ipv4serverinfo.m_ip;
        p_kindling_event->context.fdInfo.dport = fdInfo->m_sockinfo.m_ipv4serverinfo.m_port;
        break;
      case SCAP_FD_IPV6_SOCK:
        p_kindling_event->context.fdInfo.protocol = get_protocol(fdInfo->get_l4proto());
        p_kindling_event->context.fdInfo.role = fdInfo->is_role_server();
        memcpy(p_kindling_event->context.fdInfo.sip,
               fdInfo->m_sockinfo.m_ipv6info.m_fields.m_sip.m_b,
               sizeof(fdInfo->m_sockinfo.m_ipv6info.m_fields.m_sip.m_b));
        memcpy(p_kindling_event->context.fdInfo.dip,
               fdInfo->m_sockinfo.m_ipv6info.m_fields.m_dip.m_b,
               sizeof(fdInfo->m_sockinfo.m_ipv6info.m_fields.m_dip.m_b));
        p_kindling_event->context.fdInfo.sport = fdInfo->m_sockinfo.m_ipv6info.m_fields.m_sport;
        p_kindling_event->context.fdInfo.dport = fdInfo->m_sockinfo.m_ipv6info.m_fields.m_dport;
        break;
      case SCAP_FD_IPV6_SERVSOCK:
        p_kindling_event->context.fdInfo.protocol = get_protocol(fdInfo->get_l4proto());
        p_kindling_event->context.fdInfo.role = fdInfo->is_role_server();
        memcpy(p_kindling_event->context.fdInfo.dip, fdInfo->m_sockinfo.m_ipv6serverinfo.m_ip.m_b,
               sizeof(fdInfo->m_sockinfo.m_ipv6serverinfo.m_ip.m_b));
        p_kindling_event->context.fdInfo.dport = fdInfo->m_sockinfo.m_ipv6serverinfo.m_port;
        break;
      case SCAP_FD_UNIX_SOCK:
        p_kindling_event->context.fdInfo.source = fdInfo->m_sockinfo.m_unixinfo.m_fields.m_source;
        p_kindling_event->context.fdInfo.destination =
            fdInfo->m_sockinfo.m_unixinfo.m_fields.m_dest;
        break;
      default:
        break;
    }
  }

  switch (ev->get_type()) {
    case PPME_TCP_RCV_ESTABLISHED_E:
    case PPME_TCP_CLOSE_E: {
      auto pTuple = ev->get_param_value_raw("tuple");
      userAttNumber = setTuple(p_kindling_event, pTuple, userAttNumber);

      auto pRtt = ev->get_param_value_raw("srtt");
      if (pRtt != NULL) {
        strcpy(p_kindling_event->userAttributes[userAttNumber].key, "rtt");
        memcpy(p_kindling_event->userAttributes[userAttNumber].value, pRtt->m_val, pRtt->m_len);
        p_kindling_event->userAttributes[userAttNumber].valueType = UINT32;
        p_kindling_event->userAttributes[userAttNumber].len = pRtt->m_len;
        userAttNumber++;
      }
      break;
    }
    case PPME_TCP_CONNECT_X: {
      auto pTuple = ev->get_param_value_raw("tuple");
      userAttNumber = setTuple(p_kindling_event, pTuple, userAttNumber);
      auto pRetVal = ev->get_param_value_raw("retval");
      if (pRetVal != NULL) {
        strcpy(p_kindling_event->userAttributes[userAttNumber].key, "retval");
        memcpy(p_kindling_event->userAttributes[userAttNumber].value, pRetVal->m_val,
               pRetVal->m_len);
        p_kindling_event->userAttributes[userAttNumber].valueType = UINT64;
        p_kindling_event->userAttributes[userAttNumber].len = pRetVal->m_len;
        userAttNumber++;
      }
      break;
    }
    case PPME_TCP_DROP_E:
    case PPME_TCP_SET_STATE_E: {
      auto pTuple = ev->get_param_value_raw("tuple");
      userAttNumber = setTuple(p_kindling_event, pTuple, userAttNumber);
      auto old_state = ev->get_param_value_raw("old_state");
      if (old_state != NULL) {
        strcpy(p_kindling_event->userAttributes[userAttNumber].key, "old_state");
        memcpy(p_kindling_event->userAttributes[userAttNumber].value, old_state->m_val,
               old_state->m_len);
        p_kindling_event->userAttributes[userAttNumber].len = old_state->m_len;
        p_kindling_event->userAttributes[userAttNumber].valueType = INT32;
        userAttNumber++;
      }
      auto new_state = ev->get_param_value_raw("new_state");
      if (new_state != NULL) {
        strcpy(p_kindling_event->userAttributes[userAttNumber].key, "new_state");
        memcpy(p_kindling_event->userAttributes[userAttNumber].value, new_state->m_val,
               new_state->m_len);
        p_kindling_event->userAttributes[userAttNumber].valueType = INT32;
        p_kindling_event->userAttributes[userAttNumber].len = new_state->m_len;
        userAttNumber++;
      }
      break;
    }
    case PPME_TCP_SEND_RESET_E:
    case PPME_TCP_RECEIVE_RESET_E: {
      auto pTuple = ev->get_param_value_raw("tuple");
      userAttNumber = setTuple(p_kindling_event, pTuple, userAttNumber);
      break;
    }
    case PPME_TCP_RETRANCESMIT_SKB_E: {
      auto pTuple = ev->get_param_value_raw("tuple");
      userAttNumber = setTuple(p_kindling_event, pTuple, userAttNumber);

      auto segs = ev->get_param_value_raw("segs");
      if (segs != NULL) {
        strcpy(p_kindling_event->userAttributes[userAttNumber].key, "segs");
        memcpy(p_kindling_event->userAttributes[userAttNumber].value, segs->m_val, segs->m_len);
        p_kindling_event->userAttributes[userAttNumber].len = segs->m_len;
        p_kindling_event->userAttributes[userAttNumber].valueType = INT32;
        userAttNumber++;
      }
      break;
    }
    default: {
      uint16_t paramsNumber = ev->get_num_params();
      // Since current data structure specifies the maximum count of `user_attributes`
      if ((paramsNumber + userAttNumber) > MAX_USERATTR_NUM) {
        paramsNumber = MAX_USERATTR_NUM - userAttNumber;
      }
      // TODO Add another branch to verify the number of userAttNumber is less than MAX_USERATTR_NUM
      // after the program becomes more complexd
      for (auto i = 0; i < paramsNumber; i++) {
        strcpy(p_kindling_event->userAttributes[userAttNumber].key, (char*)ev->get_param_name(i));
        memcpy(p_kindling_event->userAttributes[userAttNumber].value, ev->get_param(i)->m_val,
               ev->get_param(i)->m_len);
        p_kindling_event->userAttributes[userAttNumber].len = ev->get_param(i)->m_len;
        p_kindling_event->userAttributes[userAttNumber].valueType =
            get_type(ev->get_param_info(i)->type);
        userAttNumber++;
      }
    }
  }
  p_kindling_event->paramsNumber = userAttNumber;
  strcpy(p_kindling_event->name, (char*)ev->get_name());
  char* tmp_comm;
  map<uint64_t, char*>::iterator key =
      ptid_comm.find(threadInfo->m_pid << 32 | (threadInfo->m_tid & 0xFFFFFFFF));
  if (key != ptid_comm.end()) {
    tmp_comm = key->second;
  } else {
    tmp_comm = (char*)threadInfo->m_comm.data();
  }
  strcpy(p_kindling_event->context.tinfo.comm, tmp_comm);
  strcpy(p_kindling_event->context.tinfo.containerId, (char*)threadInfo->m_container_id.data());
  evtcnt++;
  *event_count = evtcnt;
  if (ev->get_ts() - receiver_ts >= interval || *event_count > 998) {
    receiver_ts = ev->get_ts();
  }
  return 1;
}

void parse_jf(char* data_val, sinsp_evt_param data_param, kindling_event_t_for_go* p_kindling_event,
              sinsp_threadinfo* threadInfo, uint16_t& userAttNumber) {
  int val_offset = 0;
  int tmp_offset = 0;
  for (int i = 6; i < data_param.m_len; i++) {
    if (data_val[i] == '!') {
      if (val_offset == 0) {
        start_time_char[tmp_offset] = '\0';
      } else if (val_offset == 1) {
        end_time_char[tmp_offset] = '\0';
      } else if (val_offset == 2) {
        tid_char[tmp_offset] = '\0';
        break;
      }
      tmp_offset = 0;
      val_offset++;
      continue;
    }
    if (val_offset == 0) {
      start_time_char[tmp_offset] = data_val[i];
    } else if (val_offset == 1) {
      end_time_char[tmp_offset] = data_val[i];
    } else if (val_offset == 2) {
      tid_char[tmp_offset] = data_val[i];
    }
    tmp_offset++;
  }
  p_kindling_event->timestamp = atol(start_time_char);
  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "end_time");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value,
         to_string(atol(end_time_char)).data(), 19);
  p_kindling_event->userAttributes[userAttNumber].valueType = UINT64;
  p_kindling_event->userAttributes[userAttNumber].len = 19;
  userAttNumber++;
  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "data");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, data_val, data_param.m_len);
  p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
  p_kindling_event->userAttributes[userAttNumber].len = data_param.m_len;
  userAttNumber++;
  strcpy(p_kindling_event->name, "java_futex_info");
  p_kindling_event->context.tinfo.tid = threadInfo->m_tid;
  map<uint64_t, char*>::iterator key =
      ptid_comm.find(threadInfo->m_pid << 32 | (threadInfo->m_tid & 0xFFFFFFFF));
  if (key != ptid_comm.end()) {
    strcpy(p_kindling_event->context.tinfo.comm, key->second);
  } else {
    strcpy(p_kindling_event->context.tinfo.comm, (char*)threadInfo->m_comm.data());
  }
  p_kindling_event->context.tinfo.pid = threadInfo->m_pid;
  p_kindling_event->paramsNumber = userAttNumber;
}

void parse_xtid(sinsp_evt* s_evt, char* data_val, sinsp_evt_param data_param,
                kindling_event_t_for_go* p_kindling_event, sinsp_threadinfo* threadInfo,
                uint16_t& userAttNumber) {
  int val_offset = 0;
  int tmp_offset = 0;
  int traceId_offset = 0;
  int protocol_offset = 0;
  int url_offset = 0;
  for (int i = 8; i < data_param.m_len; i++) {
    if (data_val[i] == '!') {
      if (val_offset == 0) {
        traceId[tmp_offset] = '\0';
        traceId_offset = tmp_offset;
      } else if (val_offset == 1) {
        isEnter[tmp_offset] = '\0';
      } else if (val_offset == 2) {
        protocol[tmp_offset] = '\0';
        protocol_offset = tmp_offset;
      } else if (val_offset == 3) {
        url[tmp_offset] = '\0';
        url_offset = tmp_offset;
        break;
      }
      tmp_offset = 0;
      val_offset++;
      continue;
    }
    if (val_offset == 0) {
      traceId[tmp_offset] = data_val[i];
    } else if (val_offset == 1) {
      isEnter[tmp_offset] = data_val[i];
    } else if (val_offset == 2) {
      protocol[tmp_offset] = data_val[i];
    } else if (val_offset == 3) {
      url[tmp_offset] = data_val[i];
    }

    tmp_offset++;
  }
  p_kindling_event->timestamp = s_evt->get_ts();
  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "trace_id");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, traceId, traceId_offset);
  p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
  p_kindling_event->userAttributes[userAttNumber].len = traceId_offset;
  userAttNumber++;
  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "is_enter");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, isEnter, 1);
  p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
  p_kindling_event->userAttributes[userAttNumber].len = 1;
  userAttNumber++;
  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "protocol");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, protocol, protocol_offset);
  p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
  p_kindling_event->userAttributes[userAttNumber].len = protocol_offset;
  userAttNumber++;
  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "url");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, url, url_offset);
  p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
  p_kindling_event->userAttributes[userAttNumber].len = url_offset;
  userAttNumber++;
  strcpy(p_kindling_event->name, "apm_trace_id_event");
  p_kindling_event->context.tinfo.tid = threadInfo->m_tid;
  map<uint64_t, char*>::iterator key =
      ptid_comm.find(threadInfo->m_pid << 32 | (threadInfo->m_tid & 0xFFFFFFFF));
  if (key != ptid_comm.end()) {
    strcpy(p_kindling_event->context.tinfo.comm, key->second);
  } else {
    strcpy(p_kindling_event->context.tinfo.comm, (char*)threadInfo->m_comm.data());
  }
  p_kindling_event->context.tinfo.pid = threadInfo->m_pid;
  strcpy(p_kindling_event->context.tinfo.containerId, (char*)threadInfo->m_container_id.data());
  p_kindling_event->paramsNumber = userAttNumber;
}

void parse_span(sinsp_evt* s_evt, char* data_val, sinsp_evt_param data_param,
                kindling_event_t_for_go* p_kindling_event, sinsp_threadinfo* threadInfo,
                uint16_t& userAttNumber) {
  if (data_param.m_len < 10) {
    return;
  }
  int val_offset = 0;
  int tmp_offset = 0;
  int span_offset = 0;
  int traceId_offset = 0;

  bool version_new = false;
  int fromIndex = 8;
  if (data_val[8] == '1' && data_val[9] == '!') {
    version_new = true;
    fromIndex = 10;
  }
  for (int i = fromIndex; i < data_param.m_len; i++) {
    if (data_val[i] == '!') {
      if (val_offset == 0) {
        duration_char[tmp_offset] = '\0';
      } else if (val_offset == 1) {
        span_char[tmp_offset] = '\0';
        span_offset = tmp_offset;
      } else if (val_offset == 2) {
        traceId[tmp_offset] = '\0';
        traceId_offset = tmp_offset;
        break;
      }
      tmp_offset = 0;
      val_offset++;
      continue;
    }
    if (val_offset == 0) {
      duration_char[tmp_offset] = data_val[i];
    } else if (val_offset == 1) {
      span_char[tmp_offset] = data_val[i];
    } else if (val_offset == 2) {
      traceId[tmp_offset] = data_val[i];
    }
    tmp_offset++;
  }
  if (version_new) {
    // StartTime
    p_kindling_event->timestamp = atol(duration_char);
  } else {
    // EndTime - Duration
    p_kindling_event->timestamp = s_evt->get_ts() - atol(duration_char);
  }
  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "end_time");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, to_string(s_evt->get_ts()).data(),
         19);
  p_kindling_event->userAttributes[userAttNumber].valueType = UINT64;
  p_kindling_event->userAttributes[userAttNumber].len = 19;
  userAttNumber++;

  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "trace_id");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, traceId, traceId_offset);
  p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
  p_kindling_event->userAttributes[userAttNumber].len = traceId_offset;
  userAttNumber++;

  strcpy(p_kindling_event->userAttributes[userAttNumber].key, "span");
  memcpy(p_kindling_event->userAttributes[userAttNumber].value, span_char, span_offset);
  p_kindling_event->userAttributes[userAttNumber].valueType = CHARBUF;
  p_kindling_event->userAttributes[userAttNumber].len = span_offset;
  userAttNumber++;

  strcpy(p_kindling_event->name, "apm_span_event");
  p_kindling_event->context.tinfo.tid = threadInfo->m_tid;
  map<uint64_t, char*>::iterator key =
      ptid_comm.find(threadInfo->m_pid << 32 | (threadInfo->m_tid & 0xFFFFFFFF));
  if (key != ptid_comm.end()) {
    strcpy(p_kindling_event->context.tinfo.comm, key->second);
  }
  p_kindling_event->context.tinfo.pid = threadInfo->m_pid;
  p_kindling_event->paramsNumber = userAttNumber;
}

void parse_tm(char* data_val, sinsp_evt_param data_param, sinsp_threadinfo* threadInfo) {
  char* comm_char = new char[256];
  int val_offset = 0;
  int tmp_offset = 0;
  for (int i = 6; i < data_param.m_len; i++) {
    if (data_val[i] == '!') {
      if (val_offset == 0) {
        tid_char[tmp_offset] = '\0';
      } else if (val_offset == 1) {
        comm_char[tmp_offset] = '\0';
        break;
      }
      tmp_offset = 0;
      val_offset++;
      continue;
    }
    if (val_offset == 0) {
      tid_char[tmp_offset] = data_val[i];
    } else if (val_offset == 1) {
      comm_char[tmp_offset] = data_val[i];
    }
    tmp_offset++;
  }
  uint64_t v_tid = inspector->get_pid_vtid_info(threadInfo->m_pid, atol(tid_char));
  if (v_tid == 0) {
    if (ptid_comm[threadInfo->m_pid << 32 | (atol(tid_char) & 0xFFFFFFFF)] != nullptr &&
        memcmp(ptid_comm[threadInfo->m_pid << 32 | (atol(tid_char) & 0xFFFFFFFF)], comm_char,
               strlen(comm_char)) == 0) {
      delete[] comm_char;
    } else {
      ptid_comm[threadInfo->m_pid << 32 | (atol(tid_char) & 0xFFFFFFFF)] = comm_char;
    }
  } else {
    if (ptid_comm[threadInfo->m_pid << 32 | (v_tid & 0xFFFFFFFF)] != nullptr &&
        memcmp(ptid_comm[threadInfo->m_pid << 32 | (v_tid & 0xFFFFFFFF)], comm_char,
               strlen(comm_char)) == 0) {
      delete[] comm_char;
    } else {
      ptid_comm[threadInfo->m_pid << 32 | (v_tid & 0xFFFFFFFF)] = comm_char;
    }
  }
}

void init_kindling_event(kindling_event_t_for_go* p_kindling_event, void** pp_kindling_event) {
  if (nullptr == *pp_kindling_event) {
    *pp_kindling_event = (kindling_event_t_for_go*)malloc(sizeof(kindling_event_t_for_go));
    p_kindling_event = (kindling_event_t_for_go*)*pp_kindling_event;

    p_kindling_event->name = (char*)malloc(sizeof(char) * 1024);
    p_kindling_event->context.tinfo.comm = (char*)malloc(sizeof(char) * 256);
    p_kindling_event->context.tinfo.containerId = (char*)malloc(sizeof(char) * 256);
    p_kindling_event->context.fdInfo.filename = (char*)malloc(sizeof(char) * 1024);
    p_kindling_event->context.fdInfo.directory = (char*)malloc(sizeof(char) * 1024);

    for (int i = 0; i < 16; i++) {
      p_kindling_event->userAttributes[i].key = (char*)malloc(sizeof(char) * 128);
      p_kindling_event->userAttributes[i].value = (char*)malloc(sizeof(char) * EVENT_DATA_SIZE);
    }
  }
}

void print_event(sinsp_evt* s_evt) {
  if (printEvent) {
    string line;
    if (formatter->tostring(s_evt, &line)) {
      cout << line << endl;
    }
  }
}

int is_normal_event(int res, sinsp_evt* s_evt, ppm_event_category* category) {
  if (res == SCAP_TIMEOUT) {
    return -1;
  } else if (res != SCAP_SUCCESS) {
    return -1;
  }
  *category = s_evt->get_category();
  if (!inspector->is_debug_enabled() && *category & EC_INTERNAL) {
    return -1;
  }
  if (*category & EC_IO_BASE) {
    auto pres = s_evt->get_param_value_raw("res");
    if (pres && *(int64_t*)pres->m_val <= 0) {
      return -1;
    }
  }
  return 0;
}

int setTuple(kindling_event_t_for_go* p_kindling_event, const sinsp_evt_param* pTuple,
             int userAttNumber) {
  if (NULL != pTuple) {
    auto tuple = pTuple->m_val;
    if (tuple[0] == PPM_AF_INET) {
      if (pTuple->m_len == 1 + 4 + 2 + 4 + 2) {
        strcpy(p_kindling_event->userAttributes[userAttNumber].key, "sip");
        memcpy(p_kindling_event->userAttributes[userAttNumber].value, tuple + 1, 4);
        p_kindling_event->userAttributes[userAttNumber].valueType = UINT32;
        p_kindling_event->userAttributes[userAttNumber].len = 4;
        userAttNumber++;

        strcpy(p_kindling_event->userAttributes[userAttNumber].key, "sport");
        memcpy(p_kindling_event->userAttributes[userAttNumber].value, tuple + 5, 2);
        p_kindling_event->userAttributes[userAttNumber].valueType = UINT16;
        p_kindling_event->userAttributes[userAttNumber].len = 2;
        userAttNumber++;

        strcpy(p_kindling_event->userAttributes[userAttNumber].key, "dip");
        memcpy(p_kindling_event->userAttributes[userAttNumber].value, tuple + 7, 4);
        p_kindling_event->userAttributes[userAttNumber].valueType = UINT32;
        p_kindling_event->userAttributes[userAttNumber].len = 4;
        userAttNumber++;

        strcpy(p_kindling_event->userAttributes[userAttNumber].key, "dport");
        memcpy(p_kindling_event->userAttributes[userAttNumber].value, tuple + 11, 2);
        p_kindling_event->userAttributes[userAttNumber].valueType = UINT16;
        p_kindling_event->userAttributes[userAttNumber].len = 2;
        userAttNumber++;
      }
    }
  }
  return userAttNumber;
}

uint16_t get_protocol(scap_l4_proto proto) {
  switch (proto) {
    case SCAP_L4_TCP:
      return TCP;
    case SCAP_L4_UDP:
      return UDP;
    case SCAP_L4_ICMP:
      return ICMP;
    case SCAP_L4_RAW:
      return RAW;
    default:
      return UNKNOWN;
  }
}

uint16_t get_type(ppm_param_type type) {
  switch (type) {
    case PT_INT8:
      return INT8;
    case PT_INT16:
      return INT16;
    case PT_INT32:
      return INT32;
    case PT_INT64:
    case PT_FD:
    case PT_PID:
    case PT_ERRNO:
      return INT64;
    case PT_FLAGS8:
    case PT_UINT8:
    case PT_SIGTYPE:
      return UINT8;
    case PT_FLAGS16:
    case PT_UINT16:
    case PT_SYSCALLID:
      return UINT16;
    case PT_UINT32:
    case PT_FLAGS32:
    case PT_MODE:
    case PT_UID:
    case PT_GID:
    case PT_BOOL:
    case PT_SIGSET:
      return UINT32;
    case PT_UINT64:
    case PT_RELTIME:
    case PT_ABSTIME:
      return UINT64;
    case PT_CHARBUF:
    case PT_FSPATH:
      return CHARBUF;
    case PT_BYTEBUF:
      return BYTEBUF;
    case PT_DOUBLE:
      return DOUBLE;
    case PT_SOCKADDR:
    case PT_SOCKTUPLE:
    case PT_FDLIST:
    default:
      return BYTEBUF;
  }
}

uint16_t get_kindling_category(sinsp_evt* sEvt) {
  sinsp_evt::category cat;
  sEvt->get_category(&cat);
  switch (cat.m_category) {
    case EC_OTHER:
      return CAT_OTHER;
    case EC_FILE:
      return CAT_FILE;
    case EC_NET:
      return CAT_NET;
    case EC_IPC:
      return CAT_IPC;
    case EC_MEMORY:
      return CAT_MEMORY;
    case EC_PROCESS:
      return CAT_PROCESS;
    case EC_SLEEP:
      return CAT_SLEEP;
    case EC_SYSTEM:
      return CAT_SYSTEM;
    case EC_SIGNAL:
      return CAT_SIGNAL;
    case EC_USER:
      return CAT_USER;
    case EC_TIME:
      return CAT_TIME;
    case EC_IO_READ:
    case EC_IO_WRITE:
    case EC_IO_OTHER: {
      switch (cat.m_subcategory) {
        case sinsp_evt::SC_FILE:
          return CAT_FILE;
        case sinsp_evt::SC_NET:
          return CAT_NET;
        case sinsp_evt::SC_IPC:
          return CAT_IPC;
        default:
          return CAT_OTHER;
      }
    }
    default:
      return CAT_OTHER;
  }
}

uint16_t get_kindling_source(uint16_t etype) {
  if (PPME_IS_ENTER(etype)) {
    switch (etype) {
      case PPME_PROCEXIT_E:
      case PPME_SCHEDSWITCH_6_E:
      case PPME_SYSDIGEVENT_E:
      case PPME_CONTAINER_E:
      case PPME_PROCINFO_E:
      case PPME_SCHEDSWITCH_1_E:
      case PPME_DROP_E:
      case PPME_PROCEXIT_1_E:
      case PPME_CPU_HOTPLUG_E:
      case PPME_K8S_E:
      case PPME_TRACER_E:
      case PPME_MESOS_E:
      case PPME_CONTAINER_JSON_E:
      case PPME_NOTIFICATION_E:
      case PPME_INFRASTRUCTURE_EVENT_E:
      case PPME_PAGE_FAULT_E:
        return SOURCE_UNKNOWN;
      case PPME_TCP_RCV_ESTABLISHED_E:
      case PPME_TCP_CLOSE_E:
      case PPME_TCP_DROP_E:
      case PPME_TCP_RETRANCESMIT_SKB_E:
        return KRPOBE;
        // TODO add cases of tracepoint, kprobe, uprobe
      default:
        return SYSCALL_ENTER;
    }
  } else {
    switch (etype) {
      case PPME_CONTAINER_X:
      case PPME_PROCINFO_X:
      case PPME_SCHEDSWITCH_1_X:
      case PPME_DROP_X:
      case PPME_CPU_HOTPLUG_X:
      case PPME_K8S_X:
      case PPME_TRACER_X:
      case PPME_MESOS_X:
      case PPME_CONTAINER_JSON_X:
      case PPME_NOTIFICATION_X:
      case PPME_INFRASTRUCTURE_EVENT_X:
      case PPME_PAGE_FAULT_X:
        return SOURCE_UNKNOWN;
        // TODO add cases of tracepoint, kprobe, uprobe
      default:
        return SYSCALL_EXIT;
    }
  }
}

void attach_pid(char* pid, bool is_new_start, bool is_attach, bool is_all_attach, bool is_ps) {
  char result_buf[1024], command[1024];
  int rc = 0;
  FILE* fp;
  if (is_new_start) {
    sleep(10);
  }
  if (is_all_attach && is_ps) {
    const char* ps_command = "ps -ef | grep \"java\" | grep -v \"grep\" | awk '{print $2}' \0";
    snprintf(command, sizeof(command), "%s", ps_command);
  } else {
    string attach_command_prefix;
    if (is_attach) {
      attach_command_prefix = "./async-profiler/profiler.sh start ";
    } else {
      attach_command_prefix = "./async-profiler/profiler.sh stop ";
    }
    attach_command_prefix.append(pid);
    strcpy(command, attach_command_prefix.c_str());
  }

  fp = popen(command, "r");
  if (NULL == fp) {
    perror("popen execute failed!\n");
    return;
  }
  if (!is_ps && is_attach) {
    cout << "------"
         << " start attach for pid " << pid << "------" << endl;
  }
  if (!is_ps && !is_attach) {
    cout << "------"
         << " start detach for pid " << pid << "------" << endl;
  }

  while (fgets(result_buf, sizeof(result_buf), fp) != NULL) {
    if ('\n' == result_buf[strlen(result_buf) - 1]) {
      result_buf[strlen(result_buf) - 1] = '\0';
    }
    if (is_ps) {
      attach_pid(result_buf, false, is_attach, is_all_attach, false);

    } else {
      printf("%s\r\n", result_buf);
    }
  }

  rc = pclose(fp);
  if (-1 == rc) {
    perror("close command fp failed!\n");
    exit(1);
  } else {
    printf("command:【%s】command process status:【%d】command return value:【%d】\r\n", command,
           rc, WEXITSTATUS(rc));
  }

  if (!is_ps && is_attach) {
    cout << "------end attach for pid " << pid << "------" << endl;
  }
  if (!is_ps && !is_attach) {
    cout << "------end detach for pid " << pid << "------" << endl;
  }
}

void attach_agent(int64_t pid, char* error_message, bool is_attach) {
  char result_buf[1024], command[1024];
  string attach_command_prefix;
  if (is_attach) {
    attach_command_prefix = "./async-profiler/jattach.sh start ";
  } else {
    attach_command_prefix = "./async-profiler/jattach.sh stop ";
  }
  attach_command_prefix.append(std::to_string(pid));
  strcpy(command, attach_command_prefix.c_str());

  FILE* fp;
  fp = popen(command, "r");
  if (NULL == fp) {
    perror("popen execute failed!\n");
    strcpy(error_message, "popen execute failed");
    return;
  }
  if (is_attach) {
    cout << "------"
         << " start attach agent for pid " << pid << "------" << endl;
  } else {
    cout << "------"
         << " start detach agent for pid " << pid << "------" << endl;
  }

  char* error_msg;
  while (fgets(result_buf, sizeof(result_buf), fp) != NULL) {
    if ('\n' == result_buf[strlen(result_buf) - 1]) {
      result_buf[strlen(result_buf) - 1] = '\0';
    }
    error_msg = strstr(result_buf, "[ERROR] ");
    if (error_msg) {
      strcpy(error_message, error_msg);
    }
    printf("%s\r\n", result_buf);
  }

  int rc = pclose(fp);
  if (-1 == rc) {
    perror("close command fp failed!\n");
    strcpy(error_message, "close command fp failed");
    return;
  } else {
    printf("command:【%s】command process status:【%d】command return value:【%d】\r\n", command,
           rc, WEXITSTATUS(rc));
  }

  if (is_attach) {
    cout << "------end attach agent for pid " << pid << "------" << endl;
  } else {
    cout << "------end detach agent for pid " << pid << "------" << endl;
  }
}

char* start_attach_agent(int64_t pid) {
  char* error_message = (char*)malloc(1024 * sizeof(char));
  error_message[0] = '\0';
  if (!inspector) {
    strcpy(error_message, "Please start profile first");
  } else {
    attach_agent(pid, error_message, true);
  }
  return error_message;
}

char* stop_attach_agent(int64_t pid) {
  char* error_message = (char*)malloc(1024 * sizeof(char));
  error_message[0] = '\0';
  if (!inspector) {
    strcpy(error_message, "Please start profile first");
  } else {
    attach_agent(pid, error_message, false);
  }
  return error_message;
}

int start_profile() {
  if (!inspector) {
    return -1;
  }
  is_start_profile = true;
  attach_pid(nullptr, false, true, true, true);
  inspector->set_eventmask(PPME_CPU_ANALYSIS_E);

  return 0;
}

int stop_profile() {
  if (!inspector) {
    return -1;
  }
  is_start_profile = false;
  attach_pid(nullptr, false, false, true, true);
  inspector->unset_eventmask(PPME_CPU_ANALYSIS_E);

  return 0;
}

void start_profile_debug(int64_t pid, int64_t tid) {
  is_profile_debug = true;
  debug_pid = pid;
  debug_tid = tid;
}

void stop_profile_debug() {
  is_profile_debug = false;
  debug_pid = 0;
  debug_tid = 0;
}

void print_profile_debug_info(sinsp_evt* sevt) {
  if (!debug_file_log.is_open()) {
    debug_file_log.open("profile_debug.log", ios::app | ios::out);
  }
  string line;
  if (formatter->tostring(sevt, &line)) {
    if (debug_file_log.is_open()) {
      debug_file_log << sevt->get_ts() << "  ";
      debug_file_log << line;
      debug_file_log << "\n";
    }
  }
}

void get_capture_statistics(struct capture_statistics_for_go* stats) {
  scap_stats s;
  printCurrentTime();
  inspector->get_capture_stats(&s);
  stats->evts = s.n_evts;
  printf("seen by driver: %d \n", stats->evts);
  stats->drops = s.n_drops;
  printf("Number of dropped events: %d \n", stats->drops);
  stats->drops_buffer = s.n_drops_buffer;
  printf("Number of dropped events caused by full buffer: %d \n", stats->drops_buffer);
  stats->drops_pf = s.n_drops_pf;
  printf("Number of dropped events caused by invalid memory access: %d \n",
         stats->drops_pf);
  stats->drops_bug = s.n_drops_bug;
  printf(
      "Number of dropped events caused by an invalid condition in the kernel instrumentation: %d \n", stats->drops_bug);
  stats->preemptions = s.n_preemptions;
  printf("Number of preemptions: %d \n", stats->preemptions);
  stats->suppressed = s.n_suppressed;
  printf("Number of events skipped due to the tid being in a set of suppressed tids: %d \n",
         stats->suppressed);
  stats->tids_suppressed = s.n_tids_suppressed;
  printf("Number of threads currently being suppressed: %d \n", stats->tids_suppressed);
  fflush(stdout);
}

void clear_profile_map() {
  if (inspector) inspector->clear_profile_map();
}

void clear_stacks_map() {
  if (inspector) inspector->clear_stacks_map();
}

int get_profile_data(struct sample_key key, struct bpf_profile_data* sample) {
  if (!inspector) return -1;
  return inspector->get_profile_data(key, sample);
}

int get_profile_keys(struct sample_key_set *set) {
  if (!inspector) return -1;
  return inspector->get_profile_keys(set);
}

void start_cpu_sampeling() {
  if (inspector) inspector->set_cpu_sampling(PROFILING_ENABLED);
}

void start_cpu_sampeling_all() {
  if (inspector) inspector->set_cpu_sampling(PROFILING_CAPTURE_ALL);
}

void stop_cpu_sampeling() {
  if (inspector) inspector->set_cpu_sampling(PROFILING_DISABLED);
}

int set_profile_pids(const uint32_t *pids, uint32_t n) {
  if (!inspector)
    return -1;
  inspector->clear_profile_pid_config();
  if (!pids || n == 0)
    return 0;
  struct pid_config cfg = {};
  cfg.type = PROFILING_TYPE_FRAMEPOINTERS;
  cfg.collect = 1;
  for (uint32_t i = 0; i < n; i++) {
    if (inspector->set_profile_pid_config(pids[i], &cfg) != 0)
      return -1;
  }
  return 0;
}
