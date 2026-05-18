@echo off
setlocal enabledelayedexpansion

REM Initialize build environment (installs choco/ninja/MSVC if needed, sets up PATH)
call "%~dp0prepare.bat"
if errorlevel 1 (
    echo [ERROR] prepare.bat failed. Aborting.
    exit /b 1
)

for /f "delims=" %%i in ('git rev-parse --show-toplevel') do set repoRoot=%%i
cd /d "%repoRoot%"

REM Initialize only the protobuf submodule (non-recursive). Protobuf's own
REM nested submodules (third_party/googletest, third_party/benchmark on v3.x)
REM are only needed when protobuf_BUILD_TESTS=ON / benchmarks are enabled, and
REM modern protobuf (v4+/v21+) has dropped git submodules entirely in favor of
REM CMake FetchContent. Skipping --recursive saves clone time and CI bandwidth.
git submodule update --init third_party/_submodules/protobuf

REM Build and install the C++ Protocol Buffer runtime and the Protocol Buffer compiler (protoc)
cd third_party\_submodules\protobuf

REM If PROTOBUF_REF is set, switch submodule to the specified ref
if not "%PROTOBUF_REF%"=="" (
    echo Switching protobuf submodule to %PROTOBUF_REF%...
    git fetch --tags
    git checkout %PROTOBUF_REF%
)

REM ---------------------------------------------------------------------------
REM Detect protobuf major version and build the cmake command line. We compute
REM both up-front (before the fast-path check) so the signature comparison
REM below can include the exact cmake invocation we would run.
REM   - protobuf v3.x  : CMakeLists.txt is in cmake/ subdirectory
REM   - protobuf v4+   : CMakeLists.txt is in root directory
REM ---------------------------------------------------------------------------
for /f "tokens=*" %%v in ('git describe --tags --abbrev^=0 2^>nul') do set PROTOBUF_VERSION=%%v
if not defined PROTOBUF_VERSION set PROTOBUF_VERSION=unknown
echo Detected protobuf version: %PROTOBUF_VERSION%

REM Extract major version number from tag (e.g., v3.19.3 -> 3, v32.0 -> 32)
set "VER_STR=%PROTOBUF_VERSION:~1%"
for /f "tokens=1 delims=." %%a in ("%VER_STR%") do set MAJOR_VERSION=%%a

if %MAJOR_VERSION% LEQ 3 (
    REM Legacy protobuf (v3.x): CMakeLists.txt is in cmake/ subdirectory
    set "PROTOBUF_BUILD_VARIANT=legacy"
    set "CMAKE_SRC=cmake"
    set "CMAKE_FLAGS=-DCMAKE_BUILD_TYPE=Debug -DCMAKE_CXX_STANDARD=17 -DCMAKE_POLICY_VERSION_MINIMUM=3.5 -Dprotobuf_BUILD_TESTS=OFF -Dprotobuf_WITH_ZLIB=OFF -Dprotobuf_BUILD_SHARED_LIBS=OFF"
) else (
    REM Modern protobuf (v4+/v21+/v32+): CMakeLists.txt is in root directory
    REM Refer: https://github.com/protocolbuffers/protobuf/blob/v32.0/cmake/README.md#cmake-configuration
    REM - protobuf_MSVC_STATIC_RUNTIME defaults to ON, which uses static CRT (/MTd for Debug).
    REM   Our project's CMakeLists.txt also sets static CRT to match.
    REM - protobuf_WITH_ZLIB=OFF: disable ZLIB dependency to avoid ZLIB::ZLIB link requirement
    REM   in protobuf's exported CMake targets, which simplifies cross-platform builds.
    REM - protobuf_BUILD_SHARED_LIBS=OFF: build static libraries explicitly.
    set "PROTOBUF_BUILD_VARIANT=modern"
    set "CMAKE_SRC=."
    set "CMAKE_FLAGS=-DCMAKE_BUILD_TYPE=Debug -DCMAKE_CXX_STANDARD=17 -DCMAKE_POLICY_VERSION_MINIMUM=3.5 -Dprotobuf_BUILD_TESTS=OFF -Dprotobuf_WITH_ZLIB=OFF -Dprotobuf_BUILD_SHARED_LIBS=OFF -Dutf8_range_ENABLE_INSTALL=ON"
)

