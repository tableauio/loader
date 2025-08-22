@echo off
setlocal
setlocal enabledelayedexpansion

for /f "delims=" %%i in ('git rev-parse --show-toplevel') do set repoRoot=%%i
cd /d "%repoRoot%"

set "PROTOC=%repoRoot%\third_party\_submodules\protobuf\cmake\build\debug\protoc.exe"
set "PROTOBUF_PROTO=%repoRoot%\third_party\_submodules\protobuf\src"
set "TABLEAU_PROTO=%repoRoot%\third_party\_submodules\tableau\proto"
set "ROOTDIR=%repoRoot%\test\cpp-tableau-loader"
set "PLGUIN_DIR=%repoRoot%\cmd\protoc-gen-cpp-tableau-loader"
set "PROTOCONF_IN=%repoRoot%\test\proto"
set "PROTOCONF_OUT=%ROOTDIR%\src\protoconf"

REM remove old generated files
rmdir /s /q "%PROTOCONF_OUT%" 2>nul
mkdir "%PROTOCONF_OUT%"

REM build
pushd "%PLGUIN_DIR%"
go build
popd

set "PATH=%PATH%;%PLGUIN_DIR%"

set protoFiles=
pushd "%PROTOCONF_IN%"
for /R %%f in (*.proto) do (
  set protoFiles=!protoFiles! "%%f"
)
popd
"%PROTOC%" ^
--cpp-tableau-loader_out="%PROTOCONF_OUT%" ^
--cpp-tableau-loader_opt=paths=source_relative,shards=2 ^
--cpp_out="%PROTOCONF_OUT%" ^
--proto_path="%PROTOBUF_PROTO%" ^
--proto_path="%TABLEAU_PROTO%" ^
--proto_path="%PROTOCONF_IN%" ^
!protoFiles!

set "TABLEAU_IN=%TABLEAU_PROTO%\tableau\protobuf"
set "TABLEAU_OUT=%ROOTDIR%\src"
REM remove old generated files
if exist "%TABLEAU_OUT%\tableau" rmdir /s /q "%TABLEAU_OUT%\tableau"
mkdir "%TABLEAU_OUT%\tableau"

"%PROTOC%" ^
--cpp_out="%TABLEAU_OUT%" ^
--proto_path="%PROTOBUF_PROTO%" ^
--proto_path="%TABLEAU_PROTO%" ^
"%TABLEAU_IN%\tableau.proto" "%TABLEAU_IN%\wellknown.proto"

endlocal
endlocal
