"""Unit + dry-run integration tests for make.py.

Two layers:

1. **Unit tests** — import make.py directly and exercise pure logic
   (Versions, Platform, _winquote, Runner, etc.). No subprocess, no
   network, no filesystem mutation outside pytest's tmp_path.

2. **Dry-run snapshot tests** — spawn `python3 make.py --dry-run <args>`
   and assert the printed command sequence. Canonical contract test
   for "the orchestrator still emits the right cmake/ctest/buf calls."

Subprocess tests use `--dry-run`, so they never touch the toolchain or
network. The whole suite runs in ~5s on any host.

Run:
    pip install pytest
    python -m pytest test_make.py -v
"""

import json
import subprocess
import sys
from pathlib import Path

import pytest

import make

REPO_ROOT = Path(__file__).resolve().parent
MAKE_PY = REPO_ROOT / "make.py"


# ---------------------------------------------------------------------------
# Unit: Versions
# ---------------------------------------------------------------------------


class TestVersions:
    def test_loads_real_versions_env(self):
        v = make.Versions.load(REPO_ROOT)
        # Every key documented in versions.env must be present.
        for key in (
            "GO_VERSION",
            "BUF_VERSION",
            "DEFAULT_VARIANT",
            "MODERN_PROTOBUF_VERSION",
            "MODERN_VCPKG_BASELINE_COMMIT",
            "LEGACY_V3_PROTOBUF_VERSION",
            "LEGACY_V3_VCPKG_BASELINE_COMMIT",
            "DOTNET_VERSION",
            "CMAKE_VERSION",
        ):
            assert key in v.raw, f"Expected {key} in versions.env"
            assert v.raw[key], f"{key} must not be empty"

    def test_typed_accessors(self):
        v = make.Versions.load(REPO_ROOT)
        assert v.go_version == v.raw["GO_VERSION"]
        assert v.buf_version == v.raw["BUF_VERSION"]
        assert v.dotnet_version == v.raw["DOTNET_VERSION"]
        assert v.cmake_version == v.raw["CMAKE_VERSION"]
        # default_variant is lowercased.
        assert v.default_variant == v.raw["DEFAULT_VARIANT"].lower()
        # protobuf_version / vcpkg_baseline_commit resolve through DEFAULT_VARIANT
        # to the corresponding MODERN_/LEGACY_V3_-prefixed key.
        prefix = v.default_variant.upper().replace("-", "_")
        assert v.protobuf_version == v.raw[f"{prefix}_PROTOBUF_VERSION"]
        assert v.vcpkg_baseline_commit == v.raw[f"{prefix}_VCPKG_BASELINE_COMMIT"]

    def test_variants_enumerates_all_rows(self):
        v = make.Versions.load(REPO_ROOT)
        variants = v.variants()
        # Both rows declared in versions.env must be present and complete.
        assert "modern" in variants
        assert "legacy_v3" in variants
        for label, row in variants.items():
            assert row.get("protobuf_version"), f"{label} missing protobuf_version"
            assert row.get("vcpkg_baseline_commit"), f"{label} missing vcpkg_baseline_commit"

    def test_default_variant_falls_back_to_modern(self):
        # Old versions.env files without DEFAULT_VARIANT default to 'modern'.
        v = make.Versions(raw={})
        assert v.default_variant == "modern"

    def test_variant_value_falls_back_to_unprefixed_key(self):
        # Backward compat: a flat versions.env (legacy schema) still resolves.
        v = make.Versions(raw={"PROTOBUF_VERSION": "9.9.9", "VCPKG_BASELINE_COMMIT": "f" * 40})
        assert v.protobuf_version == "9.9.9"
        assert v.vcpkg_baseline_commit == "f" * 40

    def test_default_variant_label_with_dash_resolves(self):
        # 'legacy-v3' (CI label form) and 'legacy_v3' (env-key form) both work.
        v = make.Versions(raw={
            "DEFAULT_VARIANT": "legacy-v3",
            "LEGACY_V3_PROTOBUF_VERSION": "3.21.12",
            "LEGACY_V3_VCPKG_BASELINE_COMMIT": "a" * 40,
        })
        assert v.protobuf_version == "3.21.12"
        assert v.vcpkg_baseline_commit == "a" * 40

    def test_protobuf_version_looks_like_semver(self):
        v = make.Versions.load(REPO_ROOT)
        # Cheap smoke check — guards against a typo that breaks vcpkg.
        parts = v.protobuf_version.split(".")
        assert 2 <= len(parts) <= 3
        for p in parts:
            assert p.isdigit(), f"Non-numeric protobuf version segment: {p}"

    def test_vcpkg_baseline_is_full_sha(self):
        v = make.Versions.load(REPO_ROOT)
        assert len(v.vcpkg_baseline_commit) == 40
        int(v.vcpkg_baseline_commit, 16)  # raises ValueError if not hex

    def test_ignores_blank_and_comment_lines(self, tmp_path):
        env = tmp_path / ".devcontainer"
        env.mkdir()
        (env / "versions.env").write_text(
            "\n"
            "# this is a comment\n"
            "FOO=bar\n"
            "\n"
            "# another comment\n"
            "BAZ=qux\n",
            encoding="utf-8",
        )
        v = make.Versions.load(tmp_path)
        assert v.raw == {"FOO": "bar", "BAZ": "qux"}

    def test_missing_file_raises(self, tmp_path):
        with pytest.raises(FileNotFoundError):
            make.Versions.load(tmp_path)

    def test_get_with_default(self):
        v = make.Versions(raw={"FOO": "bar"})
        assert v.get("FOO") == "bar"
        assert v.get("MISSING") is None
        assert v.get("MISSING", "default") == "default"


