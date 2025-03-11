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
cd cmake
cmake .
cmake --build .

endlocal