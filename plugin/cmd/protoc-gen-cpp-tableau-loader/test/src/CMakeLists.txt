cmake_minimum_required(VERSION 2.8)

# set the project name
project(loader)
file(GLOB_RECURSE PROTO_SOURCE *.cc)
file(GLOB_RECURSE SOURCE *.cpp)

# specify the C++ standard
# set(CMAKE_CXX_STANDARD 11)
# set(CMAKE_CXX_STANDARD_REQUIRED True)
SET(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -Wall -g -fPIC -std=c++11 -Wno-literal-suffix -Wno-error=literal-suffix -Wno-deprecated -Wno-unused-variable -Wno-sign-compare -Wno-strict-aliasing -fno-strict-aliasing -DNDEBUG")

# root dir define
SET(CMAKE_ROOT_DIR ${CMAKE_CURRENT_SOURCE_DIR})
SET(THRID_PARTY_DIR ${CMAKE_CURRENT_SOURCE_DIR}/../../../../../third_party)

# google protobuf
SET(PROTOBUF_ROOT_DIR ${THRID_PARTY_DIR}/protobuf/)
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
SET(COMMON_LIB
    ${PROTOBUF_LIB}/libprotobuf.a
    pthread
    )

# inlcude
include_directories(${PROJECT_SOURCE_DIR} ${COMMON_INCLUDE_DIR})

# add the executable
add_executable(${PROJECT_NAME} ${PROTO_SOURCE} ${SOURCE})
set(EXECUTABLE_OUTPUT_PATH ${PROJECT_SOURCE_DIR}/../bin)
target_link_libraries(${PROJECT_NAME} ${COMMON_LIB})