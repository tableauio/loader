@echo off
setlocal enabledelayedexpansion

REM ===========================================================================
REM prepare.bat — bootstrap a Windows build environment for the C++ loader.
REM
REM Installs (only if missing): Chocolatey, Ninja, CMake 3.31.8, MSVC Build
REM Tools (Visual Studio 2022 Build Tools), buf CLI, and vcpkg.
REM
REM Then installs `protobuf` (and friends) into vcpkg using the static-CRT
REM triplet x64-windows-static, so that downstream cmake builds can pick it
REM up via -DCMAKE_TOOLCHAIN_FILE=%VCPKG_ROOT%\scripts\buildsystems\vcpkg.cmake.
REM
REM The vcpkg checkout is pinned to VCPKG_BASELINE_COMMIT below for
REM reproducibility — the same commit testing-cpp.yml uses in CI.
REM
REM Override `protobuf` to a specific vcpkg port version with PROTOBUF_VCPKG_VERSION:
REM   set PROTOBUF_VCPKG_VERSION=3.21.12 && .\prepare.bat
REM
REM When PROTOBUF_VCPKG_VERSION is set we install in vcpkg MANIFEST mode (a
REM rendered vcpkg.json under %LOCALAPPDATA%\loader\vcpkg-manifest\), which
REM is the only mode where `--x-version` / `overrides` actually pin the port
REM version. In that case the install root is %VCPKG_INSTALLED_DIR%, and
REM downstream cmake invocations must add:
REM     -DVCPKG_INSTALLED_DIR=%VCPKG_INSTALLED_DIR% -DVCPKG_MANIFEST_INSTALL=OFF
REM (matching the CI flow in .github/workflows/testing-cpp.yml).
REM
REM When PROTOBUF_VCPKG_VERSION is NOT set we install in vcpkg CLASSIC mode
REM (no manifest), and downstream cmake works with just the toolchain file.
REM
REM This script is idempotent: re-running it on a machine that already has
REM everything installed is a no-op (a few seconds of probing). Only the MSVC
REM environment variables are re-exported each time, since vcvarsall.bat sets
REM cmd-session-local state that does not persist.
REM ===========================================================================

REM -----------------------------------------------------------------------
REM Parse arguments
REM   --dry-run        : print what would be done, but do not install anything
REM   --simulate-clean : pretend nothing is installed (implies --dry-run)
REM -----------------------------------------------------------------------
set "DRY_RUN=0"
set "SIMULATE_CLEAN=0"
for %%A in (%*) do (
    if /i "%%A"=="--dry-run"        set "DRY_RUN=1"
    if /i "%%A"=="--simulate-clean" set "DRY_RUN=1" & set "SIMULATE_CLEAN=1"
)
if "%DRY_RUN%"=="1"        echo [DRY-RUN] No changes will be made to the system.
if "%SIMULATE_CLEAN%"=="1" echo [DRY-RUN] Simulating a clean machine (all tools treated as not installed).

echo [INFO] Preparing build environment...

