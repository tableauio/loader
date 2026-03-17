@echo off
setlocal enabledelayedexpansion

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
REM         (equivalent to CI step: choco install ninja -y)
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
REM Step 3: Ensure MSVC compiler (cl.exe) is available
REM         (equivalent to CI step: ilammy/msvc-dev-cmd@v1)
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

echo [INFO] Build environment ready.

REM Export PATH and key MSVC vars back to the caller's environment
endlocal & set "PATH=%PATH%" & set "INCLUDE=%INCLUDE%" & set "LIB=%LIB%" & set "LIBPATH=%LIBPATH%" & set "WindowsSdkDir=%WindowsSdkDir%" & set "VCToolsInstallDir=%VCToolsInstallDir%"
