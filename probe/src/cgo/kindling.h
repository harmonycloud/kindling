//
// Created by jundi zhou on 2022/6/1.
//
#pragma once
#ifndef SYSDIG_KINDLING_H
#define SYSDIG_KINDLING_H

#include <QCoreApplication>
#include <QPluginLoader>
#include <QString>
#include <QtDebug>
#include <QtPlugin>
#include <chrono>
#include <cstdlib>
#include <iomanip>
#include "KindlingInterface.h"
#include "sinsp.h"

int init_probe();

void start_perf();

void stop_perf();

void exipre_window_cache();

int getEvent(uint64_t interval, void** kindlingEvent, int* event_count);

int get_events_by_interval(uint64_t interval, void** kindlingEvent, void* count);

uint16_t get_kindling_category(sinsp_evt* sEvt);

void init_sub_label();

int start_profile();

int stop_profile();

void start_profile_debug(int64_t pid, int64_t tid);

void stop_profile_debug();

void print_profile_debug_info(sinsp_evt* sevt);

void attach_pid(char* pid, bool is_new_start, bool is_attach, bool is_all_attach, bool is_ps);

char* start_attach_agent(int64_t pid);

char* stop_attach_agent(int64_t pid);

void attach_agent(int64_t pid, char* error_message, bool is_attach);

void get_capture_statistics(struct capture_statistics_for_go* stats);

/* CPU continuous profiling (eBPF stack sampling → Pyroscope). */
void clear_profile_map();
void clear_stacks_map();
int get_profile_data(struct sample_key key, struct bpf_profile_data* sample);
int get_profile_keys(struct sample_key_set *set);
void start_cpu_sampeling();
void start_cpu_sampeling_all();
void stop_cpu_sampeling();
int set_profile_pids(const uint32_t *pids, uint32_t n);

uint16_t get_protocol(scap_l4_proto proto);
uint16_t get_type(ppm_param_type type);
uint16_t get_kindling_source(uint16_t etype);


void suppress_events_comm(string comm);

