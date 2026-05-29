# Loader

The official config loader for [Tableau](https://github.com/tableauio/tableau).

## Prerequisites

> TODO: [devcontainer](https://code.visualstudio.com/docs/devcontainers/containers)

- C++ standard: at least C++17
- A working **`protoc` + `libprotobuf`** toolchain on your machine. The same
  protobuf release **must** provide both: protobuf v22+ enforces a strict
  gencode/runtime version check via `PROTOBUF_VERSION` in the generated
  headers, so a mismatched `protoc` and `libprotobuf` will fail to link.

> **Migrating from the bundled-protobuf layout?** Loader used to vendor
> protobuf as a git submodule under `third_party/_submodules/protobuf` plus
> an `init.sh` / `init.bat` build pipeline. Both are gone. If you've checked
> the repo out before this change, clean up the orphan worktree and submodule
> metadata before building:
>
> ```sh
> git submodule deinit -f third_party/_submodules/protobuf
> rm -rf third_party/_submodules/protobuf .git/modules/third_party/_submodules/protobuf
> ```
>
> Then install protobuf via one of the channels documented in
> [Install protobuf](#install-protobuf).

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
  This installs whatever protobuf version the vcpkg checkout's baseline ships
  (currently the 6.x line). To pin a specific version, use vcpkg **manifest
  mode**: drop a `vcpkg.json` in your build directory with a `builtin-baseline`
  + `overrides`, e.g.

  ```json
  {
    "name": "loader-build",
    "version": "0.1.0",
    "dependencies": ["protobuf"],
    "overrides": [{ "name": "protobuf", "version": "3.21.12" }],
    "builtin-baseline": "<recent-vcpkg-commit-sha>"
  }
  ```

  > **Note:** classic-mode `vcpkg install --x-version=...` is silently a no-op;
  > version pinning only works in manifest mode. See
  > `.github/workflows/testing-cpp.yml` for the exact pattern CI uses.

  Then put `protoc` on `PATH` (so `buf generate` works) and pass
  `-DCMAKE_TOOLCHAIN_FILE=<vcpkg-root>/scripts/buildsystems/vcpkg.cmake` to
  CMake. See [Dev at Linux](#dev-at-linux) / [Dev at Windows](#dev-at-windows)
  for the exact commands.

- **Linux (system package):**
  ```sh
  sudo apt-get install -y protobuf-compiler libprotobuf-dev   # Debian / Ubuntu
  ```
  > **Avoid `dnf` / `yum` on RHEL-family distros.** The `protobuf-devel`
  > shipped by Fedora / RHEL / TencentOS repos is typically stuck on
  > protobuf **3.5.x**, which is far behind what loader expects and predates
  > the v22 / Abseil split. Use vcpkg or build from source instead.

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
> Setting this switches the script to vcpkg **manifest mode** — the only mode
> in which the version pin actually takes effect. The install root moves from
> `%VCPKG_ROOT%\installed\x64-windows-static\` to a manifest dir under
> `%LOCALAPPDATA%\loader\vcpkg-manifest\vcpkg_installed\`, and `prepare.bat`
> exports its path as `%VCPKG_INSTALLED_DIR%`. Your downstream CMake invocation
> must then add `-DVCPKG_INSTALLED_DIR=%VCPKG_INSTALLED_DIR%` and
> `-DVCPKG_MANIFEST_INSTALL=OFF` (see [Dev at Windows](#dev-at-windows)).

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
- CMake (vcpkg-provided protobuf, classic mode — default):
  - C++17: `cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Debug -DCMAKE_TOOLCHAIN_FILE=%VCPKG_ROOT%\scripts\buildsystems\vcpkg.cmake -DVCPKG_TARGET_TRIPLET=x64-windows-static`
  - C++20: append `-DCMAKE_CXX_STANDARD=20`
- CMake (vcpkg manifest mode — only when you ran `prepare.bat` with `PROTOBUF_VCPKG_VERSION` set):
  - C++17: `cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Debug -DCMAKE_TOOLCHAIN_FILE=%VCPKG_ROOT%\scripts\buildsystems\vcpkg.cmake -DVCPKG_TARGET_TRIPLET=x64-windows-static -DVCPKG_INSTALLED_DIR="%VCPKG_INSTALLED_DIR%" -DVCPKG_MANIFEST_INSTALL=OFF`
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
