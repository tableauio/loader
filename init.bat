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
cmake -S . -B build ^
 -G "NMake Makefiles" ^
 -DCMAKE_BUILD_TYPE=Debug ^
 -DCMAKE_POLICY_VERSION_MINIMUM="3.5" ^

REM Compile the code
cmake --build build

endlocal
