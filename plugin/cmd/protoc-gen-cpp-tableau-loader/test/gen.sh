#!/bin/bash

# set -eux
set -e
set -o pipefail

cd "$(git rev-parse --show-toplevel)"
ROOTDIR="./plugin/cmd/protoc-gen-cpp-tableau-loader"
INDIR="${ROOTDIR}/test/testdata"
OUTDIR="${ROOTDIR}/test/src"

PLUGIN="${ROOTDIR}/protoc-gen-cpp-tableau-loader"

mkdir -p "$OUTDIR"
protoc --plugin "$PLUGIN" --cpp_out="$OUTDIR" --cpp-tableau-loader_out="$OUTDIR" --cpp-tableau-loader_opt=paths=source_relative -I "$INDIR" item.proto