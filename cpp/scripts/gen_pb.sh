#!/bin/bash

# set -eux
set -e
set -o pipefail

cd "$(git rev-parse --show-toplevel)"

# bash ./scripts/gen_pb.sh

PROTOC="./third_party/protobuf/src/protoc"
# tableau_proto="./proto"
PROTOCONF_IN="./protoconf"
PROTOCONF_OUT="./cpp/src/protoconf"

# remove *.go
rm -rf $PROTOCONF_OUT
mkdir -p $PROTOCONF_OUT

for item in "$PROTOCONF_IN"/* ; do
    echo "$item"
    if [ -f "$item" ]; then
        ${PROTOC} \
        --cpp_out="$PROTOCONF_OUT" \
        --proto_path="$PROTOCONF_IN" \
        "$item"
    fi
done