# Dev Container

The recommended way to develop on `tableauio/loader`. One container, all
four target languages (C++17, Go 1.24, .NET 8, Node 20) plus protobuf
6.33.4 via vcpkg, pinned to the exact toolchain CI uses.

## Prerequisites

- Docker Desktop (Windows / macOS) or Docker Engine (Linux)
- VS Code with the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)

## Open the container

```sh
code .                # in the repo root
```

In VS Code, run **Dev Containers: Reopen in Container** from the command
palette. First build is one-time ~25 minutes (vcpkg compiles protobuf
6.33.4 from source); subsequent reopens are near-instant.

When the container is ready, the integrated terminal prints a banner with
five toolchain versions. After that, every command from the per-language
sections of the repo root [`README.md`](../README.md) works as written —
no PATH dance, no extra cmake flags.

## Pin a different protobuf version

Daily dev runs against protobuf 6.33.4 (CI's "modern" matrix entry). To
rebuild against the legacy v3 line:

```sh
LOADER_PROTOBUF_VERSION=3.21.12 code .
```

…then **Dev Containers: Reopen in Container** (or **Rebuild Container**
if the container is already running). The vcpkg layer rebuilds with the
manifest pinning protobuf 3.21.12; everything else is reused from the
cache.

## Host-OS caveats

- **Windows.** WSL2 backend required. **Check the workspace out under
  WSL2** (e.g. `\\wsl.localhost\Ubuntu\home\<user>\loader`) — not under
  `/mnt/c/...` — for good bind-mount performance. Files under `/mnt/c/`
  work but file-watching and large `cmake --build` operations are 5–10×
  slower.

- **Apple Silicon.** Docker builds the container natively as arm64. No
  Rosetta or QEMU emulation. Confirm with `docker info | grep Architecture`
  → expect `linux/arm64`.

- **Linux (native Docker Engine).** No special configuration.

## Architecture

Single-stage Dockerfile based on
`mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04`, with these layers:

1. Architecture detection (`TARGETARCH` → Go arch, buf arch, vcpkg triplet)
2. Go 1.24.0 (official tarball, multi-arch)
3. buf 1.67.0 (single-binary release, multi-arch)
4. vcpkg pinned to `dc8d75c…df932`, protobuf installed via vcpkg manifest
   mode and asserted against the requested version
5. .NET SDK 8.0 (Microsoft apt repo)
6. Node.js 20 LTS (NodeSource apt repo)
7. `ENV CMAKE_PREFIX_PATH=/opt/vcpkg/active` so `find_package(Protobuf CONFIG)`
   resolves automatically

The architecture choice is detected from BuildKit's `TARGETARCH` and fed
into Go / buf / vcpkg triplet selection. Docker auto-selects the host
arch on build.

## Falling back

If you can't run Docker (corp policy, restricted machines, etc.) the
existing manual setup paths in the [repo README](../README.md) — Windows
`prepare.bat`, per-language `Install protobuf` instructions — still work.
The devcontainer is the recommended path; the rest is the supported
fallback.
