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
REM Refer: https://github.com/protocolbuffers/protobuf/blob/3.19.x/cmake/README.md#cmake-configuration
cd third_party\_submodules\protobuf\cmake
REM use Debug version
REM - protobuf_MSVC_STATIC_RUNTIME defaults to ON, which uses static CRT (/MTd for Debug).
REM   Our project's CMakeLists.txt also sets static CRT to match.
REM - protobuf_WITH_ZLIB=OFF: disable ZLIB dependency to avoid ZLIB::ZLIB link requirement
REM   in protobuf's exported CMake targets, which simplifies cross-platform builds.
REM - protobuf_BUILD_SHARED_LIBS=OFF: build static libraries explicitly.
cmake -S . -B build -G Ninja ^
  -DCMAKE_BUILD_TYPE=Debug ^
  -DCMAKE_CXX_STANDARD=17 ^
  -DCMAKE_POLICY_VERSION_MINIMUM=3.5 ^
  -Dprotobuf_BUILD_TESTS=OFF ^
  -Dprotobuf_WITH_ZLIB=OFF ^
  -Dprotobuf_BUILD_SHARED_LIBS=OFF

REM Compile the code
cmake --build build --parallel

endlocal
