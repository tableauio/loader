# Loader

The official config loader for [Tableau](https://github.com/tableauio/tableau).

## Prerequisites

> TODO: [devcontainer](https://code.visualstudio.com/docs/devcontainers/containers)

- C++ standard: at least C++17
- A working **`protoc` + `libprotobuf`** toolchain on your machine. The same
  protobuf release **must** provide both: protobuf v22+ enforces a strict
  gencode/runtime version check via `PROTOBUF_VERSION` in the generated
  headers, so a mismatched `protoc` and `libprotobuf` will fail to link.

### Install protobuf

Pick whichever channel fits your platform; loader does not bundle protobuf.

- **vcpkg (recommended, cross-platform):**
  ```sh
  git clone https://github.com/microsoft/vcpkg.git ~/vcpkg
  ~/vcpkg/bootstrap-vcpkg.sh                       # macOS / Linux
  # .\vcpkg\bootstrap-vcpkg.bat                    # Windows
  ~/vcpkg/vcpkg install protobuf                   # Linux:   x64-linux
  # ~/vcpkg/vcpkg install protobuf:x64-osx         # macOS
  # .\vcpkg\vcpkg install protobuf:x64-windows-static  # Windows (matches loader's static CRT)
  ```
  Pin to the legacy v3 line if you need it: append `--x-version=3.21.12`.
  Then point CMake at vcpkg with `-DCMAKE_TOOLCHAIN_FILE=<vcpkg-root>/scripts/buildsystems/vcpkg.cmake`.

- **Linux (system package):**
  ```sh
  sudo apt-get install -y protobuf-compiler libprotobuf-dev   # Debian / Ubuntu
  sudo dnf install -y protobuf-compiler protobuf-devel        # Fedora / RHEL 8+ / CentOS Stream / Rocky / Alma
  sudo yum install -y epel-release \
      && sudo yum install -y protobuf-compiler protobuf-devel # CentOS 7 (via EPEL)
  ```
  > Distro packages can lag well behind upstream (e.g. CentOS 7 ships protobuf 2.5; RHEL/Rocky 8 ships 3.x). If you need protobuf v22+ (or any specific version), prefer **vcpkg** above or **build from source**.

- **macOS (Homebrew):**
  ```sh
  brew install protobuf
  ```

- **From source:** see [Protocol Buffers C++ Installation](https://github.com/protocolbuffers/protobuf/tree/master/src).
  After installing, point CMake at it with `-DCMAKE_PREFIX_PATH=/path/to/protobuf-install`
  (or `-DProtobuf_ROOT=...`).

### Windows: bootstrap the rest of the toolchain

Run `prepare.bat` **as Administrator** to install everything you need on a
fresh Windows machine: [Chocolatey](https://chocolatey.org/),
[CMake](https://github.com/Kitware/CMake/releases),
[Ninja](https://ninja-build.org/), MSVC build tools, [buf](https://buf.build/),
**vcpkg**, and `protobuf:x64-windows-static`. It also activates the MSVC
compiler environment for the current cmd session.

```bat
.\prepare.bat
```

> ⚠️ **Admin required:** This script uses Chocolatey and MSI installers that write to system-protected directories (`C:\ProgramData`, `C:\Program Files`). Right-click Command Prompt → **Run as administrator**, then execute the script.
>
> Preview what the script would do without making any changes:
> ```bat
> .\prepare.bat --dry-run
> ```
>
> Override the protobuf vcpkg port version (e.g. for the legacy v3 ABI):
> ```bat
> set PROTOBUF_VCPKG_VERSION=3.21.12 && .\prepare.bat
> ```

> **Note:** The **installation** part of `prepare.bat` only runs once per machine — it detects already-installed tools (Chocolatey, Ninja, CMake, MSVC Build Tools, buf, vcpkg, protobuf) and skips them, so no manual installation is required.
>
> However, the MSVC compiler environment (`cl.exe` on `PATH`, plus `INCLUDE` / `LIB` / `LIBPATH` / `WindowsSdkDir` / `VCToolsInstallDir`) is exported to the **current cmd session only** — `vcvarsall.bat` does not (and should not) write these into the persistent user `PATH`. You therefore need to re-run `.\prepare.bat` in **every new cmd window** before building the loader. Subsequent runs are near-instant since no installation work is repeated.

### References

- [Chocolatey](https://chocolatey.org/)
- [CMake 3.31.8](https://github.com/Kitware/CMake/releases/tag/v3.31.8)
- [Ninja](https://ninja-build.org/)
- [Visual Studio 2022](https://visualstudio.microsoft.com/downloads/)
- [Use the Microsoft C++ Build Tools from the command line](https://learn.microsoft.com/en-us/cpp/build/building-on-the-command-line?view=msvc-170)
- [vcpkg](https://github.com/microsoft/vcpkg)
- [buf CLI](https://buf.build/docs/cli/)

## C++

### Dev at Linux

- Change dir: `cd test/cpp-tableau-loader`
- Generate protoconf: `buf generate ..` (assumes `protoc` is on `PATH`; if you installed via vcpkg, prepend `<vcpkg-root>/installed/x64-linux/tools/protobuf` to `PATH`)
- CMake (system protobuf):
  - C++17: `cmake -S . -B build -DCMAKE_BUILD_TYPE=Debug`
  - C++20: `cmake -S . -B build -DCMAKE_BUILD_TYPE=Debug -DCMAKE_CXX_STANDARD=20`
  - clang: `cmake -S . -B build -DCMAKE_BUILD_TYPE=Debug -DCMAKE_CXX_COMPILER=clang++`
- CMake (vcpkg-provided protobuf):
  ```sh
  cmake -S . -B build -DCMAKE_BUILD_TYPE=Debug \
      -DCMAKE_TOOLCHAIN_FILE=<vcpkg-root>/scripts/buildsystems/vcpkg.cmake
  ```
- Build: `cmake --build build --parallel`
- Test: `ctest --test-dir build --output-on-failure`

### Dev at Windows

> **Important:** CMake with Ninja requires MSVC environment variables (`cl.exe`, `INCLUDE`, `LIB`, etc.) to be active. Run `.\prepare.bat` from the **loader** root in the **same cmd session** (use **cmd**, not PowerShell — `prepare.bat` exports vars via `endlocal & set ...` which only works for a cmd parent process) before switching to the test directory. Opening a new terminal window will lose these variables.
>
> **Build type:** vcpkg's `x64-windows-static` triplet (and our `prepare.bat`) builds protobuf as **Debug** with the static CRT (`/MTd`). To avoid LNK2038 `_ITERATOR_DEBUG_LEVEL` / `RuntimeLibrary` CRT-mismatch errors, the loader must also be built as Debug. `CMakeLists.txt` does not set a default, so always pass `-DCMAKE_BUILD_TYPE=Debug` explicitly — also required for multi-config generators (Visual Studio default = Debug, but stay explicit to match the protobuf you installed).

- Initialize MSVC environment (from loader root): `.\prepare.bat`
- Change dir: `cd test\cpp-tableau-loader`, or change directory with Drive, e.g.: `cd /D D:\GitHub\loader\test\cpp-tableau-loader`
- Generate protoconf: `buf generate ..` (the `prepare.bat` step above already puts the vcpkg-built `protoc.exe` on `PATH`)
- CMake (vcpkg-provided protobuf):
  - C++17: `cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Debug -DCMAKE_TOOLCHAIN_FILE=%VCPKG_ROOT%\scripts\buildsystems\vcpkg.cmake -DVCPKG_TARGET_TRIPLET=x64-windows-static`
  - C++20: append `-DCMAKE_CXX_STANDARD=20`
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
- Generate protoconf: `buf generate ..` (requires `protoc` on `PATH`; install protobuf as described in [Install protobuf](#install-protobuf))
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