REM -----------------------------------------------------------------------
REM Step 0: Ensure Chocolatey is installed
REM -----------------------------------------------------------------------
set "CHOCO_EXE="
set "CHOCO_BASE="
if "%SIMULATE_CLEAN%"=="0" (
    REM Try env var first, then fall back to registry (HKCU then HKLM)
    if defined ChocolateyInstall set "CHOCO_BASE=%ChocolateyInstall%"
    if not defined CHOCO_BASE (
        for /f "usebackq tokens=2*" %%a in (`reg query "HKCU\Environment" /v ChocolateyInstall 2^>nul`) do set "CHOCO_BASE=%%b"
    )
    if not defined CHOCO_BASE (
        for /f "usebackq tokens=2*" %%a in (`reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v ChocolateyInstall 2^>nul`) do set "CHOCO_BASE=%%b"
    )
    if not defined CHOCO_BASE set "CHOCO_BASE=%ALLUSERSPROFILE%\chocolatey"
    if exist "!CHOCO_BASE!\bin\choco.exe"       set "CHOCO_EXE=!CHOCO_BASE!\bin\choco.exe"
    if exist "!CHOCO_BASE!\redirects\choco.exe" set "CHOCO_EXE=!CHOCO_BASE!\redirects\choco.exe"
    if exist "!CHOCO_BASE!\tools\choco.exe"     set "CHOCO_EXE=!CHOCO_BASE!\tools\choco.exe"
)
if not defined CHOCO_EXE (
    echo [INFO] Chocolatey not found. Installing Chocolatey...
    if "%DRY_RUN%"=="0" (
        powershell -NoProfile -ExecutionPolicy Bypass -Command ^
            "[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))"
        if errorlevel 1 (
            echo [ERROR] Failed to install Chocolatey.
            exit /b 1
        )
    ) else (
        echo [DRY-RUN] Would run: powershell ... install Chocolatey
    )
    REM Add Chocolatey to current session PATH
    set "PATH=%ALLUSERSPROFILE%\chocolatey\bin;%PATH%"
    REM Persist Chocolatey bin to user PATH permanently
    if "%DRY_RUN%"=="0" (
        for /f "usebackq tokens=2*" %%a in (`reg query "HKCU\Environment" /v PATH 2^>nul`) do set "USR_PATH=%%b"
        echo !USR_PATH! | findstr /i /c:"%ALLUSERSPROFILE%\chocolatey\bin" >nul 2>&1
        if errorlevel 1 (
            setx PATH "%ALLUSERSPROFILE%\chocolatey\bin;!USR_PATH!"
            echo [INFO] Chocolatey bin added to user PATH permanently.
        )
    ) else (
        echo [DRY-RUN] Would run: setx PATH "%%ALLUSERSPROFILE%%\chocolatey\bin;..."
    )
    echo [INFO] Chocolatey installed successfully.
) else (
    echo [INFO] Chocolatey already installed.
)

REM Refresh ChocolateyInstall var if it was just installed (also read from registry)
if not defined ChocolateyInstall (
    for /f "usebackq tokens=2*" %%a in (`reg query "HKCU\Environment" /v ChocolateyInstall 2^>nul`) do set "ChocolateyInstall=%%b"
)
if not defined ChocolateyInstall set "ChocolateyInstall=%ALLUSERSPROFILE%\chocolatey"
if "%SIMULATE_CLEAN%"=="0" (
    set "PATH=%ChocolateyInstall%\bin;%ChocolateyInstall%\lib\ninja\tools;%PATH%"
)

REM -----------------------------------------------------------------------
REM Step 1: Ensure Ninja is installed via Chocolatey
REM -----------------------------------------------------------------------
set "NINJA_FOUND=0"
if "%SIMULATE_CLEAN%"=="0" (
    where ninja.exe >nul 2>&1
    if not errorlevel 1 set "NINJA_FOUND=1"
)
if "%NINJA_FOUND%"=="0" (
    echo [INFO] ninja.exe not found. Installing via choco...
    if "%DRY_RUN%"=="0" (
        choco install ninja -y --no-progress
        if errorlevel 1 (
            echo [ERROR] Failed to install ninja.
            exit /b 1
        )
    ) else (
        echo [DRY-RUN] Would run: choco install ninja -y --no-progress
    )
    REM Add ninja to current session PATH
    if defined ChocolateyInstall (
        set "NINJA_PATH=!ChocolateyInstall!\lib\ninja\tools"
    ) else (
        set "NINJA_PATH=%ALLUSERSPROFILE%\chocolatey\lib\ninja\tools"
    )
    set "PATH=!NINJA_PATH!;%PATH%"
    REM Persist ninja path to user PATH permanently
    if "%DRY_RUN%"=="0" (
        for /f "usebackq tokens=2*" %%a in (`reg query "HKCU\Environment" /v PATH 2^>nul`) do set "USR_PATH=%%b"
        echo !USR_PATH! | findstr /i /c:"ninja\tools" >nul 2>&1
        if errorlevel 1 (
            setx PATH "!NINJA_PATH!;!USR_PATH!"
            echo [INFO] ninja path added to user PATH permanently.
        )
    ) else (
        echo [DRY-RUN] Would run: setx PATH "!NINJA_PATH!;..."
    )
    echo [INFO] ninja installed successfully.
) else (
    echo [INFO] ninja.exe already in PATH.
)

