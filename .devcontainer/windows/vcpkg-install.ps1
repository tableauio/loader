# tableauio/loader — Windows-container vcpkg + protobuf install.
#
# Called from windows/Dockerfile during image build. Mirrors the Linux
# Dockerfile's vcpkg manifest-mode install + post-install version
# assertion, in PowerShell + cmd.
#
# Inputs (env vars; set in the Dockerfile from versions.env):
#   PROTOBUF_VERSION         — protobuf vcpkg port version, e.g. 6.33.4
#   VCPKG_BASELINE_COMMIT    — commit to pin in builtin-baseline
#   VCPKG_ROOT               — C:\vcpkg
#
# Side effects:
#   - Renders C:\vcpkg-manifest\vcpkg.json
#   - Runs `vcpkg install --triplet=x64-windows-static`
#   - Asserts the resolved port version matches PROTOBUF_VERSION
#
# vcpkg compiles protobuf from source under MSVC — that needs the VS Build
# Tools env (cl.exe, INCLUDE, LIB) active in the running shell. We invoke
# vsdevcmd.bat through cmd.exe and capture the resulting environment, then
# replay it into PowerShell before invoking vcpkg.

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$triplet = 'x64-windows-static'
$manifestDir = 'C:\vcpkg-manifest'
$installRoot = Join-Path $manifestDir 'vcpkg_installed'
New-Item -ItemType Directory -Force -Path $manifestDir | Out-Null

# 1. Render manifest. JSON-escape the substituted values just in case.
$pv = $env:PROTOBUF_VERSION
$bc = $env:VCPKG_BASELINE_COMMIT
if (-not $pv) { throw 'PROTOBUF_VERSION env var is empty.' }
if (-not $bc) { throw 'VCPKG_BASELINE_COMMIT env var is empty.' }

$manifest = @{
    name              = 'loader-devcontainer-windows'
    version           = '0.1.0'
    dependencies      = @('protobuf')
    overrides         = @(@{ name = 'protobuf'; version = $pv })
    'builtin-baseline' = $bc
} | ConvertTo-Json -Depth 4
Set-Content -Path (Join-Path $manifestDir 'vcpkg.json') -Value $manifest -Encoding ASCII

# 2. Activate the VS 2022 Build Tools environment for this PS session.
#    On the mcr.microsoft.com/visualstudio/buildtools base image, the
#    Build Tools live at a fixed `C:\BuildTools\` prefix, so we go straight
#    there instead of probing via vswhere. Fall back to vswhere if a future
#    base image changes this layout.
$primaryVsdevcmd = 'C:\BuildTools\Common7\Tools\VsDevCmd.bat'
if (Test-Path $primaryVsdevcmd) {
    $vsdevcmd = $primaryVsdevcmd
} else {
    $vswhere = 'C:\Program Files (x86)\Microsoft Visual Studio\Installer\vswhere.exe'
    if (-not (Test-Path $vswhere)) {
        throw "Neither $primaryVsdevcmd nor $vswhere found; VS Build Tools layer is missing."
    }
    $installPath = & $vswhere -latest -products * `
        -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 `
        -property installationPath
    if (-not $installPath) { throw 'No VS install with C++ tools detected.' }
    $vsdevcmd = Join-Path $installPath 'Common7\Tools\VsDevCmd.bat'
    if (-not (Test-Path $vsdevcmd)) { throw "VsDevCmd.bat not found: $vsdevcmd" }
}
Write-Host "Using VsDevCmd at: $vsdevcmd"

# Capture the environment that vsdevcmd produces.
$envDump = & cmd.exe /s /c "`"$vsdevcmd`" -arch=amd64 -host_arch=amd64 && set"
foreach ($line in $envDump) {
    if ($line -match '^([^=]+)=(.*)$') {
        Set-Item -Path "env:$($Matches[1])" -Value $Matches[2]
    }
}
if (-not (Get-Command cl.exe -ErrorAction SilentlyContinue)) {
    throw 'cl.exe not on PATH after VsDevCmd activation; cannot proceed.'
}

# 3. Manifest-mode install.
Push-Location $manifestDir
try {
    & "$env:VCPKG_ROOT\vcpkg.exe" install `
        "--triplet=$triplet" `
        "--x-install-root=$installRoot"
    if ($LASTEXITCODE -ne 0) {
        throw "vcpkg install failed (exit code $LASTEXITCODE)"
    }
} finally {
    Pop-Location
}

# 4. Post-install assertion — same shape as the Linux Dockerfile's case
#    statement and prepare.bat's findstr check.
$infoDir = Join-Path $installRoot 'vcpkg\info'
$marker = Get-ChildItem -Path $infoDir -Filter "protobuf_*_${triplet}.list" `
              -ErrorAction SilentlyContinue | Select-Object -First 1
if (-not $marker) {
    throw "vcpkg installed-file marker not found under $infoDir"
}
if ($marker.Name -notlike "protobuf_${pv}*") {
    Write-Error "Installed protobuf does not match requested version $pv."
    Write-Error "  vcpkg installed-file marker: $($marker.Name)"
    Write-Error "  Bump VCPKG_BASELINE_COMMIT in .devcontainer/shared/versions.env"
    Write-Error "  to a commit that knows about the requested version."
    exit 1
}

Write-Host "vcpkg protobuf $pv installed under $installRoot ($triplet)."
