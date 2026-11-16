# Copyright (C) 2022-2023 Kindling Project Authors.
# SPDX-License-Identifier: Apache-2.0

GitCommit=${shell git rev-parse --short HEAD || echo unknow}

define exec-command
$(1)
endef

LIBS_SRC := $(abspath ./probe/libs/agent-libs)

.PHONY: collector
collector: libkindling
	@echo Build Env:
	@echo Agent is build based on commit: ${GitCommit}
	cd collector && go build -o docker/kindling-collector -ldflags="-X 'github.com/Kindling-project/kindling/collector/version.CodeVersion=${GitCommit}'" ./cmd/kindling-collector/

.PHONY: libkindling
libkindling: patch-libs
	mkdir -p probe/build && cd probe/build && cmake -DBUILD_DRIVER=OFF -DPROBE_VERSION=0.1.1dev -DAGENT_LIBS_SOURCE_DIR=$(LIBS_SRC) .. && make 
	mkdir -p collector/docker/libso/. && cp -rf probe/build/src/libkindling.so collector/docker/libso/. && cp -rf probe/build/src/libkindling.so /usr/lib64/.

COLLECTOR_PATH := $(abspath ./collector)
SCRIPTS_PATH := $(abspath ./deploy/scripts)
PROBE_PATH := $(abspath ./probe)
PROBE_PATCH_FILE := $(PROBE_PATH)/cmake/modules/agent-libs.cmake

kernelVersion = $(shell uname -r)
ErrorExit = exit 1
SuccessExit = exit 0

LIBS_IMAGE ?= kindlingproject/kernel-builder:latest
KINDLING_IMAGE ?= kindlingproject/agent-builder:latest


## Check if probe build image and collector build image exists or pull from dockerhub
.PHONY: builder-images
builder-images:
	@echo "Checking builder images...";
	@if [ -z "$(shell docker images -q $(LIBS_IMAGE))" ]; then \
		echo "Downloading probe build image..."; \
		docker pull $(LIBS_IMAGE); \
	else \
		echo "Probe build image already exists"; \
	fi
	@if [ -z "$(shell docker images -q $(KINDLING_IMAGE))" ]; then \
		echo "Downloading collector build image..."; \
		docker pull $(KINDLING_IMAGE); \
	else \
		echo "Collector build image already exists"; \
	fi


## Check kernel headers exists or download, support Ubuntu, CentOS
.PHONY: kernel-headers
kernel-headers:
	@if [ $(shell cat /etc/os-release | grep Ubuntu | wc -l) -ne 0 ]; then \
		make _ubuntu-headers; \
	else \
		if [ $(shell cat /etc/os-release | grep CentOS | wc -l) -ne 0 ]; then \
			make _centos-kernel; \
		else \
			echo "System kernel is not the expected Ubuntu and CentOS.Try installing kernel headers locally"; \
			$(ErrorExit); \
		fi \
	fi

.PHONY: _ubuntu-headers
_ubuntu-headers:
	@if [ -z "$(shell dpkg -l | grep linux-headers-$(kernelVersion))" ]; then \
		echo "Downloading linux-headers-$(kernelVersion)..."; \
		sudo apt-get -y install linux-headers-$(kernelVersion); \
	else \
		echo "Linux-headers-$(kernelVersion) already installed"; \
	fi

.PHONY: _centos-kernel
_centos-kernel:
	@if [ -z "$(shell yum list installed |grep kernel-devel-$(kernelVersion))" ]; then \
		echo "Downloading kernel-devel-$(kernelVersion)..."; \
		sudo yum -y install kernel-devel-$(kernelVersion); \
	else \
		echo "Kernel-devel-$(kernelVersion) already installed"; \
	fi

## Check libs exists or download, patch cmake file
.PHONY: patch-libs
patch-libs:
	@echo "Checking libs...";
	@if [ ! "$(shell ls -A1q $(LIBS_SRC))" ]; then \
		echo "$(LIBS_SRC) not exist, downloading....."; \
		git submodule update --init --recursive; \
	else \
		echo "Libs src: $(LIBS_SRC) exist"; \
	fi

## Grant executable permissions to build.sh
.PHONY: patch-scripts
patch-scripts:
	@echo "Grant executable permissions to build.sh";
	@cd $(SCRIPTS_PATH) && chmod +x build.sh;

## Check all depends
.PHONY: depends
depends: builder-images kernel-headers patch-libs patch-scripts

## Build probe in docker with /lib/modules
.PHONY: build-probe
build-probe:
	@echo "Building kindling probe..."
	@docker run \
		--env "ENV_HEADER_VERSION=$(kernelVersion)" \
		--rm -it \
		-v /usr:/host/usr \
		-v /lib/modules:/host/lib/modules \
		-v $(LIBS_SRC):/source \
		$(LIBS_IMAGE);

.PHONY: pack-probe
pack-probe:
	@echo "Packaging kindling probe...";
	@if [ ! -d "$(LIBS_SRC)/kindling-falcolib-probe/" ]; then \
		echo "The packaged probe does not exist.try again"; \
		$(ErrorExit); \
	else \
		cd $(LIBS_SRC) && \
		tar -cvzf kindling-falcolib-probe.tar.gz kindling-falcolib-probe/ && \
		mv kindling-falcolib-probe.tar.gz $(COLLECTOR_PATH)/docker/ ;\
	fi

## Build kindling probe in docker
.PHONY: kindling-probe
kindling-probe: build-probe pack-probe

.PHONY: _build-collector
_build-collector:
	@cd $(SCRIPTS_PATH) && sh ./build.sh

## Build kindling collector in docker
.PHONY: kindling-collector
kindling-collector:
	@if [ ! -f "$(COLLECTOR_PATH)/docker/kindling-falcolib-probe.tar.gz" ]; then \
		echo "The kindling-falcolib-probe.tar.gz does not exist. make probe-libs first"; \
		$(ErrorExit); \
	fi
	@cd $(SCRIPTS_PATH) && bash -c "./run_docker.sh make _build-collector";
	@$(SuccessExit) ;

## Install depends and build probe locally and build collector
.PHONY: docker-build-all
docker-build-all: depends kindling-probe kindling-collector

## Clean build
.PHONY: clean
clean:
	sudo rm -rf probe/libs/agent-libs/kindling-falcolib-probe
	sudo rm -rf collector/docker/kindling-falcolib-probe.tar.gz
	sudo rm -rf probe/build/
	sudo rm -rf collector/docker/libso/libkindling.so
	sudo rm -rf collector/docker/kindling-collector
	@if [ "$(shell docker images -q kindling-agent:latest)" ]; then \
		docker rmi kindling-agent:latest; \
		echo "Remove old docker image: kindling-agent:latest"; \
	fi
