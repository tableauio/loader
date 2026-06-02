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
# Symmetry with linux/Dockerfile: that file COPYs versions.env into the
# image and `source`s it per-RUN. Windows has no POSIX `source`, so we
# do the equivalent by baking a small `Import-LoaderVersions` helper
# into a stable on-image path, then making every RUN's SHELL prefix call
# it before the body executes. Net effect: every RUN sees $env:GO_VERSION
# (and friends) without depending on Machine-scope env, without enumerating
# --build-arg flags from the wrapper, and without re-parsing versions.env
# in every RUN body.
#
# Build context is .devcontainer/ (the parent of this directory) so that
# `COPY shared\...` resolves. devcontainer.json sets `"context": ".."`.
#
# Why mcr.microsoft.com/dotnet/framework/runtime (not windows/servercore):
#
#   The modern VS 2022 Build Tools installer (`vs_BuildTools.exe`)
#   requires .NET Framework runtime support that `windows/servercore`
#   strips out. Microsoft's own guidance — see
#   https://learn.microsoft.com/en-us/visualstudio/install/build-tools-container
#   — explicitly says: "If you base your image directly on
#   microsoft/windowsservercore, the .NET Framework might not install
#   properly … Instead, base your image on microsoft/dotnet-framework:4.7.1
#   or later." We use 4.8-windowsservercore-ltsc2022, which is also one
#   of the images pre-cached on GitHub's windows-2022 runners.
#
#   ltsc2022 matches GitHub Actions' windows-latest runner family.
#
#   Build memory: pass `docker build -m 2GB` (or higher). vs_BuildTools
#   exhausts the default 1 GB silently. CI's windows-2022 runners have
#   plenty of RAM but you must set the flag.

FROM mcr.microsoft.com/dotnet/framework/runtime:4.8-windowsservercore-ltsc2022

# ---------------------------------------------------------------------------
# Drop versions.env on disk and a helper script that imports it into the
# current PowerShell session's $env:KEY namespace. Both are stable,
# read-only, and shared across every later RUN.
# ---------------------------------------------------------------------------
COPY shared\versions.env C:\loader\versions.env
COPY windows\Import-LoaderVersions.ps1 C:\loader\Import-LoaderVersions.ps1

# Default SHELL: PowerShell, with `Import-LoaderVersions` already invoked
# so $env:GO_VERSION etc. are live before the RUN body runs. Mirrors the
# linux/Dockerfile heredoc that begins with `. /opt/versions.env`.
SHELL ["powershell", "-NoProfile", "-NoLogo", "-ExecutionPolicy", "Bypass", "-Command", "$ErrorActionPreference='Stop'; $ProgressPreference='SilentlyContinue'; . C:\\loader\\Import-LoaderVersions.ps1; Import-LoaderVersions;"]

