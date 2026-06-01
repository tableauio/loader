# Dev Container for tableauio/loader — Design

**Status:** Approved through brainstorming. Implementation plan to follow via `writing-plans`.
**Date:** 2026-05-29
**Scope:** Add a Dev Container that pins the full multi-language toolchain
(C++17, Go 1.24, .NET 8, Node 20, buf 1.67.0, protobuf 6.33.4 via vcpkg) so
contributors on any host OS get a reproducible build environment that mirrors
CI exactly.

## Goals

1. **One-command setup** on any host (Windows, macOS, Linux) — "Reopen in Container" replaces the existing per-OS, per-language manual setup as the *recommended* path.
2. **Reproducibility** — protobuf, buf, Go, .NET, Node, and the vcpkg checkout are pinned to the exact versions / commit SHA used by CI (`testing-cpp.yml`, `testing-go.yml`, `testing-csharp.yml`).
3. **Multi-arch native** — Apple Silicon contributors build natively as arm64 (no QEMU emulation); amd64 contributors build natively as amd64. One Dockerfile, no buildx publish step required.
4. **Pinnable protobuf version** — daily dev runs against modern (6.33.4) by default; legacy-v3 (3.21.12) is reachable by setting one host env var, with no second Dockerfile.
5. **Daily commands stay unchanged** — every shell snippet currently in README's per-language sections (`buf generate ..`, `cmake -S . -B build -DCMAKE_BUILD_TYPE=Debug`, `go test ./...`, `dotnet test`) works inside the container without flag tweaks.

## Non-goals

- **Prebuilt-and-pushed image on ghcr.io.** Out of scope for v1; revisit only if first-run latency (~25 min vcpkg compile) becomes a real complaint. Adding it later is a CI-only change with no Dockerfile churn.
- **CI running inside the devcontainer.** CI keeps `lukka/run-vcpkg` for cached vcpkg installs; rebuilding the devcontainer image per matrix entry would be strictly slower with no reproducibility win.
- **Unity-side C# workflow.** Unity Editor doesn't run in a Linux container; the container covers `.NET 8 + xUnit` (which is what `test/csharp-tableau-loader/` exercises) only.
- **Replacing `prepare.bat` or per-language manual setup.** Both stay as fallback paths for contributors who can't run Docker (corp policy, restricted machines).

## File layout

Three new files, one directory; nothing existing moves.

```
.devcontainer/
├── Dockerfile          # ~95 lines, single stage, multi-arch, multi-language
├── devcontainer.json   # ~30 lines
└── README.md           # 1-pager: prerequisites, how to enter, host caveats
```

## Architecture

### Image base

`mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04` — Microsoft's officially-maintained Dev Container base image. Provides:

- gcc-13, glibc 2.39, cmake, ninja, git, sudo
- Non-root `vscode` user (uid/gid 1000) with passwordless sudo
- Multi-arch: pulls amd64 or arm64 automatically based on host
- Dev Containers shell hooks (`postCreateCommand` etc.)

### Toolchain layers (Dockerfile, ordered for cache friendliness)

Each layer is a single `RUN` block, version-pinned, named in a comment block. Order goes least-likely-to-change to most-likely-to-change so editing the latter doesn't invalidate the former.

1. **Architecture detection.** `ARG TARGETARCH` (BuildKit auto-populates from build host). One `RUN` writes the resolved arch-dependent values to `/opt/buildargs.env` for downstream layers:

   | TARGETARCH | GO_ARCH | BUF_ARCH | VCPKG_TRIPLET |
   |---|---|---|---|
   | amd64 | amd64 | x86_64 | x64-linux |
   | arm64 | arm64 | aarch64 | arm64-linux |

   Unknown arches fail the build with a clear message.

2. **Go 1.24.0** — download official tarball for `${GO_ARCH}` to `/usr/local/go`. PATH is exposed via `ENV PATH=/usr/local/go/bin:/home/vscode/go/bin:${PATH}` (not `/etc/profile.d/`) so that non-interactive shells like the `postCreateCommand` and any `RUN` in downstream Dockerfiles see Go without sourcing profile. `/home/vscode/go/bin` is included so `go install`-placed binaries land on PATH automatically.

3. **buf 1.67.0** — single binary download `buf-Linux-${BUF_ARCH}` to `/usr/local/bin/buf`.