REM -----------------------------------------------------------------------
REM Step 2: Ensure CMake 3.31.8 is installed
REM         Try Chocolatey first; fall back to direct MSI download.
REM -----------------------------------------------------------------------
set "CMAKE_FOUND=0"
if "%SIMULATE_CLEAN%"=="0" (
    where cmake.exe >nul 2>&1
    if not errorlevel 1 set "CMAKE_FOUND=1"
)
if "%CMAKE_FOUND%"=="0" (
    echo [INFO] cmake.exe not found. Installing CMake 3.31.8...
    if "%DRY_RUN%"=="0" (
        set "CMAKE_INSTALLED=0"
        REM --- Attempt 1: Chocolatey ---
        choco install cmake --version=3.31.8 --installargs "'ADD_CMAKE_TO_PATH=System'" -y --no-progress >nul 2>&1 && set "CMAKE_INSTALLED=1"
        if "!CMAKE_INSTALLED!"=="0" (
            echo [WARN] choco install cmake failed. Falling back to direct MSI download...
            set "CMAKE_MSI=%TEMP%\cmake-3.31.8-windows-x86_64.msi"
            powershell -NoProfile -Command "(New-Object Net.WebClient).DownloadFile('https://github.com/Kitware/CMake/releases/download/v3.31.8/cmake-3.31.8-windows-x86_64.msi','!CMAKE_MSI!')"
            if not exist "!CMAKE_MSI!" (
                echo [ERROR] Failed to download CMake MSI.
                exit /b 1
            )
            msiexec /i "!CMAKE_MSI!" ADD_CMAKE_TO_PATH=System /quiet /norestart
            if errorlevel 1 (
                echo [ERROR] Failed to install CMake from MSI.
                exit /b 1
            )
            del /q "!CMAKE_MSI!" 2>nul
        )
    ) else (
        echo [DRY-RUN] Would run: choco install cmake --version=3.31.8 ... (or fallback to MSI download)
    )
    REM Add cmake to current session PATH
    set "CMAKE_PATH=C:\Program Files\CMake\bin"
    set "PATH=!CMAKE_PATH!;%PATH%"
    REM Persist cmake path to user PATH permanently
    if "%DRY_RUN%"=="0" (
        for /f "usebackq tokens=2*" %%a in (`reg query "HKCU\Environment" /v PATH 2^>nul`) do set "USR_PATH=%%b"
        echo !USR_PATH! | findstr /i /c:"CMake\bin" >nul 2>&1
        if errorlevel 1 (
            setx PATH "!CMAKE_PATH!;!USR_PATH!"
            echo [INFO] cmake path added to user PATH permanently.
        )
    ) else (
        echo [DRY-RUN] Would run: setx PATH "!CMAKE_PATH!;..."
    )
    echo [INFO] cmake installed successfully.
) else (
    echo [INFO] cmake.exe already in PATH.
)

