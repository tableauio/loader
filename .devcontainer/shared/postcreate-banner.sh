#!/bin/sh
# Post-create banner for the Linux devcontainer.
# Pure echo — no installs, no version-pinning at runtime, no surprises.
# Five-line summary that prints when the container becomes ready, so the
# developer can confirm at a glance which toolchain versions landed.
set -e
printf 'tableauio/loader devcontainer ready (linux).\n'
printf '  go:     %s\n' "$(go version | cut -d' ' -f3)"
printf '  buf:    %s\n' "$(buf --version 2>&1)"
printf '  protoc: %s\n' "$(protoc --version)"
printf '  dotnet: %s\n' "$(dotnet --version)"
printf '  node:   %s\n' "$(node --version)"