struct event {
  string event_name;
  ppm_event_type event_type;
};
struct event_params_for_subscribe {
  char *name;
  char *value;
};
struct capture_statistics_for_go{
    int evts;
    int drops;
    int drops_buffer;
    int drops_pf;
    int drops_bug;
    int preemptions;
    int suppressed;
    int tids_suppressed;
};
void sub_event(char* eventName, char* category, event_params_for_subscribe params[]);
struct kindling_event_t_for_go {
  uint64_t timestamp;
  char* name;
  uint32_t category;
  uint16_t paramsNumber;
  uint64_t latency;
  struct KeyValue {
    char* key;
    char* value;
    uint32_t len;
    uint32_t valueType;
  } userAttributes[16];
  struct event_context {
    struct thread_info {
      uint32_t pid;
      uint32_t tid;
      uint32_t uid;
      uint32_t gid;
      char* comm;
      char* containerId;
    } tinfo;
    struct fd_info {
      int32_t num;
      uint32_t fdType;
      char* filename;
      char* directory;
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

void parse_jf(char* data_val, sinsp_evt_param data_param, kindling_event_t_for_go* p_kindling_event,
              sinsp_threadinfo* threadInfo, uint16_t& userAttNumber);

void parse_xtid(sinsp_evt* s_evt, char* data_val, sinsp_evt_param data_param,
                kindling_event_t_for_go* p_kindling_event, sinsp_threadinfo* threadInfo,
                uint16_t& userAttNumber);

void parse_span(sinsp_evt *s_evt, char *data_val, sinsp_evt_param data_param,
                kindling_event_t_for_go *p_kindling_event, sinsp_threadinfo* threadInfo,
                uint16_t &userAttNumber);

void parse_tm(char* data_val, sinsp_evt_param data_param, sinsp_threadinfo* threadInfo);

void init_kindling_event(kindling_event_t_for_go* p_kindling_event, void** pp_kindling_event);

void print_event(sinsp_evt* s_evt);

int is_normal_event(int res, sinsp_evt* s_evt, ppm_event_category* category);

int setTuple(kindling_event_t_for_go* kevt, const sinsp_evt_param* pTuple, int userAttNumber);

enum Category {
  CAT_NONE = 0,
  CAT_OTHER = 1,       // No specific category
  CAT_FILE = 2,        // File operations or File I/O
  CAT_NET = 3,         // Network operations or network I/O
  CAT_IPC = 4,         // IPC operations or IPC I/O
  CAT_WAIT = 5,        //
  CAT_SIGNAL = 6,      // Signal-related operations (signal)
  CAT_SLEEP = 7,       // nanosleep
  CAT_TIME = 8,        // Time-related event (e.g. gettimeofday))
  CAT_PROCESS = 9,     // Process-related event (e.g. fork, clone,
  CAT_SCHEDULER = 10,  // Scheduler event (context switch)
  CAT_MEMORY = 11,     // Memory-related operations (e.g. brk, mmap, unmap)
  CAT_USER = 12,       // User-related operations (e.g. getuid, setgid)
  CAT_SYSTEM = 13,     // System-related operations (e.g. reboot)
  Category_MAX = 14
};

enum Source {
  SOURCE_UNKNOWN = 0,
  SYSCALL_ENTER = 1,
  SYSCALL_EXIT = 2,
  TRACEPOINT = 3,
  KRPOBE = 4,
  KRETPROBE = 5,
  UPROBE = 6,
  URETPROBE = 7
};

const static event kindling_to_sysdig[PPM_EVENT_MAX] = {
    {"syscall_enter-open", PPME_SYSCALL_OPEN_E},
    {"syscall_exit-open", PPME_SYSCALL_OPEN_X},
    {"syscall_enter-close", PPME_SYSCALL_CLOSE_E},
    {"syscall_exit-close", PPME_SYSCALL_CLOSE_X},
    {"syscall_enter-read", PPME_SYSCALL_READ_E},
    {"syscall_exit-read", PPME_SYSCALL_READ_X},
    {"syscall_enter-write", PPME_SYSCALL_WRITE_E},
    {"syscall_exit-write", PPME_SYSCALL_WRITE_X},
    {"syscall_enter-brk", PPME_SYSCALL_BRK_4_E},
    {"syscall_exit-brk", PPME_SYSCALL_BRK_4_X},
    {"syscall_enter-execve", PPME_SYSCALL_EXECVE_19_E},
    {"syscall_exit-execve", PPME_SYSCALL_EXECVE_19_X},
    {"syscall_enter-clone", PPME_SYSCALL_CLONE_20_E},
    {"syscall_exit-clone", PPME_SYSCALL_CLONE_20_X},
    {"syscall_enter-socket", PPME_SOCKET_SOCKET_E},
    {"syscall_exit-socket", PPME_SOCKET_SOCKET_X},
    {"syscall_enter-bind", PPME_SOCKET_BIND_E},
    {"syscall_exit-bind", PPME_SOCKET_BIND_X},
    {"syscall_enter-connect", PPME_SOCKET_CONNECT_E},
    {"syscall_exit-connect", PPME_SOCKET_CONNECT_X},
    {"syscall_enter-listen", PPME_SOCKET_LISTEN_E},
    {"syscall_exit-listen", PPME_SOCKET_LISTEN_X},
    {"syscall_enter-accept", PPME_SOCKET_ACCEPT_5_E},
    {"syscall_exit-accept", PPME_SOCKET_ACCEPT_5_X},
    {"syscall_enter-accept4", PPME_SOCKET_ACCEPT4_5_E},
    {"syscall_exit-accept4", PPME_SOCKET_ACCEPT4_5_X},
    {"syscall_enter-sendto", PPME_SOCKET_SENDTO_E},
    {"syscall_exit-sendto", PPME_SOCKET_SENDTO_X},
    {"syscall_enter-recvfrom", PPME_SOCKET_RECVFROM_E},
    {"syscall_exit-recvfrom", PPME_SOCKET_RECVFROM_X},
    {"syscall_enter-shutdown", PPME_SOCKET_SHUTDOWN_E},
    {"syscall_exit-shutdown", PPME_SOCKET_SHUTDOWN_X},
    {"syscall_enter-getsockname", PPME_SOCKET_GETSOCKNAME_E},
    {"syscall_exit-getsockname", PPME_SOCKET_GETSOCKNAME_X},
    {"syscall_enter-getpeername", PPME_SOCKET_GETPEERNAME_E},
    {"syscall_exit-getpeername", PPME_SOCKET_GETPEERNAME_X},
    {"syscall_enter-socketpair", PPME_SOCKET_SOCKETPAIR_E},
    {"syscall_exit-socketpair", PPME_SOCKET_SOCKETPAIR_X},
    {"syscall_enter-setsockopt", PPME_SOCKET_SETSOCKOPT_E},
    {"syscall_exit-setsockopt", PPME_SOCKET_SETSOCKOPT_X},
    {"syscall_enter-getsockopt", PPME_SOCKET_GETSOCKOPT_E},
    {"syscall_exit-getsockopt", PPME_SOCKET_GETSOCKOPT_X},
    {"syscall_enter-sendmsg", PPME_SOCKET_SENDMSG_E},
    {"syscall_exit-sendmsg", PPME_SOCKET_SENDMSG_X},
    {"syscall_enter-sendmmsg", PPME_SOCKET_SENDMMSG_E},
    {"syscall_exit-sendmmsg", PPME_SOCKET_SENDMMSG_X},
    {"syscall_enter-recvmsg", PPME_SOCKET_RECVMSG_E},
    {"syscall_exit-recvmsg", PPME_SOCKET_RECVMSG_X},
    {"syscall_enter-recvmmsg", PPME_SOCKET_RECVMMSG_E},
    {"syscall_exit-recvmmsg", PPME_SOCKET_RECVMMSG_X},
    {"syscall_enter-sendfile", PPME_SYSCALL_SENDFILE_E},
    {"syscall_exit-sendfile", PPME_SYSCALL_SENDFILE_X},
    {"syscall_enter-creat", PPME_SYSCALL_CREAT_E},
    {"syscall_exit-creat", PPME_SYSCALL_CREAT_X},
    {"syscall_enter-pipe", PPME_SYSCALL_PIPE_E},
    {"syscall_exit-pipe", PPME_SYSCALL_PIPE_X},
    {"syscall_enter-pipe2", PPME_SYSCALL_PIPE_E},
    {"syscall_exit-pipe2", PPME_SYSCALL_PIPE_X},
    {"syscall_enter-eventfd", PPME_SYSCALL_EVENTFD_E},
    {"syscall_exit-eventfd", PPME_SYSCALL_EVENTFD_X},
    {"syscall_enter-eventfd2", PPME_SYSCALL_EVENTFD_E},
    {"syscall_exit-eventfd2", PPME_SYSCALL_EVENTFD_X},
    {"syscall_enter-futex", PPME_SYSCALL_FUTEX_E},
    {"syscall_exit-futex", PPME_SYSCALL_FUTEX_X},
    {"syscall_enter-stat", PPME_SYSCALL_STAT_E},
    {"syscall_exit-stat", PPME_SYSCALL_STAT_X},
    {"syscall_enter-lstat", PPME_SYSCALL_LSTAT_E},
    {"syscall_exit-lstat", PPME_SYSCALL_LSTAT_X},
    {"syscall_enter-fstat", PPME_SYSCALL_FSTAT_E},
    {"syscall_exit-fstat", PPME_SYSCALL_FSTAT_X},
    {"syscall_enter-stat64", PPME_SYSCALL_STAT64_E},
    {"syscall_exit-stat64", PPME_SYSCALL_STAT64_X},
    {"syscall_enter-lstat64", PPME_SYSCALL_LSTAT64_E},
    {"syscall_exit-lstat64", PPME_SYSCALL_LSTAT64_X},
    {"syscall_enter-fstat64", PPME_SYSCALL_FSTAT64_E},
    {"syscall_exit-fstat64", PPME_SYSCALL_FSTAT64_X},
    {"syscall_enter-epoll_wait", PPME_SYSCALL_EPOLLWAIT_E},
    {"syscall_exit-epoll_wait", PPME_SYSCALL_EPOLLWAIT_X},
    {"syscall_enter-poll", PPME_SYSCALL_POLL_E},
    {"syscall_exit-poll", PPME_SYSCALL_POLL_X},
    {"syscall_enter-ppoll", PPME_SYSCALL_PPOLL_E},
    {"syscall_exit-ppoll", PPME_SYSCALL_PPOLL_X},
    {"syscall_enter-select", PPME_SYSCALL_SELECT_E},
    {"syscall_exit-select", PPME_SYSCALL_SELECT_X},
    {"syscall_enter-lseek", PPME_SYSCALL_LSEEK_E},
    {"syscall_exit-lseek", PPME_SYSCALL_LSEEK_X},
    {"syscall_enter-llseek", PPME_SYSCALL_LLSEEK_E},
    {"syscall_exit-llseek", PPME_SYSCALL_LLSEEK_X},
    {"syscall_enter-getcwd", PPME_SYSCALL_GETCWD_E},
    {"syscall_exit-getcwd", PPME_SYSCALL_GETCWD_X},
    {"syscall_enter-chdir", PPME_SYSCALL_CHDIR_E},
    {"syscall_exit-chdir", PPME_SYSCALL_CHDIR_X},
    {"syscall_enter-fchdir", PPME_SYSCALL_FCHDIR_E},
    {"syscall_exit-fchdir", PPME_SYSCALL_FCHDIR_X},
    {"syscall_enter-mkdir", PPME_SYSCALL_MKDIR_2_E},
    {"syscall_exit-mkdir", PPME_SYSCALL_MKDIR_2_X},
    {"syscall_enter-mkdirat", PPME_SYSCALL_MKDIRAT_E},
    {"syscall_exit-mkdirat", PPME_SYSCALL_MKDIRAT_X},
    {"syscall_enter-rmdir", PPME_SYSCALL_RMDIR_2_E},
    {"syscall_exit-rmdir", PPME_SYSCALL_RMDIR_2_X},
    {"syscall_enter-unlink", PPME_SYSCALL_UNLINK_2_E},
    {"syscall_exit-unlink", PPME_SYSCALL_UNLINK_2_X},
    {"syscall_enter-unlinkat", PPME_SYSCALL_UNLINKAT_2_E},
    {"syscall_exit-unlinkat", PPME_SYSCALL_UNLINKAT_2_X},
    {"syscall_enter-openat", PPME_SYSCALL_OPENAT_2_E},
    {"syscall_exit-openat", PPME_SYSCALL_OPENAT_2_X},
    {"syscall_enter-link", PPME_SYSCALL_LINK_2_E},
    {"syscall_exit-link", PPME_SYSCALL_LINK_2_X},
    {"syscall_enter-linkat", PPME_SYSCALL_LINKAT_2_E},
    {"syscall_exit-linkat", PPME_SYSCALL_LINKAT_2_X},
    {"syscall_enter-pread", PPME_SYSCALL_PREAD_E},
    {"syscall_exit-pread", PPME_SYSCALL_PREAD_X},
    {"syscall_enter-pwrite", PPME_SYSCALL_PWRITE_E},
    {"syscall_exit-pwrite", PPME_SYSCALL_PWRITE_X},
    {"syscall_enter-readv", PPME_SYSCALL_READV_E},
    {"syscall_exit-readv", PPME_SYSCALL_READV_X},
    {"syscall_enter-writev", PPME_SYSCALL_WRITEV_E},
    {"syscall_exit-writev", PPME_SYSCALL_WRITEV_X},
    {"syscall_enter-preadv", PPME_SYSCALL_PREADV_E},
    {"syscall_exit-preadv", PPME_SYSCALL_PREADV_X},
    {"syscall_enter-pwritev", PPME_SYSCALL_PWRITEV_E},
    {"syscall_exit-pwritev", PPME_SYSCALL_PWRITEV_X},
    {"syscall_enter-dup", PPME_SYSCALL_DUP_E},
    {"syscall_exit-dup", PPME_SYSCALL_DUP_X},
    {"syscall_enter-dup2", PPME_SYSCALL_DUP_E},
    {"syscall_exit-dup2", PPME_SYSCALL_DUP_X},
    {"syscall_enter-dup3", PPME_SYSCALL_DUP_E},
    {"syscall_exit-dup3", PPME_SYSCALL_DUP_X},
    {"syscall_enter-signalfd", PPME_SYSCALL_SIGNALFD_E},
    {"syscall_exit-signalfd", PPME_SYSCALL_SIGNALFD_X},
    {"syscall_enter-signalfd4", PPME_SYSCALL_SIGNALFD_E},
    {"syscall_exit-signalfd4", PPME_SYSCALL_SIGNALFD_X},
    {"syscall_enter-kill", PPME_SYSCALL_KILL_E},
    {"syscall_exit-kill", PPME_SYSCALL_KILL_X},
    {"syscall_enter-tkill", PPME_SYSCALL_TKILL_E},
    {"syscall_exit-tkill", PPME_SYSCALL_TKILL_X},
    {"syscall_enter-tgkill", PPME_SYSCALL_TGKILL_E},
    {"syscall_exit-tgkill", PPME_SYSCALL_TGKILL_X},
    {"syscall_enter-nanosleep", PPME_SYSCALL_NANOSLEEP_E},
    {"syscall_exit-nanosleep", PPME_SYSCALL_NANOSLEEP_X},
    {"syscall_enter-timerfd_create", PPME_SYSCALL_TIMERFD_CREATE_E},
    {"syscall_exit-timerfd_create", PPME_SYSCALL_TIMERFD_CREATE_X},
    {"syscall_enter-inotify_init", PPME_SYSCALL_INOTIFY_INIT_E},
    {"syscall_exit-inotify_init", PPME_SYSCALL_INOTIFY_INIT_X},
    {"syscall_enter-inotify_init1", PPME_SYSCALL_INOTIFY_INIT_E},
    {"syscall_exit-inotify_init1", PPME_SYSCALL_INOTIFY_INIT_X},
    {"syscall_enter-getrlimit", PPME_SYSCALL_GETRLIMIT_E},
    {"syscall_exit-getrlimit", PPME_SYSCALL_GETRLIMIT_X},
    {"syscall_enter-setrlimit", PPME_SYSCALL_SETRLIMIT_E},
    {"syscall_exit-setrlimit", PPME_SYSCALL_SETRLIMIT_X},
    {"syscall_enter-prlimit", PPME_SYSCALL_PRLIMIT_E},
    {"syscall_exit-prlimit", PPME_SYSCALL_PRLIMIT_X},
    {"syscall_enter-fcntl", PPME_SYSCALL_FCNTL_E},
    {"syscall_exit-fcntl", PPME_SYSCALL_FCNTL_X},
    {"syscall_enter-ioctl", PPME_SYSCALL_IOCTL_3_E},
    {"syscall_exit-ioctl", PPME_SYSCALL_IOCTL_3_X},
    {"syscall_enter-mmap", PPME_SYSCALL_MMAP_E},
    {"syscall_exit-mmap", PPME_SYSCALL_MMAP_X},
    {"syscall_enter-mmap2", PPME_SYSCALL_MMAP2_E},
    {"syscall_exit-mmap2", PPME_SYSCALL_MMAP2_X},
    {"syscall_enter-munmap", PPME_SYSCALL_MUNMAP_E},
    {"syscall_exit-munmap", PPME_SYSCALL_MUNMAP_X},
    {"syscall_enter-splice", PPME_SYSCALL_SPLICE_E},
    {"syscall_exit-splice", PPME_SYSCALL_SPLICE_X},
    {"syscall_enter-ptrace", PPME_SYSCALL_PTRACE_E},
    {"syscall_exit-ptrace", PPME_SYSCALL_PTRACE_X},
    {"syscall_enter-rename", PPME_SYSCALL_RENAME_E},
    {"syscall_exit-rename", PPME_SYSCALL_RENAME_X},
    {"syscall_enter-renameat", PPME_SYSCALL_RENAMEAT_E},
    {"syscall_exit-renameat", PPME_SYSCALL_RENAMEAT_X},
    {"syscall_enter-symlink", PPME_SYSCALL_SYMLINK_E},
    {"syscall_exit-symlink", PPME_SYSCALL_SYMLINK_X},
    {"syscall_enter-symlinkat", PPME_SYSCALL_SYMLINKAT_E},
    {"syscall_exit-symlinkat", PPME_SYSCALL_SYMLINKAT_X},
    {"syscall_enter-fork", PPME_SYSCALL_FORK_20_E},
    {"syscall_exit-fork", PPME_SYSCALL_FORK_20_X},
    {"syscall_enter-vfork", PPME_SYSCALL_VFORK_20_E},
    {"syscall_exit-vfork", PPME_SYSCALL_VFORK_20_X},
    {"syscall_enter-quotactl", PPME_SYSCALL_QUOTACTL_E},
    {"syscall_exit-quotactl", PPME_SYSCALL_QUOTACTL_X},
    {"syscall_enter-setresuid", PPME_SYSCALL_SETRESUID_E},
    {"syscall_exit-setresuid", PPME_SYSCALL_SETRESUID_X},
    {"syscall_enter-setresgid", PPME_SYSCALL_SETRESGID_E},
    {"syscall_exit-setresgid", PPME_SYSCALL_SETRESGID_X},
    {"syscall_enter-setuid", PPME_SYSCALL_SETUID_E},
    {"syscall_exit-setuid", PPME_SYSCALL_SETUID_X},
    {"syscall_enter-setgid", PPME_SYSCALL_SETGID_E},
    {"syscall_exit-setgid", PPME_SYSCALL_SETGID_X},
    {"syscall_enter-getuid", PPME_SYSCALL_GETUID_E},
    {"syscall_exit-getuid", PPME_SYSCALL_GETUID_X},
    {"syscall_enter-geteuid", PPME_SYSCALL_GETEUID_E},
    {"syscall_exit-geteuid", PPME_SYSCALL_GETEUID_X},
    {"syscall_enter-getgid", PPME_SYSCALL_GETGID_E},
    {"syscall_exit-getgid", PPME_SYSCALL_GETGID_X},
    {"syscall_enter-getegid", PPME_SYSCALL_GETEGID_E},
    {"syscall_exit-getegid", PPME_SYSCALL_GETEGID_X},
    {"syscall_enter-getresuid", PPME_SYSCALL_GETRESUID_E},
    {"syscall_exit-getresuid", PPME_SYSCALL_GETRESUID_X},
    {"syscall_enter-getresgid", PPME_SYSCALL_GETRESGID_E},
    {"syscall_exit-getresgid", PPME_SYSCALL_GETRESGID_X},
    {"syscall_enter-getdents", PPME_SYSCALL_GETDENTS_E},
    {"syscall_exit-getdents", PPME_SYSCALL_GETDENTS_X},
    {"syscall_enter-getdents64", PPME_SYSCALL_GETDENTS64_E},
    {"syscall_exit-getdents64", PPME_SYSCALL_GETDENTS64_X},
    {"syscall_enter-setns", PPME_SYSCALL_SETNS_E},
    {"syscall_exit-setns", PPME_SYSCALL_SETNS_X},
    {"syscall_enter-flock", PPME_SYSCALL_FLOCK_E},
    {"syscall_exit-flock", PPME_SYSCALL_FLOCK_X},
    {"syscall_enter-semop", PPME_SYSCALL_SEMOP_E},
    {"syscall_exit-semop", PPME_SYSCALL_SEMOP_X},
    {"syscall_enter-semctl", PPME_SYSCALL_SEMCTL_E},
    {"syscall_exit-semctl", PPME_SYSCALL_SEMCTL_X},
    {"syscall_enter-mount", PPME_SYSCALL_MOUNT_E},
    {"syscall_exit-mount", PPME_SYSCALL_MOUNT_X},
    {"syscall_enter-umount", PPME_SYSCALL_UMOUNT_E},
    {"syscall_exit-umount", PPME_SYSCALL_UMOUNT_X},
    {"syscall_enter-semget", PPME_SYSCALL_SEMGET_E},
    {"syscall_exit-semget", PPME_SYSCALL_SEMGET_X},
    {"syscall_enter-access", PPME_SYSCALL_ACCESS_E},
    {"syscall_exit-access", PPME_SYSCALL_ACCESS_X},
    {"syscall_enter-chroot", PPME_SYSCALL_CHROOT_E},
    {"syscall_exit-chroot", PPME_SYSCALL_CHROOT_X},
    {"syscall_enter-setsid", PPME_SYSCALL_SETSID_E},
    {"syscall_exit-setsid", PPME_SYSCALL_SETSID_X},
    {"syscall_enter-setpgid", PPME_SYSCALL_SETPGID_E},
    {"syscall_exit-setpgid", PPME_SYSCALL_SETPGID_X},
    {"syscall_enter-unshare", PPME_SYSCALL_UNSHARE_E},
    {"syscall_exit-unshare", PPME_SYSCALL_UNSHARE_X},
    {"syscall_enter-bpf", PPME_SYSCALL_BPF_E},
    {"syscall_exit-bpf", PPME_SYSCALL_BPF_X},
    {"syscall_enter-seccomp", PPME_SYSCALL_SECCOMP_E},
    {"syscall_exit-seccomp", PPME_SYSCALL_SECCOMP_X},
    {"syscall_enter-fchmodat", PPME_SYSCALL_FCHMODAT_E},
    {"syscall_exit-fchmodat", PPME_SYSCALL_FCHMODAT_X},
    {"syscall_enter-chmod", PPME_SYSCALL_CHMOD_E},
    {"syscall_exit-chmod", PPME_SYSCALL_CHMOD_X},
    {"syscall_enter-fchmod", PPME_SYSCALL_FCHMOD_E},
    {"syscall_exit-fchmod", PPME_SYSCALL_FCHMOD_X},
    {"tracepoint-sched_switch", PPME_SCHEDSWITCH_6_E},
    {"tracepoint-signaldeliver", PPME_SIGNALDELIVER_E},
    {"tracepoint-signaldeliver", PPME_SIGNALDELIVER_X},
    {"syscall_enter-alarm", PPME_GENERIC_E},
    {"syscall_exit-alarm", PPME_GENERIC_X},
    {"syscall_enter-epoll_create", PPME_GENERIC_E},
    {"syscall_exit-epoll_create", PPME_GENERIC_X},
    {"syscall_enter-epoll_ctl", PPME_GENERIC_E},
    {"syscall_exit-epoll_ctl", PPME_GENERIC_X},
    {"syscall_enter-lchown", PPME_GENERIC_E},
    {"syscall_exit-lchown", PPME_GENERIC_X},
    {"syscall_enter-old_select", PPME_GENERIC_E},
    {"syscall_exit-old_select", PPME_GENERIC_X},
    {"syscall_enter-pause", PPME_GENERIC_E},
    {"syscall_exit-pause", PPME_GENERIC_X},
    {"syscall_enter-process_vm_readv", PPME_GENERIC_E},
    {"syscall_exit-process_vm_readv", PPME_GENERIC_X},
    {"syscall_enter-process_vm_writev", PPME_GENERIC_E},
    {"syscall_exit-process_vm_writev", PPME_GENERIC_X},
    {"syscall_enter-pselect6", PPME_GENERIC_E},
    {"syscall_exit-pselect6", PPME_GENERIC_X},
    {"syscall_enter-sched_getparam", PPME_GENERIC_E},
    {"syscall_exit-sched_getparam", PPME_GENERIC_X},
    {"syscall_enter-sched_setparam", PPME_GENERIC_E},
    {"syscall_exit-sched_setparam", PPME_GENERIC_X},
    {"syscall_enter-syslog", PPME_GENERIC_E},
    {"syscall_exit-syslog", PPME_GENERIC_X},
    {"syscall_enter-uselib", PPME_GENERIC_E},
    {"syscall_exit-uselib", PPME_GENERIC_X},
    {"syscall_enter-utime", PPME_GENERIC_E},
    {"syscall_exit-utime", PPME_GENERIC_X},
    {"tracepoint-ingress", PPME_NETIF_RECEIVE_SKB_E},
    {"tracepoint-egress", PPME_NET_DEV_XMIT_E},
    {"kprobe-tcp_close", PPME_TCP_CLOSE_E},
    {"kprobe-tcp_rcv_established", PPME_TCP_RCV_ESTABLISHED_E},
    {"kprobe-tcp_drop", PPME_TCP_DROP_E},
    {"kprobe-tcp_retransmit_skb", PPME_TCP_RETRANCESMIT_SKB_E},
    {"kretprobe-tcp_connect", PPME_TCP_CONNECT_X},
    {"kprobe-tcp_set_state", PPME_TCP_SET_STATE_E},
    {"tracepoint-tcp_send_reset", PPME_TCP_SEND_RESET_E},
    {"tracepoint-tcp_receive_reset", PPME_TCP_RECEIVE_RESET_E},
    {"tracepoint-cpu_analysis", PPME_CPU_ANALYSIS_E},
    {"tracepoint-procexit", PPME_PROCEXIT_1_E},
};

struct event_category {
  string cateogry_name;
  Category category_value;
};

const static event_category category_map[Category_MAX + 1] = {
    {"other", CAT_OTHER},     {"file", CAT_FILE},
    {"net", CAT_NET},         {"ipc", CAT_IPC},
    {"wait", CAT_WAIT},       {"signal", CAT_SIGNAL},
    {"sleep", CAT_SLEEP},     {"time", CAT_TIME},
    {"process", CAT_PROCESS}, {"scheduler", CAT_SCHEDULER},
    {"memory", CAT_MEMORY},   {"user", CAT_USER},
    {"system", CAT_SYSTEM},
};

enum L4Proto { UNKNOWN = 0, TCP = 1, UDP = 2, ICMP = 3, RAW = 4 };

enum ValueType {
  NONE = 0,
  INT8 = 1,      // 1 byte
  INT16 = 2,     // 2 bytes
  INT32 = 3,     // 4 bytes
  INT64 = 4,     // 8 bytes
  UINT8 = 5,     // 1 byte
  UINT16 = 6,    // 2 bytes
  UINT32 = 7,    // 4 bytes
  UINT64 = 8,    // 8 bytes
  CHARBUF = 9,   // bytes, NULL terminated
  BYTEBUF = 10,  // bytes
  FLOAT = 11,    // 4 bytes
  DOUBLE = 12,   // 8 bytes
  BOOL = 13      // 4 bytes
};

const static int EVENT_DATA_SIZE = 80960;

#endif  // SYSDIG_KINDLING_H
