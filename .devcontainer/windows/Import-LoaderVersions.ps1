# tableauio/loader devcontainer (windows) - versions.env importer.
#
# Mirrors the linux/Dockerfile pattern of `. /opt/versions.env` in every
# RUN that needs the values. PowerShell has no `source`-equivalent for
# KEY=VALUE files, so we provide one as a function.
#
# Loaded into the SHELL prefix of the windows/Dockerfile so every RUN
# layer's body executes with $env:GO_VERSION etc. already populated from
# C:\loader\versions.env.

function Import-LoaderVersions {
    [CmdletBinding()]
    param(
        [string]$Path = 'C:\loader\versions.env'
    )
    if (-not (Test-Path $Path)) {
        throw "versions.env not found at $Path"
    }
    # Strip blank lines and # comments, then JOIN into one string before
    # passing to ConvertFrom-StringData. The pipeline form
    #   Get-Content $Path | ConvertFrom-StringData
    # is a trap: it produces ONE HASHTABLE PER INPUT LINE (7 separate
    # single-key hashtables), not one merged hashtable. The string form
    # parses the entire blob as a single hashtable.
    $body = (Get-Content $Path |
        Where-Object { $_ -and $_ -notmatch '^\s*#' }) -join "`n"
    $vars = ConvertFrom-StringData -StringData $body
    foreach ($k in $vars.Keys) {
        Set-Item -Path "env:$k" -Value $vars[$k]
    }
}

