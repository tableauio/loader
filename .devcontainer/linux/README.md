# Dev Container

The recommended way to develop on `tableauio/loader`. One container, all
four target languages (C++17, Go, .NET, Node) plus protobuf via vcpkg,
pinned to the exact toolchain CI uses. All version pins live in
[`../shared/versions.env`](../shared/versions.env) — bumping any of them
is a one-line change consumed by this Dockerfile, `prepare.bat`, and
the CI workflows.

Use this variant on **every** host that can run Docker: Linux (amd64 +
arm64), macOS (Intel + Apple Silicon), and Windows + WSL2. For Windows
hosts that prefer bare-metal native dev (no Docker, no WSL2), see the
[`prepare.bat`](../../prepare.bat) bootstrap at the repo root.

## Prerequisites

- Docker Desktop (Windows / macOS) or Docker Engine (Linux)
- VS Code with the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)

## Open the container

```sh
code .                # in the repo root
```

In VS Code, run **Dev Containers: Reopen in Container** from the command
palette. First build is one-time ~25 minutes (vcpkg compiles protobuf
from source); subsequent reopens are near-instant.

When the container is ready, the integrated terminal prints a banner with
five toolchain versions. After that, every command from the per-language
sections of the repo root [`README.md`](../README.md) works as written —
no PATH dance, no extra cmake flags.

## Pin a different protobuf version

Daily dev runs against the `PROTOBUF_VERSION` set in
[`../shared/versions.env`](../shared/versions.env) (CI's "modern" matrix
entry). To rebuild against the legacy v3 line for one container only,
without editing the shared file:

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
2. Go (official tarball, multi-arch) — version from `versions.env`
3. buf (single-binary release, multi-arch) — version from `versions.env`
4. vcpkg pinned to `versions.env`'s `VCPKG_BASELINE_COMMIT`, protobuf
   installed via vcpkg manifest mode and asserted against the requested
   version
5. .NET SDK (Microsoft apt repo) — version from `versions.env`
6. Node.js LTS (NodeSource apt repo) — version from `versions.env`
7. `ENV CMAKE_PREFIX_PATH=/opt/vcpkg/active` so `find_package(Protobuf CONFIG)`
   resolves automatically

The architecture choice is detected from BuildKit's `TARGETARCH` and fed
into Go / buf / vcpkg triplet selection. Docker auto-selects the host
arch on build.

The `dockerfile` field in `devcontainer.json` is `Dockerfile`, but the
build `context` is the parent `.devcontainer/` directory so the
Dockerfile's `COPY shared/versions.env …` and `COPY shared/postcreate-banner.sh …`
lines resolve.

## Troubleshooting

### `buf generate` works but the C++ or C# build then fails with stale-codegen errors

If you're hitting errors like

- C++: `fatal error: google/protobuf/generated_message_table_driven.h: No such file or directory`
- C#: hundreds of `error CS0101: The namespace already contains a definition for ...`

your host workspace probably has generated files from a *different* protobuf
version (left over from a previous host toolchain — gitignored, so `git pull`
doesn't remove them). They shadow what the container's `protoc` produces.

Wipe and retry:

```sh
rm -rf test/cpp-tableau-loader/src/protoconf/tableau \
       test/cpp-tableau-loader/build \
       test/csharp-tableau-loader/protoconf \
       test/csharp-tableau-loader/{bin,obj}
```

Then re-run `buf generate ..` from the affected language's `test/<lang>-tableau-loader/`
directory. A fresh clone doesn't have these stale artefacts.

## Falling back

If you can't run Docker (corp policy, restricted machines, etc.) the
existing manual setup paths in the [repo README](../../README.md) —
Windows `prepare.bat`, per-language `Install protobuf` instructions —
still work. The devcontainer is the recommended path; the rest is the
supported fallback.

For a Windows host that prefers bare-metal native dev (no Docker, no
WSL2), see [`prepare.bat`](../../prepare.bat) at the repo root — it
bootstraps the C++ toolchain (MSVC, CMake, Ninja, vcpkg+protobuf, buf).
You'll additionally need Go, .NET SDK, and Node.js, which one-line winget
installs cover (see the repo root [README](../../README.md)). For macOS
hosts where you'd rather not run a Linux container at all, see the
[macOS notes](../macos/).
