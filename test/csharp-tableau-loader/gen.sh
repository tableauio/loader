#!/bin/bash

# set -eux
set -e
set -o pipefail

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
ROOTDIR="./test/csharp-tableau-loader"
PLUGIN_DIR="./cmd/protoc-gen-csharp-tableau-loader"
PROTOCONF_IN="./test/proto"
PROTOCONF_OUT="${ROOTDIR}/protoconf"
LOADER_OUT="${ROOTDIR}/tableau"

# remove old generated files
rm -rfv "$PROTOCONF_OUT" "$LOADER_OUT"
mkdir -p "$PROTOCONF_OUT" "$LOADER_OUT"

# build
cd "${PLUGIN_DIR}" && go build && cd -

export PATH="$(pwd)/${PLUGIN_DIR}:${PATH}"

# Collect all .proto files (use `find` for cross-platform compatibility).
PROTO_FILES=$(find "$PROTOCONF_IN" -name "*.proto")

${PROTOC} \
    --csharp_out="$PROTOCONF_OUT" \
    --csharp-tableau-loader_out="$LOADER_OUT" \
    --csharp-tableau-loader_opt=paths=source_relative \
    --proto_path="$PROTOBUF_PROTO" \
    --proto_path="$TABLEAU_PROTO" \
    --proto_path="$PROTOCONF_IN" \
    $PROTO_FILES

TABLEAU_IN="$TABLEAU_PROTO/tableau/protobuf"
TABLEAU_OUT="${ROOTDIR}/protoconf/tableau"
# remove old generated files
rm -rfv "$TABLEAU_OUT"
mkdir -p "$TABLEAU_OUT"

${PROTOC} \
    --csharp_out="$TABLEAU_OUT" \
    --proto_path="$PROTOBUF_PROTO" \
    --proto_path="$TABLEAU_PROTO" \
    "$TABLEAU_IN/tableau.proto" \
    "$TABLEAU_IN/wellknown.proto"
