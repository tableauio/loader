#!/usr/bin/env python3
"""
make.py — single cross-platform entrypoint for the tableauio/loader repo.

Consolidates per-language `buf generate` / `cmake` / `go test` / `dotnet test`
/ `npm test` recipes into one Python tool that works identically on
native Windows, macOS, Linux, and inside the devcontainer.

Usage (high level):
    python make.py setup    [--lang go|cpp|csharp|ts|all] [--dry-run]
    python make.py generate --lang go|cpp|csharp|ts
    python make.py build    --lang go|cpp|csharp|ts [build flags]
    python make.py test     --lang go|cpp|csharp|ts [build flags] [-k FILTER] [--smoke]
    python make.py clean    [--lang ...] [--all]
    python make.py env
    python make.py --version

Standard flags (apply to every subcommand):
    --verbose / -v   echo every subprocess
    --dry-run        print but do not execute
    --cwd <path>     repo root (default: auto-detect from versions.env)

The Windows MSVC env trick: every command that needs cl.exe / vcpkg-protoc
runs through Platform.windows_msvc_wrap(), which transparently wraps the
command in `cmd /c "<vcvarsall.bat> x64 && <cmd>"`. The user's shell
PATH/INCLUDE/LIB is never mutated.

Stdlib only — no `pip install` required. Targets Python >= 3.10.
"""

import argparse
import json
import os
import platform as _stdlib_platform
import shlex  # noqa: F401  (kept for potential future POSIX shell quoting)
import shutil
import subprocess
import sys
import urllib.request
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional

MAKE_PY_VERSION = "0.1.0"

# ---------------------------------------------------------------------------
# Versions
# ---------------------------------------------------------------------------


@dataclass
class Versions:
    """Parsed view of .devcontainer/versions.env.

    The format rules (documented in .devcontainer/README.md):
      - One KEY=VALUE per line, no quotes, no spaces around `=`.
      - Comments start with `#` at column 0.
      - Blank lines are ignored.
      - No shell expansion.
    """

    raw: dict[str, str] = field(default_factory=dict)

    @classmethod
    def load(cls, repo_root: Path) -> "Versions":
        path = repo_root / ".devcontainer" / "versions.env"
        if not path.is_file():
            raise FileNotFoundError(
                f"Missing {path}; cannot resolve pinned tool versions."
            )
        raw: dict[str, str] = {}
        for line in path.read_text(encoding="utf-8").splitlines():
            if not line or line.startswith("#"):
                continue
            if "=" not in line:
                continue
            k, v = line.split("=", 1)
            raw[k.strip()] = v.strip()
        return cls(raw=raw)

    def get(self, key: str, default: Optional[str] = None) -> Optional[str]:
        return self.raw.get(key, default)

    @property
    def go_version(self) -> Optional[str]:
        return self.raw.get("GO_VERSION")

    @property
    def buf_version(self) -> Optional[str]:
        return self.raw.get("BUF_VERSION")

    @property
    def protobuf_version(self) -> Optional[str]:
        return self.raw.get("PROTOBUF_VERSION")

    @property
    def vcpkg_baseline_commit(self) -> Optional[str]:
        return self.raw.get("VCPKG_BASELINE_COMMIT")

    @property
    def dotnet_version(self) -> Optional[str]:
        return self.raw.get("DOTNET_VERSION")

    @property
    def node_version(self) -> Optional[str]:
        return self.raw.get("NODE_VERSION")

    @property
    def cmake_version(self) -> Optional[str]:
        return self.raw.get("CMAKE_VERSION")


# ---------------------------------------------------------------------------
# Platform
# ---------------------------------------------------------------------------