REM -----------------------------------------------------------------------
REM Step 3: Ensure MSVC compiler (cl.exe) is available, then activate its
REM         environment for this cmd session via vcvarsall.bat. The CI
REM         workflow uses ilammy/msvc-dev-cmd@v1 to do the same thing.
REM -----------------------------------------------------------------------
set "CL_FOUND=0"
if "%SIMULATE_CLEAN%"=="0" (
    where cl.exe >nul 2>&1
    if not errorlevel 1 set "CL_FOUND=1"
)
set "SKIP_MSVC=0"
if "%CL_FOUND%"=="0" (
    echo [INFO] cl.exe not found. Searching for existing VS installation...
    set "VSWHERE="
    if "%SIMULATE_CLEAN%"=="0" (
        for %%d in ("%ProgramFiles(x86)%" "%ProgramFiles%") do (
            if not defined VSWHERE (
                if exist "%%~d\Microsoft Visual Studio\Installer\vswhere.exe" (
                    set "VSWHERE=%%~d\Microsoft Visual Studio\Installer\vswhere.exe"
                )
            )
        )
    )
    if not defined VSWHERE (
        echo [INFO] Visual Studio not found. Installing via choco...
        if "%DRY_RUN%"=="0" (
            choco install visualstudio2022buildtools --package-parameters "--add Microsoft.VisualStudio.Workload.VCTools --includeRecommended --passive --locale en-US" -y
            if errorlevel 1 (
                echo [ERROR] Failed to install Visual Studio Build Tools.
                exit /b 1
            )
            echo [INFO] Visual Studio Build Tools installed successfully.
            REM Re-search vswhere after installation
            for %%d in ("%ProgramFiles(x86)%" "%ProgramFiles%") do (
                if not defined VSWHERE (
                    if exist "%%~d\Microsoft Visual Studio\Installer\vswhere.exe" (
                        set "VSWHERE=%%~d\Microsoft Visual Studio\Installer\vswhere.exe"
                    )
                )
            )
        ) else (
            echo [DRY-RUN] Would run: choco install visualstudio2022buildtools ...
            echo [DRY-RUN] Would search vswhere.exe after installation.
            set "SKIP_MSVC=1"
        )
    )
    if "!SKIP_MSVC!"=="0" (
        if not defined VSWHERE (
            echo [ERROR] vswhere.exe still not found after installation. Please restart and retry.
            exit /b 1
        )
        set "VCVARSALL="
        for /f "usebackq delims=" %%p in (`"!VSWHERE!" -latest -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath`) do (
            set "VCVARSALL=%%p\VC\Auxiliary\Build\vcvarsall.bat"
        )
        if not defined VCVARSALL (
            echo [ERROR] No VS installation with C++ tools detected.
            exit /b 1
        )
        if not exist "!VCVARSALL!" (
            echo [ERROR] vcvarsall.bat not found at: !VCVARSALL!
            exit /b 1
        )
        echo [INFO] Initializing MSVC environment from: !VCVARSALL!
        call "!VCVARSALL!" x64
    )
) else (
    echo [INFO] cl.exe already in PATH, skipping MSVC environment setup.
)

REM -----------------------------------------------------------------------
REM Step 4: Ensure buf CLI is installed
REM         The CI workflow uses bufbuild/buf-action@v1 (also pinned to
REM         BUF_VERSION below) to do the same thing.
REM         buf is a single self-contained .exe; install it under
REM         %LOCALAPPDATA%\buf\bin\buf.exe to avoid requiring admin rights.
REM -----------------------------------------------------------------------
set "BUF_VERSION=1.67.0"
set "BUF_FOUND=0"
if "%SIMULATE_CLEAN%"=="0" (
    where buf.exe >nul 2>&1
    if not errorlevel 1 set "BUF_FOUND=1"
)
if "%BUF_FOUND%"=="0" (
    echo [INFO] buf.exe not found. Installing buf %BUF_VERSION%...
    set "BUF_DIR=%LOCALAPPDATA%\buf\bin"
    set "BUF_EXE=!BUF_DIR!\buf.exe"
    set "BUF_URL=https://github.com/bufbuild/buf/releases/download/v%BUF_VERSION%/buf-Windows-x86_64.exe"
    if "%DRY_RUN%"=="0" (
        if not exist "!BUF_DIR!" mkdir "!BUF_DIR!"
        powershell -NoProfile -Command "(New-Object Net.WebClient).DownloadFile('!BUF_URL!','!BUF_EXE!')"
        if not exist "!BUF_EXE!" (
            echo [ERROR] Failed to download buf from !BUF_URL!.
            exit /b 1
        )
    ) else (
        echo [DRY-RUN] Would run: download !BUF_URL! to !BUF_EXE!
    )
    REM Add buf to current session PATH
    set "PATH=!BUF_DIR!;%PATH%"
    REM Persist buf path to user PATH permanently
    if "%DRY_RUN%"=="0" (
        for /f "usebackq tokens=2*" %%a in (`reg query "HKCU\Environment" /v PATH 2^>nul`) do set "USR_PATH=%%b"
        echo !USR_PATH! | findstr /i /c:"buf\bin" >nul 2>&1
        if errorlevel 1 (
            setx PATH "!BUF_DIR!;!USR_PATH!"
            echo [INFO] buf path added to user PATH permanently.
        )
    ) else (
        echo [DRY-RUN] Would run: setx PATH "!BUF_DIR!;..."
    )
    echo [INFO] buf installed successfully.
) else (
    echo [INFO] buf.exe already in PATH.
)