# ---------------------------------------------------------------------------
# Unit: Platform
# ---------------------------------------------------------------------------


class TestPlatform:
    def test_detect_returns_one_os(self):
        p = make.Platform.detect()
        # Exactly one of is_windows / is_macos / is_linux must be True.
        flags = [p.is_windows, p.is_macos, p.is_linux]
        assert sum(flags) == 1, f"Expected exactly one OS flag, got {flags}"

    def test_vcpkg_triplet_windows(self):
        p = make.Platform(sys_platform="win32", machine="amd64", in_devcontainer=False)
        assert p.vcpkg_triplet == "x64-windows-static"

    def test_vcpkg_triplet_macos_x64(self):
        p = make.Platform(
            sys_platform="darwin", machine="x86_64", in_devcontainer=False
        )
        assert p.vcpkg_triplet == "x64-osx"

    def test_vcpkg_triplet_macos_arm64(self):
        p = make.Platform(sys_platform="darwin", machine="arm64", in_devcontainer=False)
        assert p.vcpkg_triplet == "arm64-osx"

    def test_vcpkg_triplet_linux_x64(self):
        p = make.Platform(sys_platform="linux", machine="x86_64", in_devcontainer=False)
        assert p.vcpkg_triplet == "x64-linux"

    def test_vcpkg_triplet_linux_arm64(self):
        p = make.Platform(
            sys_platform="linux", machine="aarch64", in_devcontainer=False
        )
        assert p.vcpkg_triplet == "arm64-linux"

    def test_cmake_toolchain_args_devcontainer_is_empty(self):
        p = make.Platform(sys_platform="linux", machine="x86_64", in_devcontainer=True)
        assert p.cmake_toolchain_args() == []

    def test_cmake_toolchain_args_linux_no_vcpkg_is_empty(self, monkeypatch):
        # Without vcpkg installed (no VCPKG_ROOT), cmake falls back to
        # system protobuf via find_package — no -D flags needed.
        monkeypatch.delenv("VCPKG_ROOT", raising=False)
        p = make.Platform(sys_platform="linux", machine="x86_64", in_devcontainer=False)
        assert p.cmake_toolchain_args() == []

    def test_cmake_toolchain_args_linux_with_vcpkg_emits_flags(self, tmp_path):
        # After `make.py setup --lang cpp` populates ~/vcpkg, native Linux
        # builds also route through vcpkg. Same protobuf pin as Windows.
        p = make.Platform(
            sys_platform="linux",
            machine="x86_64",
            in_devcontainer=False,
            vcpkg_root=tmp_path,
        )
        args = p.cmake_toolchain_args()
        assert any("CMAKE_TOOLCHAIN_FILE" in a for a in args)
        assert any("VCPKG_TARGET_TRIPLET=x64-linux" in a for a in args)

    def test_cmake_toolchain_args_macos_no_vcpkg_is_empty(self, monkeypatch):
        monkeypatch.delenv("VCPKG_ROOT", raising=False)
        p = make.Platform(sys_platform="darwin", machine="arm64", in_devcontainer=False)
        assert p.cmake_toolchain_args() == []

    def test_cmake_toolchain_args_macos_with_vcpkg_emits_flags(self, tmp_path):
        p = make.Platform(
            sys_platform="darwin",
            machine="arm64",
            in_devcontainer=False,
            vcpkg_root=tmp_path,
        )
        args = p.cmake_toolchain_args()
        assert any("CMAKE_TOOLCHAIN_FILE" in a for a in args)
        assert any("VCPKG_TARGET_TRIPLET=arm64-osx" in a for a in args)

    def test_cmake_toolchain_args_windows_with_vcpkg_root(self, tmp_path):
        p = make.Platform(
            sys_platform="win32",
            machine="amd64",
            in_devcontainer=False,
            vcpkg_root=tmp_path,
        )
        args = p.cmake_toolchain_args()
        assert any("CMAKE_TOOLCHAIN_FILE" in a for a in args)
        assert any("VCPKG_TARGET_TRIPLET=x64-windows-static" in a for a in args)
        # Path must include vcpkg.cmake.
        toolchain_arg = next(a for a in args if "CMAKE_TOOLCHAIN_FILE" in a)
        assert toolchain_arg.endswith("vcpkg.cmake")

    def test_cmake_toolchain_args_windows_explicit_triplet(self, tmp_path):
        p = make.Platform(
            sys_platform="win32",
            machine="amd64",
            in_devcontainer=False,
            vcpkg_root=tmp_path,
        )
        args = p.cmake_toolchain_args(triplet="x64-windows")
        assert any("VCPKG_TARGET_TRIPLET=x64-windows" in a for a in args)
        assert not any("VCPKG_TARGET_TRIPLET=x64-windows-static" in a for a in args)

    def test_cmake_toolchain_args_windows_no_vcpkg_root_returns_empty(
        self, monkeypatch
    ):
        # No VCPKG_ROOT in env, no vcpkg_root attr -> empty list (lets cmake
        # fail with a useful "Could not find Protobuf" error).
        monkeypatch.delenv("VCPKG_ROOT", raising=False)
        p = make.Platform(sys_platform="win32", machine="amd64", in_devcontainer=False)
        assert p.cmake_toolchain_args() == []


    def test_detect_ignores_dockerenv_marker(self, monkeypatch):
        """Bug fix: /.dockerenv is created by Docker for *every* container,
        not just the loader devcontainer. The detect() heuristic must NOT
        treat its presence as a devcontainer signal."""
        # Simulate: only /.dockerenv exists; /opt/vcpkg/active does NOT.
        # Use as_posix() so the literal compares correctly on Windows where
        # str(WindowsPath('/.dockerenv')) renders with backslashes.
        def fake_exists(self):
            return self.as_posix() == "/.dockerenv"
        monkeypatch.setattr(make.Path, "exists", fake_exists)
        p = make.Platform.detect()
        assert p.in_devcontainer is False, (
            "/.dockerenv alone must NOT be treated as devcontainer"
        )

    def test_detect_recognizes_vcpkg_active_marker(self, monkeypatch):
        """Positive: /opt/vcpkg/active is the marker the devcontainer's
        Dockerfile actually sets. Its presence is the sole devcontainer signal."""
        def fake_exists(self):
            return self.as_posix() == "/opt/vcpkg/active"
        monkeypatch.setattr(make.Path, "exists", fake_exists)
        p = make.Platform.detect()
        assert p.in_devcontainer is True


