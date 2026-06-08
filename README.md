# Loader

The official config loader for [Tableau](https://github.com/tableauio/tableau).

| Plugin | Language | Generated extension |
| --- | --- | --- |
| `protoc-gen-go-tableau-loader` | Go | `*.pc.go` |
| `protoc-gen-cpp-tableau-loader` | C++17 | `*.pc.h` / `*.pc.cc` |
| `protoc-gen-csharp-tableau-loader` | C# (Unity 2022.3 LTS / .NET 8) | `*.pc.cs` |

## Quick start

Use [`make.py`](./make.py) (Python 3.10+, stdlib only):

```sh
python make.py setup --lang all      # one-time host toolchain install
python make.py test  --lang go       # Go
python make.py test  --lang cpp      # C++
python make.py test  --lang csharp   # C#
python make.py test  --lang ts       # TypeScript (experimental)
```

Recommended environment: [devcontainer](./.devcontainer/) (open in VS Code → **Dev Containers: Reopen in Container**). Inside the container, `setup` is a no-op.

Native hosts: `python make.py setup` installs everything pinned to [`./.devcontainer/versions.env`](./.devcontainer/versions.env) — the same versions CI and the devcontainer use. Toolchain layout per host:

- **Go** — official tarball from go.dev to `~/.local/go/` (Linux/macOS) or winget (Windows).
- **buf** — pinned binary from GitHub releases to `~/.local/bin/` (Linux/macOS) or `%LOCALAPPDATA%\buf\bin\` (Windows).
- **protobuf** — vcpkg at `VCPKG_BASELINE_COMMIT` on every native host (Linux/macOS/Windows). Switch versions per-test with `--protobuf-version`.
- **.NET / Node / cmake / ninja** — Homebrew (macOS), Microsoft+NodeSource apt repos (Linux), winget+Chocolatey (Windows).

On Windows, run setup from **cmd as Administrator** the first time. Subsequent commands work from any shell because each subprocess sources `vcvarsall.bat` itself — your shell PATH/INCLUDE/LIB are never mutated.

## Commands

```sh
python make.py setup    [--lang go|cpp|csharp|ts|all]
python make.py generate --lang go|cpp|csharp|ts
python make.py build    --lang go|cpp|csharp|ts [--cxx-std 17|20] [--cxx-compiler msvc|clang|gcc]
                                                [--protobuf-version <ver>] [--triplet <triplet>]
python make.py test     --lang go|cpp|csharp|ts [-k <filter>] [--smoke] [--coverage] [--no-race]
                                                (+ all build flags)
python make.py clean    [--lang ...] [--all]
python make.py env                   # diagnostic JSON
python make.py --version
```

Global flags: `--verbose / -v`, `--dry-run`, `--cwd <path>`.

Examples:

```sh
python make.py test --lang go     -k Test_ActivityConf_OrderedMap
python make.py test --lang go     --race          # opt in (Windows default is off; needs cgo+MSVC)
python make.py test --lang cpp    --protobuf-version 3.21.12
python make.py test --lang csharp -k HubTest.Load
python make.py test --lang cpp    --no-clean   # skip pre-build wipe
```

The C++ flow wipes `test/cpp-tableau-loader/{build,src/tableau,src/protoconf}` by default (gitignored `*.pb.*` from a prior protobuf version shadows fresh codegen).

## Tests for `make.py` itself

```sh
pip install pytest
python -m pytest test_make.py -v
```

CI: [`.github/workflows/testing-make.yml`](.github/workflows/testing-make.yml).

## References

- [Protocol Buffers C++ Reference](https://protobuf.dev/reference/cpp/)
- [Protocol Buffers Go Reference](https://protobuf.dev/reference/go/)
- [protobuf-es](https://github.com/bufbuild/protobuf-es)
- [vcpkg](https://github.com/microsoft/vcpkg)
- [buf CLI](https://buf.build/docs/cli/)
