#!/bin/bash

# set -eux
set -e
set -o pipefail

cd "$(git rev-parse --show-toplevel)"
PROTOC="./third_party/protobuf/src/protoc"
PROTOBUF_PROTO="./third_party/protobuf/src"
TABLEAU_PROTO="./third_party/tableau/proto"
PROTOCONF_IN="./protoconf"
PROTOCONF_OUT="./cpp/src/protoconf"

# remove old generated files
rm -rfv "$PROTOCONF_OUT"
mkdir -p "$PROTOCONF_OUT"
${PROTOC} \
--cpp_out="$PROTOCONF_OUT" \
--proto_path="$PROTOBUF_PROTO" \
--proto_path="$TABLEAU_PROTO" \
--proto_path="$PROTOCONF_IN" \
"$PROTOCONF_IN"/*

TABLEAU_IN="./third_party/tableau/proto/tableau/protobuf"
TABLEAU_OUT="./cpp/src/"
# remove old generated files
rm -rfv "$TABLEAU_OUT/tableau"
mkdir -p "$TABLEAU_OUT/tableau"

${PROTOC} \
--cpp_out="$TABLEAU_OUT" \
--proto_path="$PROTOBUF_PROTO" \
--proto_path="$TABLEAU_PROTO" \
"${TABLEAU_PROTO}/tableau/protobuf/tableau.proto"