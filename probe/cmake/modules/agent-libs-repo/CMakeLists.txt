# Copyright (C) 2022-2023 Kindling Project Authors.
# SPDX-License-Identifier: Apache-2.0

cmake_minimum_required(VERSION 3.5.1)

project(. NONE)

include(ExternalProject)
message(STATUS "Driver version: ${AGENT_LIBS_VERSION}")

if (AGENT_LIBS_SOURCE_DIR)
    ExternalProject_Add(
            agent-libs
            SOURCE_DIR "${AGENT_LIBS_SOURCE_DIR}"
            CONFIGURE_COMMAND ""
            BUILD_COMMAND ""
            INSTALL_COMMAND ""
            TEST_COMMAND "")
else ()
    message(STATUS "No falcosecurity/libs found locally, please set it via AGENT_LIBS_SOURCE_DIR")
endif ()