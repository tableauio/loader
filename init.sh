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
cd third_party/_submodules/protobuf

# If PROTOBUF_REF is set, switch submodule to the specified ref
if [ -n "${PROTOBUF_REF:-}" ]; then
    echo "Switching protobuf submodule to ${PROTOBUF_REF}..."
    git fetch --tags
    git checkout "${PROTOBUF_REF}"
    git submodule update --init --recursive
fi

# Detect protobuf major version to determine cmake source directory and arguments.
# - protobuf v3.x uses cmake/ subdirectory for CMake builds with minimal options.
# - protobuf v4+ (v21+) and latest (v32+) use the root directory with additional options.
PROTOBUF_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "unknown")
echo "Detected protobuf version: ${PROTOBUF_VERSION}"

# Extract major version number from tag (e.g., v3.19.3 -> 3, v32.0 -> 32)
MAJOR_VERSION=$(echo "${PROTOBUF_VERSION}" | sed 's/^v//' | cut -d. -f1)

if [ "${MAJOR_VERSION}" -le 3 ] 2>/dev/null; then
    # Legacy protobuf (v3.x): CMakeLists.txt is in cmake/ subdirectory
    echo "Using legacy cmake/ subdirectory for protobuf ${PROTOBUF_VERSION}"
    cmake -S cmake -B .build -G Ninja \
        -DCMAKE_BUILD_TYPE=Debug \
        -DCMAKE_CXX_STANDARD=17 \
        -Dprotobuf_BUILD_TESTS=OFF \
        -Dprotobuf_WITH_ZLIB=OFF \
        -Dprotobuf_BUILD_SHARED_LIBS=OFF
else
    # Modern protobuf (v4+/v21+/v32+): CMakeLists.txt is in root directory
    # Refer: https://github.com/protocolbuffers/protobuf/blob/v32.0/cmake/README.md#cmake-configuration
    echo "Using root CMakeLists.txt for protobuf ${PROTOBUF_VERSION}"
    # - protobuf_WITH_ZLIB=OFF: disable ZLIB dependency to avoid ZLIB::ZLIB link requirement
    #   in protobuf's exported CMake targets, which simplifies cross-platform builds.
    # - protobuf_BUILD_SHARED_LIBS=OFF: build static libraries explicitly.
    cmake -S . -B .build -G Ninja \
        -DCMAKE_BUILD_TYPE=Debug \
        -DCMAKE_CXX_STANDARD=17 \
        -Dprotobuf_BUILD_TESTS=OFF \
        -Dprotobuf_WITH_ZLIB=OFF \
        -Dprotobuf_BUILD_SHARED_LIBS=OFF \
        -Dutf8_range_ENABLE_INSTALL=ON
fi

# Compile the code
cmake --build .build --parallel

# Install into .build/_install so that protobuf-config.cmake (along with
# absl and utf8_range configs) is generated for find_package(Protobuf CONFIG)
# used by downstream CMakeLists.txt.
# NOTE: .build/ is already in protobuf's .gitignore, so _install stays clean.
cmake --install .build --prefix .build/_install
