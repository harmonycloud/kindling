# Copyright (C) 2022 Kindling Project Authors.
# SPDX-License-Identifier: Apache-2.0

GitCommit=$(git rev-parse --short HEAD || echo unsupported)
echo "Git commit:" "$GitCommit"

BINARY_PATH=docker/kindling-collector
echo "The output path of the go executable file is" "$(pwd)/$BINARY_PATH"

if [ -f $BINARY_PATH ]; then
  echo "There is an old executable file. Clean it first..."
  rm -f $BINARY_PATH
  echo "Clean done."
fi

echo "Start to build the executable..."
go build -v -o $BINARY_PATH -ldflags="-X 'github.com/Kindling-project/kindling/collector/version.CodeVersion=$GitCommit'" ./cmd/kindling-collector/


