#!/bin/bash

# set -eux
set -e
set -o pipefail

shopt -s globstar

cd "$(git rev-parse --show-toplevel)"
go mod tidy

PROTOC="./third_party/_submodules/protobuf/cmake/build/protoc"
PROTOBUF_PROTO="./third_party/_submodules/protobuf/src"
TABLEAU_GOPATH="github.com/tableauio/tableau"
TABLEAU_PROTO="$(go env GOPATH)/pkg/mod/$TABLEAU_GOPATH@$(grep $TABLEAU_GOPATH go.mod | awk '{print $2}')/proto"
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
