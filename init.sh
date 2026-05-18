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

# -----------------------------------------------------------------------------
# Detect protobuf major version and build the cmake command line. We compute
# both up-front (before the fast-path check) so the signature comparison below
# can include the exact cmake invocation we would run.
#   - protobuf v3.x  : CMakeLists.txt is in cmake/ subdirectory
#   - protobuf v4+   : CMakeLists.txt is in root directory
# -----------------------------------------------------------------------------
PROTOBUF_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "unknown")
echo "Detected protobuf version: ${PROTOBUF_VERSION}"

# Extract major version number from tag (e.g., v3.19.3 -> 3, v32.0 -> 32)
MAJOR_VERSION=$(echo "${PROTOBUF_VERSION}" | sed 's/^v//' | cut -d. -f1)

if [ "${MAJOR_VERSION}" -le 3 ] 2>/dev/null; then
    # Legacy protobuf (v3.x): CMakeLists.txt is in cmake/ subdirectory
    PROTOBUF_BUILD_VARIANT="legacy"
    CMAKE_ARGS=(
        -S cmake
        -B .build
        -G Ninja
        -DCMAKE_BUILD_TYPE=Release
        -DCMAKE_CXX_STANDARD=17
        -Dprotobuf_BUILD_TESTS=OFF
        -Dprotobuf_WITH_ZLIB=OFF
        -Dprotobuf_BUILD_SHARED_LIBS=OFF
    )
else
    # Modern protobuf (v4+/v21+/v32+): CMakeLists.txt is in root directory
    # Refer: https://github.com/protocolbuffers/protobuf/blob/v32.0/cmake/README.md#cmake-configuration
    # - protobuf_WITH_ZLIB=OFF: disable ZLIB dependency to avoid ZLIB::ZLIB link requirement
    #   in protobuf's exported CMake targets, which simplifies cross-platform builds.
    # - protobuf_BUILD_SHARED_LIBS=OFF: build static libraries explicitly.
    PROTOBUF_BUILD_VARIANT="modern"
    CMAKE_ARGS=(
        -S .
        -B .build
        -G Ninja
        -DCMAKE_BUILD_TYPE=Release
        -DCMAKE_CXX_STANDARD=17
        -Dprotobuf_BUILD_TESTS=OFF
        -Dprotobuf_WITH_ZLIB=OFF
        -Dprotobuf_BUILD_SHARED_LIBS=OFF
        -Dutf8_range_ENABLE_INSTALL=ON
    )
fi

# Build a stable, multi-line signature describing the inputs that determine
# the contents of .build/_install. Any change to these values must invalidate
# the fast-path. Adding new compile-time inputs? Append a line here.
SIG_FILE=".build/_install/.build_signature"
EXPECTED_SIGNATURE=$(printf '%s\n' \
    "schema=1" \
    "version=${PROTOBUF_VERSION}" \
    "variant=${PROTOBUF_BUILD_VARIANT}" \
    "cmake_args=${CMAKE_ARGS[*]}")

# -----------------------------------------------------------------------------
# Fast path: if a previous build's _install dir is present AND its embedded
# signature matches the one we're about to use, skip the (very long) protobuf
# compile entirely.
# Set FORCE_REBUILD_PROTOBUF=1 to bypass this short-circuit unconditionally.
# -----------------------------------------------------------------------------
if [ -z "${FORCE_REBUILD_PROTOBUF:-}" ] && [ -f "${SIG_FILE}" ]; then
    ACTUAL_SIGNATURE=$(cat "${SIG_FILE}")
    if [ "${ACTUAL_SIGNATURE}" = "${EXPECTED_SIGNATURE}" ]; then
        echo "[INFO] Build signature matches; reusing existing protobuf install at .build/_install."
        echo "[INFO] Set FORCE_REBUILD_PROTOBUF=1 to force a clean rebuild."
        exit 0
    fi
    echo "[INFO] Build signature mismatch; rebuilding protobuf."
    echo "[INFO]   actual:"
    printf '%s\n' "${ACTUAL_SIGNATURE}" | sed 's/^/[INFO]     /'
    echo "[INFO]   expected:"
    printf '%s\n' "${EXPECTED_SIGNATURE}" | sed 's/^/[INFO]     /'
fi

# Wipe any stale install dir so we don't leave half-overwritten files behind
# when cmake flags change (e.g. Release -> Debug puts artifacts in different
# places, an in-place re-install would mix old and new).
rm -rf .build 2>/dev/null || true

# -----------------------------------------------------------------------------
# Configure
# -----------------------------------------------------------------------------
if [ "${PROTOBUF_BUILD_VARIANT}" = "legacy" ]; then
    echo "Using legacy cmake/ subdirectory for protobuf ${PROTOBUF_VERSION}"
else
    echo "Using root CMakeLists.txt for protobuf ${PROTOBUF_VERSION}"
fi
cmake "${CMAKE_ARGS[@]}"

# Compile the code
cmake --build .build --parallel

# Install into .build/_install so that protobuf-config.cmake (along with
# absl and utf8_range configs) is generated for find_package(Protobuf CONFIG)
# used by downstream CMakeLists.txt.
# NOTE: .build/ is already in protobuf's .gitignore, so _install stays clean.
cmake --install .build --prefix .build/_install

# Persist the signature so the next run can fast-path skip when nothing changed.
printf '%s\n' "${EXPECTED_SIGNATURE}" > "${SIG_FILE}"
echo "[INFO] Wrote build signature to ${SIG_FILE}"
