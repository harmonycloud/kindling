// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

//
// Created by jundi on 2022/12/15.
//

#include "catch_sig.h"

#include <limits.h>
#include <signal.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/types.h>
#include <unistd.h>

static size_t get_executable_path(char* processdir, char* processname, size_t len) {
  char* path_end;
  if (readlink("/proc/self/exe", processdir, len) <= 0) return -1;
  printf("process dir: %s\n", processdir);
  fflush(stdout);
  path_end = strrchr(processdir, '/');
  printf("process end: %s\n", path_end);
  fflush(stdout);
  if (path_end == NULL) return -1;
  ++path_end;
  strcpy(processname, path_end);
  printf("path end: %s, processname: %s\n", path_end, processname);
  fflush(stdout);
  *path_end = '\0';
  return (size_t)(path_end - processdir);
}

static void print_core(int signum, siginfo_t* info, void* secret, struct sigaction* oldact) {
  printf("crash signum:%d si_code:%d\n", signum, info->si_code);
  fflush(stdout);
  char cmd[50];
  sprintf(cmd, "gcore %u", getpid());
  system(cmd);
  char path[PATH_MAX];
  char processname[1024];
  printf("get pid file.\n");
  fflush(stdout);
  get_executable_path(path, processname, sizeof(path));
  // TODO!!!!
  sprintf(cmd, "sh ./gdb_print.sh ./%s ./core.%u", processname, getpid());
  system(cmd);
  oldact->sa_sigaction(signum, info, secret);
}

static struct sigaction oldabrtact;
static void abrtsigaction(int signum, siginfo_t* info, void* secret) {
  print_core(signum, info, secret, &oldabrtact);
}

static struct sigaction oldsegvact;
static void segvsigaction(int signum, siginfo_t* info, void* secret) {
  print_core(signum, info, secret, &oldsegvact);
}

static struct sigaction oldstopact;
static void stopsigaction(int signum, siginfo_t* info, void* secret) {
  print_core(signum, info, secret, &oldstopact);
}

void sig_set_up(void) {
  struct sigaction act;
  memset(&act, 0, sizeof act);
  act.sa_flags = SA_ONSTACK | SA_SIGINFO;
  act.sa_sigaction = segvsigaction;
  sigaction(SIGSEGV, &act, &oldsegvact);
  act.sa_sigaction = abrtsigaction;
  sigaction(SIGABRT, &act, &oldabrtact);
  act.sa_sigaction = stopsigaction;
  sigaction(SIGSTOP, &act, &oldstopact);
}