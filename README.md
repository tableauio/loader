<p align="center">
  <a href="https://tableauio.github.io/">
    <img alt="Tableau" src="https://avatars.githubusercontent.com/u/97329105?s=200&v=4" width="160">
  </a>
</p>

<h3 align="center">Tableau Loader</h3>

<table align="center"><tr>
<tr>
  <td align="center"><b>Package</b></td>
  <td>
    <a href="https://github.com/tableauio/loader/releases?q=protoc-gen-go-tableau-loader"><img src="https://img.shields.io/github/v/release/tableauio/loader?filter=cmd%2Fprotoc-gen-go-tableau-loader%2F*&style=flat-square&label=Go&color=00ADD8&logo=go&logoColor=white" alt="Release Go version"></a>
    <a href="https://github.com/tableauio/loader/releases?q=protoc-gen-cpp-tableau-loader"><img src="https://img.shields.io/github/v/release/tableauio/loader?filter=cmd%2Fprotoc-gen-cpp-tableau-loader%2F*&style=flat-square&label=C%2B%2B&color=00599C&logo=cplusplus&logoColor=white" alt="Release C++ version"></a>
    <a href="https://github.com/tableauio/loader/releases?q=protoc-gen-csharp-tableau-loader"><img src="https://img.shields.io/github/v/release/tableauio/loader?filter=cmd%2Fprotoc-gen-csharp-tableau-loader%2F*&style=flat-square&label=C%23&color=512BD4&logo=csharp&logoColor=white" alt="Release C# version"></a>
    <a href="https://pkg.go.dev/github.com/tableauio/loader"><img src="https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white" alt="go.dev"></a>
  </td>
</tr>
  <td align="center"><b>Release</b></td>
  <td>
    <a href="https://github.com/tableauio/loader/actions/workflows/release-go.yml"><img src="https://github.com/tableauio/loader/actions/workflows/release-go.yml/badge.svg" alt="Release Go"></a>
    <a href="https://github.com/tableauio/loader/actions/workflows/release-cpp.yml"><img src="https://github.com/tableauio/loader/actions/workflows/release-cpp.yml/badge.svg" alt="Release C++"></a>
    <a href="https://github.com/tableauio/loader/actions/workflows/release-csharp.yml"><img src="https://github.com/tableauio/loader/actions/workflows/release-csharp.yml/badge.svg" alt="Release C#"></a>
  </td>
</tr><tr>
  <td align="center"><b>Testing</b></td>
  <td>
    <a href="https://github.com/tableauio/loader/actions/workflows/testing-go.yml"><img src="https://github.com/tableauio/loader/actions/workflows/testing-go.yml/badge.svg" alt="Testing Go"></a>
    <a href="https://github.com/tableauio/loader/actions/workflows/testing-cpp.yml"><img src="https://github.com/tableauio/loader/actions/workflows/testing-cpp.yml/badge.svg" alt="Testing C++"></a>
    <a href="https://github.com/tableauio/loader/actions/workflows/testing-csharp.yml"><img src="https://github.com/tableauio/loader/actions/workflows/testing-csharp.yml/badge.svg" alt="Testing C#"></a>
    <a href="https://github.com/tableauio/loader/actions/workflows/testing-make.yml"><img src="https://github.com/tableauio/loader/actions/workflows/testing-make.yml/badge.svg" alt="Testing make.py"></a>
  </td>
</tr>
<tr>
  <td align="center"><b>License</b></td>
  <td>
    <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/github/license/tableauio/loader?style=flat-square" alt="License"></a>
  </td>
</tr></table>

| Plugin | Language | Generated extension |
| --- | --- | --- |
| `protoc-gen-go-tableau-loader` | Go | `*.pc.go` |
| `protoc-gen-cpp-tableau-loader` | C++17 | `*.pc.h` / `*.pc.cc` |
| `protoc-gen-csharp-tableau-loader` | C# (Unity 2022.3 LTS / .NET 8) | `*.pc.cs` |

## Quick start

Use [`make.py`](./make.py) (Python 3.10+, stdlib only):

```sh
python3 make.py setup --lang all      # one-time host toolchain install
python3 make.py test  --lang go       # Go
python3 make.py test  --lang cpp      # C++
python3 make.py test  --lang csharp   # C#
python3 make.py test  --lang ts       # TypeScript (experimental)
```

Recommended environment: [devcontainer](./.devcontainer/) (open in VS Code → **Dev Containers: Reopen in Container**). Inside the container, `setup` is a no-op.

Native hosts: `python3 make.py setup` installs everything pinned to [`./.devcontainer/versions.env`](./.devcontainer/versions.env) — the same versions CI and the devcontainer use. Toolchain layout per host:

- **Go** — official tarball from go.dev to `~/.local/go/` (Linux/macOS) or winget (Windows).
- **buf** — pinned binary from GitHub releases to `~/.local/bin/` (Linux/macOS) or `%LOCALAPPDATA%\buf\bin\` (Windows).
- **protobuf** — vcpkg at `VCPKG_BASELINE_COMMIT` on every native host (Linux/macOS/Windows). Switch versions per-test with `--protobuf-version`.
- **.NET / Node / cmake / ninja** — Homebrew (macOS), Microsoft+NodeSource apt repos (Linux), winget+Chocolatey (Windows).

On Windows, run setup from **cmd as Administrator** the first time. Subsequent commands work from any shell because each subprocess sources `vcvarsall.bat` itself — your shell PATH/INCLUDE/LIB are never mutated.

## Commands

```sh
python3 make.py setup    [--lang go|cpp|csharp|ts|all]
python3 make.py generate --lang go|cpp|csharp|ts
python3 make.py build    --lang go|cpp|csharp|ts [--cxx-std 17|20] [--cxx-compiler msvc|clang|gcc]
                                                [--protobuf-version <ver>] [--triplet <triplet>]
python3 make.py test     --lang go|cpp|csharp|ts [-k <filter>] [--smoke] [--coverage] [--no-race]
                                                (+ all build flags)
python3 make.py clean    [--lang ...] [--all]
python3 make.py env                   # diagnostic JSON
python3 make.py --version
```

Global flags: `--verbose / -v`, `--dry-run`, `--cwd <path>`.

Examples:

```sh
python3 make.py test --lang go     -k Test_ActivityConf_OrderedMap
python3 make.py test --lang go     --race          # opt in (Windows default is off; needs cgo+MSVC)
python3 make.py test --lang cpp    --protobuf-version 3.21.12
python3 make.py test --lang csharp -k HubTest.Load
python3 make.py test --lang cpp    --no-clean   # skip pre-build wipe
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
