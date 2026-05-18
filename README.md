# Loader

The official config loader for [Tableau](https://github.com/tableauio/tableau).

## Prerequisites

> TODO: [devcontainer](https://code.visualstudio.com/docs/devcontainers/containers)

- C++ standard: at least C++17
- Prepare and init:
  - macOS or Linux: `bash init.sh`
  - Windows:
    1. Run `prepare.bat` **as Administrator** to automatically install all build dependencies ([Chocolatey](https://chocolatey.org/), [CMake](https://github.com/Kitware/CMake/releases), [Ninja](https://ninja-build.org/), MSVC build tools, and [buf](https://buf.build/)), configure `PATH`, and initialize the MSVC compiler environment:
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
    > **Note:** The **installation** part of `prepare.bat` only runs once per machine — it detects already-installed tools (Chocolatey, Ninja, CMake, MSVC Build Tools, buf) and skips them, so no manual installation is required.
    >
    > However, the MSVC compiler environment (`cl.exe` on `PATH`, plus `INCLUDE` / `LIB` / `LIBPATH` / `WindowsSdkDir` / `VCToolsInstallDir`) is exported to the **current cmd session only** — `vcvarsall.bat` does not (and should not) write these into the persistent user `PATH`. You therefore need to re-run `.\prepare.bat` in **every new cmd window** before invoking `init.bat` or building the loader. Subsequent runs are near-instant since no installation work is repeated.

> **Fast path (idempotent re-runs):** Building protobuf takes 5–15 minutes. To make repeated runs cheap, both `init.sh` and `init.bat` short-circuit and exit immediately when `third_party/_submodules/protobuf/.build/_install` already contains a valid `protobuf-config.cmake` (the marker that the previous build finished). This means:
> - Re-running `init.sh` / `init.bat` after a successful first run is a no-op (a second or two).
> - CI workflows cache `.build/_install` (see `.github/workflows/testing-cpp.yml`) and the fast path then turns the "build protobuf" step into a near-instant cache restore.
> - To force a clean rebuild (e.g. after changing protobuf flags or switching `PROTOBUF_REF` to a version whose previously-installed artefacts are still around), set `FORCE_REBUILD_PROTOBUF=1`:
>   ```sh
>   FORCE_REBUILD_PROTOBUF=1 bash init.sh         # macOS / Linux
>   ```
>   ```bat
>   set FORCE_REBUILD_PROTOBUF=1 && .\init.bat    :: Windows (cmd)
>   ```
>   Or simply delete `third_party/_submodules/protobuf/.build/` before rerunning.

### References

- [Chocolatey](https://chocolatey.org/)
- [CMake 3.31.8](https://github.com/Kitware/CMake/releases/tag/v3.31.8)
- [Ninja](https://ninja-build.org/)
- [Visual Studio 2022](https://visualstudio.microsoft.com/downloads/)
- [Use the Microsoft C++ Build Tools from the command line](https://learn.microsoft.com/en-us/cpp/build/building-on-the-command-line?view=msvc-170)
- [buf CLI](https://buf.build/docs/cli/)

## C++

### Dev at Linux

- Change dir: `cd test/cpp-tableau-loader`
- Generate protoconf: `PATH=../../third_party/_submodules/protobuf/.build/_install/bin:$PATH buf generate ..`
- CMake (the project's `CMakeLists.txt` defaults `CMAKE_BUILD_TYPE` to `Release` for single-config generators when unset, so `-DCMAKE_BUILD_TYPE=...` is omitted below):
  - C++17: `cmake -S . -B build`
  - C++20: `cmake -S . -B build -DCMAKE_CXX_STANDARD=20`
  - clang: `cmake -S . -B build -DCMAKE_CXX_COMPILER=clang++`
- Build: `cmake --build build --parallel`
- Test: `ctest --test-dir build --output-on-failure`

### Dev at Windows

> **Important:** CMake with Ninja requires MSVC environment variables (`cl.exe`, `INCLUDE`, `LIB`, etc.) to be active. Run `.\prepare.bat` from the **loader** root in the **same cmd session** (use **cmd**, not PowerShell — `prepare.bat` exports vars via `endlocal & set ...` which only works for a cmd parent process) before switching to the test directory. Opening a new terminal window will lose these variables.
>
> **Build type:** The protobuf submodule is built as **Release** (`/MT`) by `init.bat`. To avoid LNK2038 `_ITERATOR_DEBUG_LEVEL` / `RuntimeLibrary` CRT-mismatch errors, the loader must also be built as Release. The `CMakeLists.txt` defaults `CMAKE_BUILD_TYPE` to `Release` when it is unset, so the commands below work out of the box.

- Initialize MSVC environment (from loader root): `.\prepare.bat`
- Change dir: `cd test\cpp-tableau-loader`, or change directory with Drive, e.g.: `cd /D D:\GitHub\loader\test\cpp-tableau-loader`
- Generate protoconf:
  - cmd: `cmd /C "set PATH=..\..\third_party\_submodules\protobuf\.build\_install\bin;%PATH% && buf generate .."`
  - PowerShell: `$env:PATH = "..\..\third_party\_submodules\protobuf\.build\_install\bin;" + $env:PATH; buf generate ..`
- CMake (Ninja is single-config; the project's `CMakeLists.txt` defaults `CMAKE_BUILD_TYPE` to `Release` when unset, so `-DCMAKE_BUILD_TYPE=...` is omitted below):
  - C++17: `cmake -S . -B build -G "Ninja"`
  - C++20: `cmake -S . -B build -G "Ninja" -DCMAKE_CXX_STANDARD=20`
- Build: `cmake --build build --parallel`
- Test: `ctest --test-dir build --output-on-failure`

> **Note:** Tests are written with [GoogleTest](https://github.com/google/googletest), pulled in via CMake `FetchContent` (no manual installation needed).

### References

- [Protocol Buffers C++ Installation](https://github.com/protocolbuffers/protobuf/tree/master/src)
- [Protocol Buffers C++ Reference](https://protobuf.dev/reference/cpp/)

## Go

- Install: **go1.24** or above
- Change dir: `cd test/go-tableau-loader`
- Generate protoconf: `buf generate ..`
- Test: `go test ./...`

### References

- [Protocol Buffers Go Reference](https://protobuf.dev/reference/go/)

## C#

### Requirements

- Unity 2022.3 LTS (C# 9)
- dotnet-sdk-8.0

### Test

- Install: **dotnet-sdk-8.0**
- Change dir: `cd test/csharp-tableau-loader`
- Generate protoconf:
  - macOS / Linux: `PATH=../../third_party/_submodules/protobuf/.build/_install/bin:$PATH buf generate ..`
  - Windows (cmd): `cmd /C "set PATH=..\..\third_party\_submodules\protobuf\.build\_install\bin;%PATH% && buf generate .."`
  - Windows (PowerShell): `$env:PATH = "..\..\third_party\_submodules\protobuf\.build\_install\bin;" + $env:PATH; buf generate ..`
- Test: `dotnet test`

> **Note:** Tests are written with [xUnit](https://xunit.net/).

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
