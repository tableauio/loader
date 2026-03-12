#!/bin/bash

# set -eux
set -e
set -o pipefail

cd "$(git rev-parse --show-toplevel)"

# Allow overriding protoc via environment variable.
# Default to the locally compiled protoc for local development.
PROTOC="${PROTOC:-./third_party/_submodules/protobuf/cmake/build/protoc}"
PROTOBUF_PROTO="./third_party/_submodules/protobuf/src"
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
