cmake_minimum_required(VERSION 3.22)

# set the project name
project(loader)
file(GLOB_RECURSE PROTO_SOURCE ${PROJECT_SOURCE_DIR}/src/*.cc)
file(GLOB_RECURSE SOURCE ${PROJECT_SOURCE_DIR}/src/*.cpp)

# check C++ standard requirement
set(MIN_CXX_STANDARD 17)
if (NOT CMAKE_CXX_STANDARD)
    set(CMAKE_CXX_STANDARD ${MIN_CXX_STANDARD}) # default min C++ standard
else()
    if(CMAKE_CXX_STANDARD LESS ${MIN_CXX_STANDARD})
        message(FATAL_ERROR "CMAKE_CXX_STANDARD is less than the minimum required standard: C++${MIN_CXX_STANDARD}.")
    endif()
endif()
set(CMAKE_CXX_STANDARD_REQUIRED True)
message(STATUS "Using C++${CMAKE_CXX_STANDARD} standard")

if (MSVC)
    SET(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} /W4 /DNDEBUG")
    # Use static CRT (/MT or /MTd) to match protobuf's default static runtime build.
    set(CMAKE_MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>")
else()
    SET(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -g -fPIC -Wno-deprecated -Wno-unused-variable -Wno-sign-compare -Wno-strict-aliasing -fno-strict-aliasing -DNDEBUG")
endif()

# google protobuf
# Try to find protobuf from local submodule first, then fallback to system.
# Protobuf's install.cmake generates config files under:
#   - Linux:   build/lib64/cmake/protobuf/ (via CMAKE_INSTALL_LIBDIR)
#   - Windows: build/cmake/                (MSVC uses "cmake" directly)
# Setting CMAKE_PREFIX_PATH to the build dir allows CMake to search both layouts.
set(LOCAL_PROTOBUF_BUILD_DIR "${PROJECT_SOURCE_DIR}/../../third_party/_submodules/protobuf/cmake/build")
set(LOCAL_PROTOBUF_SRC_DIR "${PROJECT_SOURCE_DIR}/../../third_party/_submodules/protobuf/src")
if(EXISTS "${LOCAL_PROTOBUF_BUILD_DIR}" AND EXISTS "${LOCAL_PROTOBUF_SRC_DIR}")
    message(STATUS "Found local protobuf submodule, using it preferentially.")
    list(PREPEND CMAKE_PREFIX_PATH "${LOCAL_PROTOBUF_BUILD_DIR}")
endif()

# Use CONFIG mode explicitly to pick up protobuf's own protobuf-config.cmake
# instead of CMake's built-in FindProtobuf.cmake module, which may not handle
# the local build directory layout correctly.
find_package(Protobuf CONFIG REQUIRED)
message(STATUS "Using protobuf ${Protobuf_VERSION}")

# loader
SET(LOADER_SRC_DIR ${PROJECT_SOURCE_DIR}/src/)

# include
# protobuf::libprotobuf target already provides its include directories via
# INTERFACE_INCLUDE_DIRECTORIES, so we only need to add our own source dirs.
include_directories(${LOADER_SRC_DIR} ${LOADER_SRC_DIR}/protoconf)

# add the executable
add_executable(${PROJECT_NAME} ${PROTO_SOURCE} ${SOURCE})
set(EXECUTABLE_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/bin)

# link libraries
target_link_libraries(${PROJECT_NAME} protobuf::libprotobuf)
if(NOT MSVC)
    # When using clang with a GCC toolchain (e.g. gcc-toolset-13 on RHEL/CentOS),
    # clang selects GCC 13's C++ headers but the C compiler (GCC 8) may inject an
    # older libstdc++ search path into the link line. This causes undefined references
    # to symbols like std::__throw_bad_array_new_length() and std::filesystem::*.
    # Detect the GCC library directory used by the CXX compiler and link stdc++/stdc++fs
    # from there explicitly.
    if(CMAKE_CXX_COMPILER_ID MATCHES "Clang")
        execute_process(
            COMMAND ${CMAKE_CXX_COMPILER} -print-libgcc-file-name
            OUTPUT_VARIABLE _LIBGCC_FILE
            OUTPUT_STRIP_TRAILING_WHITESPACE
        )
        if(_LIBGCC_FILE)
            get_filename_component(_GCC_LIB_DIR "${_LIBGCC_FILE}" DIRECTORY)
            message(STATUS "Detected GCC library directory for clang: ${_GCC_LIB_DIR}")
            # Prepend GCC toolchain lib dir so it is searched before the system GCC 8 path.
            target_link_options(${PROJECT_NAME} PRIVATE "-L${_GCC_LIB_DIR}")
        endif()
    endif()
    target_link_libraries(${PROJECT_NAME} pthread stdc++fs)
endif()
