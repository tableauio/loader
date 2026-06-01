# tableauio/loader — `.devcontainer/shared/`

Files in this directory are consumed by **multiple** devcontainer paths
(`linux/`, `windows/`) and by host-side scripts (`prepare.bat`, CI
workflows). Edit them with cross-platform parsing in mind.

## Files

| File | Consumers | Format |
| --- | --- | --- |
| [`versions.env`](./versions.env) | All Dockerfiles, `prepare.bat`, every `.github/workflows/*.yml` | `KEY=VALUE`, one per line, no quotes, no `$VAR` expansion |
| [`postcreate-banner.sh`](./postcreate-banner.sh) | `linux/devcontainer.json` `postCreateCommand` | POSIX `sh` |
| [`postcreate-banner.ps1`](./postcreate-banner.ps1) | `windows/devcontainer.json` `postCreateCommand` | PowerShell |

## `versions.env` parsing rules

Every consumer needs to read this file with at most a one-liner. The format
is therefore extremely conservative:

- **One assignment per line**, exactly `KEY=VALUE`.
- **No quotes**, no spaces around `=`, no inline comments after the value.
- **Comments start at column 0** with `#`.
- **Blank lines** are ignored.
- **No shell expansion** — values are bare literals.

Quick parsers per language:

```sh
# POSIX shell (Linux Dockerfile)
. .devcontainer/shared/versions.env
echo "$GO_VERSION"
```

```powershell
# PowerShell (Windows Dockerfile)
$v = Get-Content .devcontainer/shared/versions.env |
     Where-Object { $_ -and $_ -notmatch '^\s*#' } |
     ConvertFrom-StringData
$v.GO_VERSION
```

```cmd
:: Windows cmd (prepare.bat)
for /f "tokens=1,2 delims==" %%a in (.devcontainer\shared\versions.env) do (
    if not "%%a"=="" if not "%%a:~0,1%"=="#" set "%%a=%%b"
)
echo %GO_VERSION%
```

```yaml
# GitHub Actions (read once, export to $GITHUB_ENV)
- name: Read pinned versions
  shell: bash
  run: |
    while IFS='=' read -r k v; do
      [ -z "$k" ] && continue
      [ "${k#\#}" != "$k" ] && continue
      printf '%s=%s\n' "$k" "$v" >> "$GITHUB_ENV"
    done < .devcontainer/shared/versions.env
```

## Lockstep rule

`VCPKG_BASELINE_COMMIT` is the trickiest pin: it must know about every
`PROTOBUF_VERSION` value used anywhere (devcontainer default + every
`testing-cpp.yml` matrix entry). Bumping `PROTOBUF_VERSION` to a value
the current baseline doesn't know is caught at build time by the
post-install assertion in both Dockerfiles and `prepare.bat` — fail loud,
no silent wrong-version installs.
