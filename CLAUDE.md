# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository.

## What this repo is

`github.com/tableauio/loader` is the official config-loader generator for [Tableau](https://github.com/tableauio/tableau). It ships **three `protoc` plugins** (Go) that read protobuf files annotated with the `tableau.workbook` / `tableau.worksheet` / `tableau.field` extensions and emit strongly-typed loader code:

| Plugin (under `cmd/`) | Output | Generated extension |
| --- | --- | --- |
| `protoc-gen-go-tableau-loader` | Go | `*.pc.go` |
| `protoc-gen-cpp-tableau-loader` | C++17 | `*.pc.h` / `*.pc.cc` |
| `protoc-gen-csharp-tableau-loader` | C# (Unity 2022.3 LTS / .NET 8) | `*.pc.cs` |

Generated code is opinionated: every worksheet message becomes a `Messager` with `Load`/`Store`/`Get*`/index/ordered-map accessors; all messagers register into a singleton-ish `Hub`; runtime delegates file IO + protobuf (un)marshaling to `github.com/tableauio/tableau`'s `format`/`load`/`store` packages.

## Common commands

Build/test happens **per language** under `test/<lang>-tableau-loader/` (TS lives under `_lab/ts/`). The repo root only hosts the Go module + plugin sources; `go test ./...` from root only exercises shared packages (`internal/index`, `internal/loadutil`, `pkg/treemap`, `pkg/udiff`).

The single cross-platform driver is **`make.py`** (Python 3.10+, stdlib only). It works on Windows, macOS, Linux, and inside the devcontainer, and is what CI calls.

```sh
python make.py setup    --lang all       # one-time host toolchain install (no-op in container)
python make.py generate --lang go        # buf generate ..
python make.py build    --lang cpp
python make.py test     --lang go
python make.py test     --lang cpp -k HubTest.Load
python make.py env                       # diagnostic JSON
python make.py --version
```

C++ wipes `test/cpp-tableau-loader/{build,src/tableau,src/protoconf}` before regenerating (gitignored `*.pb.*` shadows fresh codegen). `--no-clean` skips it. A leftover `vcpkg.json` from a previous `--protobuf-version` run is auto-removed in classic mode so cmake doesn't accidentally re-enter manifest mode.

On Windows, `make.py` wraps every C++ subprocess in `cmd /c "call vcvarsall.bat x64 >nul && <cmd>"` so MSVC env lives per-subprocess; the shell PATH is never mutated.

`make.py setup` pins every toolchain dimension (matches CI + devcontainer): Go via official go.dev tarball to `~/.local/go/`, buf via GitHub release binary, protobuf via **vcpkg at `VCPKG_BASELINE_COMMIT` on every host (macOS/Linux/Windows)**, .NET / Node via Microsoft+NodeSource (Linux) or Homebrew (macOS) or winget (Windows), cmake/ninja via the host's package manager. Resolved paths cached in `~/.loader-env.json` so subsequent `make.py test --lang cpp` invocations pick them up without re-running setup.

### Dev container

- `.devcontainer/` → **Dev Containers: Reopen in Container**. Ubuntu 24.04 + all toolchains pinned. First build ~25 min; reopens instant.
- Inside: `python make.py setup` is a no-op. `python make.py test --lang <X>` works for all languages — Dockerfile presets `CMAKE_PREFIX_PATH=/opt/vcpkg/active`.
- Override protobuf version: `LOADER_PROTOBUF_VERSION=3.21.12 code .` then **Rebuild Container**.
- Single source of truth for all toolchain versions: **`.devcontainer/versions.env`**.

CI primary tests (`testing-{cpp,go,csharp}.yml`) use `lukka/run-vcpkg` for cached vcpkg installs + `python make.py test --lang <X>` for build/test. `devcontainer-smoke.yml` builds the image on `.devcontainer/**` PRs (amd64 + arm64). `testing-make.yml` runs the make.py unit + dry-run regression suite on every push.

### Plugin development (Go module at repo root)

```sh
go vet ./...
go test ./...                                            # internal/index, internal/loadutil, pkg/treemap, pkg/udiff
go test ./internal/index -run Test_ParseIndexDescriptor  # single test
go build -o /tmp/p ./cmd/protoc-gen-go-tableau-loader    # smoke-build a plugin
```

Plugins are invoked through `buf generate` from a test directory (`buf.gen.yaml` runs them via `local: ["go", "run", "../../cmd/protoc-gen-go-tableau-loader"]`), so plugin changes take effect on the next `buf generate` without an explicit install step.

### Per-language

```sh
# Go
python make.py test --lang go                                        # full
python make.py test --lang go -k Test_ActivityConf_OrderedMap        # filter
python make.py test --lang go --smoke                                # plugin-only `go vet` (devcontainer-smoke)
python make.py test --lang go --coverage                             # CI: -coverprofile=coverage.txt -covermode=atomic
python make.py test --lang go --race                                 # opt in to -race (default off on Windows; needs cgo+MSVC)

# C++ (requires matching protoc + libprotobuf — protobuf v22+ enforces gencode/runtime check)
python make.py test --lang cpp                                       # full
python make.py test --lang cpp -k HubTest.Load                       # filter
python make.py test --lang cpp --cxx-std 20                          # C++20
python make.py test --lang cpp --cxx-compiler clang                  # clang++
python make.py test --lang cpp --protobuf-version 3.21.12            # legacy v3 (vcpkg manifest mode)

# C#
python make.py test --lang csharp                                    # full
python make.py test --lang csharp -k HubTest.Load                    # FullyQualifiedName~HubTest.Load

# TypeScript (experimental, not in CI)
python make.py test --lang ts                                        # npm install + generate + test
```

GoogleTest is fetched via CMake `FetchContent` — no manual install.

### make.py regression tests

```sh
pip install pytest && python -m pytest test_make.py -v
```

`test_make.py` (next to `make.py`) covers: pure-logic unit tests (`Versions`, `Platform`, `windows_msvc_wrap`, `_winquote`, `Runner`, repo-root discovery) + dry-run snapshot tests (assert exact subprocess sequences for every `<subcommand> --lang <X>` combo). Runs in <5s; CI: `.github/workflows/testing-make.yml` (ubuntu/macos/windows).

## Big-picture architecture

### Plugin pipeline (the part you'll modify most)

Every plugin's `main.go` is the same shape: parse flags, set protogen feature bits (proto2 → editions 2024, FEATURE_PROTO3_OPTIONAL), iterate `gen.Files`, decide what to generate per file/message. Decision logic is centralized in **`internal/options`**:

- `options.NeedGenFile(f)` — file-level gate: must have `(tableau.workbook)` set and at least one message with `(tableau.worksheet)`.
- `options.NeedGenOrderedMap` / `NeedGenIndex` / `NeedGenOrderedIndex` — message-level gates that *also* honour the `lang_options` map on `WorksheetOptions` (e.g. `lang_options: { key: "Index" value: "go" }` = "only generate index accessors for Go"). Language IDs (`cpp`, `go`, `cs`) are in `internal/options/options.go`.

Each plugin splits work between two passes:

1. **Per-message generation** (`messager.go` per plugin) — emits one `*.pc.{go,h,cc,cs}` per `.proto`. Delegates ordered-map field/method emission to `cmd/<plugin>/orderedmap/`, index emission to `cmd/<plugin>/indexes/`. **The shared semantic model — what the index syntax means — lives in `internal/index`, not per-plugin**: `ParseIndexDescriptor` walks the message tree and returns a `LevelMessage` linked-list describing what indices apply at each nesting level (map → list → map → list, etc.). Plugins consume this descriptor; do not duplicate parsing per-language.
2. **Cross-message ("embed") generation** — emits `hub.pc.*`, `messager_container.pc.go`, `util.pc.*`. Driven by `cmd/<plugin>/embed.go`, which `//go:embed`s templates under `cmd/<plugin>/embed/templates/` (Go) or `cmd/<plugin>/embed/` (C++/C# also include verbatim `*.pc.{h,cc,cs}` files emitted unchanged). The "all messagers in source order" iteration source is **`internal/xproto.ParseProtoFiles`**.

C++ plugin sharding: `--shards=N` in `buf.gen.yaml` makes `xproto.ProtoFiles.SplitShards(N)` partition messagers across N `hub_shard*.pc.cc` files to parallelize the (heavy) compile. Tri-state `--mode` (`default` / `hub` / `messager`) lets users split protoconf generation across separate `buf generate` invocations.

### Index syntax (`internal/index/index.go`)

`worksheet.index` / `worksheet.ordered_index` strings use a compact mini-language parsed by one regex:

```
ID                                         # single-column index on field "ID"
ID@Item                                    # named "Item"
ID<SortedCol>@Item                         # sort within group by SortedCol
(ID, Name)<SortedCol1, SortedCol2>@Item    # composite index
CountryItemAttrName                        # CamelCase concatenation of nested-field path Country.Item.Attr.Name
```

Multi-level nested maps/lists are flattened: `CountryItemAttrName` reaches into `country_list[].item_map[].attr_list[].name`. Generated indexes return leveled key structs whose names are derived in `helper.ParseLeveledMapPrefix` — a 3-level map with indexes only at the 2nd level produces fewer key structs than one with indexes at every level (compare `Fruit5Conf` vs `Fruit4Conf` in `test/proto/index_conf.proto`).

### Hub and Messager runtime (Go)

- Hub state is held in `atomic.Pointer[MessagerContainer]` so `Load(...)` swaps the snapshot atomically while `Get*()` callers race-freely read the previous one.
- Container is generated (`messager_container.pc.go`) with both generic `GetMessager(name)` and typed `Get<Name>()`. `Hub.NewContext` / `FromContext` propagate the snapshot through `context.Context`.
- `hub.WithMutableCheck` is opt-in: enabling it flips `enableBackup()` on each messager, then a goroutine periodically `proto.Equal`s `originalMessage()` vs `Message()` and calls `OnMutate(name, original, current)` (default handler prints a unified diff via `pkg/udiff`).
- Two extension points: `processAfterLoad` (per-messager, runs as `Load` finishes) and `ProcessAfterLoadAll(hub *Hub)` (cross-messager, against a temporary hub). `test/go-tableau-loader/customconf/custom_item_conf.go` shows the canonical pattern: hand-written messager registers via `tableau.Register(func() Messager { ... })` and uses `ProcessAfterLoadAll` to consume data from another messager — that's how derived/computed configs are built.

### Test data flow

1. `test/proto/*.proto` — hand-written annotated protos (source of truth for what generators are exercised).
2. `test/testdata/conf/*.json`, `patchconf/`, `patchconf2/`, `patchresult/` — JSON inputs the upstream `tableau` toolchain produces from spreadsheets; loader tests read them at runtime.
3. Each `test/<lang>-tableau-loader/` runs `buf generate ..` → language-specific stubs → native test runner against `test/testdata/`.

Patch tests verify three semantics defined upstream (merge / replace / recursive_patch) — loader's job is just to wire `patch_paths`/`patch_dirs` through `load.MessagerOptions`.

## Versioning and releases

Each plugin has its own `version` constant in its `main.go`, released independently via tags shaped `cmd/protoc-gen-{go,cpp,csharp}-tableau-loader/<semver>`. Matching workflows in `.github/workflows/release-*.yml` build cross-platform binaries and attach them to the GitHub release. Bump `const version = "..."` in lockstep with the tag.

## Style and conventions

- C++ / proto: `clang-format` per `.clang-format` (Google base, 120-col).
- Go: standard `gofmt` / `go vet`; CI runs `go vet ./...` and `go test -race`.
- Generated files end in `.pc.<ext>` (the `extensions.PC` constant). Anything matching `*.pb.*` is `.gitignore`d — never commit generated proto-runtime files.
- Worksheet language gating: when adding a feature only some target languages support, gate it via `lang_options` in `internal/options` rather than baking the rule into per-plugin `messager.go`.
