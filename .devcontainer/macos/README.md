# Dev Container — macOS notes

> **There is no macOS devcontainer.** Apple's macOS SLA prohibits
> virtualising macOS on non-Apple hardware, and Microsoft / Docker
> publish no `mcr.microsoft.com/macos/...` base image. This directory is
> documentation only — it intentionally has no `devcontainer.json` so
> the VS Code Dev Containers picker does not list it.

## Recommended path on macOS

Use the [Linux devcontainer](../linux/). Docker Desktop on macOS runs
Linux containers natively:

| Host | Container arch the Linux Dockerfile builds | Native? |
| --- | --- | --- |
| Apple Silicon (M-series) | `linux/arm64` | yes |
| Intel Mac | `linux/amd64` | yes |

Confirm with:

```sh
docker info | grep Architecture
```

→ `aarch64` on Apple Silicon, `x86_64` on Intel. Neither uses Rosetta or
QEMU emulation.

## If you want to build on macOS *without* a container

The same toolchain versions live in
[`../shared/versions.env`](../shared/versions.env) — read them and
install via Homebrew. There is no automated script (yet) because the
Linux container is the recommended path; this is documented for parity
with Windows's `prepare.bat` fallback.

```sh
# Read pinned versions
. .devcontainer/shared/versions.env

# Toolchain via Homebrew
brew install \
    "go@${GO_VERSION%.*}" \
    "buf" \
    "protobuf" \
    "dotnet@${DOTNET_VERSION%.*}" \
    "node@${NODE_VERSION}" \
    "cmake" \
    "ninja"
```

> **Caveats**
>
> - Homebrew tracks the latest formula version, not `versions.env`'s
>   exact pin. For protobuf in particular, this means you may end up
>   with whichever 6.x or later release Homebrew currently ships, not
>   `${PROTOBUF_VERSION}`. If the gencode/runtime version check bites,
>   either install via vcpkg manifest mode (see the repo
>   [README → Install protobuf](../../README.md#install-protobuf)) or
>   switch to the Linux devcontainer.
>
> - `dotnet@8` and `node@20` are versioned brews; the unversioned `dotnet`
>   / `node` formulae track their respective LTS lines and may drift.
>
> - Apple Silicon users: nothing in the loader requires Rosetta. If you
>   see x86_64 Homebrew complaints, run `arch -arm64 brew ...`.

## Why no macOS container?

Three blocking reasons:

1. **Apple licensing.** macOS may only run on Apple-branded hardware.
   Container images are by design hardware-portable; the licence
   forbids that.
2. **No base image.** Microsoft (the publisher of nearly every Windows /
   Linux base image used by Docker) doesn't publish a macOS image, and
   no third party does either.
3. **No runtime.** Even if a base image existed, `dockerd` on macOS runs
   a Linux VM under the hood (LinuxKit on Intel, virtualization-framework
   on Apple Silicon). It cannot run a macOS container.

If your build genuinely needs Apple-specific toolchains (Xcode,
codesigning, `security` keychain), use a real macOS host — that's what
GitHub Actions' `macos-latest` runners are. We don't have one of those
in our matrix today, so this isn't a parity loss.