@dataclass
class Platform:
    """Host OS / arch / devcontainer detection plus a few helpers.

    The two helpers worth highlighting:

      - cmake_toolchain_args(): returns the right `-DCMAKE_TOOLCHAIN_FILE=...
        -DVCPKG_TARGET_TRIPLET=...` flags on Windows; returns [] inside the
        devcontainer (CMAKE_PREFIX_PATH=/opt/vcpkg/active is preset by the
        Dockerfile) and on macOS / Linux (system protobuf is on PATH).

      - windows_msvc_wrap(cmd): on Windows wraps the command in
        `cmd /c "<vcvarsall> x64 && <cmd>"` so MSVC env lives only in the
        single child process. On all other OSes returns cmd unchanged.
    """

    sys_platform: str
    machine: str
    in_devcontainer: bool
    vcpkg_root: Optional[Path] = None
    vcvarsall_path: Optional[Path] = None
    protoc_tools_dir: Optional[Path] = None
    vcpkg_installed_dir: Optional[Path] = None  # manifest mode only

    @classmethod
    def detect(cls) -> "Platform":
        sys_platform = sys.platform
        machine = _stdlib_platform.machine().lower()
        in_devcontainer = (
            Path("/opt/vcpkg/active").exists() or Path("/.dockerenv").exists()
        )
        return cls(
            sys_platform=sys_platform,
            machine=machine,
            in_devcontainer=in_devcontainer,
        )

    @property
    def is_windows(self) -> bool:
        return self.sys_platform.startswith("win")

    @property
    def is_macos(self) -> bool:
        return self.sys_platform == "darwin"

    @property
    def is_linux(self) -> bool:
        return self.sys_platform.startswith("linux")

    @property
    def vcpkg_triplet(self) -> str:
        """Default vcpkg triplet for the host. Mirrors the Dockerfile lines 32-36."""
        if self.is_windows:
            return "x64-windows-static"
        if self.is_macos:
            if self.machine in ("arm64", "aarch64"):
                return "arm64-osx"
            return "x64-osx"
        if self.is_linux:
            if self.machine in ("arm64", "aarch64"):
                return "arm64-linux"
            return "x64-linux"
        return "x64-linux"

    def cmake_toolchain_args(
        self, triplet: Optional[str] = None, force_vcpkg: bool = False
    ) -> list[str]:
        """Extra cmake -D flags to pick up vcpkg's protobuf, when applicable.

        Default behaviour (force_vcpkg=False):
          - devcontainer:  []  (Dockerfile presets CMAKE_PREFIX_PATH)
          - macOS/Linux:   []  (system protobuf via brew/apt)
          - Windows:       toolchain + triplet flags (always uses vcpkg)

        With force_vcpkg=True, every host (incl. Linux/macOS) gets the
        toolchain flags. Used when --protobuf-version is set, because
        manifest-mode means we ARE using vcpkg regardless of host.
        """
        if self.in_devcontainer and not force_vcpkg:
            return []  # Dockerfile presets CMAKE_PREFIX_PATH=/opt/vcpkg/active.
        if not self.is_windows and not force_vcpkg:
            # macOS/Linux native: system protobuf or homebrew/apt resolves via
            # find_package(Protobuf) without a toolchain file.
            return []
        # Locate VCPKG_ROOT (cached attr or env). Required on Windows always,
        # and on every OS when force_vcpkg=True (manifest mode).
        vcpkg_root = self.vcpkg_root or _env_path("VCPKG_ROOT")
        if vcpkg_root is None:
            # Best-effort fallback so cmake configure fails with a useful
            # message rather than us emitting an empty -D flag.
            return []
        toolchain = vcpkg_root / "scripts" / "buildsystems" / "vcpkg.cmake"
        args = [
            f"-DCMAKE_TOOLCHAIN_FILE={toolchain}",
            f"-DVCPKG_TARGET_TRIPLET={triplet or self.vcpkg_triplet}",
        ]
        return args

    def windows_msvc_wrap(self, cmd: list[str]) -> list[str]:
        """Wrap a command so it runs inside an MSVC-environment subshell.

        On Windows, returns a single-element list containing one cmd-shell
        command string of the form:
            'call "<vcvarsall>" x64 >nul && <quoted cmd>'

        Runner.run() detects the [windows-shell-string] pattern (via the
        first element having no path separators but containing spaces) and
        passes it to subprocess with shell=True so cmd parses it natively
        — bypassing Python's Win32 CreateProcess quoting that otherwise
        backslash-escapes the inner double quotes and breaks cmd's parser.

        On every other OS this returns `cmd` unchanged.
        """
        if not self.is_windows:
            return cmd
        vcvars = self.vcvarsall_path or locate_vcvarsall()
        if vcvars is None:
            # No MSVC available — return cmd unchanged so the caller's
            # subprocess fails with a useful "cl.exe not found" error rather
            # than a more confusing wrapping failure.
            return cmd
        self.vcvarsall_path = vcvars
        inner = " ".join(_winquote(arg) for arg in cmd)
        # Single string: `call "<vcvars>" x64 >nul && <inner>`.
        # `call` is required so vcvarsall returns control to our `&&`.
        # `>nul` swallows vcvarsall's banner (cosmetic).
        line = f'call "{vcvars}" x64 >nul && {inner}'
        return [_WIN_SHELL_MARKER + line]


# Sentinel prefix used to flag a Runner.run() argv as "single shell string,
# please run via cmd". Using a prefix instead of a separate kwarg keeps
# windows_msvc_wrap() composable with the rest of the runner pipeline.
_WIN_SHELL_MARKER = "\x00CMDSHELL\x00"


def _env_path(name: str) -> Optional[Path]:
    v = os.environ.get(name)
    return Path(v) if v else None


def _winquote(arg: str) -> str:
    """Quote an argument for the Windows cmd shell.

    cmd's quoting is famously bad. Wrap in double quotes if there is any
    whitespace, ampersand, or other shell metacharacter.
    """
    if not arg:
        return '""'
    if any(c in arg for c in ' &|<>^"()'):
        # Escape embedded quotes by doubling them (cmd convention).
        return '"' + arg.replace('"', '""') + '"'
    return arg


def locate_vcvarsall() -> Optional[Path]:
    """Find vcvarsall.bat on a Windows host.

    Strategy:
      1. Try vswhere.exe (the canonical method since VS 2017).
      2. Fall back to a few well-known install paths.
    Returns None if MSVC isn't installed.
    """
    if sys.platform != "win32":
        return None
    pf86 = os.environ.get("ProgramFiles(x86)", r"C:\Program Files (x86)")
    pf = os.environ.get("ProgramFiles", r"C:\Program Files")
    for base in (pf86, pf):
        vswhere = Path(base) / "Microsoft Visual Studio" / "Installer" / "vswhere.exe"
        if vswhere.is_file():
            try:
                out = subprocess.run(
                    [
                        str(vswhere),
                        "-latest",
                        "-products",
                        "*",
                        "-requires",
                        "Microsoft.VisualStudio.Component.VC.Tools.x86.x64",
                        "-property",
                        "installationPath",
                    ],
                    capture_output=True,
                    text=True,
                    check=False,
                )
                inst = out.stdout.strip().splitlines()
                if inst:
                    candidate = (
                        Path(inst[0]) / "VC" / "Auxiliary" / "Build" / "vcvarsall.bat"
                    )
                    if candidate.is_file():
                        return candidate
            except OSError:
                pass
    # Hardcoded fallbacks for VS 2022 Build Tools / Community / Enterprise.
    for base in (pf86, pf):
        for edition in ("BuildTools", "Community", "Professional", "Enterprise"):
            candidate = (
                Path(base)
                / "Microsoft Visual Studio"
                / "2022"
                / edition
                / "VC"
                / "Auxiliary"
                / "Build"
                / "vcvarsall.bat"
            )
            if candidate.is_file():
                return candidate
    return None


# ---------------------------------------------------------------------------
# Runner — subprocess helper
# ---------------------------------------------------------------------------