REM -----------------------------------------------------------------------
REM Step 5: Ensure vcpkg is installed and `protobuf` is provisioned
REM
REM Resolution order for the vcpkg install location:
REM   1. Existing %VCPKG_ROOT% if it points at a usable classic-mode bootstrap.
REM   2. Existing %VCPKG_INSTALLATION_ROOT% (set on GitHub-hosted runners).
REM   3. Existing %USERPROFILE%\vcpkg (a previous run of this script).
REM   4. Fresh clone into %USERPROFILE%\vcpkg.
REM
REM A "usable" vcpkg root must contain BOTH vcpkg.exe AND bootstrap-vcpkg.bat.
REM This deliberately rejects the manifest-only vcpkg shipped under
REM   C:\Program Files\Microsoft Visual Studio\2022\<edition>\VC\vcpkg
REM which has no bootstrap script and refuses classic-mode `vcpkg install
REM <port>:<triplet>` with: "Could not locate a manifest (vcpkg.json) above
REM the current working directory. This vcpkg distribution does not have a
REM classic mode instance."
REM
REM When we install vcpkg ourselves we pin it to VCPKG_BASELINE_COMMIT so
REM the protobuf port version is reproducible. When we adopt a pre-existing
REM user-managed vcpkg, we leave its checkout state alone (the user owns it).
REM
REM We then run `vcpkg install protobuf:x64-windows-static` so that the
REM static-CRT libprotobuf + protoc match the loader build (CMakeLists.txt
REM forces /MT[d] via CMAKE_MSVC_RUNTIME_LIBRARY).
REM
REM Override the protobuf port version (e.g. for the legacy v3 line) with:
REM   set PROTOBUF_VCPKG_VERSION=3.21.12 && .\prepare.bat
REM In that case we switch to manifest mode (see header comment above).
REM -----------------------------------------------------------------------
REM Pre-flight: vcpkg compiles protobuf from source via MSVC, so cl.exe must
REM resolve here. If Step 3 didn't activate the MSVC environment (e.g. choco
REM just installed VS Build Tools and the registration hasn't fully landed
REM in this shell), fail fast with an actionable message rather than letting
REM vcpkg emit cryptic compiler-detection errors deep into the install.
where cl.exe >nul 2>&1
if errorlevel 1 (
    if "%DRY_RUN%"=="0" (
        echo [ERROR] cl.exe not on PATH; the MSVC environment is not active in this shell.
        echo [ERROR] Open a new "Developer Command Prompt for VS 2022" or rerun this script
        echo [ERROR] in a fresh cmd window so vcvarsall.bat can take effect, then retry.
        exit /b 1
    ) else (
        echo [DRY-RUN] [WARN] cl.exe not on PATH; would error out before vcpkg install.
    )
)

REM Pin both the vcpkg checkout and the manifest's builtin-baseline to the
REM same commit testing-cpp.yml uses. Bumping vcpkg? Bump both this value
REM and VCPKG_COMMIT in .github/workflows/testing-cpp.yml in lockstep.
set "VCPKG_BASELINE_COMMIT=dc8d75cfc3281b8e2a4ed8ee4163c891190df932"
set "VCPKG_TRIPLET=x64-windows-static"
set "VCPKG_EXE="

