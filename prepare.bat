@echo off
setlocal enabledelayedexpansion

echo [INFO] Preparing build environment...

REM -----------------------------------------------------------------------
REM Step 0: Ensure Chocolatey is installed
REM -----------------------------------------------------------------------
set "CHOCO_EXE="
set "CHOCO_BASE="
if defined ChocolateyInstall set "CHOCO_BASE=%ChocolateyInstall%"
if not defined CHOCO_BASE set "CHOCO_BASE=%ALLUSERSPROFILE%\chocolatey"
if exist "%CHOCO_BASE%\bin\choco.exe"       set "CHOCO_EXE=%CHOCO_BASE%\bin\choco.exe"
if exist "%CHOCO_BASE%\redirects\choco.exe" set "CHOCO_EXE=%CHOCO_BASE%\redirects\choco.exe"
if not defined CHOCO_EXE (
    echo [INFO] Chocolatey not found. Installing Chocolatey...
    powershell -NoProfile -ExecutionPolicy Bypass -Command ^
        "[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))"
    if errorlevel 1 (
        echo [ERROR] Failed to install Chocolatey.
        exit /b 1
    )
    REM Add Chocolatey to current session PATH
    set "PATH=%ALLUSERSPROFILE%\chocolatey\bin;%PATH%"
    REM Persist Chocolatey bin to user PATH permanently
    for /f "usebackq tokens=2*" %%a in (`reg query "HKCU\Environment" /v PATH 2^>nul`) do set "USR_PATH=%%b"
    echo !USR_PATH! | findstr /i /c:"%ALLUSERSPROFILE%\chocolatey\bin" >nul 2>&1
    if errorlevel 1 (
        setx PATH "%ALLUSERSPROFILE%\chocolatey\bin;!USR_PATH!"
        echo [INFO] Chocolatey bin added to user PATH permanently.
    )
    echo [INFO] Chocolatey installed successfully.
) else (
    echo [INFO] Chocolatey already installed.
)

REM Refresh ChocolateyInstall var if it was just installed
if not defined ChocolateyInstall set "ChocolateyInstall=%ALLUSERSPROFILE%\chocolatey"
set "PATH=%ChocolateyInstall%\bin;%ChocolateyInstall%\lib\ninja\tools;%PATH%"

REM -----------------------------------------------------------------------
REM Step 1: Ensure Ninja is installed via Chocolatey
REM         (equivalent to CI step: choco install ninja -y)
REM -----------------------------------------------------------------------
where ninja.exe >nul 2>&1
if errorlevel 1 (
    echo [INFO] ninja.exe not found. Installing via choco...
    choco install ninja -y --no-progress
    if errorlevel 1 (
        echo [ERROR] Failed to install ninja.
        exit /b 1
    )
    REM Add ninja to current session PATH
    if defined ChocolateyInstall (
        set "NINJA_PATH=!ChocolateyInstall!\lib\ninja\tools"
    ) else (
        set "NINJA_PATH=%ALLUSERSPROFILE%\chocolatey\lib\ninja\tools"
    )
    set "PATH=!NINJA_PATH!;%PATH%"
    REM Persist ninja path to user PATH permanently
    for /f "usebackq tokens=2*" %%a in (`reg query "HKCU\Environment" /v PATH 2^>nul`) do set "USR_PATH=%%b"
    echo !USR_PATH! | findstr /i /c:"ninja\tools" >nul 2>&1
    if errorlevel 1 (
        setx PATH "!NINJA_PATH!;!USR_PATH!"
        echo [INFO] ninja path added to user PATH permanently.
    )
    echo [INFO] ninja installed successfully.
) else (
    echo [INFO] ninja.exe already in PATH.
)

REM -----------------------------------------------------------------------
REM Step 2: Ensure MSVC compiler (cl.exe) is available
REM         (equivalent to CI step: ilammy/msvc-dev-cmd@v1)
REM -----------------------------------------------------------------------
where cl.exe >nul 2>&1
if errorlevel 1 (
    echo [INFO] cl.exe not found. Searching for existing VS installation...
    set "VSWHERE="
    for %%d in ("%ProgramFiles(x86)%" "%ProgramFiles%") do (
        if not defined VSWHERE (
            if exist "%%~d\Microsoft Visual Studio\Installer\vswhere.exe" (
                set "VSWHERE=%%~d\Microsoft Visual Studio\Installer\vswhere.exe"
            )
        )
    )
    if not defined VSWHERE (
        echo [INFO] Visual Studio not found. Installing via choco...
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
    )
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
) else (
    echo [INFO] cl.exe already in PATH, skipping MSVC environment setup.
)

echo [INFO] Build environment ready.

REM Export PATH and key MSVC vars back to the caller's environment
endlocal & set "PATH=%PATH%" & set "INCLUDE=%INCLUDE%" & set "LIB=%LIB%" & set "LIBPATH=%LIBPATH%" & set "WindowsSdkDir=%WindowsSdkDir%" & set "VCToolsInstallDir=%VCToolsInstallDir%"