@dataclass
class Runner:
    """Thin wrapper around subprocess.run with verbose / dry-run support."""

    verbose: bool = False
    dry_run: bool = False

    def run(
        self,
        cmd: list[str],
        cwd: Optional[Path] = None,
        env: Optional[dict[str, str]] = None,
        check: bool = True,
        shell: bool = False,
    ) -> int:
        # Detect the windows_msvc_wrap sentinel: a single-element argv whose
        # value starts with _WIN_SHELL_MARKER. Strip the marker and run via
        # cmd's native parser (shell=True) so Python's Win32 CreateProcess
        # quoting doesn't mangle the embedded double quotes.
        if (
            len(cmd) == 1
            and isinstance(cmd[0], str)
            and cmd[0].startswith(_WIN_SHELL_MARKER)
        ):
            line = cmd[0][len(_WIN_SHELL_MARKER) :]
            location = f" (cwd={cwd})" if cwd else ""
            if self.dry_run:
                print(f"[dry-run] {line}{location}")
                return 0
            if self.verbose:
                print(f"[run-shell] {line}{location}")
            proc = subprocess.run(
                line,
                cwd=str(cwd) if cwd else None,
                env=env,
                check=False,
                shell=True,
            )
            if check and proc.returncode != 0:
                raise SystemExit(
                    f"[error] command failed with exit code {proc.returncode}: {line}"
                )
            return proc.returncode

        printable = " ".join(_winquote(c) if " " in c else c for c in cmd)
        location = f" (cwd={cwd})" if cwd else ""
        if self.dry_run:
            print(f"[dry-run] {printable}{location}")
            return 0
        if self.verbose:
            print(f"[run] {printable}{location}")
        proc = subprocess.run(
            cmd,
            cwd=str(cwd) if cwd else None,
            env=env,
            check=False,
            shell=shell,
        )
        if check and proc.returncode != 0:
            raise SystemExit(
                f"[error] command failed with exit code {proc.returncode}: {printable}"
            )
        return proc.returncode

    def rmtree(self, path: Path) -> None:
        if self.dry_run:
            # Always announce the wipe in dry-run, even if path is missing —
            # it's the orchestrator's *intent* we want to capture.
            print(f"[dry-run] rm -rf {path}")
            return
        if not path.exists():
            return
        if self.verbose:
            print(f"[rm-rf] {path}")
        shutil.rmtree(path, ignore_errors=True)

    def mkdirp(self, path: Path) -> None:
        if self.dry_run:
            print(f"[dry-run] mkdir -p {path}")
            return
        if path.exists():
            return
        if self.verbose:
            print(f"[mkdir-p] {path}")
        path.mkdir(parents=True, exist_ok=True)


# ---------------------------------------------------------------------------
# Repo root discovery
# ---------------------------------------------------------------------------


def find_repo_root(start: Optional[Path] = None) -> Path:
    """Locate the repo root by walking upward to find .devcontainer/versions.env."""
    here = (start or Path(__file__).resolve()).parent if start is None else start
    here = here.resolve()
    for candidate in [here, *here.parents]:
        if (candidate / ".devcontainer" / "versions.env").is_file():
            return candidate
    raise SystemExit(
        "[error] Could not locate repo root (no .devcontainer/versions.env found)."
    )


# ---------------------------------------------------------------------------
# ~/.loader-env.json — Windows toolchain cache
# ---------------------------------------------------------------------------


LOADER_ENV_PATH = Path.home() / ".loader-env.json"


def load_loader_env() -> dict:
    if not LOADER_ENV_PATH.is_file():
        return {}
    try:
        return json.loads(LOADER_ENV_PATH.read_text(encoding="utf-8"))
    except (OSError, ValueError):
        return {}


def save_loader_env(data: dict, runner: Runner) -> None:
    text = json.dumps(data, indent=2, sort_keys=True)
    if runner.dry_run:
        print(f"[dry-run] write {LOADER_ENV_PATH}: {text}")
        return
    LOADER_ENV_PATH.write_text(text, encoding="utf-8")


def hydrate_platform_from_env(plat: Platform) -> None:
    """Populate Platform from $VCPKG_ROOT and ~/.loader-env.json (Windows)."""
    if not plat.is_windows:
        return
    if plat.vcpkg_root is None:
        plat.vcpkg_root = _env_path("VCPKG_ROOT")
    cache = load_loader_env()
    if plat.vcpkg_root is None and cache.get("vcpkg_root"):
        plat.vcpkg_root = Path(cache["vcpkg_root"])
    if plat.vcvarsall_path is None and cache.get("vcvarsall_path"):
        candidate = Path(cache["vcvarsall_path"])
        if candidate.is_file():
            plat.vcvarsall_path = candidate
    if plat.protoc_tools_dir is None and cache.get("protoc_tools_dir"):
        plat.protoc_tools_dir = Path(cache["protoc_tools_dir"])
    if plat.vcpkg_installed_dir is None and cache.get("vcpkg_installed_dir"):
        plat.vcpkg_installed_dir = Path(cache["vcpkg_installed_dir"])


# ---------------------------------------------------------------------------
# Setup commands
# ---------------------------------------------------------------------------


LANGS_ALL = ("go", "cpp", "csharp", "ts")


def _which(name: str) -> Optional[str]:
    return shutil.which(name)


def cmd_setup(args, ctx: "Context") -> int:
    """Install host toolchains. OS-dispatched. Idempotent."""
    if ctx.platform.in_devcontainer:
        print("[info] Running inside devcontainer; toolchain already installed.")
        return 0

    langs = _resolve_langs(args.lang)
    print(f"[info] Setting up host toolchain for: {', '.join(langs)}")
    print(f"[info] Pinned versions: {ctx.versions.raw}")

    if ctx.platform.is_macos:
        return _setup_macos(langs, ctx)
    if ctx.platform.is_linux:
        return _setup_linux(langs, ctx)
    if ctx.platform.is_windows:
        return _setup_windows(langs, ctx, skip_vcpkg=getattr(args, "skip_vcpkg", False))
    print(
        f"[error] Unsupported host platform: {ctx.platform.sys_platform}",
        file=sys.stderr,
    )
    return 1


def _resolve_langs(lang: str) -> list[str]:
    if lang in (None, "all"):
        return list(LANGS_ALL)
    return [lang]


