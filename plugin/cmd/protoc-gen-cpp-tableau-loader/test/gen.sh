#!/bin/bash

# set -eux
set -e
set -o pipefail

cd "$(git rev-parse --show-toplevel)"
PROTOC="./third_party/protobuf/src/protoc"
PROTOBUF_PROTO="./third_party/protobuf/src"
TABLEAU_PROTO="./third_party/tableau/proto"
PROTOCONF_IN="./protoconf"

ROOTDIR="./plugin/cmd/protoc-gen-cpp-tableau-loader"
PROTOCONF_OUT="${ROOTDIR}/test/src/protoconf"
PLUGIN="${ROOTDIR}/protoc-gen-cpp-tableau-loader"

# remove old generated files
rm -rfv "$PROTOCONF_OUT"
mkdir -p "$PROTOCONF_OUT"

# build
cd "${ROOTDIR}" && go build && cd -

# generate
for item in "$PROTOCONF_IN"/* ; do
    echo "$item"
    if [ -f "$item" ]; then
        ${PROTOC} \
        --plugin "$PLUGIN" \
        --cpp-tableau-loader_out="$PROTOCONF_OUT" \
        --cpp-tableau-loader_opt=paths=source_relative \
        --cpp_out="$PROTOCONF_OUT" \
        --proto_path="$PROTOBUF_PROTO" \
        --proto_path="$TABLEAU_PROTO" \
        --proto_path="$PROTOCONF_IN" \
        "$item"
    fi
done

TABLEAU_IN="./third_party/tableau/proto/tableau/protobuf"
TABLEAU_OUT="${ROOTDIR}/test/src"
# remove old generated files
rm -rfv "$TABLEAU_OUT/tableau"
mkdir -p "$TABLEAU_OUT/tableau"

${PROTOC} \
--cpp_out="$TABLEAU_OUT" \
--proto_path="$PROTOBUF_PROTO" \
--proto_path="$TABLEAU_PROTO" \
"${TABLEAU_PROTO}/tableau/protobuf/tableau.proto"
    
