#!/bin/bash

# set -eux
set -e
set -o pipefail

shopt -s globstar

cd "$(git rev-parse --show-toplevel)"

# Allow overriding protoc via environment variable.
# Default to locally compiled protoc for local development; fallback to system protoc.
LOCAL_PROTOC="./third_party/_submodules/protobuf/cmake/build/protoc"
if [ -z "$PROTOC" ]; then
    if [ -x "$LOCAL_PROTOC" ]; then
        PROTOC="$LOCAL_PROTOC"
    else
        PROTOC="$(which protoc 2>/dev/null || true)"
    fi
fi
if [ -z "$PROTOC" ]; then
    echo "Error: protoc not found. Please build protobuf submodule or install protoc." >&2
    exit 1
fi
# Allow overriding protobuf include path via environment variable.
# Default to local submodule source; fallback to system include path.
LOCAL_PROTOBUF_PROTO="./third_party/_submodules/protobuf/src"
if [ -z "$PROTOBUF_PROTO" ]; then
    if [ -d "$LOCAL_PROTOBUF_PROTO/google/protobuf" ]; then
        PROTOBUF_PROTO="$LOCAL_PROTOBUF_PROTO"
    else
        PROTOBUF_PROTO="$(pkg-config --variable=includedir protobuf 2>/dev/null || echo /usr/include)"
    fi
fi
TABLEAU_PROTO="./third_party/_submodules/tableau/proto"
ROOTDIR="./test/cpp-tableau-loader"
PLGUIN_DIR="./cmd/protoc-gen-cpp-tableau-loader"
PROTOCONF_IN="./test/proto"
PROTOCONF_OUT="${ROOTDIR}/src/protoconf"

# remove old generated files
rm -rfv "$PROTOCONF_OUT"
mkdir -p "$PROTOCONF_OUT"

# build protoc plugin of loader
cd "${PLGUIN_DIR}" && go build && cd -

export PATH="${PATH}:${PLGUIN_DIR}"

${PROTOC} \
    --cpp-tableau-loader_out="$PROTOCONF_OUT" \
    --cpp-tableau-loader_opt=paths=source_relative,shards=2 \
    --cpp_out="$PROTOCONF_OUT" \
    --proto_path="$PROTOBUF_PROTO" \
    --proto_path="$TABLEAU_PROTO" \
    --proto_path="$PROTOCONF_IN" \
    "$PROTOCONF_IN"/**/*.proto

TABLEAU_IN="$TABLEAU_PROTO/tableau/protobuf"
TABLEAU_OUT="${ROOTDIR}/src"
# remove old generated files
rm -rfv "$TABLEAU_OUT/tableau"
mkdir -p "$TABLEAU_OUT/tableau"

${PROTOC} \
    --cpp_out="$TABLEAU_OUT" \
    --proto_path="$PROTOBUF_PROTO" \
    --proto_path="$TABLEAU_PROTO" \
    "${TABLEAU_PROTO}/tableau/protobuf/tableau.proto" \
    "${TABLEAU_PROTO}/tableau/protobuf/wellknown.proto"