REM Honor pre-existing VCPKG_ROOT / VCPKG_INSTALLATION_ROOT only if they
REM point at a classic-mode-capable vcpkg (i.e. bootstrap-vcpkg.bat is present).
if "%SIMULATE_CLEAN%"=="0" (
    if defined VCPKG_ROOT (
        if exist "%VCPKG_ROOT%\vcpkg.exe" (
            if exist "%VCPKG_ROOT%\bootstrap-vcpkg.bat" (
                set "VCPKG_EXE=%VCPKG_ROOT%\vcpkg.exe"
            ) else (
                echo [WARN] %VCPKG_ROOT% looks like a manifest-only vcpkg ^(no bootstrap-vcpkg.bat^); ignoring.
                set "VCPKG_ROOT="
            )
        )
    )
    if not defined VCPKG_EXE (
        if defined VCPKG_INSTALLATION_ROOT (
            if exist "%VCPKG_INSTALLATION_ROOT%\vcpkg.exe" (
                if exist "%VCPKG_INSTALLATION_ROOT%\bootstrap-vcpkg.bat" (
                    set "VCPKG_ROOT=%VCPKG_INSTALLATION_ROOT%"
                    set "VCPKG_EXE=%VCPKG_INSTALLATION_ROOT%\vcpkg.exe"
                ) else (
                    echo [WARN] %VCPKG_INSTALLATION_ROOT% looks like a manifest-only vcpkg; ignoring.
                )
            )
        )
    )
    if not defined VCPKG_EXE (
        if exist "%USERPROFILE%\vcpkg\vcpkg.exe" (
            if exist "%USERPROFILE%\vcpkg\bootstrap-vcpkg.bat" (
                set "VCPKG_ROOT=%USERPROFILE%\vcpkg"
                set "VCPKG_EXE=%USERPROFILE%\vcpkg\vcpkg.exe"
            )
        )
    )
)

if not defined VCPKG_EXE (
    echo [INFO] vcpkg not found. Installing into %USERPROFILE%\vcpkg ...
    set "VCPKG_ROOT=%USERPROFILE%\vcpkg"
    if "%DRY_RUN%"=="0" (
        if not exist "!VCPKG_ROOT!" (
            REM Full clone (no --depth 1) so we can `git checkout` an arbitrary
            REM commit below for reproducibility.
            git clone https://github.com/microsoft/vcpkg.git "!VCPKG_ROOT!"
            if errorlevel 1 (
                echo [ERROR] Failed to clone vcpkg.
                exit /b 1
            )
        )
        REM Pin the checkout so port versions are reproducible. Safe to run
        REM repeatedly: a no-op when we're already on VCPKG_BASELINE_COMMIT.
        git -C "!VCPKG_ROOT!" fetch --quiet origin %VCPKG_BASELINE_COMMIT%
        git -C "!VCPKG_ROOT!" checkout --quiet %VCPKG_BASELINE_COMMIT%
        if errorlevel 1 (
            echo [ERROR] Failed to checkout vcpkg @ %VCPKG_BASELINE_COMMIT%.
            exit /b 1
        )
        call "!VCPKG_ROOT!\bootstrap-vcpkg.bat" -disableMetrics
        if errorlevel 1 (
            echo [ERROR] Failed to bootstrap vcpkg.
            exit /b 1
        )
    ) else (
        echo [DRY-RUN] Would run: git clone https://github.com/microsoft/vcpkg.git "!VCPKG_ROOT!"
        echo [DRY-RUN] Would run: git -C "!VCPKG_ROOT!" checkout %VCPKG_BASELINE_COMMIT%
        echo [DRY-RUN] Would run: "!VCPKG_ROOT!\bootstrap-vcpkg.bat" -disableMetrics
    )
    set "VCPKG_EXE=!VCPKG_ROOT!\vcpkg.exe"
    REM Persist VCPKG_ROOT and PATH to user environment
    if "%DRY_RUN%"=="0" (
        setx VCPKG_ROOT "!VCPKG_ROOT!"
        for /f "usebackq tokens=2*" %%a in (`reg query "HKCU\Environment" /v PATH 2^>nul`) do set "USR_PATH=%%b"
        echo !USR_PATH! | findstr /i /c:"!VCPKG_ROOT!" >nul 2>&1
        if errorlevel 1 (
            setx PATH "!VCPKG_ROOT!;!USR_PATH!"
            echo [INFO] vcpkg path added to user PATH permanently.
        )
    ) else (
        echo [DRY-RUN] Would run: setx VCPKG_ROOT "!VCPKG_ROOT!"
        echo [DRY-RUN] Would run: setx PATH "!VCPKG_ROOT!;..."
    )
    set "PATH=!VCPKG_ROOT!;%PATH%"
    echo [INFO] vcpkg installed at !VCPKG_ROOT! ^(pinned to %VCPKG_BASELINE_COMMIT%^).
) else (
    echo [INFO] vcpkg already available at !VCPKG_ROOT! ^(user-managed; not re-pinning^).
)

