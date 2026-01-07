# Loader

The official config loader for [Tableau](https://github.com/tableauio/tableau).

## Prerequisites

> TODO: [devcontainer](https://code.visualstudio.com/docs/devcontainers/containers)

- C++ standard: at least C++17
- Install: [CMake 3.22](https://github.com/Kitware/CMake/releases/tag/v3.31.8) or above
- Init protobuf:
  - macOS or Linux: `bash init.sh`
  - Windows: 
    - Install [Visual Studio 2022](https://visualstudio.microsoft.com/downloads/)
    - Environment Setup: Open the appropriate `Developer Command Prompt for VS 2022` from the *Start* menu to ensure `cl.exe` and other build tools are in your `PATH`.
    - Change dir to **loader** repo
    - Run: `.\init.bat`

## C++

### Dev at Linux

- Change dir: `cd test/cpp-tableau-loader`
- Generate protoconf: `bash ./gen.sh`
- CMake:
  - C++17: `cmake -S . -B build`
  - C++20: `cmake -S . -B build -DCMAKE_CXX_STANDARD=20`
  - clang: `cmake -S . -B build -DCMAKE_CXX_COMPILER=clang++`
- Build: `cmake --build build -j16`
- Run: `./bin/loader`

### Dev at Windows

- Change dir: `cd test\cpp-tableau-loader`
- Generate protoconf: `.\gen.bat`
- CMake:
  - C++17: `cmake -S . -B build -G "NMake Makefiles"`
  - C++20: `cmake -S . -B build -G "NMake Makefiles" -DCMAKE_CXX_STANDARD=20`
- Build: `cmake --build build`
- Run: `.\bin\loader.exe`

### References

- [Protocol Buffers C++ Installation](https://github.com/protocolbuffers/protobuf/tree/master/src)
- [Protocol Buffers C++ Reference](https://protobuf.dev/reference/cpp/)

## Go

- Install: **go1.21** or above
- Install protoc-gen-go: `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
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


> [protobufjs: Reflection vs. static code](https://github.com/protobufjs/protobuf.js/blob/master/cli/README.md#reflection-vs-static-code) 

If using reflection (`.proto` or `JSON`) but not static code, and for well-known types support, then [proto3-json-serializer](https://github.com/googleapis/proto3-json-serializer-nodejs) is a good option. This library implements proto3 JSON serialization and deserialization for
[protobuf.js](https://www.npmjs.com/package/protobufjs) protobuf objects
according to the [spec](https://protobuf.dev/programming-guides/proto3/#json).

### References:

- [How to Setup a TypeScript + Node.js Project](https://khalilstemmler.com/blogs/typescript/node-starter-project/)
- [proto3-json-serializer](https://github.com/googleapis/proto3-json-serializer-nodejs)
