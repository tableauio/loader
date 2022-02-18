#!/bin/bash

# set -eux
set -e
set -o pipefail

cd "$(git rev-parse --show-toplevel)"
PROTOC="./third_party/_submodules/protobuf/src/protoc"
PROTOBUF_PROTO="./third_party/_submodules/protobuf/src"
TABLEAU_PROTO="./third_party/_submodules/tableau/proto"
PROTOCONF_IN="./test/proto"
PROTOCONF_OUT="./_lab/cpp/src/protoconf"

# remove old generated files
rm -rfv "$PROTOCONF_OUT"
mkdir -p "$PROTOCONF_OUT"
${PROTOC} \
--cpp_out="$PROTOCONF_OUT" \
--proto_path="$PROTOBUF_PROTO" \
--proto_path="$TABLEAU_PROTO" \
--proto_path="$PROTOCONF_IN" \
"$PROTOCONF_IN"/*

TABLEAU_IN="./third_party/_submodules/tableau/proto/tableau/protobuf"
TABLEAU_OUT="./_lab/cpp/src/"
# remove old generated files
rm -rfv "$TABLEAU_OUT/tableau"
mkdir -p "$TABLEAU_OUT/tableau"

${PROTOC} \
--cpp_out="$TABLEAU_OUT" \
--proto_path="$PROTOBUF_PROTO" \
--proto_path="$TABLEAU_PROTO" \
"${TABLEAU_PROTO}/tableau/protobuf/tableau.proto"