def _setup_macos(langs: list[str], ctx: "Context") -> int:
    if _which("brew") is None:
        print(
            "[error] Homebrew is required on macOS. Install from https://brew.sh and re-run.",
            file=sys.stderr,
        )
        return 1
    pkgs: list[str] = []
    if "go" in langs:
        pkgs.append("go")
    if "cpp" in langs:
        pkgs.extend(["protobuf", "cmake", "ninja"])
    if "csharp" in langs:
        # Homebrew dotnet@8 cask covers .NET 8.
        pkgs.append(f"dotnet@{ctx.versions.dotnet_version or '8'}")
    if "ts" in langs:
        pkgs.append(f"node@{ctx.versions.node_version or '20'}")
    pkgs.append("buf")
    pkgs = list(dict.fromkeys(pkgs))  # de-dup, preserving order
    ctx.runner.run(["brew", "update"], check=False)
    ctx.runner.run(["brew", "install", *pkgs], check=False)
    print("[info] macOS toolchain ready.")
    return 0


def _setup_linux(langs: list[str], ctx: "Context") -> int:
    apt = _which("apt-get") is not None
    dnf = _which("dnf") is not None
    if not (apt or dnf):
        print(
            "[error] Neither apt-get nor dnf found. Install your toolchain manually.",
            file=sys.stderr,
        )
        return 1

    base_pkgs: list[str] = []
    if "cpp" in langs:
        if apt:
            base_pkgs.extend(
                [
                    "protobuf-compiler",
                    "libprotobuf-dev",
                    "cmake",
                    "ninja-build",
                    "build-essential",
                    "git",
                ]
            )
        else:
            base_pkgs.extend(
                [
                    "protobuf-compiler",
                    "protobuf-devel",
                    "cmake",
                    "ninja-build",
                    "gcc-c++",
                    "git",
                ]
            )
    if "go" in langs and apt:
        base_pkgs.append("golang")
    if "go" in langs and dnf:
        base_pkgs.append("golang")

    if base_pkgs:
        if apt:
            sudo_prefix = ["sudo"] if os.geteuid() != 0 else []
            ctx.runner.run([*sudo_prefix, "apt-get", "update"], check=False)
            ctx.runner.run(
                [*sudo_prefix, "apt-get", "install", "-y", *base_pkgs], check=False
            )
        else:
            sudo_prefix = ["sudo"] if os.geteuid() != 0 else []
            ctx.runner.run(
                [*sudo_prefix, "dnf", "install", "-y", *base_pkgs], check=False
            )

    # buf, .NET, Node are not always packaged at our pinned version; install
    # via direct download / vendor scripts to ~/.local/.
    if "go" in langs or "cpp" in langs or "csharp" in langs or "ts" in langs:
        _ensure_buf_linux(ctx)
    if "csharp" in langs:
        _ensure_dotnet_linux(ctx)
    if "ts" in langs:
        _ensure_node_linux(ctx)

    print("[info] Linux toolchain ready.")
    return 0


def _ensure_buf_linux(ctx: "Context") -> None:
    if _which("buf") is not None:
        return
    ver = ctx.versions.buf_version
    if not ver:
        return
    arch = (
        "x86_64"
        if _stdlib_platform.machine().lower() in ("x86_64", "amd64")
        else "aarch64"
    )
    url = f"https://github.com/bufbuild/buf/releases/download/v{ver}/buf-Linux-{arch}"
    target_dir = Path.home() / ".local" / "bin"
    ctx.runner.mkdirp(target_dir)
    target = target_dir / "buf"
    print(f"[info] Downloading buf {ver} -> {target}")
    if not ctx.runner.dry_run:
        urllib.request.urlretrieve(url, str(target))
        target.chmod(0o755)


def _ensure_dotnet_linux(ctx: "Context") -> None:
    if _which("dotnet") is not None:
        return
    ver = ctx.versions.dotnet_version or "8.0"
    script = Path.home() / ".local" / "bin" / "dotnet-install.sh"
    ctx.runner.mkdirp(script.parent)
    print(f"[info] Bootstrapping .NET {ver} via dotnet-install.sh")
    if not ctx.runner.dry_run:
        urllib.request.urlretrieve("https://dot.net/v1/dotnet-install.sh", str(script))
        script.chmod(0o755)
    install_dir = Path.home() / ".dotnet"
    ctx.runner.run(
        [str(script), "--channel", ver, "--install-dir", str(install_dir)],
        check=False,
    )


def _ensure_node_linux(ctx: "Context") -> None:
    if _which("node") is not None:
        return
    print("[info] Node not found; install via your distro or NodeSource manually.")


