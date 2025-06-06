cmake_minimum_required(VERSION 3.22)

# set the project name
project(loader)
file(GLOB_RECURSE PROTO_SOURCE *.cc)
file(GLOB_RECURSE SOURCE *.cpp)

# Add option to switch between C++11 and C++17
option(USE_CPP17 "Use C++17 standard (otherwise use C++11)" OFF)

# specify the C++ standard
if(USE_CPP17)
    set(CMAKE_CXX_STANDARD 17)
    message(STATUS "Using C++17 standard")
else()
    set(CMAKE_CXX_STANDARD 11)
    message(STATUS "Using C++11 standard")
endif()
set(CMAKE_CXX_STANDARD_REQUIRED True)

if (MSVC)
    if(USE_CPP17)
        SET(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} /W4 /DNDEBUG /std:c++17")
    else()
        SET(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} /W4 /DNDEBUG")
    endif()
else()
    if(USE_CPP17)
        SET(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -g -fPIC -std=c++17 -Wno-deprecated -Wno-unused-variable -Wno-sign-compare -Wno-strict-aliasing -fno-strict-aliasing -DNDEBUG")
        # For older GCC versions that need explicit linking with stdc++fs
        if(CMAKE_CXX_COMPILER_ID STREQUAL "GNU" AND CMAKE_CXX_COMPILER_VERSION VERSION_LESS "9.0")
            set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -lstdc++fs")
        endif()
    else()
        SET(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -g -fPIC -std=c++11 -Wno-deprecated -Wno-unused-variable -Wno-sign-compare -Wno-strict-aliasing -fno-strict-aliasing -DNDEBUG")
    endif()
endif()

# root dir define
SET(CMAKE_ROOT_DIR ${CMAKE_CURRENT_SOURCE_DIR})
SET(THRID_PARTY_DIR ${CMAKE_CURRENT_SOURCE_DIR}/../../../third_party)

# google protobuf
SET(PROTOBUF_ROOT_DIR ${THRID_PARTY_DIR}/_submodules/protobuf/)
SET(PROTOBUF_DIR ${PROTOBUF_ROOT_DIR}/src/)
# SET(PROTOBUF_INCLUDE ${PROTOBUF_DIR}/google/protobuf)
SET(PROTOBUF_LIB ${PROTOBUF_DIR}/.libs/)

# protoconf
SET(PROTOCONF_DIR ${TOOLS_DIR}/protoconf)
SET(PROTOCONF_TOOL ${PROTOCONF_DIR}/protoconf)

# common include
SET(COMMON_INCLUDE_DIR
    ${PROTOBUF_DIR}
    )

# common lib
if(MSVC)
    SET(COMMON_LIB
        ${PROTOBUF_ROOT_DIR}/cmake/Debug/libprotobufd.lib
        )
    SET(CMAKE_MSVC_RUNTIME_LIBRARY "MultiThreadedDebug")
else()
    SET(COMMON_LIB
        ${PROTOBUF_LIB}/libprotobuf.a
        pthread
        $<$<BOOL:${USE_CPP17}>:stdc++fs>
        )
endif()

# include
include_directories(${PROJECT_SOURCE_DIR} ${COMMON_INCLUDE_DIR} ${PROJECT_SOURCE_DIR}/protoconf)

# add the executable
add_executable(${PROJECT_NAME} ${PROTO_SOURCE} ${SOURCE})
set(EXECUTABLE_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/../bin)
target_link_libraries(${PROJECT_NAME} ${COMMON_LIB})