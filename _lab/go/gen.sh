#!/bin/bash

# set -eux
set -e
set -o pipefail

cd "$(git rev-parse --show-toplevel)"
PROTOC="./third_party/_submodules/protobuf/src/protoc"
PROTOBUF_PROTO="./third_party/_submodules/protobuf/src"
TABLEAU_PROTO="./third_party/_submodules/tableau/proto"
PROTOCONF_IN="./test/proto"
PROTOCONF_OUT="./_lab/go/protoconf"

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
