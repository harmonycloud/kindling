#
# Copyright (C) 2013-2018 Draios Inc dba Sysdig.
#
# This file is part of sysdig .
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

cmake_minimum_required(VERSION 3.16)

add_compile_options(-fPIC)
add_compile_options(-pie)
option(MINIMAL_BUILD "Produce a minimal sysdig binary with only the essential features (no eBPF probe driver, no kubernetes, no mesos, no marathon and no container metadata)" ON)
option(USE_BUNDLED_DEPS "Enable bundled dependencies instead of using the system ones" ON)

# Add path for custom CMake modules.
list(APPEND CMAKE_MODULE_PATH
	"${CMAKE_CURRENT_SOURCE_DIR}/cmake/modules")

if(NOT CMAKE_BUILD_TYPE)
	SET(CMAKE_BUILD_TYPE Release)
endif()

project(kindling-probe)
set(PACKAGE_NAME "kindling")

add_definitions(-DPLATFORM_NAME="${CMAKE_SYSTEM_NAME}")

option(BUILD_WARNINGS_AS_ERRORS "Enable building with -Wextra -Werror flags")

if(MINIMAL_BUILD)
  set(MINIMAL_BUILD_FLAGS "-DMINIMAL_BUILD")
endif()

set(KINDLING_DEBUG_FLAGS "-D_DEBUG")
set(CMAKE_COMMON_FLAGS "-Wall -ggdb ${MINIMAL_BUILD_FLAGS}")

if (BUILD_WARNINGS_AS_ERRORS)
	set(CMAKE_SUPPRESSED_WARNINGS "-Wno-unused-parameter -Wno-missing-field-initializers -Wno-sign-compare -Wno-type-limits -Wno-implicit-fallthrough -Wno-format-truncation")
	set(CMAKE_COMMON_FLAGS "${CMAKE_COMMON_FLAGS} -Wextra -Werror ${CMAKE_SUPPRESSED_WARNINGS}")
endif ()

set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${CMAKE_COMMON_FLAGS} -std=gnu99")
set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} ${CMAKE_COMMON_FLAGS} -std=c++0x")

set(CMAKE_C_FLAGS_DEBUG "${KINDLING_DEBUG_FLAGS}")
set(CMAKE_CXX_FLAGS_DEBUG "${KINDLING_DEBUG_FLAGS}")

set(CMAKE_C_FLAGS_RELEASE "-O3 -fno-strict-aliasing -DNDEBUG")
set(CMAKE_CXX_FLAGS_RELEASE "-O3 -fno-strict-aliasing -DNDEBUG")

if (CMAKE_SYSTEM_NAME MATCHES "SunOS")
	set(CMD_MAKE gmake)
else ()
	set(CMD_MAKE make)
endif ()

##set(CMAKE_PREFIX_PATH "/opt/homebrew/Cellar/qt/6.3.1_3/bin/qt-cmake")
find_package(Qt5 COMPONENTS
		Core
		Gui
		Widgets
		REQUIRED)

include(agent-libs)
include(bcc_minimal)
add_subdirectory(src)

