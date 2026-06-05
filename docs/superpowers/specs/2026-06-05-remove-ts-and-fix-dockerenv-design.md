# Remove TypeScript support from `make.py` and fix `/.dockerenv` heuristic

**Date:** 2026-06-05
**Status:** Approved

## Background

Two real bugs surfaced while testing `python make.py setup --lang all` on a fresh `ubuntu:24.04` container:

1. **`/.dockerenv` triggers a silent devcontainer short-circuit.** `Platform.detect()` (`make.py` lines 146-148) treats *any* container as the devcontainer:

   ```python
   in_devcontainer = (
       Path("/opt/vcpkg/active").exists() or Path("/.dockerenv").exists()
   )
   ```

   `/.dockerenv` is created by Docker for **every** container, so `make.py setup` is a no-op in any non-devcontainer Docker image. The intended marker is `/opt/vcpkg/active`, which the devcontainer Dockerfile actually sets.

2. **`_ensure_node_linux` is a no-op.** `make.py` lines 710-713 only print
   `[info] Node not found; install via your distro or NodeSource manually.`
   and return. Setup for `--lang ts` therefore silently leaves the host without Node, contradicting CLAUDE.md's claim that Node is installed via "Microsoft+NodeSource (Linux)".

Two further bugs were found inside the TS lab itself (`_lab/ts/package.json` is missing a `build` script before its `node build/index.js` test command, and `_lab/ts/src/index.ts` references `../../test/testdata/ThemeConf.json` instead of `../../test/testdata/conf/ThemeConf.json`). CLAUDE.md already documents the TS pipeline as "experimental, not in CI", and grep confirms **no CI workflow references TS**. Rather than maintain a broken experimental path, this work removes TS from `make.py`'s tool surface entirely. The `_lab/ts/` scratchpad stays in the repo for anyone who wants to experiment manually; the lab's two bugs are intentionally left unfixed.

## Goals

- Drop TypeScript from the `make.py` tool surface (`setup` / `build` / `test` / `clean` / `env`).
- Fix the `/.dockerenv` heuristic so `make.py setup` actually runs in arbitrary Docker containers.
- Keep `_lab/ts/` in the tree as an untooled scratchpad.
- Keep the C++/Go/.NET workflows green; `python -m pytest test_make.py -v` must pass.

## Non-goals

- Fixing the two `_lab/ts/` bugs (missing `build` script, stale data path) — out of scope, the lab is no longer tooled.
- Touching plugin source under `cmd/protoc-gen-{go,cpp,csharp}-tableau-loader/`.
- CI workflow changes — confirmed no workflow references TS.

## Architecture

`make.py` keeps its current shape. Three platform installers (`_setup_linux` / `_setup_macos` / `_setup_windows`), per-language helpers (`_lang_dir`, `_build_or_test`, `cmd_clean`, `cmd_env`), and `LANGS_ALL` all stay. The only structural change is shrinking the language axis from 4 → 3 by removing `"ts"`.

## Changes by file

### `make.py`

1. **Module docstring (lines 1-15):** scrub the `npm test` mention on line 6, and remove `|ts` from each of the `--lang go|cpp|csharp|ts[|all]` Usage examples (lines 10-13).
2. **`Versions` (lines 105-106):** delete the `node_version` property.
3. **`Platform.detect` (lines 146-148):** the `/.dockerenv` clause is dropped; only `Path("/opt/vcpkg/active").exists()` remains as the devcontainer signal.
4. **`LANGS_ALL` (line 491):** drop `"ts"`. Resulting tuple: `("go", "cpp", "csharp")`.
5. **`_setup_macos` (lines 547-549):** delete the `if "ts" in langs:` branch that appends `node@N` to the brew package list.
6. **`_setup_linux` (lines 615-616):** delete the `if "ts" in langs: _ensure_node_linux(ctx)` call.
7. **`_ensure_node_linux` (lines 710-713):** delete the function body entirely.
8. **`_setup_windows` (lines 824-827):** delete the `if "ts" in langs and _which("node") is None:` winget block.
9. **`_lang_dir` (lines 932-933):** delete the `if lang == "ts": return repo_root / "_lab" / "ts"` branch.
10. **`generate` dispatch (lines 941-944):** delete the `if lang == "ts":` arm that runs `npm run generate`.
11. **`_build_or_test` (lines 982-983):** delete the `if lang == "ts":` dispatch to `_ts_build_or_test`.
12. **`_ts_build_or_test` (lines 1217-1227):** delete the function entirely.
13. **`cmd_clean` (lines 1251-1252):** delete the `elif lang == "ts": rmtree(node_modules)` branch.
14. **`cmd_env` (lines 1281-1282):** delete the `"node"` and `"npm"` entries from the `tools` dict.

