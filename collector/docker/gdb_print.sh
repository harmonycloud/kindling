#!/bin/sh
# Copyright (C) 2022 Kindling Project Authors.
# SPDX-License-Identifier: Apache-2.0


binname=$1
corefile=$2

(gdb $binname $corefile > ./$corefile.log 2>&1) <<GDBEOF
thread apply all bt
GDBEOF

cat ./$corefile.log
rm -f ./$corefile.log ./$corefile