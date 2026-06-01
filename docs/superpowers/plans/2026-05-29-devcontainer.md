# Dev Container Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Dev Container under `.devcontainer/` so contributors on any host (Windows/macOS/Linux) get a one-command, reproducible build environment that mirrors CI's exact toolchain (C++17, Go 1.24, .NET 8, Node 20, buf 1.67.0, protobuf 6.33.4 via vcpkg).

**Architecture:** Single-stage `Dockerfile` based on `mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04`, layered Go → buf → vcpkg/protobuf → .NET/Node. Multi-arch via BuildKit `TARGETARCH` (amd64 + arm64 native, no QEMU). Protobuf version pinnable via the `LOADER_PROTOBUF_VERSION` host env var, flowing through `devcontainer.json` build args into a vcpkg manifest-mode install with a post-install version assertion.

**Tech Stack:** Docker (BuildKit), `mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04`, vcpkg manifest mode, VS Code Dev Containers spec.

**Spec:** [`docs/superpowers/specs/2026-05-29-devcontainer-design.md`](./../specs/2026-05-29-devcontainer-design.md)

---

## Files Created / Modified

| Path | Action | Purpose |
|---|---|---|
| `.devcontainer/Dockerfile` | Create | Multi-arch, multi-language toolchain image |
| `.devcontainer/devcontainer.json` | Create | VS Code config: build args, named volume, extensions, banner |
| `.devcontainer/README.md` | Create | Host prerequisites, how-to, host-OS caveats |
| `README.md` | Modify | Add "Recommended: Dev Container" subsection at top of Prerequisites; add "If you can't / don't want to use the devcontainer…" lead-in to existing Windows + per-language blocks |

CI workflows (`.github/workflows/*.yml`), `prepare.bat`, `buf.gen.yaml` files, and `CMakeLists.txt` are **not** touched.

---

## Conventions for this plan

- Each Dockerfile change is **one logical layer**, built and verified before the next is added. The "test" for a layer is `docker build` + a `docker run --rm` smoke check that the binary on PATH reports the expected version.
- Build target image tag: `loader-devcontainer:dev` (overwritten each build).
- Build context is `.devcontainer/` so all `docker build` commands use `docker build -t loader-devcontainer:dev .devcontainer/`.
- Smoke checks use `docker run --rm loader-devcontainer:dev <cmd>`.
- After Task 5 the image takes ~25 minutes to rebuild from scratch on first run because vcpkg compiles protobuf from source. Subsequent builds reuse layers and only re-run the layer you changed.

---

### Task 1: Stub Dockerfile with the base image only

**Files:**
- Create: `.devcontainer/Dockerfile`

- [ ] **Step 1: Create the file**

`.devcontainer/Dockerfile`:

```dockerfile
# syntax=docker/dockerfile:1.7
# tableauio/loader devcontainer
#
# Single-stage, multi-arch (amd64 + arm64) image bringing the full
# C++/Go/.NET/Node toolchain plus protobuf 6.33.4 (via vcpkg) at the
# exact versions CI uses. See docs/superpowers/specs/2026-05-29-devcontainer-design.md.

FROM mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04
```

- [ ] **Step 2: Build the base layer**

```sh
docker build -t loader-devcontainer:dev .devcontainer/
```

Expected: build completes in <1 min on a warm Docker, ending with
`=> => naming to docker.io/library/loader-devcontainer:dev`.

- [ ] **Step 3: Smoke-check the base image runs and gives us the vscode user**

```sh
docker run --rm loader-devcontainer:dev id
```

