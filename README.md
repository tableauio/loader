# Loader

The official config loader for [Tableau](https://github.com/tableauio/tableau).

## Prerequisites

> TODO: docker container-dev

- Development OS: linux
- Init protobuf: `bash init.sh`

## C++

- Install: **CMake 3.22** or above
- Change dir: `cd test/cpp-tableau-loader`
- Generate protoconf: `bash ./gen.sh`
- Create build dir: `mkdir build && cd build`
- Run cmake: `cmake ../src/`
- Build: `make -j8`, then the **bin** dir will be generated at `test/cpp-tableau-loader/bin`.

### References

- [Protocol Buffers C++ Installation](https://github.com/protocolbuffers/protobuf/tree/master/src)
- [Protocol Buffers C++ Reference](https://protobuf.dev/reference/cpp/)

## Go

- Change dir: `cd test/go-tableau-loader`
- Generate protoconf: `bash ./gen.sh`
- Build: `go build`

### References

- [Protocol Buffers Go Reference](https://protobuf.dev/reference/go/)

## TypeScript

### Requirements

- nodejs v16.0.0
- protobufjs v7.2.3

### Test

- Change dir: `cd test/ts-tableau-loader`
- Install depedencies: `npm install`
- Generate protoconf: `npm run generate`
- Test: `npm run test`

### Problems in [protobufjs](https://github.com/protobufjs/protobuf.js):

- [Unable to use Google well known types](https://github.com/protobufjs/protobuf.js/issues/1042)
- [google.protobuf.Timestamp deserialization incompatible with canonical JSON representation](https://github.com/protobufjs/protobuf.js/issues/893)
- [Implement wrapper for google.protobuf.Timestamp, and correctly generate wrappers for static target.](https://github.com/protobufjs/protobuf.js/pull/1258)

### References:

- [How to Setup a TypeScript + Node.js Project](https://khalilstemmler.com/blogs/typescript/node-starter-project/)
