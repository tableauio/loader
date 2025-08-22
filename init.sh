#!/bin/bash
# set -eux
set -e
set -o pipefail

cd "$(git rev-parse --show-toplevel)"

git submodule update --init --recursive

# prerequisites
# On Ubuntu/Debian, you can install them with:
# sudo apt-get install autoconf automake libtool curl make g++ unzip

### reference: https://github.com/protocolbuffers/protobuf/tree/v3.19.3/src

# google protobuf
cd third_party/_submodules/protobuf
git checkout v3.19.3
git submodule update --init --recursive

# Build and install the C++ Protocol Buffer runtime and the Protocol Buffer compiler (protoc)
# Refer: https://github.com/protocolbuffers/protobuf/blob/3.19.x/cmake/README.md#cmake-configuration
cd cmake
# use Debug version
cmake -S . -B build \
 -DCMAKE_BUILD_TYPE=Debug \
 -DCMAKE_POLICY_VERSION_MINIMUM="3.5"

# Compile the code
cmake --build build