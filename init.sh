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
cd third_party/protobuf
git checkout v3.19.3
git submodule update --init --recursive
./autogen.sh
# Build and install the C++ Protocol Buffer runtime and the Protocol Buffer compiler (protoc)
./configure
make -j64
make check
# sudo make install
# sudo ldconfig # refresh shared library cache.