# ---------------------------------------------------------------------------
# Unit: Platform.windows_msvc_wrap
# ---------------------------------------------------------------------------


class TestWindowsMsvcWrap:
    """Exercises the central insight of make.py: the cmd-shell-string
    sentinel that bypasses Python's CreateProcess quoting on Windows."""

    def test_passthrough_on_linux(self):
        p = make.Platform(sys_platform="linux", machine="x86_64", in_devcontainer=False)
        cmd = ["cmake", "-S", ".", "-B", "build"]
        assert p.windows_msvc_wrap(cmd) == cmd

    def test_passthrough_on_macos(self):
        p = make.Platform(sys_platform="darwin", machine="arm64", in_devcontainer=False)
        cmd = ["cmake", "-S", ".", "-B", "build"]
        assert p.windows_msvc_wrap(cmd) == cmd

    def test_windows_with_no_vcvarsall_returns_unchanged(self):
        # When MSVC isn't installed, return cmd unchanged so subprocess
        # fails with a useful "cl.exe not found" error rather than a
        # confusing wrapping failure.
        p = make.Platform(
            sys_platform="win32",
            machine="amd64",
            in_devcontainer=False,
            vcvarsall_path=None,
        )
        # Stub locate_vcvarsall to return None.
        original = make.locate_vcvarsall
        try:
            make.locate_vcvarsall = lambda: None
            cmd = ["cmake", "-S", "."]
            assert p.windows_msvc_wrap(cmd) == cmd
        finally:
            make.locate_vcvarsall = original

    def test_windows_with_vcvarsall_emits_sentinel(self, tmp_path):
        fake_vcvars = tmp_path / "vcvarsall.bat"
        fake_vcvars.write_text("rem fake", encoding="utf-8")
        p = make.Platform(
            sys_platform="win32",
            machine="amd64",
            in_devcontainer=False,
            vcvarsall_path=fake_vcvars,
        )
        wrapped = p.windows_msvc_wrap(["buf", "generate", ".."])
        assert len(wrapped) == 1
        line = wrapped[0]
        assert line.startswith(make._WIN_SHELL_MARKER)
        # The shell line must `call` vcvarsall and then `&&` the inner cmd.
        body = line[len(make._WIN_SHELL_MARKER) :]
        assert body.startswith(f'call "{fake_vcvars}" x64')
        assert " >nul && buf generate .." in body

    def test_windows_quotes_paths_with_spaces(self, tmp_path):
        fake_vcvars = tmp_path / "vcvarsall.bat"
        fake_vcvars.write_text("rem fake", encoding="utf-8")
        p = make.Platform(
            sys_platform="win32",
            machine="amd64",
            in_devcontainer=False,
            vcvarsall_path=fake_vcvars,
        )
        wrapped = p.windows_msvc_wrap(["cmake", "-DPATH=C:\\Program Files\\foo"])
        body = wrapped[0][len(make._WIN_SHELL_MARKER) :]
        # The arg with a space MUST be quoted, but the bare cmake should not.
        assert (
            ' cmake "-DPATH=C:\\Program Files\\foo"' in body
            or ' cmake -DPATH="C:\\Program Files\\foo"' in body
            or ' cmake "-DPATH=C:\\Program Files\\foo"' in body
            or '"-DPATH=C:\\Program Files\\foo"' in body
        )  # at least one of these forms


