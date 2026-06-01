# escape=`
# tableauio/loader devcontainer (windows)
#
# Windows-container variant. Runs ONLY on Windows hosts (Hyper-V or
# process isolation, Docker Desktop in Windows-containers mode).
# windows/amd64 ONLY — Microsoft does not publish arm64 Windows base
# images, so arm64 hosts must use the Linux container under ../linux/.
#
# All version pins are read from ../shared/versions.env (the single source
# of truth shared with the Linux container, prepare.bat, and CI). To bump
# Go / buf / protobuf / .NET / Node / vcpkg-baseline / cmake, edit that
# file — not this one.
#
# Build context is .devcontainer/ (the parent of this directory) so that
# `COPY shared\...` resolves. devcontainer.json sets `"context": ".."`.
#
# Why mcr.microsoft.com/visualstudio/buildtools (not windows/servercore):
#
#   The modern VS 2022 Build Tools installer (`vs_BuildTools.exe`)
#   requires Windows servicing components that `windows/servercore`
#   strips out. Attempting to install via Chocolatey or directly fails
#   with installer exit code -2146232797 deep into the install. Microsoft
#   publishes `mcr.microsoft.com/visualstudio/buildtools:ltsc2022`
#   specifically to solve this — the image is based on servercore but
#   ships the VS 2022 Build Tools (VC++ workload) pre-installed, with
#   `C:\BuildTools\Common7\Tools\VsDevCmd.bat` ready to source.
#
#   ltsc2022 matches GitHub Actions' windows-latest runner family.

FROM mcr.microsoft.com/visualstudio/buildtools:ltsc2022

SHELL ["powershell", "-NoProfile", "-NoLogo", "-ExecutionPolicy", "Bypass", "-Command", "$ErrorActionPreference='Stop'; $ProgressPreference='SilentlyContinue';"]

# ---------------------------------------------------------------------------
# Pull pinned versions from the shared file. PowerShell consumers parse it
# with `ConvertFrom-StringData` after stripping comment / blank lines.
# We render the parsed values into machine-wide environment variables so
# subsequent RUN layers (and the running container) see them as $env:KEY.
# ---------------------------------------------------------------------------
COPY shared\versions.env C:\loader\versions.env

RUN $raw = Get-Content C:\loader\versions.env | `
        Where-Object { $_ -and $_ -notmatch '^\s*#' }; `
    $vars = $raw | ConvertFrom-StringData; `
    foreach ($k in $vars.Keys) { `
        [System.Environment]::SetEnvironmentVariable($k, $vars[$k], 'Machine'); `
    }

# Re-import machine env into this build session so each RUN below sees
# $env:GO_VERSION etc. without re-parsing the file.
SHELL ["powershell", "-NoProfile", "-NoLogo", "-ExecutionPolicy", "Bypass", "-Command", "$ErrorActionPreference='Stop'; $ProgressPreference='SilentlyContinue'; foreach ($k in @('GO_VERSION','BUF_VERSION','PROTOBUF_VERSION','VCPKG_BASELINE_COMMIT','DOTNET_VERSION','NODE_VERSION','CMAKE_VERSION')) { Set-Item -Path \"env:$k\" -Value ([System.Environment]::GetEnvironmentVariable($k,'Machine')) };"]

# ---------------------------------------------------------------------------
# Chocolatey — bootstrap the package manager that installs Git, CMake,
# Ninja, .NET, and Node. (MSVC Build Tools come from the base image.)
# ---------------------------------------------------------------------------
RUN [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; `
    Invoke-Expression ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

ENV ChocolateyInstall=C:\ProgramData\chocolatey
RUN $env:PATH = \"$env:ChocolateyInstall\bin;$env:PATH\"; `
    [System.Environment]::SetEnvironmentVariable('PATH', $env:PATH, 'Machine')

# ---------------------------------------------------------------------------
# Git, CMake, Ninja — same versions/sources prepare.bat uses.
# Git is needed for `git clone` of vcpkg below.
# ---------------------------------------------------------------------------
RUN choco install -y --no-progress git ninja; `
    choco install -y --no-progress cmake --version=$env:CMAKE_VERSION --installargs '\"ADD_CMAKE_TO_PATH=System\"'

# Refresh PATH so the new tools are visible in this build session.
RUN $env:PATH = [System.Environment]::GetEnvironmentVariable('PATH','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('PATH','User'); `
    [System.Environment]::SetEnvironmentVariable('PATH', $env:PATH, 'Process')

# ---------------------------------------------------------------------------
# Go SDK — official MSI to C:\Go. Pinned to versions.env's GO_VERSION.
# ---------------------------------------------------------------------------
RUN $url = \"https://go.dev/dl/go${env:GO_VERSION}.windows-amd64.msi\"; `
    Invoke-WebRequest -Uri $url -OutFile C:\go.msi; `
    Start-Process msiexec.exe -ArgumentList '/i','C:\go.msi','/qn','/norestart' -Wait; `
    Remove-Item C:\go.msi
