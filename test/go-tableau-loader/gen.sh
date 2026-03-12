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
PLUGIN_DIR="./cmd/protoc-gen-go-tableau-loader"
PROTOCONF_IN="./test/proto"
PROTOCONF_OUT="./test/go-tableau-loader/protoconf"
LOADER_OUT="$PROTOCONF_OUT/loader"

# remove old generated files
rm -rfv "$PROTOCONF_OUT" "$LOADER_OUT"
mkdir -p "$PROTOCONF_OUT" "$LOADER_OUT"

# build
cd "${PLUGIN_DIR}" && go build && cd -

export PATH="$(pwd)/${PLUGIN_DIR}:${PATH}"

# Collect all .proto files (use `find` for cross-platform compatibility).
PROTO_FILES=$(find "$PROTOCONF_IN" -name "*.proto")

${PROTOC} \
    --go-tableau-loader_out="$LOADER_OUT" \
    --go-tableau-loader_opt=paths=source_relative,pkg=loader \
    --go_out="$PROTOCONF_OUT" \
    --go_opt=paths=source_relative \
    --proto_path="$PROTOBUF_PROTO" \
    --proto_path="$TABLEAU_PROTO" \
    --proto_path="$PROTOCONF_IN" \
    $PROTO_FILES
