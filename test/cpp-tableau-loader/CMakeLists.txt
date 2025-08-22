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
else()
    SET(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -g -fPIC -Wno-deprecated -Wno-unused-variable -Wno-sign-compare -Wno-strict-aliasing -fno-strict-aliasing -DNDEBUG")
endif()

# root dir define
SET(CMAKE_ROOT_DIR ${CMAKE_CURRENT_SOURCE_DIR})
SET(THRID_PARTY_DIR ${CMAKE_CURRENT_SOURCE_DIR}/../../third_party)

# google protobuf
SET(PROTOBUF_ROOT_DIR ${THRID_PARTY_DIR}/_submodules/protobuf/)
SET(PROTOBUF_SRC_DIR ${PROTOBUF_ROOT_DIR}/src/)
SET(PROTOBUF_BUILD_DIR ${PROTOBUF_ROOT_DIR}/cmake/build/)
# loader
SET(LOADER_SRC_DIR ${PROJECT_SOURCE_DIR}/src/)

# protoconf
SET(PROTOCONF_DIR ${TOOLS_DIR}/protoconf)
SET(PROTOCONF_TOOL ${PROTOCONF_DIR}/protoconf)

# common include
SET(COMMON_INCLUDE_DIR
    ${PROTOBUF_SRC_DIR}
    ${LOADER_SRC_DIR}
    )

# common lib
if(MSVC)
    SET(COMMON_LIB
        ${PROTOBUF_BUILD_DIR}/libprotobufd.lib
        )
    SET(CMAKE_MSVC_RUNTIME_LIBRARY "MultiThreadedDebug")
else()
    SET(COMMON_LIB
        ${PROTOBUF_BUILD_DIR}/libprotobufd.a
        pthread
        stdc++fs
        )
endif()

# include
include_directories(${COMMON_INCLUDE_DIR} ${LOADER_SRC_DIR}/protoconf)

# add the executable
add_executable(${PROJECT_NAME} ${PROTO_SOURCE} ${SOURCE})
set(EXECUTABLE_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/bin)
target_link_libraries(${PROJECT_NAME} ${COMMON_LIB})