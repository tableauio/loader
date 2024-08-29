#!/bin/bash

# set -eux
set -e
set -o pipefail

cd "$(git rev-parse --show-toplevel)"
PROTOC="./third_party/_submodules/protobuf/src/protoc"
PROTOBUF_PROTO="./third_party/_submodules/protobuf/src"
TABLEAU_PROTO="./third_party/_submodules/tableau/proto"
ROOTDIR="./test/cpp-tableau-loader"
PLGUIN_DIR="./cmd/protoc-gen-cpp-tableau-loader"
PROTOCONF_IN="./test/proto"
PROTOCONF_OUT="${ROOTDIR}/src/protoconf"

# remove old generated files
rm -rfv "$PROTOCONF_OUT"
mkdir -p "$PROTOCONF_OUT"

# build
cd "${PLGUIN_DIR}" && go build && cd -

# generate
# for item in "$PROTOCONF_IN"/* ; do
#     echo "$item"
#     if [ -f "$item" ]; then
#         ${PROTOC} \
#         --cpp-tableau-loader_out="$PROTOCONF_OUT" \
#         --cpp-tableau-loader_opt=paths=source_relative \
#         --cpp_out="$PROTOCONF_OUT" \
#         --proto_path="$PROTOBUF_PROTO" \
#         --proto_path="$TABLEAU_PROTO" \
#         --proto_path="$PROTOCONF_IN" \
#         "$item"
#     fi
# done

export PATH="${PATH}:${PLGUIN_DIR}"

${PROTOC} \
--plugin "$PLUGIN" \
--cpp-tableau-loader_out="$PROTOCONF_OUT" \
--cpp-tableau-loader_opt=paths=source_relative,registry-shards=2 \
--cpp_out="$PROTOCONF_OUT" \
--proto_path="$PROTOBUF_PROTO" \
--proto_path="$TABLEAU_PROTO" \
--proto_path="$PROTOCONF_IN" \
"$PROTOCONF_IN"/*.proto "$PROTOCONF_IN"/base/*.proto

TABLEAU_IN="./third_party/_submodules/tableau/proto/tableau/protobuf"
TABLEAU_OUT="${ROOTDIR}/src"
# remove old generated files
rm -rfv "$TABLEAU_OUT/tableau"
mkdir -p "$TABLEAU_OUT/tableau"

${PROTOC} \
--cpp_out="$TABLEAU_OUT" \
--proto_path="$PROTOBUF_PROTO" \
--proto_path="$TABLEAU_PROTO" \
"${TABLEAU_PROTO}/tableau/protobuf/tableau.proto"
    
