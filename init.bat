@echo off
setlocal

for /f "delims=" %%i in ('git rev-parse --show-toplevel') do set repoRoot=%%i
cd /d "%repoRoot%"

git submodule update --init --recursive

REM google protobuf
cd third_party\_submodules\protobuf
git checkout v3.19.3
git submodule update --init --recursive

REM Build and install the C++ Protocol Buffer runtime and the Protocol Buffer compiler (protoc)
REM Refer: https://github.com/protocolbuffers/protobuf/blob/3.19.x/cmake/README.md#cmake-configuration
cd cmake
REM use Debug version
REM -Dprotobuf_MSVC_STATIC_RUNTIME=OFF: use dynamic CRT (/MDd) to match the default MSVC runtime library setting.
REM -DCMAKE_MSVC_RUNTIME_LIBRARY: CMake 3.15+ standard way to control MSVC runtime library.
cmake -S . -B build -G Ninja -DCMAKE_BUILD_TYPE=Debug -DCMAKE_POLICY_VERSION_MINIMUM=3.5 -Dprotobuf_BUILD_TESTS=OFF -Dprotobuf_MSVC_STATIC_RUNTIME=OFF -DCMAKE_MSVC_RUNTIME_LIBRARY="MultiThreaded$<$<CONFIG:Debug>:Debug>DLL"

REM Compile the code
cmake --build build --parallel

endlocal