def _setup_windows(langs: list[str], ctx: "Context", skip_vcpkg: bool) -> int:
    """Windows host setup."""
    cache = load_loader_env()

    # Step 0: Chocolatey
    choco = _which("choco")
    if choco is None:
        print("[info] Chocolatey not found. Installing...")
        ctx.runner.run(
            [
                "powershell",
                "-NoProfile",
                "-ExecutionPolicy",
                "Bypass",
                "-Command",
                "[System.Net.ServicePointManager]::SecurityProtocol = "
                "[System.Net.ServicePointManager]::SecurityProtocol -bor 3072; "
                "iex ((New-Object System.Net.WebClient).DownloadString("
                "'https://community.chocolatey.org/install.ps1'))",
            ],
            check=False,
        )
    else:
        print(f"[info] Chocolatey already installed at: {choco}")

    if "cpp" in langs:
        # Step 1: Ninja
        if _which("ninja") is None:
            print(f"[info] Installing ninja via choco...")
            ctx.runner.run(
                ["choco", "install", "ninja", "-y", "--no-progress"], check=False
            )
        else:
            print("[info] ninja already on PATH.")

        # Step 2: CMake
        if _which("cmake") is None:
            ver = ctx.versions.cmake_version or "3.31.8"
            print(f"[info] Installing CMake {ver} via choco...")
            ctx.runner.run(
                [
                    "choco",
                    "install",
                    "cmake",
                    f"--version={ver}",
                    "--installargs",
                    "ADD_CMAKE_TO_PATH=System",
                    "-y",
                    "--no-progress",
                ],
                check=False,
            )
        else:
            print("[info] cmake already on PATH.")

        # Step 3: MSVC Build Tools
        vcvars = locate_vcvarsall()
        if vcvars is None:
            print(
                "[info] MSVC Build Tools not found. Installing visualstudio2022buildtools..."
            )
            ctx.runner.run(
                [
                    "choco",
                    "install",
                    "visualstudio2022buildtools",
                    "--package-parameters",
                    "--add Microsoft.VisualStudio.Workload.VCTools --includeRecommended --passive --locale en-US",
                    "-y",
                ],
                check=False,
            )
            vcvars = locate_vcvarsall()
        if vcvars is not None:
            print(f"[info] vcvarsall.bat: {vcvars}")
            cache["vcvarsall_path"] = str(vcvars)
            ctx.platform.vcvarsall_path = vcvars

    # Step 4: buf
    if _which("buf") is None:
        ver = ctx.versions.buf_version
        if ver:
            buf_dir = (
                Path(os.environ.get("LOCALAPPDATA", str(Path.home()))) / "buf" / "bin"
            )
            ctx.runner.mkdirp(buf_dir)
            buf_exe = buf_dir / "buf.exe"
            url = f"https://github.com/bufbuild/buf/releases/download/v{ver}/buf-Windows-x86_64.exe"
            print(f"[info] Downloading buf {ver} -> {buf_exe}")
            if not ctx.runner.dry_run:
                urllib.request.urlretrieve(url, str(buf_exe))
            print(f"[info] buf installed at {buf_exe}; add to PATH manually if needed.")
    else:
        print("[info] buf already on PATH.")

    # Step 5: vcpkg + protobuf
    if "cpp" in langs and not skip_vcpkg:
        _setup_vcpkg_windows(ctx, cache)

    # Optional: Go / .NET / Node
    if "go" in langs and _which("go") is None:
        ctx.runner.run(
            ["winget", "install", "--id", "GoLang.Go.1.24", "-e"], check=False
        )
    if "csharp" in langs and _which("dotnet") is None:
        ctx.runner.run(
            ["winget", "install", "--id", "Microsoft.DotNet.SDK.8", "-e"], check=False
        )
    if "ts" in langs and _which("node") is None:
        ctx.runner.run(
            ["winget", "install", "--id", "OpenJS.NodeJS.LTS", "-e"], check=False
        )

    save_loader_env(cache, ctx.runner)
    print("[info] Windows toolchain ready.")
    print(
        "[info] Build/test commands run vcvarsall.bat per-process; your shell PATH is unchanged."
    )
    return 0


def _setup_vcpkg_windows(ctx: "Context", cache: dict) -> None:
    triplet = ctx.platform.vcpkg_triplet
    baseline = ctx.versions.vcpkg_baseline_commit

    vcpkg_root = ctx.platform.vcpkg_root or _env_path("VCPKG_ROOT")
    if vcpkg_root is None and cache.get("vcpkg_root"):
        vcpkg_root = Path(cache["vcpkg_root"])

    # Reject manifest-only vcpkg under VS install dir (no bootstrap-vcpkg.bat).
    if vcpkg_root is not None and not (vcpkg_root / "bootstrap-vcpkg.bat").is_file():
        print(f"[warn] {vcpkg_root} looks like a manifest-only vcpkg; ignoring.")
        vcpkg_root = None

    if vcpkg_root is None:
        candidate = Path(os.environ.get("USERPROFILE", str(Path.home()))) / "vcpkg"
        if (candidate / "bootstrap-vcpkg.bat").is_file():
            vcpkg_root = candidate

    if vcpkg_root is None:
        vcpkg_root = Path(os.environ.get("USERPROFILE", str(Path.home()))) / "vcpkg"
        print(f"[info] Cloning vcpkg into {vcpkg_root}...")
        ctx.runner.run(
            ["git", "clone", "https://github.com/microsoft/vcpkg.git", str(vcpkg_root)],
            check=False,
        )
        if baseline:
            ctx.runner.run(
                ["git", "-C", str(vcpkg_root), "fetch", "--quiet", "origin", baseline],
                check=False,
            )
            ctx.runner.run(
                ["git", "-C", str(vcpkg_root), "checkout", "--quiet", baseline],
                check=False,
            )
        # Bootstrap requires MSVC env on Windows.
        bootstrap = vcpkg_root / "bootstrap-vcpkg.bat"
        ctx.runner.run(
            ctx.platform.windows_msvc_wrap([str(bootstrap), "-disableMetrics"]),
            check=False,
        )

    cache["vcpkg_root"] = str(vcpkg_root)
    ctx.platform.vcpkg_root = vcpkg_root

    vcpkg_exe = vcpkg_root / "vcpkg.exe"
    print(f"[info] vcpkg at: {vcpkg_root}")

    print(f"[info] Installing protobuf:{triplet} via vcpkg (classic mode)...")
    ctx.runner.run(
        ctx.platform.windows_msvc_wrap(
            [str(vcpkg_exe), "install", f"protobuf:{triplet}"]
        ),
        check=False,
    )
    protoc_dir = vcpkg_root / "installed" / triplet / "tools" / "protobuf"
    if (protoc_dir / "protoc.exe").is_file():
        cache["protoc_tools_dir"] = str(protoc_dir)
        ctx.platform.protoc_tools_dir = protoc_dir


# ---------------------------------------------------------------------------
# generate / build / test / clean
# ---------------------------------------------------------------------------


def _lang_dir(repo_root: Path, lang: str) -> Path:
    if lang == "ts":
        return repo_root / "_lab" / "ts"
    return repo_root / "test" / f"{lang}-tableau-loader"


def _buf_generate(
    ctx: "Context", lang: str, protoc_dir_override: Optional[Path] = None
) -> None:
    cwd = _lang_dir(ctx.repo_root, lang)
    if lang == "ts":
        # The TypeScript scratchpad has its own `npm run generate` script.
        ctx.runner.run(
            ["npm", "run", "generate"], cwd=cwd, shell=ctx.platform.is_windows
        )
        return
    cmd = ctx.platform.windows_msvc_wrap(["buf", "generate", ".."])
    env = os.environ.copy()
    # Pick the protoc to put on PATH:
    #   - protoc_dir_override (manifest-mode protoc from this build's
    #     vcpkg_installed/) wins — required when --protobuf-version is set,
    #     so codegen matches the libprotobuf headers cmake will use.
    #   - Else the classic-mode protoc cached in ~/.loader-env.json.
    protoc_dir = protoc_dir_override or ctx.platform.protoc_tools_dir
    if protoc_dir is not None and ctx.platform.is_windows:
        env["PATH"] = f"{protoc_dir}{os.pathsep}{env.get('PATH', '')}"
    elif protoc_dir is not None and not ctx.platform.is_windows:
        # On Linux/macOS the manifest-mode protoc dir matters for vcpkg
        # manifest builds too (CI testing-cpp.yml does this on Linux).
        env["PATH"] = f"{protoc_dir}{os.pathsep}{env.get('PATH', '')}"
    ctx.runner.run(cmd, cwd=cwd, env=env)


