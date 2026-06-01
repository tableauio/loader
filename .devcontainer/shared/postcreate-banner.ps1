# Post-create banner for the Windows devcontainer.
# Same shape as ../shared/postcreate-banner.sh (Linux) — five lines,
# pure version queries, no installs.
$ErrorActionPreference = 'Stop'
Write-Host 'tableauio/loader devcontainer ready (windows).'
Write-Host ('  go:     {0}' -f ((go version) -split '\s+')[2])
Write-Host ('  buf:    {0}' -f (buf --version 2>&1))
Write-Host ('  protoc: {0}' -f (protoc --version))
Write-Host ('  dotnet: {0}' -f (dotnet --version))
Write-Host ('  node:   {0}' -f (node --version))