# ---------------------------------------------------------------------------
# Unit: _winquote
# ---------------------------------------------------------------------------


class TestQuoting:
    def test_empty_string(self):
        assert make._winquote("") == '""'

    def test_simple_arg_unchanged(self):
        assert make._winquote("foo") == "foo"
        assert make._winquote("--flag=value") == "--flag=value"

    def test_arg_with_space_is_quoted(self):
        assert make._winquote("hello world") == '"hello world"'

    def test_arg_with_paren_is_quoted(self):
        # The (x86) parser footgun from xxx.bat — must be quoted.
        assert make._winquote("C:\\Program Files (x86)\\foo").startswith('"')

    def test_arg_with_embedded_quote_is_doubled(self):
        # cmd quotes embedded " by doubling.
        assert make._winquote('say "hi"') == '"say ""hi"""'


# ---------------------------------------------------------------------------
# Unit: Runner
# ---------------------------------------------------------------------------


class TestRunner:
    def test_dry_run_does_not_execute(self, capsys, tmp_path):
        runner = make.Runner(verbose=False, dry_run=True)
        # If this were actually executed, it would create a file we can detect.
        marker = tmp_path / "should_not_exist.txt"
        if sys.platform == "win32":
            cmd = ["cmd", "/c", f"echo hi > {marker}"]
        else:
            cmd = ["sh", "-c", f"echo hi > {marker}"]
        rc = runner.run(cmd)
        assert rc == 0
        assert not marker.exists(), "dry-run should not have executed"
        out = capsys.readouterr().out
        assert "[dry-run]" in out

    def test_dry_run_prints_cwd(self, capsys, tmp_path):
        runner = make.Runner(verbose=False, dry_run=True)
        runner.run(["echo", "hi"], cwd=tmp_path)
        out = capsys.readouterr().out
        assert f"cwd={tmp_path}" in out

    def test_sentinel_routes_through_shell(self, capsys):
        runner = make.Runner(verbose=False, dry_run=True)
        sentinel_cmd = [make._WIN_SHELL_MARKER + "echo hello-world"]
        rc = runner.run(sentinel_cmd)
        assert rc == 0
        out = capsys.readouterr().out
        # The marker must be stripped from the printed line.
        assert make._WIN_SHELL_MARKER not in out
        assert "echo hello-world" in out

    def test_rmtree_dry_run_does_not_remove(self, capsys, tmp_path):
        target = tmp_path / "doomed"
        target.mkdir()
        (target / "file.txt").write_text("x")
        runner = make.Runner(verbose=False, dry_run=True)
        runner.rmtree(target)
        assert target.exists()
        out = capsys.readouterr().out
        assert "[dry-run] rm -rf" in out

    def test_rmtree_real_removes(self, tmp_path):
        target = tmp_path / "doomed"
        target.mkdir()
        (target / "file.txt").write_text("x")
        runner = make.Runner(verbose=False, dry_run=False)
        runner.rmtree(target)
        assert not target.exists()

    def test_rmtree_no_op_on_missing(self, tmp_path):
        runner = make.Runner(verbose=False, dry_run=False)
        runner.rmtree(tmp_path / "does-not-exist")  # should not raise

    def test_check_raises_on_failure(self):
        runner = make.Runner(verbose=False, dry_run=False)
        # `python -c "raise SystemExit(7)"` will exit 7.
        with pytest.raises(SystemExit):
            runner.run([sys.executable, "-c", "raise SystemExit(7)"])

    def test_check_false_suppresses_failure(self):
        runner = make.Runner(verbose=False, dry_run=False)
        rc = runner.run([sys.executable, "-c", "raise SystemExit(7)"], check=False)
        assert rc == 7


