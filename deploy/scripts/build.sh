# Copyright (C) 2022-2023 Kindling Project Authors.
# SPDX-License-Identifier: Apache-2.0

cd ../../
mkdir -p probe/build
cd probe/build
agent_libs_src="/kindling/probe/libs/agent-libs/"
cflag="-DBUILD_DRIVER=OFF -DPROBE_VERSION=0.1.1dev"
if [ -d "$agent_libs_src" ]; then
  cflag=$cflag" -DAGENT_LIBS_SOURCE_DIR=$agent_libs_src"
fi
cmake $cflag ..
make
libKindlingPath="./src/libkindling.so"
if [ ! -f "$libKindlingPath" ]; then
  echo "compiler libkindling failed! exit!"

else
  mkdir -p ../../collector/docker/libso && cp -rf ./src/libkindling.so ../../collector/docker/libso/
  cp -rf ./src/libkindling.so /usr/lib64/
  cd ../../collector/
  go mod tidy
  sh collector-version-build.sh
  collectorPath="./docker/kindling-collector"
  if [ ! -f "$collectorPath" ]; then
    echo "compiler collector failed! exit!"
  else
    cd docker
    if [ $1 ]; then
      docker build -t kindling-agent . -f $1;
    else
      docker build -t kindling-agent . -f DockerfileLocalProbe;
    fi

  fi
fi
