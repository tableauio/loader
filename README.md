# Loader

The official config loader for [Tableau](https://github.com/tableauio/tableau).

## Prerequisites

> TODO: [devcontainer](https://code.visualstudio.com/docs/devcontainers/containers)

- C++ standard: at least C++17
- Prepare and init:
  - macOS or Linux: `bash init.sh`
  - Windows:
    1. Run `prepare.bat` **as Administrator** to automatically install all build dependencies ([Chocolatey](https://chocolatey.org/), [CMake](https://github.com/Kitware/CMake/releases), [Ninja](https://ninja-build.org/), and MSVC build tools), configure `PATH`, and initialize the MSVC compiler environment:
       ```bat
       .\prepare.bat
       ```
       > ⚠️ **Admin required:** This script uses Chocolatey and MSI installers that write to system-protected directories (`C:\ProgramData`, `C:\Program Files`). Right-click Command Prompt → **Run as administrator**, then execute the script.
       >
       > Preview what the script would do without making any changes:
       > ```bat
       > .\prepare.bat --dry-run
       > ```
    2. Run `init.bat` to initialize submodules and build protobuf:
       ```bat
       .\init.bat
       ```
    > **Note:** `prepare.bat` only needs to be run once per machine. It detects already-installed tools and skips them — no manual Visual Studio, CMake, or Ninja installation required.

### References

- [Chocolatey](https://chocolatey.org/)
- [CMake 3.31.8](https://github.com/Kitware/CMake/releases/tag/v3.31.8)
- [Ninja](https://ninja-build.org/)
- [Visual Studio 2022](https://visualstudio.microsoft.com/downloads/)
- [Use the Microsoft C++ Build Tools from the command line](https://learn.microsoft.com/en-us/cpp/build/building-on-the-command-line?view=msvc-170)

## C++

### Dev at Linux

- Change dir: `cd test/cpp-tableau-loader`
- Generate protoconf: `PATH=../../third_party/_submodules/protobuf/cmake/build:$PATH buf generate ..`
- CMake:
  - C++17: `cmake -S . -B build`
  - C++20: `cmake -S . -B build -DCMAKE_CXX_STANDARD=20`
  - clang: `cmake -S . -B build -DCMAKE_CXX_COMPILER=clang++`
- Build: `cmake --build build --parallel`
- Run: `./bin/loader`

### Dev at Windows

> **Important:** CMake with Ninja requires MSVC environment variables (`cl.exe`, `INCLUDE`, `LIB`, etc.) to be active. Run `.\prepare.bat` from the **loader** root in the **same cmd session** before switching to the test directory. Opening a new terminal window will lose these variables.

- Initialize MSVC environment (from loader root): `.\prepare.bat`
- Change dir: `cd test\cpp-tableau-loader`, or change directory with Drive, e.g.: `cd /D D:\GitHub\loader\test\cpp-tableau-loader`
- Generate protoconf: `cmd /C "set PATH=..\..\third_party\_submodules\protobuf\cmake\build;%PATH% && buf generate .."`
- CMake:
  - C++17: `cmake -S . -B build -G "Ninja"`
  - C++20: `cmake -S . -B build -G "Ninja" -DCMAKE_CXX_STANDARD=20`
- Build: `cmake --build build --parallel`
- Run: `.\bin\loader.exe`

### References

- [Protocol Buffers C++ Installation](https://github.com/protocolbuffers/protobuf/tree/master/src)
- [Protocol Buffers C++ Reference](https://protobuf.dev/reference/cpp/)

## Go

- Install: **go1.21** or above
- Change dir: `cd test/go-tableau-loader`
- Generate protoconf: `buf generate .. `
- Run: `go run .`

### References

- [Protocol Buffers Go Reference](https://protobuf.dev/reference/go/)

## C#

### Requirements

- Unity 2022.3 LTS (C# 9)
- dotnet-sdk-8.0

### Test

- Install: **dotnet-sdk-8.0**
- Change dir: `cd test/csharp-tableau-loader`
- Generate protoconf: `PATH=../third_party/_submodules/protobuf/cmake/build:$PATH buf generate .. `
- Test: `dotnet run`

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