Expected output: `uid=0(root) gid=0(root) groups=0(root)` (RUN context defaults to root; the `vscode` user is set later via `devcontainer.json`'s `remoteUser`). Confirm with:

```sh
docker run --rm loader-devcontainer:dev id vscode
```

Expected: `uid=1000(vscode) gid=1000(vscode) groups=1000(vscode),...` — confirming the base image ships the `vscode` user.

- [ ] **Step 4: Commit**

```sh
git add .devcontainer/Dockerfile
git commit -m "$(cat <<'EOF'
feat(devcontainer): add base Dockerfile (Ubuntu 24.04 cpp image)

Bootstrap the devcontainer image with Microsoft's multi-arch
mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04 base. Subsequent
commits layer Go, buf, vcpkg/protobuf, .NET, and Node on top.

Refs: docs/superpowers/specs/2026-05-29-devcontainer-design.md
EOF
)"
```

---

### Task 2: Add architecture-detection layer

Resolves `TARGETARCH` (BuildKit auto-populates it from the host) into the per-arch values that downstream layers need: Go's tarball arch, buf's release-asset arch, and the vcpkg triplet. Writes them to `/opt/buildargs.env` so later `RUN` commands can `source` them — Dockerfile `ARG`s don't persist across `RUN` blocks the way ENV does, but a shell-readable file does.

**Files:**
- Modify: `.devcontainer/Dockerfile`

- [ ] **Step 1: Append the architecture detection block**

Append to `.devcontainer/Dockerfile`:

```dockerfile
# ---------------------------------------------------------------------------
# Architecture detection. BuildKit auto-populates TARGETARCH; we resolve it
# into per-arch download-name fragments (Go's tarball, buf's release asset,
# vcpkg triplet) and persist them to /opt/buildargs.env so later RUN layers
# can `source` them — Dockerfile ARGs don't survive across RUN boundaries.
# ---------------------------------------------------------------------------
ARG TARGETARCH
RUN <<EOF
set -eux
case "${TARGETARCH}" in
  amd64) GO_ARCH=amd64; BUF_ARCH=x86_64;  TRIPLET=x64-linux   ;;
  arm64) GO_ARCH=arm64; BUF_ARCH=aarch64; TRIPLET=arm64-linux ;;
  *)     echo "unsupported TARGETARCH: ${TARGETARCH}"; exit 1 ;;
esac
mkdir -p /opt
printf 'GO_ARCH=%s\nBUF_ARCH=%s\nVCPKG_TRIPLET=%s\n' \
  "${GO_ARCH}" "${BUF_ARCH}" "${TRIPLET}" > /opt/buildargs.env
EOF
```

- [ ] **Step 2: Build**

```sh
docker build -t loader-devcontainer:dev .devcontainer/
```

Expected: success in <1 min. The architecture-detection layer is just a `case` + `printf`.

- [ ] **Step 3: Smoke-check the resolved values**

```sh
docker run --rm loader-devcontainer:dev cat /opt/buildargs.env
```

Expected (on an amd64 host):
```
GO_ARCH=amd64
BUF_ARCH=x86_64
VCPKG_TRIPLET=x64-linux
```

(arm64 host would print `GO_ARCH=arm64`, `BUF_ARCH=aarch64`, `VCPKG_TRIPLET=arm64-linux`.)

- [ ] **Step 4: Commit**

```sh
git add .devcontainer/Dockerfile
git commit -m "$(cat <<'EOF'
feat(devcontainer): add architecture detection

Resolves TARGETARCH (amd64 or arm64) into per-arch values
(Go tarball arch, buf release-asset arch, vcpkg triplet) and
writes them to /opt/buildargs.env for downstream RUN layers
to source. Unknown arches fail the build.
EOF
)"
```

---

### Task 3: Add Go 1.24.0

**Files:**
- Modify: `.devcontainer/Dockerfile`

- [ ] **Step 1: Append the Go layer**

Append to `.devcontainer/Dockerfile`:

```dockerfile
# ---------------------------------------------------------------------------
# Go 1.24.0 — official tarball into /usr/local/go.
#
# PATH is set via ENV (not /etc/profile.d/) so non-interactive shells like
# the postCreateCommand and downstream RUNs see Go without sourcing profile.
# /home/vscode/go/bin lands `go install`-placed binaries on PATH automatically.
# ---------------------------------------------------------------------------
ARG GO_VERSION=1.24.0
RUN <<EOF
set -eux
. /opt/buildargs.env
curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" \
  | tar -C /usr/local -xz
EOF
ENV PATH=/usr/local/go/bin:/home/vscode/go/bin:${PATH}
```

- [ ] **Step 2: Build**

```sh
docker build -t loader-devcontainer:dev .devcontainer/
```

Expected: success in 1–3 min (downloads ~70 MB Go tarball).

- [ ] **Step 3: Smoke-check Go**

```sh
docker run --rm loader-devcontainer:dev go version
```

Expected (amd64 host): `go version go1.24.0 linux/amd64`.

- [ ] **Step 4: Smoke-check the PATH**

```sh
docker run --rm loader-devcontainer:dev sh -c 'echo "$PATH"'
```

Expected: starts with `/usr/local/go/bin:/home/vscode/go/bin:`.

- [ ] **Step 5: Commit**

```sh
git add .devcontainer/Dockerfile
git commit -m "$(cat <<'EOF'
feat(devcontainer): add Go 1.24.0

Install Go from the official multi-arch tarball into /usr/local/go.
PATH is exposed via ENV (not /etc/profile.d) so non-interactive shells
(postCreateCommand, downstream RUNs) see `go` without sourcing profile.
EOF
)"
```

---

### Task 4: Add buf 1.67.0

**Files:**
- Modify: `.devcontainer/Dockerfile`

- [ ] **Step 1: Append the buf layer**

Append to `.devcontainer/Dockerfile`:

```dockerfile
# ---------------------------------------------------------------------------
# buf 1.67.0 — single-binary release on PATH.
# ---------------------------------------------------------------------------
ARG BUF_VERSION=1.67.0
RUN <<EOF
set -eux
. /opt/buildargs.env
curl -fsSL -o /usr/local/bin/buf \
  "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-Linux-${BUF_ARCH}"
chmod +x /usr/local/bin/buf
EOF
```

- [ ] **Step 2: Build**

```sh
docker build -t loader-devcontainer:dev .devcontainer/
```

Expected: success in <1 min (single ~30 MB binary).

- [ ] **Step 3: Smoke-check buf**

```sh
docker run --rm loader-devcontainer:dev buf --version
```

Expected: `1.67.0`.

- [ ] **Step 4: Commit**

```sh
git add .devcontainer/Dockerfile
git commit -m "$(cat <<'EOF'
feat(devcontainer): add buf 1.67.0

Single-binary release into /usr/local/bin/buf. Pinned to the same
version testing-cpp.yml / testing-csharp.yml use.
EOF
)"
```

---

### Task 5: Add vcpkg + protobuf via manifest mode (the heavy layer)

This is the longest layer (~25 min on first build because vcpkg compiles protobuf from source). It also carries the post-install version assertion that protects against vcpkg's manifest-resolution silently picking a wrong port version. Subsequent builds that don't change this layer reuse the cache and complete instantly.

**Files:**
- Modify: `.devcontainer/Dockerfile`

- [ ] **Step 1: Append the vcpkg + protobuf layer**

Append to `.devcontainer/Dockerfile`:

```dockerfile
# ---------------------------------------------------------------------------
# vcpkg + protobuf via manifest mode.
#
# Pins:
#   VCPKG_BASELINE_COMMIT — same commit testing-cpp.yml's VCPKG_COMMIT and
#       prepare.bat's VCPKG_BASELINE_COMMIT use. Bumping vcpkg = bump all three.
#   PROTOBUF_VERSION    — defaults to the modern 6.33.4 line (CI's primary).
#       Override at build time:
#         docker build --build-arg PROTOBUF_VERSION=3.21.12 ...
#       devcontainer.json wires this to the LOADER_PROTOBUF_VERSION host env var.
#
# Why manifest mode: classic-mode `vcpkg install --x-version=...` is silently
# a no-op; only manifest mode + overrides actually pin the port version. The
# post-install assertion catches the case where vcpkg's resolution still picks
# a different port revision than requested.
# ---------------------------------------------------------------------------
ARG VCPKG_BASELINE_COMMIT=dc8d75cfc3281b8e2a4ed8ee4163c891190df932
ARG PROTOBUF_VERSION=6.33.4
ENV VCPKG_ROOT=/opt/vcpkg

RUN <<EOF
set -eux
. /opt/buildargs.env

# 1. Bring up vcpkg pinned to the baseline commit.
git clone https://github.com/microsoft/vcpkg.git "${VCPKG_ROOT}"
git -C "${VCPKG_ROOT}" checkout "${VCPKG_BASELINE_COMMIT}"
"${VCPKG_ROOT}/bootstrap-vcpkg.sh" -disableMetrics

# 2. Render a minimal manifest with builtin-baseline + the protobuf override.
mkdir -p /opt/vcpkg-manifest
cat > /opt/vcpkg-manifest/vcpkg.json <<MANIFEST
{
  "name": "loader-devcontainer",
  "version": "0.1.0",
  "dependencies": ["protobuf"],
  "overrides": [{ "name": "protobuf", "version": "${PROTOBUF_VERSION}" }],
  "builtin-baseline": "${VCPKG_BASELINE_COMMIT}"
}
MANIFEST

# 3. Manifest-mode install. Triplet comes from /opt/buildargs.env
#    (x64-linux on amd64, arm64-linux on arm64).
cd /opt/vcpkg-manifest
"${VCPKG_ROOT}/vcpkg" install \
    --triplet="${VCPKG_TRIPLET}" \
    --x-install-root=/opt/vcpkg-manifest/vcpkg_installed

# 4. Post-install assertion: vcpkg writes a per-port file whose name encodes
#    the resolved version. If the prefix doesn't match what we asked for,
#    fail loudly — silently producing a wrong-version image is the bug we
#    are explicitly defending against.
INFO_FILE=$(ls /opt/vcpkg-manifest/vcpkg_installed/vcpkg/info/protobuf_*_${VCPKG_TRIPLET}.list 2>/dev/null | head -n1)
case "$(basename "${INFO_FILE:-/missing}" 2>/dev/null)" in
  protobuf_${PROTOBUF_VERSION}*)
    ;;
  *)
    echo "ERROR: installed protobuf does not match requested version ${PROTOBUF_VERSION}."
    echo "       vcpkg installed-file marker: ${INFO_FILE:-<none>}"
    echo "       Bump VCPKG_BASELINE_COMMIT (in this Dockerfile, prepare.bat,"
    echo "       and testing-cpp.yml) to a commit that knows about the requested version."
    exit 1
    ;;
esac

# 5. Stable symlinks so ENV CMAKE_PREFIX_PATH (last layer) doesn't have to
#    care about the underlying triplet.
ln -s /opt/vcpkg-manifest/vcpkg_installed/${VCPKG_TRIPLET} /opt/vcpkg/active
ln -s /opt/vcpkg/active/tools/protobuf/protoc /usr/local/bin/protoc
EOF
```

- [ ] **Step 2: Build (this is the slow one, ~25 minutes on cold cache)**

```sh
docker build -t loader-devcontainer:dev .devcontainer/
```

Expected: success after ~25 minutes on first run; subsequent builds reuse the layer instantly. The build log includes lines like `protobuf:x64-linux@6.33.4 -- Building`, then a long CMake/ninja compile, then `Total install time:`.

- [ ] **Step 3: Smoke-check protoc**

```sh
docker run --rm loader-devcontainer:dev protoc --version
```

Expected: `libprotoc 33.4` (the protoc binary reports the umbrella protoc release tag, which is `33.4` for the libprotobuf C++ 6.33.4 line — same mapping `testing-csharp.yml` documents).

- [ ] **Step 4: Smoke-check the symlinks**

```sh
docker run --rm loader-devcontainer:dev sh -c 'readlink -f /usr/local/bin/protoc; readlink -f /opt/vcpkg/active'
```

Expected (on amd64):
```
/opt/vcpkg-manifest/vcpkg_installed/x64-linux/tools/protobuf/protoc
/opt/vcpkg-manifest/vcpkg_installed/x64-linux
```

- [ ] **Step 5: Smoke-check the version-assertion marker file**

```sh
docker run --rm loader-devcontainer:dev sh -c 'ls /opt/vcpkg-manifest/vcpkg_installed/vcpkg/info/protobuf_*.list'
```

Expected: a single file named like `protobuf_6.33.4_x64-linux.list` (or with a `#N` port-revision suffix — the assertion uses prefix matching so either passes).

- [ ] **Step 6: Commit**

```sh
git add .devcontainer/Dockerfile
git commit -m "$(cat <<'EOF'
feat(devcontainer): add vcpkg + protobuf via manifest mode

Pin vcpkg to commit dc8d75c…df932 (lock-step with prepare.bat and
testing-cpp.yml's VCPKG_COMMIT). Render a minimal vcpkg.json manifest
with the protobuf override + builtin-baseline, install via
manifest mode (the only mode where the version pin actually takes
effect), and post-assert that the resolved port version starts with
the requested PROTOBUF_VERSION. Default is 6.33.4; legacy v3 reachable
via --build-arg PROTOBUF_VERSION=3.21.12.

Symlink /opt/vcpkg/active → installed/<triplet> and /usr/local/bin/protoc
→ active/tools/protobuf/protoc so downstream ENV/PATH stays
arch-independent.
EOF
)"
```

---

### Task 6: Verify the protobuf version-pinning path (no Dockerfile change)

This task is verification-only: a deliberate counterexample build that proves the `LOADER_PROTOBUF_VERSION` knob works end-to-end. No commit produced.

- [ ] **Step 1: Build with protobuf 3.21.12 (legacy v3 line)**

```sh
docker build \
    --build-arg PROTOBUF_VERSION=3.21.12 \
    -t loader-devcontainer:legacy-v3 .devcontainer/
```

Expected: vcpkg layer rebuilds (~10 min on cache hit for everything before that layer; the protobuf compile itself takes the whole time). Final assertion succeeds because the resolved port matches `3.21.12`.

- [ ] **Step 2: Smoke-check protoc reports the legacy version**

```sh
docker run --rm loader-devcontainer:legacy-v3 protoc --version
```

Expected: `libprotoc 3.21.12` (the legacy-v3 line still uses `3.x.y` in `protoc --version`; the umbrella tag is `v21.12`).

- [ ] **Step 3: Verify the assertion fires when given an impossible version**

```sh
docker build \
    --build-arg PROTOBUF_VERSION=999.0.0 \
    -t loader-devcontainer:never .devcontainer/ 2>&1 | tail -20
```

Expected: build fails. The vcpkg manifest install errors out (or the assertion catches it), and the last lines include either `error: while looking up version 999.0.0` (vcpkg-side rejection) or `ERROR: installed protobuf does not match requested version 999.0.0` (our assertion). Either failure is acceptable; both prove the silent-wrong-version regression is impossible.

- [ ] **Step 4: Drop the temporary tags**

```sh
docker rmi loader-devcontainer:legacy-v3 loader-devcontainer:never 2>/dev/null || true
```

(No commit — this task is purely verification of behaviour established in Task 5.)

---

### Task 7: Add .NET SDK 8.0 + Node.js 20 LTS

**Files:**
- Modify: `.devcontainer/Dockerfile`

- [ ] **Step 1: Append the .NET + Node layer**

Append to `.devcontainer/Dockerfile`:

```dockerfile
# ---------------------------------------------------------------------------
# .NET SDK 8.0 + Node.js 20 LTS — apt-based installs from the official
# Microsoft and NodeSource repositories. apt-get clean + rm /var/lib/apt/lists
# at the end keeps the layer small.
# ---------------------------------------------------------------------------
RUN <<EOF
set -eux
curl -fsSL https://packages.microsoft.com/config/ubuntu/24.04/packages-microsoft-prod.deb \
    -o /tmp/ms.deb
dpkg -i /tmp/ms.deb
rm /tmp/ms.deb
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y --no-install-recommends \
    dotnet-sdk-8.0 \
    nodejs
apt-get clean
rm -rf /var/lib/apt/lists/*
EOF
```

- [ ] **Step 2: Build**

```sh
docker build -t loader-devcontainer:dev .devcontainer/
```

Expected: success in 2–4 min (apt downloads ~300 MB of .NET + ~50 MB of Node).

- [ ] **Step 3: Smoke-check .NET**

```sh
docker run --rm loader-devcontainer:dev dotnet --version
```

Expected: `8.0.<patch>` (latest 8.0.x at apt-cache time; e.g. `8.0.404`).

- [ ] **Step 4: Smoke-check Node**

```sh
docker run --rm loader-devcontainer:dev node --version
```

Expected: `v20.<minor>.<patch>` (latest v20 LTS at NodeSource setup time).

- [ ] **Step 5: Commit**

```sh
git add .devcontainer/Dockerfile
git commit -m "$(cat <<'EOF'
feat(devcontainer): add .NET SDK 8.0 and Node.js 20 LTS

Microsoft apt repo for dotnet-sdk-8.0 (matches testing-csharp.yml's
target). NodeSource setup_20.x for Node 20 LTS (covers the
experimental _lab/ts/ workflow; not currently in CI). Both clean
their apt caches to keep the layer small.
EOF
)"
```

---

### Task 8: Final ENV (CMAKE_PREFIX_PATH)

This is the single line that lets every existing per-language daily command from the README work inside the container without flag tweaks. `find_package(Protobuf CONFIG)` resolves to vcpkg's pinned protobuf because `CMAKE_PREFIX_PATH` includes `/opt/vcpkg/active`.

**Files:**
- Modify: `.devcontainer/Dockerfile`

- [ ] **Step 1: Append the final ENV block**

Append to `.devcontainer/Dockerfile`:

```dockerfile
# ---------------------------------------------------------------------------
# Final environment: with CMAKE_PREFIX_PATH pointing at vcpkg's installed
# tree, find_package(Protobuf CONFIG) resolves automatically — the existing
# README's "Dev at Linux → CMake (system protobuf)" recipe works as written
# inside the container.
# VCPKG_ROOT is set above (Task 5) for users who want to invoke the toolchain
# file explicitly: -DCMAKE_TOOLCHAIN_FILE=$VCPKG_ROOT/scripts/buildsystems/vcpkg.cmake.
# ---------------------------------------------------------------------------
ENV CMAKE_PREFIX_PATH=/opt/vcpkg/active
```

- [ ] **Step 2: Build**

```sh
docker build -t loader-devcontainer:dev .devcontainer/
```

Expected: success in <30 s (only the ENV layer changes).

- [ ] **Step 3: Smoke-check the env**

```sh
docker run --rm loader-devcontainer:dev sh -c 'echo "VCPKG_ROOT=$VCPKG_ROOT"; echo "CMAKE_PREFIX_PATH=$CMAKE_PREFIX_PATH"'
```

Expected:
```
VCPKG_ROOT=/opt/vcpkg
CMAKE_PREFIX_PATH=/opt/vcpkg/active
```

- [ ] **Step 4: Commit**

```sh
git add .devcontainer/Dockerfile
git commit -m "$(cat <<'EOF'
feat(devcontainer): finalize CMAKE_PREFIX_PATH

CMAKE_PREFIX_PATH=/opt/vcpkg/active lets the existing README cmake
recipe (-DCMAKE_BUILD_TYPE=Debug, no toolchain file) resolve protobuf
inside the container without any flag changes from the host workflow.
EOF
)"
```

---

### Task 9: Add `devcontainer.json`

**Files:**
- Create: `.devcontainer/devcontainer.json`

- [ ] **Step 1: Create the file**

`.devcontainer/devcontainer.json`:

```jsonc
{
  // tableauio/loader Dev Container.
  // See docs/superpowers/specs/2026-05-29-devcontainer-design.md for the design.
  "name": "tableauio/loader",

  // Build args wire host env to Dockerfile ARGs:
  //   LOADER_PROTOBUF_VERSION on the host -> PROTOBUF_VERSION inside.
  // Default 6.33.4 (CI's modern matrix entry). To rebuild against legacy v3:
  //   LOADER_PROTOBUF_VERSION=3.21.12 code .   # then Reopen in Container.
  "build": {
    "dockerfile": "Dockerfile",
    "args": {
      "PROTOBUF_VERSION": "${localEnv:LOADER_PROTOBUF_VERSION:6.33.4}"
    }
  },

  // Persist the Go module cache across container rebuilds. Workspace itself
  // uses VS Code's default bind-mount so edits sync to the host.
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
        // Don't auto-install gopls and friends on first open — let the user
        // do it explicitly from the Go extension's command palette.
        "go.toolsManagement.autoUpdate": false,
        // Don't auto-cmake-configure on workspace open; we run cmake manually
        // per the existing README recipes.
        "cmake.configureOnOpen": false
      }
    }
  },

  // One-line ready banner so the developer knows the container is healthy.
  // Pure echo — no installs, no version-pinning at runtime, no surprises.
  "postCreateCommand": "printf 'tableauio/loader devcontainer ready.\\n  go: %s\\n  buf: %s\\n  protoc: %s\\n  dotnet: %s\\n  node: %s\\n' \"$(go version | cut -d' ' -f3)\" \"$(buf --version)\" \"$(protoc --version)\" \"$(dotnet --version)\" \"$(node --version)\""
}
```

- [ ] **Step 2: Validate the JSON parses (with jsonc comments stripped)**

```sh
python3 - <<'PY'
import json, re, pathlib
src = pathlib.Path('.devcontainer/devcontainer.json').read_text()
# Strip // line comments (devcontainer.json is jsonc).
stripped = re.sub(r'(^|[^:])//.*$', r'\1', src, flags=re.MULTILINE)
parsed = json.loads(stripped)
assert parsed['name'] == 'tableauio/loader'
assert parsed['build']['dockerfile'] == 'Dockerfile'
assert parsed['build']['args']['PROTOBUF_VERSION'] == '${localEnv:LOADER_PROTOBUF_VERSION:6.33.4}'
print('devcontainer.json OK')
PY
```

Expected: `devcontainer.json OK`.

- [ ] **Step 3: Commit**

```sh
git add .devcontainer/devcontainer.json
git commit -m "$(cat <<'EOF'
feat(devcontainer): add devcontainer.json

Wires the Dockerfile under build.args, mounts a named volume for the
Go module cache, declares the VS Code extension set, and prints a
one-line ready-banner via postCreateCommand. PROTOBUF_VERSION flows
from the host LOADER_PROTOBUF_VERSION env var (default 6.33.4) so
contributors can rebuild against the legacy v3 line via:
  LOADER_PROTOBUF_VERSION=3.21.12 code .
EOF
)"
```

---

### Task 10: Add `.devcontainer/README.md`

**Files:**
- Create: `.devcontainer/README.md`

- [ ] **Step 1: Create the file**

`.devcontainer/README.md`:

````markdown
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
````

- [ ] **Step 2: Verify the file renders as expected**

```sh
ls -la .devcontainer/README.md && wc -l .devcontainer/README.md
```

Expected: file exists, ~70 lines.

- [ ] **Step 3: Commit**

```sh
git add .devcontainer/README.md
git commit -m "$(cat <<'EOF'
docs(devcontainer): add .devcontainer/README.md

One-pager covering prerequisites (Docker Desktop / Engine + VS Code
Dev Containers extension), how to open the container, the
LOADER_PROTOBUF_VERSION knob, host-OS caveats (WSL2 workspace
location, Apple Silicon native arm64), and the layered Dockerfile
architecture. Points users at prepare.bat / per-language manual setup
as the explicit fallback path.
EOF
)"
```

---

### Task 11: Update repo root `README.md`

Adds the "Recommended: Dev Container" subsection at the top of `Prerequisites` and prefixes the existing Windows + per-language blocks with a short opt-out lead-in.

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add the "Recommended: Dev Container" subsection**

Open `README.md` and find the Prerequisites bullet list that ends with `…will fail to link.`. Immediately after the existing migration callout (the `> **Migrating from the bundled-protobuf layout?**` block), insert this new subsection — i.e. right before `### Install protobuf`:

```markdown
### Recommended: Dev Container (any host OS)

The fastest way to get a reproducible build environment is to open the
repo in VS Code and choose **Reopen in Container**. The devcontainer
under [`.devcontainer/`](./.devcontainer/) has everything pinned to the
exact versions CI uses (Go 1.24, buf 1.67.0, protobuf 6.33.4 via vcpkg,
.NET 8.0, Node 20). First container build is one-time ~25 minutes (vcpkg
compiles protobuf from source); subsequent reopens are near-instant.

After the container starts you can skip the per-language setup below and
jump straight to **[C++](#c)** / **[Go](#go)** / **[C#](#c-1)** /
**[TypeScript](#typescript)**.

Requirements: Docker Desktop (Windows + macOS) or Docker Engine (Linux),
and the VS Code "Dev Containers" extension. See
[`.devcontainer/README.md`](./.devcontainer/README.md) for the longer
how-to.
```

- [ ] **Step 2: Add an opt-out lead-in to the `Install protobuf` section**

In `README.md`, find the line that currently reads:

```markdown
### Install protobuf

Pick whichever channel fits your platform; loader does not bundle protobuf.
```

Replace it with:

```markdown
### Install protobuf

> **Skip this section if you're using the [devcontainer](#recommended-dev-container-any-host-os).**
> The instructions below cover the manual fallback for hosts where
> Docker isn't available.

Pick whichever channel fits your platform; loader does not bundle protobuf.
```

- [ ] **Step 3: Add an opt-out lead-in to the Windows bootstrap section**

Find:

```markdown
### Windows: bootstrap the rest of the toolchain

Run `prepare.bat` **as Administrator** to install everything you need on a
fresh Windows machine: [Chocolatey](https://chocolatey.org/),
```

Replace with:

```markdown
### Windows: bootstrap the rest of the toolchain

> **Skip this section if you're using the [devcontainer](#recommended-dev-container-any-host-os).**
> `prepare.bat` is the manual fallback for Windows hosts that can't run
> Docker.

Run `prepare.bat` **as Administrator** to install everything you need on a
fresh Windows machine: [Chocolatey](https://chocolatey.org/),
```

- [ ] **Step 4: Verify the README still renders sanely**

```sh
grep -nE '^#{1,4} ' README.md | head -30
```

Expected: shows the heading skeleton, with the new `### Recommended: Dev Container (any host OS)` heading appearing between the Prerequisites bullets and `### Install protobuf`.

- [ ] **Step 5: Commit**

```sh
git add README.md
git commit -m "$(cat <<'EOF'
docs: recommend the devcontainer in repo README

Add a new "Recommended: Dev Container (any host OS)" subsection at the
top of Prerequisites pointing contributors at .devcontainer/. Add a
"Skip this section if you're using the devcontainer" lead-in to the
existing "Install protobuf" and "Windows: bootstrap" blocks so the
manual paths are clearly the fallback, not the primary route.
EOF
)"
```

---

### Task 12: End-to-end integration check (verification only, no commit)

Final smoke test: bring the container up exactly the way a contributor would, and run the four E2E test commands from the README to confirm the toolchain inside the container actually exercises the repo. No commit produced.

- [ ] **Step 1: Build the final container**

```sh
docker build -t loader-devcontainer:dev .devcontainer/
```

Expected: success. If Tasks 1–10 were committed individually, this should be all-cache-hits and finish in <10 s.

- [ ] **Step 2: Run the postCreate banner manually**

```sh
docker run --rm loader-devcontainer:dev sh -c "
printf 'tableauio/loader devcontainer ready.\n  go: %s\n  buf: %s\n  protoc: %s\n  dotnet: %s\n  node: %s\n' \"\$(go version | cut -d' ' -f3)\" \"\$(buf --version)\" \"\$(protoc --version)\" \"\$(dotnet --version)\" \"\$(node --version)\"
"
```

Expected output (versions may vary slightly):
```
tableauio/loader devcontainer ready.
  go: go1.24.0
  buf: 1.67.0
  protoc: libprotoc 33.4
  dotnet: 8.0.404
  node: v20.18.0
```

- [ ] **Step 3: Run the Go E2E inside the container**

```sh
docker run --rm -v "$(pwd):/workspaces/loader" -w /workspaces/loader/test/go-tableau-loader \
    loader-devcontainer:dev sh -c "buf generate .. && go test ./..."
```

Expected: `buf generate` regenerates the Go protoconf and loader stubs, then `go test ./...` reports `ok` for `test/go-tableau-loader`, `internal/index`, `pkg/treemap`, `pkg/udiff`, etc.

- [ ] **Step 4: Run the C++ E2E inside the container**

```sh
docker run --rm -v "$(pwd):/workspaces/loader" -w /workspaces/loader/test/cpp-tableau-loader \
    loader-devcontainer:dev sh -c "
        buf generate .. &&
        cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Debug &&
        cmake --build build --parallel &&
        ctest --test-dir build --output-on-failure
    "
```

Expected: `buf generate` regenerates `*.pb.*` and `*.pc.*`, cmake configure picks up `protobuf::libprotobuf` via `CMAKE_PREFIX_PATH`, build succeeds, ctest reports all tests passed.

- [ ] **Step 5: Run the C# E2E inside the container**

```sh
docker run --rm -v "$(pwd):/workspaces/loader" -w /workspaces/loader/test/csharp-tableau-loader \
    loader-devcontainer:dev sh -c "buf generate .. && dotnet test --nologo --logger 'console;verbosity=normal'"
```

Expected: protobuf C# stubs regenerated, dotnet builds the project, xUnit reports passed.

- [ ] **Step 6: Confirm the named volume mount works (interactive)**

```sh
docker volume create loader-go-mod >/dev/null
docker run --rm -v "$(pwd):/workspaces/loader" \
    -v loader-go-mod:/home/vscode/go \
    -w /workspaces/loader/test/go-tableau-loader \
    loader-devcontainer:dev go test ./...
```

Run the command twice. The second run should be noticeably faster (Go's module cache is warm in `/home/vscode/go/pkg/mod`).

Cleanup:

```sh
docker volume rm loader-go-mod
```

- [ ] **Step 7: Push the branch (no commit; everything is already committed in earlier tasks)**

```sh
git push origin HEAD
```

Expected: branch is pushed to remote with all 10 task commits visible in `git log`.

---

## Self-Review

**Spec coverage** — every Goals item in the spec maps to a task:

| Spec goal | Implementing task(s) |
|---|---|
| One-command setup on any host (Reopen in Container) | Tasks 9 (devcontainer.json) + 10 (.devcontainer/README.md) + 11 (repo README) |
| Reproducibility — pinned versions matching CI | Tasks 3 (Go 1.24), 4 (buf 1.67.0), 5 (vcpkg + protobuf 6.33.4 with assertion), 7 (.NET 8.0, Node 20) |
| Multi-arch native (amd64 + arm64) | Tasks 2 (TARGETARCH detection), 3, 4, 5 (per-arch downloads + triplet) |
| Pinnable protobuf version via LOADER_PROTOBUF_VERSION | Tasks 5 (Dockerfile ARG + manifest mode) + 9 (devcontainer.json `${localEnv:...}`) + 6 (verification) |
| Daily commands stay unchanged (CMAKE_PREFIX_PATH) | Task 8 |

Non-goals (no ghcr.io publish, no CI inside container, no Unity, no replacing prepare.bat) — none of them produce a task, by design.

**Placeholder scan** — none. Every step has runnable code/commands; commit messages are concrete; expected outputs are spelled out.

**Type/path consistency** — `loader-devcontainer:dev` is the consistent image tag across all tasks; `.devcontainer/Dockerfile` and `.devcontainer/devcontainer.json` are referenced with consistent paths; `/opt/vcpkg/active`, `/opt/vcpkg-manifest/...`, `/opt/buildargs.env` all used identically across tasks.

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-05-29-devcontainer.md`. Two execution options:

**1. Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration with two-stage review.

**2. Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints.

Which approach?
