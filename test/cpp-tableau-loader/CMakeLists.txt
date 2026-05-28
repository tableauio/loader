cmake_minimum_required(VERSION 3.22)

# set the project name
project(loader)

# Glob the loader sources (proto-generated .cc and the hub/custom .cpp files
# under src/, but NOT files under tests/ — those are owned by the test target).
file(GLOB_RECURSE PROTO_SOURCE ${PROJECT_SOURCE_DIR}/src/*.cc)
file(GLOB_RECURSE SOURCE ${PROJECT_SOURCE_DIR}/src/*.cpp)

# Globbed test sources.
file(GLOB_RECURSE TEST_SOURCE ${PROJECT_SOURCE_DIR}/tests/*.cpp)

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
    SET(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} /W4")
    # Use static CRT (/MT or /MTd) to match protobuf's default static runtime build.
    set(CMAKE_MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>")
else()
    SET(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -g -fPIC -Wno-deprecated -Wno-unused-variable -Wno-sign-compare -Wno-strict-aliasing -fno-strict-aliasing")
endif()

# google protobuf
# Try to find protobuf from local submodule first, then fallback to system.
# Protobuf's install.cmake generates config files under:
#   - Linux:   build/lib64/cmake/protobuf/ (via CMAKE_INSTALL_LIBDIR)
#   - Windows: build/cmake/                (MSVC uses "cmake" directly)
# Setting CMAKE_PREFIX_PATH to the build dir allows CMake to search both layouts.
set(LOCAL_PROTOBUF_INSTALL_DIR "${PROJECT_SOURCE_DIR}/../../third_party/_submodules/protobuf/.build/_install")
set(LOCAL_PROTOBUF_SRC_DIR "${PROJECT_SOURCE_DIR}/../../third_party/_submodules/protobuf/src")
if(EXISTS "${LOCAL_PROTOBUF_INSTALL_DIR}" AND EXISTS "${LOCAL_PROTOBUF_SRC_DIR}")
    message(STATUS "Found local protobuf submodule, using it preferentially.")
    list(PREPEND CMAKE_PREFIX_PATH "${LOCAL_PROTOBUF_INSTALL_DIR}")
endif()

# Use CONFIG mode explicitly to pick up protobuf's own protobuf-config.cmake
# instead of CMake's built-in FindProtobuf.cmake module, which may not handle
# the local build directory layout correctly.
find_package(Protobuf CONFIG REQUIRED)
message(STATUS "Using protobuf ${Protobuf_VERSION}")

# GoogleTest via FetchContent, pinned to a stable release. Using FetchContent
# (instead of add_subdirectory on protobuf's bundled googletest) gives a
# consistent test framework regardless of which protobuf version is in use.
#
# gtest_force_shared_crt:
#   OFF  -> let CMAKE_MSVC_RUNTIME_LIBRARY decide (we set it to MultiThreaded[Debug]
#           above, i.e. static CRT /MT or /MTd).
#   ON   -> force googletest to use the DYNAMIC CRT (/MD or /MDd).
# Our protobuf submodule is built with protobuf_MSVC_STATIC_RUNTIME=ON (static
# CRT), and our loader_lib also targets static CRT. Forcing gtest to a different
# CRT would mix two C runtimes inside loader.exe; the link may still succeed but
# global-destructor sequencing at process exit can hit an Access Violation
# (gtest's STL objects holding handles allocated by a different heap). Keep gtest
# on the same static CRT as everything else.
include(FetchContent)
set(gtest_force_shared_crt OFF CACHE BOOL "" FORCE)
FetchContent_Declare(
    googletest
    GIT_REPOSITORY https://github.com/google/googletest.git
    GIT_TAG        v1.14.0
    GIT_SHALLOW    TRUE
)
FetchContent_MakeAvailable(googletest)

enable_testing()

# loader
SET(LOADER_SRC_DIR ${PROJECT_SOURCE_DIR}/src/)

# include
# protobuf::libprotobuf target already provides its include directories via
# INTERFACE_INCLUDE_DIRECTORIES, so we only need to add our own source dirs.
include_directories(${LOADER_SRC_DIR} ${LOADER_SRC_DIR}/protoconf ${PROJECT_SOURCE_DIR})

# Static library bundling all proto-generated and hub sources, reused by the
# test executable. This avoids re-compiling the (very large) generated proto
# code if more test binaries are added in the future.
add_library(loader_lib STATIC ${PROTO_SOURCE} ${SOURCE})
target_link_libraries(loader_lib PUBLIC protobuf::libprotobuf)
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
            target_link_options(loader_lib PUBLIC "-L${_GCC_LIB_DIR}")
        endif()
    endif()
    target_link_libraries(loader_lib PUBLIC pthread stdc++fs)
endif()

add_executable(${PROJECT_NAME} ${TEST_SOURCE})
set_target_properties(${PROJECT_NAME} PROPERTIES OUTPUT_NAME loader_test)
set(EXECUTABLE_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/bin)
target_link_libraries(${PROJECT_NAME} PRIVATE loader_lib GTest::gtest_main)

include(GoogleTest)
gtest_discover_tests(${PROJECT_NAME}
    WORKING_DIRECTORY ${PROJECT_SOURCE_DIR}
    DISCOVERY_TIMEOUT 60
)
