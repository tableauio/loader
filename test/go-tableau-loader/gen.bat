@echo off
setlocal
setlocal enabledelayedexpansion

for /f "delims=" %%i in ('git rev-parse --show-toplevel') do set repoRoot=%%i
cd /d "%repoRoot%"

REM Allow overriding protoc via environment variable.
REM Default to locally compiled protoc for local development; fallback to system protoc.
if not defined PROTOC (
    if exist "%repoRoot%\third_party\_submodules\protobuf\cmake\build\protoc.exe" (
        set "PROTOC=%repoRoot%\third_party\_submodules\protobuf\cmake\build\protoc.exe"
    ) else (
        where protoc >nul 2>nul
        if !errorlevel! equ 0 (
            for /f "delims=" %%p in ('where protoc') do set "PROTOC=%%p"
        ) else (
            echo Error: protoc not found. Please build protobuf submodule or install protoc. >&2
            exit /b 1
        )
    )
)
REM Allow overriding protobuf include path via environment variable.
REM Default to local submodule source; fallback to system include path.
if not defined PROTOBUF_PROTO (
    if exist "%repoRoot%\third_party\_submodules\protobuf\src\google\protobuf" (
        set "PROTOBUF_PROTO=%repoRoot%\third_party\_submodules\protobuf\src"
    ) else (
        for /f "delims=" %%p in ('where protoc 2^>nul') do set "_PROTOC_DIR=%%~dpp"
        if defined _PROTOC_DIR (
            set "PROTOBUF_PROTO=!_PROTOC_DIR!..\include"
        ) else (
            set "PROTOBUF_PROTO=%repoRoot%\third_party\_submodules\protobuf\src"
        )
    )
)
set "TABLEAU_PROTO=%repoRoot%\third_party\_submodules\tableau\proto"
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