def cmd_generate(args, ctx: "Context") -> int:
    _buf_generate(ctx, args.lang)
    return 0


def cmd_build(args, ctx: "Context") -> int:
    return _build_or_test(args, ctx, run_tests=False)


def cmd_test(args, ctx: "Context") -> int:
    return _build_or_test(args, ctx, run_tests=True)


def _build_or_test(args, ctx: "Context", run_tests: bool) -> int:
    lang = args.lang
    if lang == "go":
        return _go_build_or_test(args, ctx, run_tests)
    if lang == "cpp":
        return _cpp_build_or_test(args, ctx, run_tests)
    if lang == "csharp":
        return _csharp_build_or_test(args, ctx, run_tests)
    if lang == "ts":
        return _ts_build_or_test(args, ctx, run_tests)
    print(f"[error] unknown --lang {lang}", file=sys.stderr)
    return 2


# ----- Go -----


def _go_build_or_test(args, ctx: "Context", run_tests: bool) -> int:
    cwd = _lang_dir(ctx.repo_root, "go")

    if run_tests and getattr(args, "smoke", False):
        # Devcontainer-smoke equivalent — vet plugin packages only, skip ./test/...
        # No buf-generate needed: those packages don't depend on freshly
        # generated *.pb.go.
        ctx.runner.run(
            [
                "go",
                "vet",
                "./cmd/...",
                "./pkg/...",
                "./internal/options/...",
                "./internal/loadutil/...",
                "./internal/xproto/...",
            ],
            cwd=ctx.repo_root,
        )
        return 0

    # Always regenerate, mirroring CI.
    if not getattr(args, "no_generate", False):
        _buf_generate(ctx, "go")

    if not run_tests:
        ctx.runner.run(["go", "build", "./..."], cwd=cwd)
        return 0

    cmd = ["go", "test", "-v", "-timeout", "30m"]
    # Resolve --race default based on host OS. Windows requires cgo (and a
    # C compiler) for -race; Linux/macOS work out of the box.
    race = getattr(args, "race", None)
    if race is None:
        race = not ctx.platform.is_windows
    if race:
        cmd.append("-race")
    if getattr(args, "coverage", False):
        cmd.extend(["-coverprofile=coverage.txt", "-covermode=atomic"])
    cmd.append("./...")
    if getattr(args, "k", None):
        cmd.extend(["-run", args.k])
    ctx.runner.run(cmd, cwd=cwd)
    return 0


# ----- C++ -----