### `test_make.py`

- Remove `"NODE_VERSION"` from the two parser-test key lists (lines 49, 479).
- Remove the `assert v.node_version == v.raw["NODE_VERSION"]` line (line 62).
- Delete `test_ts_lives_under_lab` (lines 429-431).
- Audit dry-run snapshot tests for any `--lang ts` cases and remove them. (Inventory step during implementation: grep for `"ts"` and `_lang_dir.*ts` across `test_make.py`.)

### `.devcontainer/versions.env`

- Delete the `# Node.js LTS major. NodeSource apt repo is `setup_${NODE_VERSION}.x`.` comment block and the `NODE_VERSION=20` line.

### `.devcontainer/Dockerfile`

- Remove the NodeSource setup curl + `nodejs` apt install (lines around 196-199).
- Trim comment headers (lines 5, 9, 185, 186) that mention Node so they accurately describe what the image installs.

### `CLAUDE.md`

- Line 19: rewrite to drop `(TS lives under _lab/ts/)`.
- Lines 80-81: delete the `# TypeScript (experimental, not in CI)` block.

## What we explicitly do NOT touch

- `_lab/ts/` directory contents (per the brainstorming decision — broken scratchpad, documented as experimental).
- CI workflows (`grep -lriE 'typescript|_lab/ts|--lang ts|"ts"|\bts\b' .github/workflows/` returns empty — no references to TS in any workflow).
- The `_lab/ts/package.json` build-script bug (finding 3) and the stale `ThemeConf.json` path (finding 4).

## Testing strategy

1. **Unit + dry-run regression:** `python -m pytest test_make.py -v` — must be green.
2. **Argparse rejection:** `python make.py setup --lang ts` must exit non-zero with `argparse: error: argument --lang: invalid choice: 'ts' (choose from ...)`.
3. **End-to-end on fresh ubuntu:24.04:** repeat the original Docker test from the previous session. Pass criteria:
   - `/.dockerenv` does **not** short-circuit setup (the workaround `rm /.dockerenv` should no longer be needed).
   - `python make.py setup --lang all` exits 0 in a fresh container.
   - `~/.loader-env.json` is populated.
   - `python make.py test --lang go`, `--lang cpp`, `--lang csharp` all pass.

## Risks

- **Hidden TS references** — possible we miss a TS reference in `make.py` or `test_make.py`. Mitigation: final `grep -nE '"ts"|_ts_|node|npm|NODE_VERSION|_lab/ts|_ensure_node|TypeScript' make.py test_make.py` after the edits, expecting empty (or at most comments unrelated to the lab).
- **Devcontainer image rebuild** — removing Node from the Dockerfile invalidates cached layers, causing a one-time slow rebuild for anyone using the devcontainer. Acceptable; documented in CLAUDE.md.
- **Loose end with `_lab/ts/`** — directory remains, but no tooling path drives it. Anyone who tries `cd _lab/ts && npm install && npm test` will hit findings 3 and 4. Acceptable per user decision; CLAUDE.md is updated to no longer claim make.py supports it.
