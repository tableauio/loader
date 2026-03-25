#!/bin/bash
# set -eux
set -e
set -o pipefail

cd "$(git rev-parse --show-toplevel)"

git submodule update --init --recursive

# prerequisites
# On Ubuntu/Debian, you can install them with:
# sudo apt-get install autoconf automake libtool curl make g++ unzip

# Build and install the C++ Protocol Buffer runtime and the Protocol Buffer compiler (protoc)
# Refer: https://github.com/protocolbuffers/protobuf/blob/3.19.x/cmake/README.md#cmake-configuration
cd third_party/_submodules/protobuf/cmake
# use Debug version
# - protobuf_MSVC_STATIC_RUNTIME defaults to ON, which uses static CRT (/MTd for Debug).
#   Our project's CMakeLists.txt also sets static CRT to match.
# - protobuf_WITH_ZLIB=OFF: disable ZLIB dependency to avoid ZLIB::ZLIB link requirement
#   in protobuf's exported CMake targets, which simplifies cross-platform builds.
# - protobuf_BUILD_SHARED_LIBS=OFF: build static libraries explicitly.
cmake -S . -B build -G Ninja \
    -DCMAKE_BUILD_TYPE=Debug \
    -DCMAKE_CXX_STANDARD=17 \
    -Dprotobuf_BUILD_TESTS=OFF \
    -Dprotobuf_WITH_ZLIB=OFF \
    -Dprotobuf_BUILD_SHARED_LIBS=OFF

# Compile the code
cmake --build build --parallel
