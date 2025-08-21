@echo off
setlocal
setlocal enabledelayedexpansion

for /f "delims=" %%i in ('git rev-parse --show-toplevel') do set repoRoot=%%i
cd /d "%repoRoot%"
go mod tidy

set "PROTOC=%repoRoot%\third_party\_submodules\protobuf\cmake\build\protoc.exe"
set "PROTOBUF_PROTO=%repoRoot%\third_party\_submodules\protobuf\src"
set "PROTOBUF_PROTO=%repoRoot%\third_party\_submodules\protobuf\src"
set "TABLEAU_GOPATH=github.com/tableauio/tableau"
for /f "delims=" %%G in ('go env GOPATH') do set "GOPATH=%%G"
for /f "tokens=2" %%V in ('findstr /c:"%TABLEAU_GOPATH%" go.mod') do set "VERSION=%%V"
set "TABLEAU_PROTO=%GOPATH%\pkg\mod\%TABLEAU_GOPATH%@%VERSION%\proto"
set "PLGUIN_DIR=%repoRoot%\cmd\protoc-gen-go-tableau-loader"
set "PROTOCONF_IN=%repoRoot%\test\proto"
set "PROTOCONF_OUT=%repoRoot%\test\go-tableau-loader\protoconf"
set "LOADER_OUT=%PROTOCONF_OUT%\loader"

REM remove old generated files
rmdir /s /q "%PROTOCONF_OUT%" "%LOADER_OUT%" 2>nul
mkdir "%PROTOCONF_OUT%" "%LOADER_OUT%"

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
--go-tableau-loader_out="%LOADER_OUT%" ^
--go-tableau-loader_opt=paths=source_relative,pkg=loader ^
--go_out="%PROTOCONF_OUT%" ^
--go_opt=paths=source_relative ^
--proto_path="%PROTOBUF_PROTO%" ^
--proto_path="%TABLEAU_PROTO%" ^
--proto_path="%PROTOCONF_IN%" ^
!protoFiles!

endlocal
endlocal
