#!/bin/bash

# set -eux
set -e
set -o pipefail

shopt -s globstar

cd "$(git rev-parse --show-toplevel)"
PROTOC="./third_party/_submodules/protobuf/src/protoc"
PROTOBUF_PROTO="./third_party/_submodules/protobuf/src"
TABLEAU_PROTO="./third_party/_submodules/tableau/proto"
ROOTDIR="./test/csharp-tableau-loader"
PLGUIN_DIR="./cmd/protoc-gen-csharp-tableau-loader"
PROTOCONF_IN="./test/proto"
PROTOCONF_OUT="${ROOTDIR}/protoconf"
LOADER_OUT="${ROOTDIR}/tableau"

# remove old generated files
rm -rfv "$PROTOCONF_OUT" "$LOADER_OUT"
mkdir -p "$PROTOCONF_OUT" "$LOADER_OUT"

# build
cd "${PLGUIN_DIR}" && go build && cd -

export PATH="${PATH}:${PLGUIN_DIR}"

${PROTOC} \
    --csharp_out="$PROTOCONF_OUT" \
    --csharp-tableau-loader_out="$LOADER_OUT" \
    --csharp-tableau-loader_opt=paths=source_relative \
    --proto_path="$PROTOBUF_PROTO" \
    --proto_path="$TABLEAU_PROTO" \
    --proto_path="$PROTOCONF_IN" \
    "$PROTOCONF_IN"/**/*.proto

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