REM Build a stable, multi-line signature describing the inputs that determine
REM the contents of .build\_install. Any change to these values must
REM invalidate the fast-path. Adding new compile-time inputs? Append a line.
set "SIG_FILE=.build\_install\.build_signature"
set "SIG_LINE_1=schema=1"
set "SIG_LINE_2=version=!PROTOBUF_VERSION!"
set "SIG_LINE_3=variant=!PROTOBUF_BUILD_VARIANT!"
set "SIG_LINE_4=cmake_args=-S !CMAKE_SRC! -B .build -G Ninja !CMAKE_FLAGS!"

REM ---------------------------------------------------------------------------
REM Fast path: if a previous build's signature file matches the one we're
REM about to use, skip the (very long) protobuf compile entirely.
REM Set FORCE_REBUILD_PROTOBUF=1 to bypass this short-circuit unconditionally.
REM ---------------------------------------------------------------------------
if not "%FORCE_REBUILD_PROTOBUF%"=="" goto :no_fast_path
if not exist "!SIG_FILE!" goto :no_fast_path

REM Read the file 4 lines at a time and compare with the expected signature.
set "ACTUAL_LINE_1="
set "ACTUAL_LINE_2="
set "ACTUAL_LINE_3="
set "ACTUAL_LINE_4="
set "_LINE_NO=0"
for /f "usebackq delims=" %%L in ("!SIG_FILE!") do (
    set /a _LINE_NO+=1
    set "ACTUAL_LINE_!_LINE_NO!=%%L"
)

if not "!ACTUAL_LINE_1!"=="!SIG_LINE_1!" goto :sig_mismatch
if not "!ACTUAL_LINE_2!"=="!SIG_LINE_2!" goto :sig_mismatch
if not "!ACTUAL_LINE_3!"=="!SIG_LINE_3!" goto :sig_mismatch
if not "!ACTUAL_LINE_4!"=="!SIG_LINE_4!" goto :sig_mismatch

echo [INFO] Build signature matches; reusing existing protobuf install at .build\_install.
echo [INFO] Set FORCE_REBUILD_PROTOBUF=1 to force a clean rebuild.
goto :eof

:sig_mismatch
echo [INFO] Build signature mismatch; rebuilding protobuf.
echo [INFO]   actual:
echo [INFO]     !ACTUAL_LINE_1!
echo [INFO]     !ACTUAL_LINE_2!
echo [INFO]     !ACTUAL_LINE_3!
echo [INFO]     !ACTUAL_LINE_4!
echo [INFO]   expected:
echo [INFO]     !SIG_LINE_1!
echo [INFO]     !SIG_LINE_2!
echo [INFO]     !SIG_LINE_3!
echo [INFO]     !SIG_LINE_4!

:no_fast_path
REM Wipe any stale install dir so we don't leave half-overwritten files behind
REM when cmake flags change (e.g. Debug -> Release puts artifacts in different
REM places, an in-place re-install would mix old and new).
if exist .build rmdir /s /q .build

REM ---------------------------------------------------------------------------
REM Configure
REM ---------------------------------------------------------------------------
if "!PROTOBUF_BUILD_VARIANT!"=="legacy" (
    echo Using legacy cmake\ subdirectory for protobuf %PROTOBUF_VERSION%
) else (
    echo Using root CMakeLists.txt for protobuf %PROTOBUF_VERSION%
)
cmake -S !CMAKE_SRC! -B .build -G Ninja !CMAKE_FLAGS!
if errorlevel 1 exit /b 1

REM Compile the code
cmake --build .build --parallel
if errorlevel 1 exit /b 1

REM Install into .build/_install so that protobuf-config.cmake (along with
REM absl and utf8_range configs) is generated for find_package(Protobuf CONFIG)
REM used by downstream CMakeLists.txt.
REM NOTE: .build/ is already in protobuf's .gitignore, so _install stays clean.
cmake --install .build --prefix .build\_install
if errorlevel 1 exit /b 1

REM Persist the signature so the next run can fast-path skip when nothing changed.
> "!SIG_FILE!" (
    echo !SIG_LINE_1!
    echo !SIG_LINE_2!
    echo !SIG_LINE_3!
    echo !SIG_LINE_4!
)
echo [INFO] Wrote build signature to !SIG_FILE!

endlocal