# ---------------------------------------------------------------------------
# Unit: repo root discovery
# ---------------------------------------------------------------------------


class TestRepoRoot:
    def test_finds_root_from_repo(self):
        assert make.find_repo_root() == REPO_ROOT

    def test_finds_root_from_subdir(self):
        # Walk into a deep subdir and ensure we still locate it.
        deep = REPO_ROOT / "internal" / "options"
        if deep.is_dir():
            assert make.find_repo_root(deep) == REPO_ROOT

    def test_raises_outside_repo(self, tmp_path):
        with pytest.raises(SystemExit):
            make.find_repo_root(tmp_path)


# ---------------------------------------------------------------------------
# Unit: lang directory mapping
# ---------------------------------------------------------------------------


class TestLangDir:
    def test_go(self):
        assert (
            make._lang_dir(REPO_ROOT, "go") == REPO_ROOT / "test" / "go-tableau-loader"
        )

    def test_cpp(self):
        assert (
            make._lang_dir(REPO_ROOT, "cpp")
            == REPO_ROOT / "test" / "cpp-tableau-loader"
        )

    def test_csharp(self):
        assert (
            make._lang_dir(REPO_ROOT, "csharp")
            == REPO_ROOT / "test" / "csharp-tableau-loader"
        )


# ---------------------------------------------------------------------------
# Subprocess helpers
# ---------------------------------------------------------------------------


def run_make(*args: str, cwd: Path = REPO_ROOT) -> subprocess.CompletedProcess:
    """Spawn `python3 make.py <args...>` and capture stdout+stderr."""
    return subprocess.run(
        [sys.executable, str(MAKE_PY), *args],
        cwd=str(cwd),
        capture_output=True,
        text=True,
        check=False,
    )


# ---------------------------------------------------------------------------
# Integration: top-level invocations
# ---------------------------------------------------------------------------


class TestTopLevel:
    def test_help_exits_zero(self):
        proc = run_make("--help")
        assert proc.returncode == 0
        assert "make.py" in proc.stdout
        assert "setup" in proc.stdout
        assert "test" in proc.stdout

    def test_no_args_prints_help(self):
        proc = run_make()
        assert proc.returncode == 0
        assert "usage:" in proc.stdout.lower()

    def test_version_prints_versions_env(self):
        proc = run_make("--version")
        assert proc.returncode == 0
        assert f"make.py {make.MAKE_PY_VERSION}" in proc.stdout
        # Must echo every key from versions.env.
        for key in (
            "GO_VERSION",
            "BUF_VERSION",
            "DEFAULT_VARIANT",
            "MODERN_PROTOBUF_VERSION",
            "MODERN_VCPKG_BASELINE_COMMIT",
            "LEGACY_V3_PROTOBUF_VERSION",
            "LEGACY_V3_VCPKG_BASELINE_COMMIT",
            "DOTNET_VERSION",
            "CMAKE_VERSION",
        ):
            assert key in proc.stdout, f"--version missing {key}"

    def test_env_emits_valid_json(self):
        proc = run_make("env")
        assert proc.returncode == 0
        info = json.loads(proc.stdout)
        # Sanity-check structure.
        assert info["make_py_version"] == make.MAKE_PY_VERSION
        assert "sys_platform" in info
        assert "vcpkg_triplet" in info
        assert "tools" in info
        assert "versions_env" in info
        assert info["versions_env"]["GO_VERSION"]


# ---------------------------------------------------------------------------
# Integration: dry-run snapshots
# ---------------------------------------------------------------------------