ENV GOPATH=C:\Users\ContainerUser\go
RUN [System.Environment]::SetEnvironmentVariable('PATH', `
        'C:\Program Files\Go\bin;C:\Users\ContainerUser\go\bin;' + `
        [System.Environment]::GetEnvironmentVariable('PATH','Machine'), `
        'Machine')

# ---------------------------------------------------------------------------
# buf CLI — single-binary Windows release. Pinned to versions.env's
# BUF_VERSION. Mirrors prepare.bat Step 4.
# ---------------------------------------------------------------------------
RUN New-Item -ItemType Directory -Force -Path C:\Tools\buf | Out-Null; `
    $url = \"https://github.com/bufbuild/buf/releases/download/v${env:BUF_VERSION}/buf-Windows-x86_64.exe\"; `
    Invoke-WebRequest -Uri $url -OutFile C:\Tools\buf\buf.exe; `
    [System.Environment]::SetEnvironmentVariable('PATH', `
        'C:\Tools\buf;' + [System.Environment]::GetEnvironmentVariable('PATH','Machine'), `
        'Machine')

# ---------------------------------------------------------------------------
# vcpkg + protobuf via manifest mode.
#
# Triplet: x64-windows-static. Same triplet prepare.bat uses, so a
# user moving between native (prepare.bat) and container builds avoids
# the LNK2038 _ITERATOR_DEBUG_LEVEL CRT-mismatch trap.
#
# Why manifest mode: classic-mode `vcpkg install --x-version=...` is
# silently a no-op; only manifest mode + overrides actually pin the port
# version. The post-install assertion catches the case where vcpkg's
# resolution still picks a different port revision than requested.
# ---------------------------------------------------------------------------
ARG PROTOBUF_VERSION=
ENV VCPKG_ROOT=C:\vcpkg `
    VCPKG_DEFAULT_TRIPLET=x64-windows-static

RUN git clone https://github.com/microsoft/vcpkg.git $env:VCPKG_ROOT; `
    git -C $env:VCPKG_ROOT checkout $env:VCPKG_BASELINE_COMMIT; `
    & \"$env:VCPKG_ROOT\bootstrap-vcpkg.bat\" -disableMetrics

# Render the manifest (PROTOBUF_VERSION build-arg overrides versions.env if set),
# then install via vcpkg, then assert the resolved version matches what we asked for.
# vcpkg compiles protobuf from source under MSVC; vcpkg-install.ps1 activates
# the VS Build Tools env via vsdevcmd.bat (now at C:\BuildTools\... since the
# base image is mcr.microsoft.com/visualstudio/buildtools).
COPY windows\vcpkg-install.ps1 C:\loader\vcpkg-install.ps1
RUN $effective = if ($env:PROTOBUF_VERSION) { $env:PROTOBUF_VERSION } else { `
        ([System.Environment]::GetEnvironmentVariable('PROTOBUF_VERSION','Machine')) }; `
    $env:PROTOBUF_VERSION = $effective; `
    & C:\loader\vcpkg-install.ps1

# Stable PATH entries so `protoc` and the vcpkg tools are visible.
RUN [System.Environment]::SetEnvironmentVariable('PATH', `
        'C:\vcpkg;C:\vcpkg-manifest\vcpkg_installed\x64-windows-static\tools\protobuf;' + `
        [System.Environment]::GetEnvironmentVariable('PATH','Machine'), `
        'Machine')
ENV CMAKE_PREFIX_PATH=C:\vcpkg-manifest\vcpkg_installed\x64-windows-static

# ---------------------------------------------------------------------------
# .NET SDK + Node.js LTS. Versions from versions.env. Chocolatey is the
# install channel for both.
# ---------------------------------------------------------------------------
RUN $dotnetPkg = \"dotnet-${env:DOTNET_VERSION}-sdk\"; `
    choco install -y --no-progress $dotnetPkg; `
    choco install -y --no-progress nodejs-lts --version=\"${env:NODE_VERSION}.0.0\"

# ---------------------------------------------------------------------------
# Post-create banner script (parallel to the Linux variant).
# ---------------------------------------------------------------------------
COPY shared\postcreate-banner.ps1 C:\loader\postcreate-banner.ps1

# ---------------------------------------------------------------------------
# Workspace dir. Matches devcontainer.json's workspaceFolder.
# ---------------------------------------------------------------------------
WORKDIR C:\workspaces\loader

# Drop back to a normal shell for the running container — the build-time
# version-import shell is no longer needed. Override the base image's
# ENTRYPOINT (which auto-runs vsdevcmd) so a plain `docker run` lands
# the user in a clean PowerShell — vsdevcmd is invoked on demand by
# vcpkg-install.ps1 and any user shell that needs cl.exe.
SHELL ["powershell", "-NoProfile", "-NoLogo", "-Command"]
ENTRYPOINT []
CMD ["powershell.exe", "-NoLogo"]