# ---------------------------------------------------------------------------
# Visual Studio 2022 Build Tools — VC++ workload only.
#
# Mirrors Microsoft's documented Dockerfile from
# https://learn.microsoft.com/en-us/visualstudio/install/build-tools-container
# but with `--add Microsoft.VisualStudio.Workload.VCTools` (we don't need
# the AzureBuildTools workload) and the same `--remove` flags for old
# Windows SDKs that can't install in a container.
#
# Exit code 3010 is "success, reboot required" — fine inside a container.
# ---------------------------------------------------------------------------
RUN curl.exe -SL --output C:\vs_buildtools.exe https://aka.ms/vs/17/release/vs_buildtools.exe; `
    $proc = Start-Process C:\vs_buildtools.exe `
        -ArgumentList @( `
            '--quiet','--wait','--norestart','--nocache', `
            '--installPath','C:\BuildTools', `
            '--add','Microsoft.VisualStudio.Workload.VCTools', `
            '--add','Microsoft.VisualStudio.Component.VC.Tools.x86.x64', `
            '--add','Microsoft.VisualStudio.Component.Windows11SDK.22621', `
            '--remove','Microsoft.VisualStudio.Component.Windows10SDK.10240', `
            '--remove','Microsoft.VisualStudio.Component.Windows10SDK.10586', `
            '--remove','Microsoft.VisualStudio.Component.Windows10SDK.14393', `
            '--remove','Microsoft.VisualStudio.Component.Windows81SDK' `
        ) -Wait -NoNewWindow -PassThru; `
    if ($proc.ExitCode -ne 0 -and $proc.ExitCode -ne 3010) { `
        Write-Error \"vs_buildtools failed with exit code $($proc.ExitCode)\"; `
        exit $proc.ExitCode `
    }; `
    Remove-Item C:\vs_buildtools.exe -Force

# ---------------------------------------------------------------------------
# Chocolatey — bootstrap the package manager that installs Git, CMake,
# Ninja, .NET, and Node.
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
# Go SDK — official MSI to C:\Program Files\Go.
# ---------------------------------------------------------------------------
RUN $url = \"https://go.dev/dl/go${env:GO_VERSION}.windows-amd64.msi\"; `
    Write-Host \"Downloading $url\"; `
    Invoke-WebRequest -Uri $url -OutFile C:\go.msi; `
    Start-Process msiexec.exe -ArgumentList '/i','C:\go.msi','/qn','/norestart' -Wait; `
    Remove-Item C:\go.msi
ENV GOPATH=C:\Users\ContainerUser\go
RUN [System.Environment]::SetEnvironmentVariable('PATH', `
        'C:\Program Files\Go\bin;C:\Users\ContainerUser\go\bin;' + `
        [System.Environment]::GetEnvironmentVariable('PATH','Machine'), `
        'Machine')

# ---------------------------------------------------------------------------
# buf CLI — single-binary Windows release.
# ---------------------------------------------------------------------------
RUN New-Item -ItemType Directory -Force -Path C:\Tools\buf | Out-Null; `
    $url = \"https://github.com/bufbuild/buf/releases/download/v${env:BUF_VERSION}/buf-Windows-x86_64.exe\"; `
    Write-Host \"Downloading $url\"; `
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
# ---------------------------------------------------------------------------
ENV VCPKG_ROOT=C:\vcpkg `
    VCPKG_DEFAULT_TRIPLET=x64-windows-static

RUN git clone https://github.com/microsoft/vcpkg.git $env:VCPKG_ROOT; `
    git -C $env:VCPKG_ROOT checkout $env:VCPKG_BASELINE_COMMIT; `
    & \"$env:VCPKG_ROOT\bootstrap-vcpkg.bat\" -disableMetrics

# Render the manifest, install via vcpkg, assert the resolved version
# matches PROTOBUF_VERSION. vcpkg compiles protobuf from source under MSVC;
# vcpkg-install.ps1 activates the VS Build Tools env via vsdevcmd.bat.
COPY windows\vcpkg-install.ps1 C:\loader\vcpkg-install.ps1
RUN & C:\loader\vcpkg-install.ps1

# Stable PATH entries so `protoc` and the vcpkg tools are visible.
RUN [System.Environment]::SetEnvironmentVariable('PATH', `
        'C:\vcpkg;C:\vcpkg-manifest\vcpkg_installed\x64-windows-static\tools\protobuf;' + `
        [System.Environment]::GetEnvironmentVariable('PATH','Machine'), `
        'Machine')
ENV CMAKE_PREFIX_PATH=C:\vcpkg-manifest\vcpkg_installed\x64-windows-static

# ---------------------------------------------------------------------------
# .NET SDK + Node.js LTS via Chocolatey.
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

# Final SHELL: drop the auto-import prefix for the running container —
# end-users don't need versions.env in their shell session, and keeping
# it in the SHELL prefix would slow every interactive command. The
# helper is still on disk at C:\loader\Import-LoaderVersions.ps1 for any
# script that wants it.
SHELL ["powershell", "-NoProfile", "-NoLogo", "-Command"]
ENTRYPOINT []
CMD ["powershell.exe", "-NoLogo"]