class TestDryRunGo:
    def test_test_lang_go(self):
        proc = run_make("--dry-run", "test", "--lang", "go")
        assert proc.returncode == 0
        out = proc.stdout
        assert "buf generate .." in out
        assert "go test" in out
        assert "-timeout" in out and "30m" in out
        # -race default depends on host OS; covered separately in
        # test_test_lang_go_race_default_matches_host.

    def test_test_lang_go_no_race(self):
        proc = run_make("--dry-run", "test", "--lang", "go", "--no-race")
        assert proc.returncode == 0
        # When --no-race is passed, "-race" must NOT appear in the go test line.
        for line in proc.stdout.splitlines():
            if "go test" in line:
                assert "-race" not in line, f"Expected --no-race to drop -race: {line}"

    def test_test_lang_go_race_default_matches_host(self):
        # Default --race depends on host OS:
        #   - Windows -> default OFF (-race needs cgo + C compiler)
        #   - Linux/macOS -> default ON
        proc = run_make("--dry-run", "test", "--lang", "go")
        assert proc.returncode == 0
        go_test_lines = [l for l in proc.stdout.splitlines() if "go test" in l]
        assert go_test_lines, "no go test line"
        joined = "\n".join(go_test_lines)
        if sys.platform == "win32":
            assert "-race" not in joined, "Windows default should NOT include -race"
        else:
            assert "-race" in joined, "Linux/macOS default should include -race"

    def test_test_lang_go_explicit_race_forces_on(self):
        # Even on Windows, --race explicitly enables it (user opt-in).
        proc = run_make("--dry-run", "test", "--lang", "go", "--race")
        assert proc.returncode == 0
        go_test_lines = [l for l in proc.stdout.splitlines() if "go test" in l]
        joined = "\n".join(go_test_lines)
        assert "-race" in joined, "--race must force -race on regardless of host"

    def test_test_lang_go_smoke_skips_buf_generate(self):
        proc = run_make("--dry-run", "test", "--lang", "go", "--smoke")
        assert proc.returncode == 0
        # Smoke must NOT call buf generate; only go vet on the plugin packages.
        assert "buf generate" not in proc.stdout
        assert "go vet" in proc.stdout
        # Specific plugin packages required.
        for pkg in (
            "./cmd/...",
            "./pkg/...",
            "./internal/options/...",
            "./internal/loadutil/...",
            "./internal/xproto/...",
        ):
            assert pkg in proc.stdout, f"smoke missing {pkg}"

    def test_test_lang_go_filter(self):
        proc = run_make(
            "--dry-run", "test", "--lang", "go", "-k", "Test_ActivityConf_OrderedMap"
        )
        assert proc.returncode == 0
        assert "-run" in proc.stdout
        assert "Test_ActivityConf_OrderedMap" in proc.stdout

    def test_test_lang_go_coverage(self):
        proc = run_make("--dry-run", "test", "--lang", "go", "--coverage")
        assert proc.returncode == 0
        assert "-coverprofile=coverage.txt" in proc.stdout
        assert "-covermode=atomic" in proc.stdout


class TestDryRunCpp:
    def test_test_lang_cpp_default(self):
        proc = run_make("--dry-run", "test", "--lang", "cpp")
        assert proc.returncode == 0
        out = proc.stdout
        assert "buf generate .." in out
        assert "cmake -S . -B build" in out
        assert "-DCMAKE_BUILD_TYPE=Debug" in out
        assert "-DCMAKE_CXX_STANDARD=17" in out
        assert "cmake --build build --parallel" in out
        assert "ctest --test-dir build --output-on-failure" in out

    def test_test_lang_cpp_no_clean_skips_wipe(self):
        proc_clean = run_make("--dry-run", "test", "--lang", "cpp")
        proc_no_clean = run_make("--dry-run", "test", "--lang", "cpp", "--no-clean")
        # Clean run must mention rm-rf the build/src dirs; --no-clean must not.
        assert "rm -rf" in proc_clean.stdout
        assert "build" in proc_clean.stdout
        assert "rm -rf" not in proc_no_clean.stdout

    def test_test_lang_cpp_cxx_std_20(self):
        proc = run_make("--dry-run", "test", "--lang", "cpp", "--cxx-std", "20")
        assert proc.returncode == 0
        assert "-DCMAKE_CXX_STANDARD=20" in proc.stdout
        assert "-DCMAKE_CXX_STANDARD=17" not in proc.stdout

    def test_test_lang_cpp_cxx_compiler_clang(self):
        proc = run_make("--dry-run", "test", "--lang", "cpp", "--cxx-compiler", "clang")
        assert proc.returncode == 0
        assert "-DCMAKE_CXX_COMPILER=clang++" in proc.stdout

    def test_test_lang_cpp_protobuf_version_manifest_mode(self):
        proc = run_make(
            "--dry-run",
            "test",
            "--lang",
            "cpp",
            "--protobuf-version",
            "3.21.12",
            "--triplet",
            "x64-windows-static",
            "--no-clean",
        )
        assert proc.returncode == 0
        out = proc.stdout
        # Manifest-mode flags MUST appear when --protobuf-version is set.
        assert "-DVCPKG_INSTALLED_DIR=" in out
        assert "-DVCPKG_MANIFEST_INSTALL=OFF" in out
        # Must also `vcpkg install` the manifest before cmake configure.
        assert "vcpkg" in out and "install" in out
        # Triplet must be plumbed through.
        assert "--triplet=x64-windows-static" in out

    def test_test_lang_cpp_manifest_no_vcpkg_root_dry_run_ok(self, monkeypatch):
        # On a host with no VCPKG_ROOT (e.g. CI runner of testing-make.yml),
        # --dry-run must still print the command sequence rather than abort.
        # Real (non-dry-run) execution still hard-errors via a separate code
        # path; that's covered by the unit test in TestPlatform.
        monkeypatch.delenv("VCPKG_ROOT", raising=False)
        proc = run_make(
            "--dry-run",
            "test",
            "--lang",
            "cpp",
            "--protobuf-version",
            "3.21.12",
            "--triplet",
            "x64-linux",
            "--no-clean",
        )
        assert proc.returncode == 0, proc.stderr
        assert "-DVCPKG_INSTALLED_DIR=" in proc.stdout
        assert "-DVCPKG_MANIFEST_INSTALL=OFF" in proc.stdout

    def test_test_lang_cpp_manifest_forces_toolchain_on_linux(self, monkeypatch):
        # Manifest mode must emit -DCMAKE_TOOLCHAIN_FILE even on Linux:
        # find_package(Protobuf) needs vcpkg's vcpkg.cmake to resolve the
        # manifest-installed protobuf. Plain Linux (no --protobuf-version)
        # would correctly get [] (system protobuf via apt).
        monkeypatch.setenv("VCPKG_ROOT", "/tmp/fake-vcpkg")
        proc = run_make(
            "--dry-run",
            "test",
            "--lang",
            "cpp",
            "--protobuf-version",
            "3.21.12",
            "--triplet",
            "x64-linux",
            "--no-clean",
        )
        assert proc.returncode == 0, proc.stderr
        # Must have toolchain flags for the manifest install to be picked up.
        assert "-DCMAKE_TOOLCHAIN_FILE=" in proc.stdout
        assert "-DVCPKG_TARGET_TRIPLET=x64-linux" in proc.stdout

    def test_test_lang_cpp_no_vcpkg_install_skips_install(self):
        proc = run_make(
            "--dry-run",
            "test",
            "--lang",
            "cpp",
            "--protobuf-version",
            "3.21.12",
            "--triplet",
            "x64-windows-static",
            "--no-clean",
            "--no-vcpkg-install",
        )
        assert proc.returncode == 0
        # No `vcpkg ... install ...` line when --no-vcpkg-install is set.
        # cmake configure still runs, with the manifest-mode flags.
        assert "-DVCPKG_INSTALLED_DIR=" in proc.stdout
        for line in proc.stdout.splitlines():
            # `vcpkg install` must not appear; `cmake` calls are fine.
            if "vcpkg.exe install" in line or "vcpkg install" in line.replace(
                ".exe", ""
            ):
                pytest.fail(f"--no-vcpkg-install did not skip vcpkg install: {line}")

    def test_test_lang_cpp_filter(self):
        proc = run_make(
            "--dry-run", "test", "--lang", "cpp", "-k", "HubTest.Load", "--no-clean"
        )
        assert proc.returncode == 0
        # ctest -R <pattern>
        ctest_lines = [l for l in proc.stdout.splitlines() if "ctest" in l]
        assert ctest_lines, "no ctest line in output"
        joined = "\n".join(ctest_lines)
        assert "-R" in joined and "HubTest.Load" in joined


