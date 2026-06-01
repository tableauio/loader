#!/bin/sh
# Post-create banner for the Linux devcontainer.
# Pure echo — no installs, no version-pinning at runtime, no surprises.
# Mirrors the same five-line summary the previous inline postCreateCommand
# emitted; extracted to a script so the Windows container can have a
# parallel postcreate-banner.ps1 with the same shape.
set -e
printf 'tableauio/loader devcontainer ready (linux).\n'
printf '  go:     %s\n' "$(go version | cut -d' ' -f3)"
printf '  buf:    %s\n' "$(buf --version 2>&1)"
printf '  protoc: %s\n' "$(protoc --version)"
printf '  dotnet: %s\n' "$(dotnet --version)"
printf '  node:   %s\n' "$(node --version)"
