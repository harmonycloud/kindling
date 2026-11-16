// Copyright (C) 2022 Kindling Project Authors.
// SPDX-License-Identifier: Apache-2.0

#ifndef KINDLING_INTERFACE_H
#define KINDLING_INTERFACE_H
#include <QtCore>
#include "sinsp.h"
class KindlingInterface {
 public:
  virtual int addCache(sinsp_evt* sevt, sinsp* inspector) = 0;
  virtual string getCache(int64_t tid, vector<pair<uint64_t, uint64_t>>& periods,
                          vector<uint8_t> off_type, bool is_off, bool is_log, bool is_stack) = 0;
};
Q_DECLARE_INTERFACE(KindlingInterface, "sdafasdfsadfsadfsadfsadf");
#endif  // MYINTERFACE_H