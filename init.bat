@echo off
setlocal

REM Initialize build environment (installs choco/ninja/MSVC if needed, sets up PATH)
call "%~dp0prepare.bat"
if errorlevel 1 (
    echo [ERROR] prepare.bat failed. Aborting.
    exit /b 1
)

for /f "delims=" %%i in ('git rev-parse --show-toplevel') do set repoRoot=%%i
cd /d "%repoRoot%"

git submodule update --init --recursive

REM Build and install the C++ Protocol Buffer runtime and the Protocol Buffer compiler (protoc)
cd third_party\_submodules\protobuf

REM If PROTOBUF_REF is set, switch submodule to the specified ref
if not "%PROTOBUF_REF%"=="" (
    echo Switching protobuf submodule to %PROTOBUF_REF%...
    git fetch --tags
    git checkout %PROTOBUF_REF%
    git submodule update --init --recursive
)

REM Detect protobuf major version to determine cmake source directory and arguments.
for /f "tokens=*" %%v in ('git describe --tags --abbrev^=0 2^>nul') do set PROTOBUF_VERSION=%%v
if not defined PROTOBUF_VERSION set PROTOBUF_VERSION=unknown
echo Detected protobuf version: %PROTOBUF_VERSION%

REM Extract major version number from tag (e.g., v3.19.3 -> 3, v32.0 -> 32)
set "VER_STR=%PROTOBUF_VERSION:~1%"
for /f "tokens=1 delims=." %%a in ("%VER_STR%") do set MAJOR_VERSION=%%a

if %MAJOR_VERSION% LEQ 3 (
    REM Legacy protobuf (v3.x): CMakeLists.txt is in cmake/ subdirectory
    echo Using legacy cmake\ subdirectory for protobuf %PROTOBUF_VERSION%
    cmake -S cmake -B .build -G Ninja ^
      -DCMAKE_BUILD_TYPE=Debug ^
      -DCMAKE_CXX_STANDARD=17 ^
      -DCMAKE_POLICY_VERSION_MINIMUM=3.5 ^
      -Dprotobuf_BUILD_TESTS=OFF ^
      -Dprotobuf_WITH_ZLIB=OFF ^
      -Dprotobuf_BUILD_SHARED_LIBS=OFF
) else (
    REM Modern protobuf (v4+/v21+/v32+): CMakeLists.txt is in root directory
    REM Refer: https://github.com/protocolbuffers/protobuf/blob/v32.0/cmake/README.md#cmake-configuration
    echo Using root CMakeLists.txt for protobuf %PROTOBUF_VERSION%
    REM - protobuf_MSVC_STATIC_RUNTIME defaults to ON, which uses static CRT (/MTd for Debug).
    REM   Our project's CMakeLists.txt also sets static CRT to match.
    REM - protobuf_WITH_ZLIB=OFF: disable ZLIB dependency to avoid ZLIB::ZLIB link requirement
    REM   in protobuf's exported CMake targets, which simplifies cross-platform builds.
    REM - protobuf_BUILD_SHARED_LIBS=OFF: build static libraries explicitly.
    cmake -S . -B .build -G Ninja ^
      -DCMAKE_BUILD_TYPE=Debug ^
      -DCMAKE_CXX_STANDARD=17 ^
      -DCMAKE_POLICY_VERSION_MINIMUM=3.5 ^
      -Dprotobuf_BUILD_TESTS=OFF ^
      -Dprotobuf_WITH_ZLIB=OFF ^
      -Dprotobuf_BUILD_SHARED_LIBS=OFF ^
      -Dutf8_range_ENABLE_INSTALL=ON
)

REM Compile the code
cmake --build .build --parallel

REM Install into .build/_install so that protobuf-config.cmake (along with
REM absl and utf8_range configs) is generated for find_package(Protobuf CONFIG)
REM used by downstream CMakeLists.txt.
REM NOTE: .build/ is already in protobuf's .gitignore, so _install stays clean.
cmake --install .build --prefix .build\_install

endlocal