REM Install protobuf into vcpkg.
REM
REM Branching on PROTOBUF_VCPKG_VERSION:
REM   - unset: CLASSIC mode. Installs into %VCPKG_ROOT%\installed\<triplet>\,
REM            auto-discovered by the vcpkg cmake toolchain — no extra cmake
REM            flags needed downstream. Whatever protobuf the pinned vcpkg
REM            checkout (VCPKG_BASELINE_COMMIT) ships is what you get.
REM   - set:   MANIFEST mode. Renders a vcpkg.json under
REM            %LOCALAPPDATA%\loader\vcpkg-manifest\ with builtin-baseline +
REM            an override pinning protobuf to the requested version. This
REM            is the only mode in which `--x-version` / `overrides` actually
REM            take effect; classic-mode `--x-version` is silently a no-op.
REM            Install root: <manifest-dir>\vcpkg_installed\<triplet>\.
REM
REM Both modes are idempotent: vcpkg detects already-installed packages and
REM skips them.
if defined PROTOBUF_VCPKG_VERSION (
    set "VCPKG_MANIFEST_DIR=%LOCALAPPDATA%\loader\vcpkg-manifest"
    set "VCPKG_INSTALLED_DIR=!VCPKG_MANIFEST_DIR!\vcpkg_installed"
    if "%DRY_RUN%"=="0" (
        if not exist "!VCPKG_MANIFEST_DIR!" mkdir "!VCPKG_MANIFEST_DIR!"
        REM Render vcpkg.json. The `>` redirection truncates on first line and
        REM `>>` appends the rest, mirroring a here-doc.
        > "!VCPKG_MANIFEST_DIR!\vcpkg.json" echo {
        >>"!VCPKG_MANIFEST_DIR!\vcpkg.json" echo   "name": "loader-prepare",
        >>"!VCPKG_MANIFEST_DIR!\vcpkg.json" echo   "version": "0.1.0",
        >>"!VCPKG_MANIFEST_DIR!\vcpkg.json" echo   "dependencies": ["protobuf"],
        >>"!VCPKG_MANIFEST_DIR!\vcpkg.json" echo   "overrides": [{ "name": "protobuf", "version": "%PROTOBUF_VCPKG_VERSION%" }],
        >>"!VCPKG_MANIFEST_DIR!\vcpkg.json" echo   "builtin-baseline": "%VCPKG_BASELINE_COMMIT%"
        >>"!VCPKG_MANIFEST_DIR!\vcpkg.json" echo }
        echo [INFO] Installing protobuf %PROTOBUF_VCPKG_VERSION% into vcpkg ^(manifest mode, triplet !VCPKG_TRIPLET!^)...
        pushd "!VCPKG_MANIFEST_DIR!"
        "!VCPKG_EXE!" install --triplet=!VCPKG_TRIPLET! --x-install-root="!VCPKG_INSTALLED_DIR!"
        set "VCPKG_INSTALL_RC=!ERRORLEVEL!"
        popd
        if not "!VCPKG_INSTALL_RC!"=="0" (
            echo [ERROR] vcpkg failed to install protobuf %PROTOBUF_VCPKG_VERSION%.
            exit /b 1
        )
        REM Sanity check: assert the resolved version actually starts with
        REM the requested one. vcpkg port versions can have a `#N` port-revision
        REM suffix, so we match by prefix rather than equality. This is the
        REM safety net that catches future regressions in vcpkg's manifest
        REM resolution silently producing the wrong version.
        set "VCPKG_INFO_FILE=!VCPKG_INSTALLED_DIR!\vcpkg\info\.unused"
        for /f "delims=" %%f in ('dir /b /a-d "!VCPKG_INSTALLED_DIR!\vcpkg\info\protobuf_*_!VCPKG_TRIPLET!.list" 2^>nul') do (
            set "VCPKG_INFO_FILE=%%f"
        )
        echo !VCPKG_INFO_FILE! | findstr /c:"protobuf_%PROTOBUF_VCPKG_VERSION%" >nul 2>&1
        if errorlevel 1 (
            echo [ERROR] Installed protobuf does not match requested version %PROTOBUF_VCPKG_VERSION%.
            echo [ERROR]   vcpkg installed file marker: !VCPKG_INFO_FILE!
            echo [ERROR] This usually means VCPKG_BASELINE_COMMIT is too old to know about that
            echo [ERROR] version. Bump the commit at the top of Step 5 and retry.
            exit /b 1
        )
    ) else (
        echo [DRY-RUN] Would render: !VCPKG_MANIFEST_DIR!\vcpkg.json ^(protobuf %PROTOBUF_VCPKG_VERSION%, baseline %VCPKG_BASELINE_COMMIT%^)
        echo [DRY-RUN] Would run: pushd "!VCPKG_MANIFEST_DIR!" ^&^& "!VCPKG_EXE!" install --triplet=!VCPKG_TRIPLET! --x-install-root="!VCPKG_INSTALLED_DIR!"
    )
    set "PROTOC_TOOLS_DIR=!VCPKG_INSTALLED_DIR!\!VCPKG_TRIPLET!\tools\protobuf"
) else (
    if "%DRY_RUN%"=="0" (
        echo [INFO] Installing protobuf into vcpkg ^(classic mode, triplet !VCPKG_TRIPLET!^)...
        "!VCPKG_EXE!" install "protobuf:!VCPKG_TRIPLET!"
        if errorlevel 1 (
            echo [ERROR] vcpkg failed to install protobuf.
            exit /b 1
        )
    ) else (
        echo [DRY-RUN] Would run: "!VCPKG_EXE!" install "protobuf:!VCPKG_TRIPLET!"
    )
    set "VCPKG_INSTALLED_DIR="
    set "PROTOC_TOOLS_DIR=!VCPKG_ROOT!\installed\!VCPKG_TRIPLET!\tools\protobuf"
)