class TestDryRunCsharp:
    def test_test_lang_csharp_default(self):
        proc = run_make("--dry-run", "test", "--lang", "csharp")
        assert proc.returncode == 0
        assert "buf generate .." in proc.stdout
        assert "dotnet test" in proc.stdout
        assert "--nologo" in proc.stdout

    def test_test_lang_csharp_filter(self):
        proc = run_make("--dry-run", "test", "--lang", "csharp", "-k", "HubTest.Load")
        assert proc.returncode == 0
        # dotnet test --filter "FullyQualifiedName~<x>"
        assert "FullyQualifiedName~HubTest.Load" in proc.stdout


class TestDryRunGenerateAndBuild:
    def test_generate_lang_go(self):
        proc = run_make("--dry-run", "generate", "--lang", "go")
        assert proc.returncode == 0
        assert "buf generate .." in proc.stdout

    def test_generate_lang_cpp(self):
        proc = run_make("--dry-run", "generate", "--lang", "cpp")
        assert proc.returncode == 0
        assert "buf generate .." in proc.stdout

    def test_build_lang_go_no_test(self):
        proc = run_make("--dry-run", "build", "--lang", "go")
        assert proc.returncode == 0
        # build (not test) must call `go build`, not `go test`.
        assert "go build" in proc.stdout
        # Must NOT run tests.
        for line in proc.stdout.splitlines():
            assert "go test" not in line, f"build invoked go test: {line}"

    def test_build_lang_csharp_calls_dotnet_build(self):
        proc = run_make("--dry-run", "build", "--lang", "csharp")
        assert proc.returncode == 0
        assert "dotnet build" in proc.stdout
        for line in proc.stdout.splitlines():
            assert "dotnet test" not in line


