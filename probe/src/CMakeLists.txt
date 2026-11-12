include(libscap)
include(libsinsp)
include(spdlog)

include_directories(.)
include_directories("${LIBSINSP_INCLUDE_DIRS}")
include_directories("../cgo")
include_directories("../converter")
include_directories("./core")

set(SOURCE_FILES
		cgo/cgo_func.cpp
		cgo/profiling_cgo.cpp
		core/profiling/profile_map_writer.cpp
		core/profiling/profile_trace_reader.cpp
		core/profiling/pid_filter.cpp
		cgo/kindling.cpp
		converter/cpu_converter.cpp
		cgo/catch_sig.cpp cgo/catch_sig.h
		cgo/utils.cpp cgo/utils.h
		core/symbolizer/bcc_per_pid_symbolizer.cpp
		core/symbolizer/kindling_symbolizer.cpp)

add_library(kindling SHARED ${SOURCE_FILES})
set(CMAKE_POSITION_INDEPENDENT_CODE ON)
if(TARGET spdlog)
 add_dependencies(kindling spdlog)
endif()

target_link_libraries(kindling
		sinsp
		Qt5::Core
		bcc_syms_minimal)
install(TARGETS kindling
		DESTINATION bin)