4. **vcpkg + protobuf via manifest mode** *(the heavy layer; ~25 min on first build).*
   - Pinned commit: `VCPKG_BASELINE_COMMIT=dc8d75cfc3281b8e2a4ed8ee4163c891190df932` (lock-step with `prepare.bat` and `testing-cpp.yml`'s `VCPKG_COMMIT`).
   - Pinned port: `PROTOBUF_VERSION=6.33.4` (a `Dockerfile ARG`, override-friendly — see "Pinnable protobuf version" below).
   - Cloned into `/opt/vcpkg`, checked out to the pinned commit, bootstrapped with `-disableMetrics`.
   - A small `vcpkg.json` is rendered into `/opt/vcpkg-manifest/` carrying `dependencies: ["protobuf"]`, `overrides: [{name: protobuf, version: ${PROTOBUF_VERSION}}]`, `builtin-baseline: ${VCPKG_BASELINE_COMMIT}`.
   - `vcpkg install --triplet=${VCPKG_TRIPLET} --x-install-root=/opt/vcpkg-manifest/vcpkg_installed` runs from that directory.
   - **Post-install assertion:** the same `dir-bin/findstr` pattern from `prepare.bat`, ported to bash — fail the build if the resolved port version doesn't have `${PROTOBUF_VERSION}` as a prefix. This is the safety net against future vcpkg-resolution regressions silently producing the wrong version.
   - `ln -s /opt/vcpkg-manifest/vcpkg_installed/${VCPKG_TRIPLET} /opt/vcpkg/active` so the architecture-dependent path collapses behind a stable symlink.
   - `ln -s /opt/vcpkg/active/tools/protobuf/protoc /usr/local/bin/protoc` so `buf generate` finds protoc with no PATH dance.

5. **.NET SDK 8.0 + Node.js 20 LTS** — Microsoft's `packages-microsoft-prod.deb` and NodeSource's `setup_20.x` apt repos; `apt-get install -y dotnet-sdk-8.0 nodejs`; clean `/var/lib/apt/lists` to keep the layer trim.

6. **Final environment.**
   ```dockerfile
   ENV CMAKE_PREFIX_PATH=/opt/vcpkg/active
   ENV VCPKG_ROOT=/opt/vcpkg
   ```
   Stable paths, no triplet leakage.

### `devcontainer.json`

```jsonc
{
  "name": "tableauio/loader",
  "build": {
    "dockerfile": "Dockerfile",
    "args": {
      "PROTOBUF_VERSION": "${localEnv:LOADER_PROTOBUF_VERSION:6.33.4}"
    }
  },

  // Keep Go's module cache across container rebuilds.
  "mounts": [
    "source=loader-go-mod,target=/home/vscode/go,type=volume"
  ],

  "remoteUser": "vscode",
  "workspaceFolder": "/workspaces/loader",

  "customizations": {
    "vscode": {
      "extensions": [
        "golang.go",
        "ms-vscode.cmake-tools",
        "ms-vscode.cpptools",
        "ms-dotnettools.csharp",
        "bufbuild.vscode-buf",
        "zxh404.vscode-proto3"
      ],
      "settings": {
        "go.toolsManagement.autoUpdate": false,
        "cmake.configureOnOpen": false
      }
    }
  },

  "postCreateCommand": "printf 'tableauio/loader devcontainer ready.\\n  go: %s\\n  buf: %s\\n  protoc: %s\\n  dotnet: %s\\n  node: %s\\n' \"$(go version | cut -d' ' -f3)\" \"$(buf --version)\" \"$(protoc --version)\" \"$(dotnet --version)\" \"$(node --version)\""
}
```

Three intentional choices:

1. **`${localEnv:LOADER_PROTOBUF_VERSION:6.33.4}`** — host env var picked up at container-build time. Workflow: `LOADER_PROTOBUF_VERSION=3.21.12 code .` → Reopen in Container → legacy-v3 image. No second devcontainer.json.
2. **Named volume `loader-go-mod` for `~/go`** — Go module cache persists across rebuilds. Workspace itself uses VS Code's default bind-mount (edits sync to host).
3. **`go.toolsManagement.autoUpdate: false`, `cmake.configureOnOpen: false`** — stops both extensions from auto-running their setup actions on first open, which would fight manual `cmake -S . -B build` invocations and trigger a 2-minute background `gopls` install.

The `postCreateCommand` is pure echo — it prints the five tool versions so the contributor immediately knows the container is healthy. No installs, no conditionals.

## Integration with existing flows

### README change (additive, surgical)

A new subsection at the top of `Prerequisites`, **above** "Install protobuf":

> ### Recommended: Dev Container (any host OS)
>
> The fastest way to get a reproducible build environment is to open the
> repo in VS Code and choose **Reopen in Container**. The devcontainer
> under `.devcontainer/` has everything pinned to the exact versions CI
> uses (Go 1.24, buf 1.67.0, protobuf 6.33.4 via vcpkg, .NET 8.0,
> Node 20). First container build is one-time ~25 minutes (vcpkg
> compiles protobuf from source); subsequent reopens are instant.
>
> After the container starts you can skip the per-language setup below
> and jump straight to **C++** / **Go** / **C#** / **TypeScript**.
>
> Requirements: Docker Desktop (Windows + macOS) or Docker Engine (Linux),
> and the VS Code "Dev Containers" extension. See `.devcontainer/README.md`
> for the longer how-to.

The existing `Windows: bootstrap…` block and per-language `Install protobuf` block both stay as written. Each gains a one-line lead-in: *"If you can't or don't want to use the devcontainer (corp Docker policy, etc.), follow the steps below."*

### Container-side env so daily commands stay flag-free

The Dockerfile's final `ENV CMAKE_PREFIX_PATH=/opt/vcpkg/active` is the *only* mechanism needed to make the existing "Dev at Linux → CMake (system protobuf)" recipe work in the container — `find_package(Protobuf CONFIG)` resolves to vcpkg's pinned protobuf without any toolchain-file flag. The contributor types the same commands they would on a host with system-installed protobuf; they happen to land on vcpkg's pinned 6.33.4. None of the four `buf.gen.yaml` files change.

### Things explicitly NOT changed by this design

- `prepare.bat` (already correct; stays as Windows-host fallback)
- Any `buf.gen.yaml`
- `test/cpp-tableau-loader/CMakeLists.txt`
- `.github/workflows/*.yml` (CI keeps `lukka/run-vcpkg` directly)

## Verification matrix

| Host | Container arch | First-run cost | Daily-cmd cost | Notes |
|---|---|---|---|---|
| Linux x86 | amd64 native | ~25 min build | bind-mount IO native | reference path |
| macOS Apple Silicon | **arm64 native** | ~25 min build | bind-mount IO native | no Rosetta tax |
| macOS Intel | amd64 native | ~25 min build | bind-mount IO native | |
| Windows + WSL2 | amd64 native | ~25 min build | bind-mount IO good if workspace is under WSL2 (`\\wsl.localhost\Ubuntu\…`), poor under `/mnt/c/...` | flagged in `.devcontainer/README.md` |

**Acceptance gates:**

1. `docker build .devcontainer/` succeeds clean on amd64 and arm64 hosts (gate 1: image actually builds).
2. Container start runs `postCreateCommand` and prints all five tool versions (gate 2: toolchain is wired correctly).
3. Inside the container, all four E2E paths from the README run green:
   - `cd test/go-tableau-loader && buf generate .. && go test ./...`
   - `cd test/cpp-tableau-loader && buf generate .. && cmake -S . -B build -DCMAKE_BUILD_TYPE=Debug && cmake --build build && ctest --test-dir build --output-on-failure`
   - `cd test/csharp-tableau-loader && buf generate .. && dotnet test`
   - `cd _lab/ts && npm install && npm run generate && npm run test` *(stretch — TS lab isn't in CI)*
4. `LOADER_PROTOBUF_VERSION=3.21.12 code .` → Reopen in Container → all C++ test paths still green (gate 3: protobuf version pinning works end-to-end).

## Trade-offs and explicit deferrals

- **First-run latency.** ~25 min on cold build (vcpkg compiles protobuf from source). Consequences: every change to the protobuf-installation `RUN` invalidates that layer for everyone who pulls the change. Mitigated by ordering it late in the layer chain (Go/buf changes don't trigger it). If pain becomes acute, prebuild-on-ghcr.io is the escape hatch.
- **No multi-arch publishing.** The Dockerfile is arch-agnostic, but `docker build` only builds the host arch. Apple Silicon contributors get arm64 natively because Docker Desktop's BuildKit picks `linux/arm64`. We do **not** run `docker buildx build --platform linux/amd64,linux/arm64 --push` anywhere. If we ever publish to ghcr.io, that's the moment to add buildx.
- **Modern protobuf default.** Daily dev runs against 6.33.4. Legacy-v3 contributors must rebuild the container with `LOADER_PROTOBUF_VERSION=3.21.12`. Acceptable because (a) CI catches legacy-v3 regressions automatically and (b) the container rebuild is incremental — only the vcpkg layer reruns.
- **Named volume for `~/go`.** Persists the Go module cache across rebuilds (~30s saved per first `go test` post-rebuild). If a contributor wants pure isolation, they `docker volume rm loader-go-mod`.

## Implementation outline (for the writing-plans step that follows)

1. Add `.devcontainer/Dockerfile` (multi-arch, manifest-mode vcpkg, version assertion).
2. Add `.devcontainer/devcontainer.json` (build args, named volume, extensions, postCreate banner).
3. Add `.devcontainer/README.md` (Docker prereqs, host-OS caveats, the `LOADER_PROTOBUF_VERSION` knob).
4. Update repo-root `README.md` Prerequisites section: add the new "Recommended: Dev Container" subsection; lead the existing Windows / per-language blocks with the "If you can't / don't want to use the devcontainer…" line.
5. Verification: `docker build` locally for amd64; container start produces the banner; all four E2E test commands green; `LOADER_PROTOBUF_VERSION=3.21.12` rebuild produces a working legacy-v3 image.