def _cpp_build_or_test(args, ctx: "Context", run_tests: bool) -> int:
    cwd = _lang_dir(ctx.repo_root, "cpp")
    triplet = getattr(args, "triplet", None) or ctx.platform.vcpkg_triplet
    protobuf_version = getattr(args, "protobuf_version", None)
    cxx_std = getattr(args, "cxx_std", "17")
    cxx_compiler = getattr(args, "cxx_compiler", None)

    # Stale-codegen wipe (gitignored *.pb.* files left over from a previous
    # protoc version shadow fresh codegen). Skip with --no-clean.
    if not getattr(args, "no_clean", False):
        ctx.runner.rmtree(cwd / "build")
        ctx.runner.rmtree(cwd / "src" / "tableau")
        ctx.runner.rmtree(cwd / "src" / "protoconf")

    # Classic mode: a stale vcpkg.json from a previous --protobuf-version run
    # would silently switch cmake's vcpkg toolchain into manifest mode and
    # build the wrong libprotobuf into build/vcpkg_installed/. Always remove
    # it here unless we're about to render a fresh one below.
    if not protobuf_version:
        manifest_path = cwd / "vcpkg.json"
        if manifest_path.is_file():
            if ctx.runner.dry_run:
                print(f"[dry-run] rm {manifest_path}")
            else:
                manifest_path.unlink()

    # Classic mode: a stale vcpkg.json from a previous --protobuf-version run
    # would silently switch cmake's vcpkg toolchain into manifest mode and
    # build the wrong libprotobuf into build/vcpkg_installed/. Always remove
    # it here unless we're about to render a fresh one below.
    if not protobuf_version:
        manifest_path = cwd / "vcpkg.json"
        if manifest_path.is_file():
            if ctx.runner.dry_run:
                print(f"[dry-run] rm {manifest_path}")
            else:
                manifest_path.unlink()

    # Manifest mode: render vcpkg.json pinning the requested protobuf-version,
    # then run `vcpkg install` to populate vcpkg_installed/. This matches CI's
    # testing-cpp.yml flow (which uses lukka/run-vcpkg with runVcpkgInstall:
    # true) and means switching --protobuf-version Just Works without
    # re-running `make.py setup`. Idempotent: vcpkg detects already-installed
    # packages and skips them.
    cmake_extra: list[str] = []
    if protobuf_version:
        baseline = ctx.versions.vcpkg_baseline_commit or ""
        manifest = {
            "name": "loader-cpp-test",
            "version": "0.1.0",
            "dependencies": ["protobuf"],
            "overrides": [{"name": "protobuf", "version": protobuf_version}],
            "builtin-baseline": baseline,
        }
        manifest_path = cwd / "vcpkg.json"
        if not ctx.runner.dry_run:
            manifest_path.write_text(json.dumps(manifest, indent=2), encoding="utf-8")
        installed_dir = Path(
            os.environ.get("VCPKG_INSTALLED_DIR", str(cwd / "vcpkg_installed"))
        )

        # Locate vcpkg.exe / vcpkg. On Windows we hydrated VCPKG_ROOT from
        # ~/.loader-env.json; on CI it's set by lukka/run-vcpkg; on Linux
        # it's the system or devcontainer vcpkg.
        vcpkg_root = ctx.platform.vcpkg_root or _env_path("VCPKG_ROOT")
        if vcpkg_root is None:
            if ctx.runner.dry_run:
                # Dry-run: print what would happen with a placeholder so the
                # snapshot test can still verify the command sequence.
                vcpkg_root = Path("<VCPKG_ROOT>")
            else:
                print(
                    "[error] --protobuf-version requires VCPKG_ROOT to be set "
                    "(run `python make.py setup --lang cpp` first, or set "
                    "VCPKG_ROOT in your environment).",
                    file=sys.stderr,
                )
                return 1
        # Surface the resolved root on the platform so cmake_toolchain_args
        # picks it up for the configure command.
        ctx.platform.vcpkg_root = vcpkg_root
        vcpkg_exe = vcpkg_root / ("vcpkg.exe" if ctx.platform.is_windows else "vcpkg")

        # Install the manifest: must `cd` into the manifest dir for vcpkg to
        # discover it. Skip when --no-vcpkg-install is passed (CI's
        # lukka/run-vcpkg already did it).
        if not getattr(args, "no_vcpkg_install", False):
            install_cmd = [
                str(vcpkg_exe),
                "install",
                f"--triplet={triplet}",
                f"--x-install-root={installed_dir}",
            ]
            ctx.runner.run(
                ctx.platform.windows_msvc_wrap(install_cmd),
                cwd=cwd,
            )

        cmake_extra.extend(
            [
                f"-DVCPKG_INSTALLED_DIR={installed_dir}",
                "-DVCPKG_MANIFEST_INSTALL=OFF",
            ]
        )

    if not getattr(args, "no_generate", False):
        # In manifest mode, route buf-generate's protoc to the manifest's
        # tools dir so codegen matches the libprotobuf cmake links against.
        protoc_dir_override: Optional[Path] = None
        if protobuf_version:
            protoc_dir_override = installed_dir / triplet / "tools" / "protobuf"
        _buf_generate(ctx, "cpp", protoc_dir_override=protoc_dir_override)

    configure_cmd = [
        "cmake",
        "-S",
        ".",
        "-B",
        "build",
        "-DCMAKE_BUILD_TYPE=Debug",
        f"-DCMAKE_CXX_STANDARD={cxx_std}",
    ]
    if cxx_compiler:
        # Translate friendly names.
        compiler = {"msvc": "cl", "clang": "clang++", "gcc": "g++"}.get(
            cxx_compiler, cxx_compiler
        )
        configure_cmd.append(f"-DCMAKE_CXX_COMPILER={compiler}")
    if shutil.which("ninja") is not None:
        configure_cmd.extend(["-G", "Ninja"])
    configure_cmd.extend(
        ctx.platform.cmake_toolchain_args(
            triplet=triplet,
            # Manifest mode means we ARE using vcpkg regardless of host;
            # force the toolchain flags even on Linux/macOS so cmake's
            # find_package(Protobuf) resolves against vcpkg_installed/.
            force_vcpkg=bool(protobuf_version),
        )
    )
    configure_cmd.extend(cmake_extra)

    ctx.runner.run(ctx.platform.windows_msvc_wrap(configure_cmd), cwd=cwd)
    ctx.runner.run(
        ctx.platform.windows_msvc_wrap(["cmake", "--build", "build", "--parallel"]),
        cwd=cwd,
    )

    if run_tests:
        ctest_cmd = ["ctest", "--test-dir", "build", "--output-on-failure"]
        if getattr(args, "k", None):
            ctest_cmd.extend(["-R", args.k])
        ctx.runner.run(ctx.platform.windows_msvc_wrap(ctest_cmd), cwd=cwd)
    return 0


# ----- C# -----


def _csharp_build_or_test(args, ctx: "Context", run_tests: bool) -> int:
    cwd = _lang_dir(ctx.repo_root, "csharp")
    if not getattr(args, "no_generate", False):
        _buf_generate(ctx, "csharp")

    if not run_tests:
        ctx.runner.run(["dotnet", "build", "--nologo"], cwd=cwd)
        return 0

    cmd = ["dotnet", "test", "--nologo", "--logger", "console;verbosity=normal"]
    if getattr(args, "k", None):
        cmd.extend(["--filter", f"FullyQualifiedName~{args.k}"])
    ctx.runner.run(cmd, cwd=cwd)
    return 0


# ----- TypeScript -----


def _ts_build_or_test(args, ctx: "Context", run_tests: bool) -> int:
    cwd = _lang_dir(ctx.repo_root, "ts")
    if not (cwd / "node_modules").is_dir():
        ctx.runner.run(["npm", "install"], cwd=cwd, shell=ctx.platform.is_windows)
    if not getattr(args, "no_generate", False):
        ctx.runner.run(
            ["npm", "run", "generate"], cwd=cwd, shell=ctx.platform.is_windows
        )
    if run_tests:
        ctx.runner.run(["npm", "run", "test"], cwd=cwd, shell=ctx.platform.is_windows)
    return 0


# ----- clean / env -----


def cmd_clean(args, ctx: "Context") -> int:
    targets: list[str] = []
    if args.all:
        targets = list(LANGS_ALL)
    else:
        targets = _resolve_langs(args.lang)
    for lang in targets:
        cwd = _lang_dir(ctx.repo_root, lang)
        if lang == "cpp":
            ctx.runner.rmtree(cwd / "build")
            ctx.runner.rmtree(cwd / "src" / "tableau")
            ctx.runner.rmtree(cwd / "src" / "protoconf")
        elif lang == "csharp":
            ctx.runner.rmtree(cwd / "bin")
            ctx.runner.rmtree(cwd / "obj")
            ctx.runner.rmtree(cwd / "protoconf")
        elif lang == "go":
            ctx.runner.rmtree(cwd / "protoconf")
        elif lang == "ts":
            ctx.runner.rmtree(cwd / "node_modules")
            ctx.runner.rmtree(cwd / "dist")
    return 0