class TestDryRunClean:
    def test_clean_cpp(self):
        proc = run_make("--dry-run", "clean", "--lang", "cpp")
        assert proc.returncode == 0
        out = proc.stdout
        # Must wipe all three cpp dirs.
        assert "build" in out
        assert "src\\tableau" in out or "src/tableau" in out
        assert "src\\protoconf" in out or "src/protoconf" in out

    def test_clean_csharp(self):
        proc = run_make("--dry-run", "clean", "--lang", "csharp")
        assert proc.returncode == 0
        out = proc.stdout
        assert "bin" in out
        assert "obj" in out
        assert "protoconf" in out

    def test_clean_all(self):
        proc = run_make("--dry-run", "clean", "--all")
        assert proc.returncode == 0
        out = proc.stdout
        # Should mention dirs from at least cpp, csharp, go.
        assert "cpp-tableau-loader" in out
        assert "csharp-tableau-loader" in out
        assert "go-tableau-loader" in out


# ---------------------------------------------------------------------------
# Integration: setup is a no-op in devcontainer
# ---------------------------------------------------------------------------


class TestCrossPlatformPinning:
    """Unit tests for the macOS/Linux pinning helpers (Go tarball, buf
    binary, vcpkg). Network calls are guarded by --dry-run."""

    def test_go_arch_macos_intel(self, monkeypatch):
        monkeypatch.setattr(make._stdlib_platform, "machine", lambda: "x86_64")
        assert make._go_arch_macos() == "amd64"

    def test_go_arch_macos_apple_silicon(self, monkeypatch):
        monkeypatch.setattr(make._stdlib_platform, "machine", lambda: "arm64")
        assert make._go_arch_macos() == "arm64"

    def test_go_arch_linux_x64(self, monkeypatch):
        monkeypatch.setattr(make._stdlib_platform, "machine", lambda: "x86_64")
        assert make._go_arch_linux() == "amd64"

    def test_go_arch_linux_arm64(self, monkeypatch):
        monkeypatch.setattr(make._stdlib_platform, "machine", lambda: "aarch64")
        assert make._go_arch_linux() == "arm64"

    def test_setup_macos_dispatches_to_vcpkg(self, monkeypatch, capsys):
        # Simulate macOS host with brew available. Dry-run so no real install.
        # The handler's _setup_vcpkg call should print the cloning hint.
        monkeypatch.setattr(make.Platform, "detect", classmethod(
            lambda cls: make.Platform(sys_platform="darwin", machine="x86_64",
                                      in_devcontainer=False)))
        monkeypatch.setattr(make, "_which", lambda name: "/usr/local/bin/brew" if name == "brew" else None)
        ctx = make.Context(
            repo_root=REPO_ROOT,
            versions=make.Versions.load(REPO_ROOT),
            platform=make.Platform.detect(),
            runner=make.Runner(verbose=False, dry_run=True),
        )
        args = type("Args", (), {"lang": "cpp", "skip_vcpkg": False})()
        rc = make.cmd_setup(args, ctx)
        assert rc == 0
        out = capsys.readouterr().out
        # Must mention vcpkg (cross-platform install path).
        assert "vcpkg" in out.lower()

    def test_setup_linux_dispatches_to_vcpkg(self, monkeypatch, capsys):
        monkeypatch.setattr(make.Platform, "detect", classmethod(
            lambda cls: make.Platform(sys_platform="linux", machine="x86_64",
                                      in_devcontainer=False)))
        # Stub _which to make apt-get appear available, brew/dnf absent.
        which_table = {"apt-get": "/usr/bin/apt-get"}
        monkeypatch.setattr(make, "_which", lambda name: which_table.get(name))
        # Stub os.geteuid (not present on Windows test host).
        monkeypatch.setattr(make.os, "geteuid", lambda: 1000, raising=False)
        ctx = make.Context(
            repo_root=REPO_ROOT,
            versions=make.Versions.load(REPO_ROOT),
            platform=make.Platform.detect(),
            runner=make.Runner(verbose=False, dry_run=True),
        )
        args = type("Args", (), {"lang": "cpp", "skip_vcpkg": False})()
        rc = make.cmd_setup(args, ctx)
        assert rc == 0
        out = capsys.readouterr().out
        assert "vcpkg" in out.lower()



    def test_setup_skips_inside_devcontainer(self, monkeypatch, tmp_path):
        # Simulate devcontainer detection by patching Platform.detect.
        original_detect = make.Platform.detect

        @classmethod
        def fake_detect(cls):
            return make.Platform(
                sys_platform="linux",
                machine="x86_64",
                in_devcontainer=True,
            )

        monkeypatch.setattr(make.Platform, "detect", fake_detect)
        try:
            # Build a minimal Context and call cmd_setup directly.
            ctx = make.Context(
                repo_root=REPO_ROOT,
                versions=make.Versions.load(REPO_ROOT),
                platform=make.Platform.detect(),
                runner=make.Runner(verbose=False, dry_run=True),
            )
            args = type("Args", (), {"lang": "all"})()
            rc = make.cmd_setup(args, ctx)
            assert rc == 0
        finally:
            make.Platform.detect = original_detect
