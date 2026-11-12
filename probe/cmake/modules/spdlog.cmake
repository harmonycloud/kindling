option(USE_BUNDLED_SPDLOG "Enable building of the bundled spdlog" ${USE_BUNDLED_DEPS})

if(USE_BUNDLED_SPDLOG)
    set(SPDLOG_BUNDLE_DIR "${PROJECT_BINARY_DIR}/spdlog-prefix/src/spdlog")
    set(SPDLOG_INCLUDE_DIR "${SPDLOG_BUNDLE_DIR}/include")
    message(STATUS "SPDLOG_INCLUDE_DIR= '${SPDLOG_INCLUDE_DIR}'")
    if(NOT TARGET spdlog)

        message(STATUS "Using bundled spdlog in '${SPDLOG_INCLUDE_DIR}'")

        include(ExternalProject)
        ExternalProject_Add(spdlog
                PREFIX "${PROJECT_BINARY_DIR}/spdlog-prefix"
                URL "https://github.com/gabime/spdlog/archive/refs/tags/v1.14.0.tar.gz"
                CONFIGURE_COMMAND ""
                BUILD_COMMAND ""
                INSTALL_COMMAND ""
                BUILD_IN_SOURCE 1
        )
    endif()
    include_directories(${SPDLOG_INCLUDE_DIR})
else()

    find_path(SPDLOG_INCLUDE_DIR spdlog/spdlog.h)
    if(NOT SPDLOG_INCLUDE_DIR)
        message(FATAL_ERROR "Could not find spdlog include directory")
    endif()
    include_directories(${SPDLOG_INCLUDE_DIR})
endif()
