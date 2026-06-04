# Dev Container

The recommended way to develop on `tableauio/loader`. One container, all four target languages (C++17, Go, .NET, Node) plus protobuf via vcpkg, pinned to the same toolchain CI uses. Version pins live in [`./versions.env`](./versions.env), consumed by the Dockerfile, [`make.py`](../make.py), and the CI workflows.

For native dev (no Docker, no WSL2), use [`make.py`](../make.py) directly.

## Prerequisites

- Docker Desktop (Windows / macOS) or Docker Engine (Linux)
- VS Code with the [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)

## Open the container

```sh
code .
```

Then **Dev Containers: Reopen in Container** from the command palette. First build is one-time ~25 min (vcpkg compiles protobuf); reopens are near-instant.

Inside the container, `python make.py setup` is a no-op. Use `python make.py test --lang <X>` for any language.

## Pin a different protobuf version

```sh
LOADER_PROTOBUF_VERSION=3.21.12 code .
```

Then **Rebuild Container**. Only the vcpkg layer rebuilds.

## Host-OS caveats

- **Windows.** WSL2 backend required. Check the workspace out under WSL2 (`\\wsl.localhost\Ubuntu\home\<user>\loader`) — not `/mnt/c/...` — for good bind-mount performance.
- **Apple Silicon.** Docker builds the image natively as arm64. Confirm with `docker info | grep Architecture`.
- **Linux (native Docker Engine).** No special configuration.

## Architecture

Single-stage Dockerfile on `mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04`:

1. Architecture detection (`TARGETARCH` → Go arch, buf arch, vcpkg triplet)
2. Go — version from `versions.env`
3. buf — version from `versions.env`
4. vcpkg pinned to `VCPKG_BASELINE_COMMIT`, protobuf installed via manifest mode (asserts version)
5. .NET SDK — version from `versions.env`
6. Node.js LTS — version from `versions.env`
7. `ENV CMAKE_PREFIX_PATH=/opt/vcpkg/active` so `find_package(Protobuf CONFIG)` resolves automatically

Build context is `.devcontainer/`, so `COPY versions.env …` resolves directly.

## `versions.env` format

- One `KEY=VALUE` per line.
- No quotes, no spaces around `=`, no inline comments.
- Comments at column 0 with `#`. Blank lines ignored. No shell expansion.

Consumers: Dockerfile (sourced as a shell file), [`make.py`](../make.py) (`Versions.load()`), and `.github/actions/load-versions` (exports to `$GITHUB_ENV`).

Use `python make.py env` for a JSON dump of resolved values.

## Troubleshooting

### Stale-codegen errors after `buf generate`

Symptoms:
- C++: `fatal error: google/protobuf/generated_message_table_driven.h: No such file or directory`
- C#: hundreds of `error CS0101: The namespace already contains a definition for ...`

The host workspace has gitignored `*.pb.*` from a previous protobuf version that `git pull` didn't remove. Wipe with `python make.py clean --lang cpp` (or `--lang csharp`), then rerun the test.

## Falling back

No Docker? Use [`make.py`](../make.py) directly: `python make.py setup --lang all` then `python make.py test --lang <X>`. Works on macOS / Linux / Windows native.
