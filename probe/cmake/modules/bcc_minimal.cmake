cmake_minimum_required(VERSION 3.16)

option(BCC_FETCH "Fetch bcc sources at configure time" ON)

include(ExternalProject)

## zlib: static -fPIC
#set(_ZLIB_BUILD ${CMAKE_BINARY_DIR}/ext/zlib/build)
#ExternalProject_Add(ext_zlib
#        URL https://zlib.net/zlib-1.3.1.tar.gz
#        SOURCE_DIR  ${CMAKE_BINARY_DIR}/ext/zlib/src
#        BINARY_DIR  ${_ZLIB_BUILD}
#        CMAKE_ARGS -DBUILD_SHARED_LIBS=OFF -DCMAKE_POSITION_INDEPENDENT_CODE=ON -DCMAKE_BUILD_TYPE=Release
#        BUILD_COMMAND ${CMAKE_COMMAND} --build . -- -j
#        INSTALL_COMMAND "")
#add_library(ZLIB::z STATIC IMPORTED)
#set_target_properties(ZLIB::z PROPERTIES IMPORTED_LOCATION ${_ZLIB_BUILD}/libz.a)
#add_dependencies(ZLIB::z ext_zlib)

## elfutils libelf: static -fPIC，仅构建 libelf
#set(_ELF_SRC ${CMAKE_BINARY_DIR}/ext/elfutils/src)
#ExternalProject_Add(ext_elfutils
#        URL https://sourceware.org/elfutils/ftp/0.190/elfutils-0.190.tar.bz2
#        SOURCE_DIR ${_ELF_SRC}
#        CONFIGURE_COMMAND ./configure --enable-static --disable-shared
#        --disable-debuginfod --disable-libdebuginfod
#        --without-lzma --without-bzlib --without-zstd
#        CFLAGS=-fPIC
#        BUILD_COMMAND make -j
#        INSTALL_COMMAND ""
#        BUILD_IN_SOURCE 1)
#add_library(LIBELF::elf STATIC IMPORTED)
#set_target_properties(LIBELF::elf PROPERTIES IMPORTED_LOCATION ${_ELF_SRC}/libelf/.libs/libelf.a)
#add_dependencies(LIBELF::elf ext_elfutils)

# 获取 bcc 源码（配置期可见），只编最小子集
include(FetchContent)
FetchContent_Declare(bcc_src
        URL https://github.com/iovisor/bcc/archive/refs/tags/v0.30.0.tar.gz)
FetchContent_GetProperties(bcc_src)
if(NOT bcc_src_POPULATED)
    FetchContent_Populate(bcc_src)  # 仅下载/解压，不 add_subdirectory
endif()
set(BCC_SRC_DIR ${bcc_src_SOURCE_DIR})
set(BCC_INCLUDE_DIR "${BCC_SRC_DIR}/src/cc")


add_library(bcc_syms_minimal STATIC
        ${BCC_SRC_DIR}/src/cc/bcc_syms.cc
        ${BCC_SRC_DIR}/src/cc/bcc_elf.c
        ${BCC_SRC_DIR}/src/cc/bcc_perf_map.c
        ${BCC_SRC_DIR}/src/cc/bcc_proc.c
        ${BCC_SRC_DIR}/src/cc/bcc_zip.c)
set_target_properties(bcc_syms_minimal PROPERTIES POSITION_INDEPENDENT_CODE ON)
target_include_directories(bcc_syms_minimal PUBLIC "${BCC_INCLUDE_DIR}")
#target_link_libraries(bcc_syms_minimal PUBLIC LIBELF::elf ZLIB::z)
find_library(LIBELF_LIB elf REQUIRED)
find_library(ZLIB_LIB z   REQUIRED)
target_link_libraries(bcc_syms_minimal PUBLIC ${LIBELF_LIB} ${ZLIB_LIB})
