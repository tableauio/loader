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
    # Use dynamic CRT (/MD or /MDd) to match protobuf's build configuration.
    set(CMAKE_MSVC_RUNTIME_LIBRARY "MultiThreaded$<$<CONFIG:Debug>:Debug>DLL")
else()
    SET(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -g -fPIC -Wno-deprecated -Wno-unused-variable -Wno-sign-compare -Wno-strict-aliasing -fno-strict-aliasing -DNDEBUG")
endif()

# google protobuf
# Try to find protobuf from local submodule first, then fallback to system.
set(LOCAL_PROTOBUF_BUILD_DIR "${PROJECT_SOURCE_DIR}/../../third_party/_submodules/protobuf/cmake/build")
set(LOCAL_PROTOBUF_SRC_DIR "${PROJECT_SOURCE_DIR}/../../third_party/_submodules/protobuf/src")
if(EXISTS "${LOCAL_PROTOBUF_BUILD_DIR}" AND EXISTS "${LOCAL_PROTOBUF_SRC_DIR}")
    message(STATUS "Found local protobuf submodule, using it preferentially.")
    list(PREPEND CMAKE_PREFIX_PATH "${LOCAL_PROTOBUF_BUILD_DIR}")
    set(Protobuf_INCLUDE_DIR "${LOCAL_PROTOBUF_SRC_DIR}" CACHE PATH "Protobuf include dir" FORCE)
    # Use static protobuf to avoid dynamic linking issues across all platforms.
    set(Protobuf_USE_STATIC_LIBS ON)
endif()
find_package(Protobuf REQUIRED)
message(STATUS "Using protobuf ${Protobuf_VERSION}")

# When linking protobuf statically, PROTOBUF_USE_DLLS must not be defined,
# otherwise symbols will be decorated with __declspec(dllimport) and cause LNK2019.
if(Protobuf_USE_STATIC_LIBS)
    remove_definitions(-DPROTOBUF_USE_DLLS)
endif()

# loader
SET(LOADER_SRC_DIR ${PROJECT_SOURCE_DIR}/src/)

# include
include_directories(${Protobuf_INCLUDE_DIRS} ${LOADER_SRC_DIR} ${LOADER_SRC_DIR}/protoconf)

# add the executable
add_executable(${PROJECT_NAME} ${PROTO_SOURCE} ${SOURCE})
set(EXECUTABLE_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/bin)

# link libraries
if(MSVC)
    target_link_libraries(${PROJECT_NAME} protobuf::libprotobuf)
else()
    target_link_libraries(${PROJECT_NAME} protobuf::libprotobuf pthread stdc++fs)
endif()
