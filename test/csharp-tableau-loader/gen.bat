@echo off
setlocal
setlocal enabledelayedexpansion

for /f "delims=" %%i in ('git rev-parse --show-toplevel') do set repoRoot=%%i
cd /d "%repoRoot%"

set "PROTOC=%repoRoot%\third_party\_submodules\protobuf\cmake\build\protoc.exe"
set "PROTOBUF_PROTO=%repoRoot%\third_party\_submodules\protobuf\src"
set "TABLEAU_PROTO=%repoRoot%\third_party\_submodules\tableau\proto"
set "ROOTDIR=%repoRoot%\test\csharp-tableau-loader"
set "PLGUIN_DIR=%repoRoot%\cmd\protoc-gen-csharp-tableau-loader"
set "PROTOCONF_IN=%repoRoot%\test\proto"
set "PROTOCONF_OUT=%ROOTDIR%\protoconf"
set "LOADER_OUT=%ROOTDIR%\tableau"

REM remove old generated files
rmdir /s /q "%PROTOCONF_OUT%" "%LOADER_OUT%" 2>nul
mkdir "%PROTOCONF_OUT%" "%LOADER_OUT%"

REM build protoc plugin of loader
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
--csharp_out="%PROTOCONF_OUT%" ^
--csharp-tableau-loader_out="%LOADER_OUT%" ^
--csharp-tableau-loader_opt=paths=source_relative ^
--proto_path="%PROTOBUF_PROTO%" ^
--proto_path="%TABLEAU_PROTO%" ^
--proto_path="%PROTOCONF_IN%" ^
!protoFiles!

set "TABLEAU_IN=%TABLEAU_PROTO%\tableau\protobuf"
set "TABLEAU_OUT=%ROOTDIR%\protoconf\tableau"
REM remove old generated files
rmdir /s /q "%TABLEAU_OUT%" 2>nul
mkdir "%TABLEAU_OUT%"

"%PROTOC%" ^
--csharp_out="%TABLEAU_OUT%" ^
--proto_path="%PROTOBUF_PROTO%" ^
--proto_path="%TABLEAU_PROTO%" ^
"%TABLEAU_IN%\tableau.proto" "%TABLEAU_IN%\wellknown.proto"

endlocal
endlocal
