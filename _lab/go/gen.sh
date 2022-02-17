#!/bin/bash

# set -eux
set -e
set -o pipefail

cd "$(git rev-parse --show-toplevel)"
PROTOC="./third_party/protobuf/src/protoc"
PROTOBUF_PROTO="./third_party/protobuf/src"
TABLEAU_PROTO="./third_party/tableau/proto"
PROTOCONF_IN="./protoconf"
PROTOCONF_OUT="./_lab/go/testpb"

# remove old generated files
rm -rfv "$PROTOCONF_OUT"
mkdir -p "$PROTOCONF_OUT"
${PROTOC} \
--go_out="$PROTOCONF_OUT" \
--go_opt=paths=source_relative \
--proto_path="$PROTOBUF_PROTO" \
--proto_path="$TABLEAU_PROTO" \
--proto_path="$PROTOCONF_IN" \
"$PROTOCONF_IN"/*