REM Expose vcpkg-installed protoc on PATH so `buf generate` finds it.
if exist "!PROTOC_TOOLS_DIR!\protoc.exe" (
    set "PATH=!PROTOC_TOOLS_DIR!;%PATH%"
    echo [INFO] vcpkg protoc on PATH: !PROTOC_TOOLS_DIR!
)

if defined VCPKG_INSTALLED_DIR (
    echo [INFO] Manifest-mode install root: !VCPKG_INSTALLED_DIR!
    echo [INFO] When invoking cmake, also pass:
    echo [INFO]     -DVCPKG_INSTALLED_DIR="!VCPKG_INSTALLED_DIR!" -DVCPKG_MANIFEST_INSTALL=OFF
)

echo [INFO] Build environment ready.

REM Export PATH and key MSVC vars back to the caller's environment.
REM Also export VCPKG_ROOT so subsequent `cmake -DCMAKE_TOOLCHAIN_FILE=%VCPKG_ROOT%\...`
REM invocations resolve in this same cmd session even before the persisted
REM setx value takes effect in newly-spawned processes. VCPKG_INSTALLED_DIR
REM is set only in manifest mode (PROTOBUF_VCPKG_VERSION pinned).
endlocal & set "PATH=%PATH%" & set "INCLUDE=%INCLUDE%" & set "LIB=%LIB%" & set "LIBPATH=%LIBPATH%" & set "WindowsSdkDir=%WindowsSdkDir%" & set "VCToolsInstallDir=%VCToolsInstallDir%" & set "VCPKG_ROOT=%VCPKG_ROOT%" & set "VCPKG_INSTALLED_DIR=%VCPKG_INSTALLED_DIR%"
