# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

`github.com/tableauio/loader` is the official config-loader generator for [Tableau](https://github.com/tableauio/tableau). It ships **three `protoc` plugins** written in Go that read protobuf files annotated with the `tableau.workbook` / `tableau.worksheet` / `tableau.field` extensions and emit strongly-typed loader code in three target languages:

| Plugin (under `cmd/`) | Output language | Generated file extension |
| --- | --- | --- |
| `protoc-gen-go-tableau-loader` | Go | `*.pc.go` |
| `protoc-gen-cpp-tableau-loader` | C++17 | `*.pc.h` / `*.pc.cc` |
| `protoc-gen-csharp-tableau-loader` | C# (Unity 2022.3 LTS / .NET 8) | `*.pc.cs` |

Generated code is opinionated: every worksheet message becomes a `Messager` with `Load`/`Store`/`Get*`/index/ordered-map accessors, all messagers register into a singleton-ish `Hub`, and the runtime delegates file IO + protobuf (un)marshaling to the upstream `github.com/tableauio/tableau` Go module's `format`/`load`/`store` packages.

## Common commands

All build and test work happens **per language** inside `test/<lang>-tableau-loader/`. The repo root only hosts the Go module and the plugin sources; running `go test ./...` from the root only exercises the small shared packages.

### Dev Container (recommended for most contributors)

`.devcontainer/` ships a single Ubuntu 24.04 image with the entire toolchain pinned to CI's exact versions: Go 1.24, buf 1.67.0, protobuf 6.33.4 (via vcpkg, manifest mode), .NET 8, Node 20. Open the repo in VS Code and run **Dev Containers: Reopen in Container** — first build is one-time ~25 min (vcpkg compiles protobuf from source); reopens after that are instant. Inside the container every command in the per-language sections below works as written, *and* the C++ section's `-DCMAKE_TOOLCHAIN_FILE=...` flag is unnecessary (the image sets `CMAKE_PREFIX_PATH=/opt/vcpkg/active`). The container is the recommended path on every host OS; `prepare.bat` and the per-language `Install protobuf` recipes in the repo README stay as the explicit fallback for hosts that can't run Docker.

To rebuild against the legacy v3 protobuf line: `LOADER_PROTOBUF_VERSION=3.21.12 code .` then **Reopen in Container**. The Dockerfile ARG `PROTOBUF_VERSION` is wired to that host env var via `devcontainer.json` build args; vcpkg manifest mode pins the override and the post-install assertion fails the build if anything resolves to the wrong version. See `.devcontainer/README.md` for the longer how-to and host-OS caveats (notably: Windows users should check the workspace out under WSL2, not `/mnt/c/`, for usable bind-mount perf).

The container is **not** used in CI. CI workflows still run `lukka/run-vcpkg` directly (faster cached vcpkg installs); the container exists for local-dev parity only. `VCPKG_BASELINE_COMMIT` in `.devcontainer/Dockerfile` is in lock-step with `prepare.bat`'s `VCPKG_BASELINE_COMMIT` and `testing-cpp.yml`'s `VCPKG_COMMIT` — bump all three together.

### Plugin development (Go module at repo root)

```sh
go vet ./...
go test ./...                                    # tests live in internal/index, internal/loadutil, pkg/treemap, pkg/udiff
go test ./internal/index -run Test_ParseIndexDescriptor  # single test
go build -o /tmp/p ./cmd/protoc-gen-go-tableau-loader     # smoke-build a plugin
```

The plugins are always invoked through `buf generate` from a test directory — `buf.gen.yaml` runs them via `local: ["go", "run", "../../cmd/protoc-gen-go-tableau-loader"]`, so plugin changes take effect on the next `buf generate` without an explicit install step.

### Go end-to-end (`test/go-tableau-loader/`)

```sh
cd test/go-tableau-loader
buf generate ..                                  # regenerate *.pb.go + protoconf/loader/*.pc.go
go test ./...                                    # runs hub, index, and main test suites
go test -run Test_ActivityConf_OrderedMap ./...  # single test
```

CI mirrors this with `go test -v -timeout 30m -race ./... -coverprofile=coverage.txt`.

### C++ end-to-end (`test/cpp-tableau-loader/`)

Requires a matching `protoc` + `libprotobuf` toolchain (protobuf v22+ enforces a strict gencode/runtime version check). Loader does **not** vendor protobuf — bring your own via vcpkg, system package, or source build.

```sh
cd test/cpp-tableau-loader
buf generate ..                                  # writes src/protoconf/*.pb.* and *.pc.*
cmake -S . -B build -DCMAKE_BUILD_TYPE=Debug \
    -DCMAKE_TOOLCHAIN_FILE=<vcpkg-root>/scripts/buildsystems/vcpkg.cmake
cmake --build build --parallel
ctest --test-dir build --output-on-failure
ctest --test-dir build -R HubTest.Load --output-on-failure  # single test
```

Windows additionally requires running `.\prepare.bat` from the repo root (in **cmd**, not PowerShell, **as Administrator** the first time) in every new shell — it installs the toolchain on first run and re-loads MSVC env vars (`cl.exe`, `INCLUDE`, `LIB`, `VCPKG_ROOT`, …) every run because `vcvarsall.bat` does not persist them. Use `-DCMAKE_BUILD_TYPE=Debug -DVCPKG_TARGET_TRIPLET=x64-windows-static` to match the static-CRT protobuf installed by `prepare.bat`; mismatched CRTs surface as `LNK2038 _ITERATOR_DEBUG_LEVEL` errors. GoogleTest is fetched via CMake `FetchContent` — no manual install. (None of this applies inside the Dev Container — Windows users on that path interact only with Docker Desktop + WSL2.)

### C# end-to-end (`test/csharp-tableau-loader/`)

```sh
cd test/csharp-tableau-loader
buf generate ..                                  # writes protoconf/ and tableau/*.pc.cs
dotnet test                                      # xUnit
dotnet test --filter "FullyQualifiedName~HubTest.Load"  # single test
```

### TypeScript scratch (`_lab/ts/`)

Experimental, not wired into CI. `npm install && npm run generate && npm run test`.

## Big-picture architecture

### Plugin pipeline (the part you'll modify most)

Every plugin's `main.go` is the same shape: parse flags, set protogen feature bits (proto2 → editions 2024, FEATURE_PROTO3_OPTIONAL), iterate `gen.Files`, and for each file/message decide what to generate. The decision logic is centralized in `internal/options`:

- `options.NeedGenFile(f)` — file-level gate: must have `(tableau.workbook)` set and at least one message with `(tableau.worksheet)`.
- `options.NeedGenOrderedMap` / `NeedGenIndex` / `NeedGenOrderedIndex` — message-level gates that *also* honour the `lang_options` map on `WorksheetOptions`, e.g. `lang_options: { key: "Index" value: "go" }` means "only generate the index accessors for Go". `internal/options/options.go` defines the language IDs (`cpp`, `go`, `cs`).

Each plugin then splits work between two passes:

1. **Per-message generation** (`messager.go` in each plugin) — emits one `*.pc.{go,h,cc,cs}` per `.proto`. Delegates ordered-map field/method emission to `cmd/<plugin>/orderedmap/` and index field/method emission to `cmd/<plugin>/indexes/`. The shared semantic model — what the index syntax means — lives in **`internal/index`**, not in each plugin: `ParseIndexDescriptor` walks the message tree and returns a `LevelMessage` linked-list describing what indices apply at which nesting level (map → list → map → list, etc.). Plugins consume this descriptor to emit language-specific code; do not duplicate this parsing per-language.
2. **Cross-message ("embed") generation** — emits `hub.pc.*`, `messager_container.pc.go`, `util.pc.*`, etc. Driven by `cmd/<plugin>/embed.go`, which `//go:embed`s the templates under `cmd/<plugin>/embed/templates/` (Go) or `cmd/<plugin>/embed/` (C++/C# also include verbatim `*.pc.{h,cc,cs}` files that are emitted unchanged). The list of "all messagers in source order" the templates iterate over comes from **`internal/xproto.ParseProtoFiles`**.

The C++ plugin is the only one with sharding: `--shards=N` in `buf.gen.yaml` makes `xproto.ProtoFiles.SplitShards(N)` partition messagers across N `hub_shard*.pc.cc` files to parallelize the (very heavy) compile. It also has a tri-state `--mode` flag (`default` / `hub` / `messager`) for users who want to split protoconf generation across separate `buf generate` invocations.

### Index syntax (`internal/index/index.go`)

The `worksheet.index` / `worksheet.ordered_index` strings use a compact mini-language parsed by one regex. Knowing the shape helps when reading or writing tests in `test/proto/index_conf.proto`:

```
ID                              # single-column index on field "ID"
ID@Item                         # named "Item"
ID<SortedCol>@Item              # sort within group by SortedCol
(ID, Name)<SortedCol1, SortedCol2>@Item   # composite index
CountryItemAttrName             # CamelCase concatenation of nested-field path Country.Item.Attr.Name
```

Multi-level nested maps/lists are flattened: `CountryItemAttrName` reaches into `country_list[].item_map[].attr_list[].name`. Generated indexes return leveled key structs whose names are derived in `helper.ParseLeveledMapPrefix` — a 3-level map with indexes only at the 2nd level produces fewer key structs than one with indexes at every level (compare `Fruit5Conf` vs `Fruit4Conf` in `test/proto/index_conf.proto`).

### Hub and Messager runtime (Go)

Hub state is held in `atomic.Pointer[MessagerContainer]` so `Load(...)` can swap the entire snapshot atomically while `Get*()` callers race-freely read the previous one. The container is generated (`messager_container.pc.go`) and exposes both a generic `GetMessager(name)` and a typed `Get<Name>()` per messager. `Hub.NewContext` / `FromContext` propagate the snapshot through `context.Context` so request handlers see a stable view.

Mutability detection (`hub.WithMutableCheck`) is opt-in: enabling it flips `enableBackup()` on each messager so they retain the originally-loaded `proto.Message`, then a goroutine periodically `proto.Equal`s `originalMessage()` vs `Message()` and calls `OnMutate(name, original, current)` (default handler prints a unified diff via `pkg/udiff`).

`processAfterLoad` (per-messager, runs as `Load` finishes) and `ProcessAfterLoadAll(hub *Hub)` (cross-messager, runs after all are loaded against a temporary hub) are the two extension points. `test/go-tableau-loader/customconf/custom_item_conf.go` shows how a hand-written messager registers via `tableau.Register(func() Messager { ... })` and uses `ProcessAfterLoadAll` to consume data from another messager — that's the canonical pattern for derived/computed configs.

### Test data flow

1. `test/proto/*.proto` — hand-written annotated protos (the source of truth for what generators are exercised).
2. `test/testdata/conf/*.json`, `patchconf/`, `patchconf2/`, `patchresult/` — JSON inputs the upstream `tableau` toolchain would produce from spreadsheets; loader tests read them at runtime.
3. Each `test/<lang>-tableau-loader/` runs `buf generate ..` to produce its language-specific stubs into a per-language output dir, then runs that language's native test runner against `test/testdata/`.

Patch tests verify three semantics defined upstream (merge / replace / recursive_patch) — the loader's job is just to wire `patch_paths`/`patch_dirs` through `load.MessagerOptions`.

## Versioning and releases

Each plugin has its own `version` constant in its `main.go` and is released independently via tags shaped `cmd/protoc-gen-{go,cpp,csharp}-tableau-loader/<semver>`. The matching workflows in `.github/workflows/release-*.yml` build cross-platform binaries and attach them to the GitHub release. Bump the relevant `const version = "..."` in lockstep with the tag.

## Style and conventions

- C++ / proto: `clang-format` with the rules in `.clang-format` (Google base, 120-col).
- Go: standard `gofmt` / `go vet`; CI runs `go vet ./...` and `go test -race`.
- Generated files end in `.pc.<ext>` (the `extensions.PC` constant). Anything matching `*.pb.*` is `.gitignore`d — never check generated proto-runtime files in.
- Worksheet language gating: when adding a feature that only some target languages support, add it behind a `lang_options` check in `internal/options` rather than making the per-plugin `messager.go` know the rules.