def cmd_env(args, ctx: "Context") -> int:
    info = {
        "make_py_version": MAKE_PY_VERSION,
        "repo_root": str(ctx.repo_root),
        "sys_platform": ctx.platform.sys_platform,
        "machine": ctx.platform.machine,
        "in_devcontainer": ctx.platform.in_devcontainer,
        "vcpkg_triplet": ctx.platform.vcpkg_triplet,
        "vcpkg_root": str(ctx.platform.vcpkg_root) if ctx.platform.vcpkg_root else None,
        "vcvarsall_path": (
            str(ctx.platform.vcvarsall_path) if ctx.platform.vcvarsall_path else None
        ),
        "protoc_tools_dir": (
            str(ctx.platform.protoc_tools_dir)
            if ctx.platform.protoc_tools_dir
            else None
        ),
        "tools": {
            "go": _which("go"),
            "buf": _which("buf"),
            "protoc": _which("protoc"),
            "cmake": _which("cmake"),
            "ninja": _which("ninja"),
            "dotnet": _which("dotnet"),
            "node": _which("node"),
            "npm": _which("npm"),
        },
        "versions_env": ctx.versions.raw,
    }
    print(json.dumps(info, indent=2))
    return 0


# ---------------------------------------------------------------------------
# Context + arg parsing + main
# ---------------------------------------------------------------------------


@dataclass
class Context:
    repo_root: Path
    versions: Versions
    platform: Platform
    runner: Runner


def build_arg_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(
        prog="make.py",
        description="Cross-platform build/test driver for tableauio/loader.",
    )
    p.add_argument(
        "--version", action="store_true", help="print make.py version + versions.env"
    )
    p.add_argument("-v", "--verbose", action="store_true")
    p.add_argument("--dry-run", action="store_true")
    p.add_argument("--cwd", type=str, default=None, help="repo root override")

    sub = p.add_subparsers(dest="command")

    # setup
    sp = sub.add_parser("setup", help="Install host toolchain")
    sp.add_argument("--lang", choices=[*LANGS_ALL, "all"], default="all")
    sp.add_argument(
        "--skip-vcpkg",
        action="store_true",
        help="(Windows) skip vcpkg install (CI uses lukka/run-vcpkg)",
    )

    # generate
    sp = sub.add_parser("generate", help="Run buf generate for a language")
    sp.add_argument("--lang", choices=LANGS_ALL, required=True)

    # build
    sp = sub.add_parser("build", help="Compile generated code for a language")
    _add_build_flags(sp)

    # test
    sp = sub.add_parser("test", help="Run tests for a language")
    _add_build_flags(sp)
    sp.add_argument(
        "-k",
        type=str,
        default=None,
        help="test filter (go -run / ctest -R / dotnet --filter)",
    )
    sp.add_argument(
        "--smoke", action="store_true", help="(go only) smoke vet, no full test run"
    )
    # --race / --no-race: tri-state. Default depends on host OS:
    #   Linux/macOS -> default ON (-race works out of the box).
    #   Windows     -> default OFF (-race needs cgo+C compiler; users opt in
    #                  explicitly via --race once they have MSVC/MinGW).
    sp.add_argument(
        "--race",
        dest="race",
        action="store_true",
        default=None,
        help="enable -race (default on Linux/macOS, off on Windows)",
    )
    sp.add_argument(
        "--no-race", dest="race", action="store_false", help="disable -race"
    )
    sp.add_argument("--coverage", action="store_true")

    # clean
    sp = sub.add_parser("clean", help="Wipe generated code + build outputs")
    sp.add_argument("--lang", choices=[*LANGS_ALL, "all"], default="all")
    sp.add_argument("--all", action="store_true")

    # env
    sub.add_parser("env", help="Print resolved environment as JSON")

    return p


def _add_build_flags(sp: argparse.ArgumentParser) -> None:
    sp.add_argument("--lang", choices=LANGS_ALL, required=True)
    sp.add_argument("--cxx-std", choices=["17", "20"], default="17")
    sp.add_argument("--cxx-compiler", choices=["msvc", "clang", "gcc"], default=None)
    sp.add_argument(
        "--protobuf-version",
        type=str,
        default=None,
        help="(cpp) pin vcpkg protobuf port to this version (manifest mode)",
    )
    sp.add_argument(
        "--triplet", type=str, default=None, help="(cpp) vcpkg triplet override"
    )
    sp.add_argument(
        "--no-clean",
        action="store_true",
        help="(cpp) skip pre-build wipe of build/ + generated codegen",
    )
    sp.add_argument(
        "--no-vcpkg-install",
        action="store_true",
        help="(cpp manifest mode) skip `vcpkg install` (CI uses lukka/run-vcpkg)",
    )
    sp.add_argument(
        "--no-generate", action="store_true", help="skip the buf-generate step"
    )


def main(argv: Optional[list[str]] = None) -> int:
    parser = build_arg_parser()
    args = parser.parse_args(argv)

    if args.version:
        repo = (
            find_repo_root(Path(args.cwd).resolve()) if args.cwd else find_repo_root()
        )
        v = Versions.load(repo)
        print(f"make.py {MAKE_PY_VERSION}")
        for k, val in v.raw.items():
            print(f"  {k}={val}")
        return 0

    if args.command is None:
        parser.print_help()
        return 0

    repo = find_repo_root(Path(args.cwd).resolve()) if args.cwd else find_repo_root()
    versions = Versions.load(repo)
    plat = Platform.detect()
    hydrate_platform_from_env(plat)
    runner = Runner(verbose=args.verbose, dry_run=args.dry_run)
    ctx = Context(repo_root=repo, versions=versions, platform=plat, runner=runner)

    dispatch = {
        "setup": cmd_setup,
        "generate": cmd_generate,
        "build": cmd_build,
        "test": cmd_test,
        "clean": cmd_clean,
        "env": cmd_env,
    }
    handler = dispatch.get(args.command)
    if handler is None:
        parser.print_help()
        return 2
    return handler(args, ctx)


if __name__ == "__main__":
    raise SystemExit(main())
