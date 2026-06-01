# Dev Container — Windows variant

A **Windows-container** image (windows/amd64 only) for Windows hosts that
want a native MSVC build environment inside the container, without
WSL2. The Linux container under [`../linux/`](../linux/) is the
recommended path for almost everyone — use this variant only when you
specifically need Windows-native tooling.

## When to use this variant

| Goal | Use |
| --- | --- |
| Develop on Windows + WSL2 | `../linux/` |
| Develop on Windows without WSL2, native MSVC | **this** |
| Develop on macOS / Linux | `../linux/` |
| Develop on arm64 Windows (Surface Pro X) | `../linux/` (linux/arm64 native) |
| Match CI's `windows-latest` toolchain locally | **this** |

## Prerequisites

- **Windows 10/11 Pro or Enterprise.** Home edition cannot run Windows
  containers (Hyper-V missing).
- **Docker Desktop in Windows-containers mode.** Right-click the tray
  icon → *Switch to Windows containers*. This is a per-host toggle:
  Linux and Windows containers cannot run simultaneously.
- **Disk space.** ~12 GB for MSVC Build Tools alone. Total image is
  ~15 GB.
- **VS Code** with the Dev Containers extension.

## Open the container

```sh
code .                # in the repo root
```

In VS Code, run **Dev Containers: Reopen in Container** from the command
palette. VS Code shows a picker; choose **tableauio/loader (windows)**.

First build is one-time, ~45 minutes (vcpkg compiles protobuf from
source under MSVC). Subsequent reopens are near-instant.

## Pin a different protobuf version

```cmd
set LOADER_PROTOBUF_VERSION=3.21.12
code .
```

…then **Dev Containers: Rebuild Container**. Same knob as the Linux
variant, same default sourced from
[`../shared/versions.env`](../shared/versions.env).

## Architecture

Single Dockerfile based on `mcr.microsoft.com/windows/servercore:ltsc2022`:

1. Imports `versions.env` into machine-wide env so each `RUN` sees `$env:KEY`.
2. Chocolatey bootstrap.
3. Git, Ninja, CMake (CMake version pinned by `versions.env`).
4. Visual Studio 2022 Build Tools — VC++ workload.
5. Go SDK — official MSI, version from `versions.env`.
6. buf CLI — single-binary release, version from `versions.env`.
7. vcpkg (pinned to `VCPKG_BASELINE_COMMIT`) + protobuf via manifest mode
   with the `x64-windows-static` triplet — same triplet `prepare.bat`
   uses on bare metal, so a developer moving between native and
   container builds avoids the LNK2038 `_ITERATOR_DEBUG_LEVEL` CRT-mismatch
   trap. Post-install assertion catches version drift.
8. .NET SDK + Node.js LTS via Chocolatey.
9. `ENV CMAKE_PREFIX_PATH=C:\vcpkg-manifest\vcpkg_installed\x64-windows-static`
   so `find_package(Protobuf CONFIG)` resolves automatically.

## Why is the image so large?

`visualstudio2022buildtools` with the VC++ workload is ~6 GB on disk; the
Windows base image is another ~5 GB. There is no nanoserver path because
vcpkg compiles protobuf from source under MSVC and that needs the full
Win32 environment.

## Why isolation=hyperv?

Process isolation requires the **host's** Windows build to be ≥ the
container image's Windows build. ltsc2022 is conservative enough to
work on most Windows 10 21H2+ and Windows 11 hosts under Hyper-V
isolation. If your host is Windows 11 22H2+ you can drop `runArgs`
or change to `--isolation=process` for slightly faster startup.

## Falling back

If the container build is too slow or too large for your machine, the
existing manual setup paths still work:

- Native Windows: [`prepare.bat`](../../prepare.bat) at the repo root.
  Same vcpkg baseline and triplet as this container.
- Linux container under WSL2: [`../linux/`](../linux/). Same toolchain
  versions, much smaller image.

## Limitations

- **windows/amd64 only.** No arm64 Windows base image exists. Use the
  Linux container on arm64 Windows hosts (it builds linux/arm64 natively
  under Docker's Linux-container engine).
- **Windows containers don't run on macOS or Linux.** Don't try to build
  this image on a non-Windows host; it won't work.
- **Heavier than the Linux variant.** Both build time and disk footprint
  are roughly 2× the Linux container. If you don't specifically need
  Windows-native tooling, use the Linux variant.
