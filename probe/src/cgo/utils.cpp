// Copyright (C) 2023 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

//
// Created by Daxin Wang on 2023/3/17.
//

#include "utils.h"
#include <cstdio>
#include <cstdlib>
#include <time.h>

char *date = (char *)(malloc(sizeof(char) * 30));
void printCurrentTime() {
  time_t now = time(nullptr);
  strftime(date, 30, "%Y/%m/%d %H:%M:%S", localtime(&now));
  printf("%s ", date);
}