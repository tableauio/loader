# syntax=docker/dockerfile:1.7
# tableauio/loader devcontainer
#
# Single-stage, multi-arch (amd64 + arm64) image bringing the full
# C++/Go/.NET/Node toolchain plus protobuf at the exact versions CI uses.
#
# All version pins are read from ./versions.env (the single source of
# truth shared with prepare.bat and the CI workflows). To bump Go / buf /
# protobuf / .NET / Node / vcpkg-baseline, edit that file — not this one.
#
# Build context is .devcontainer/ (the directory containing this file),
# so `COPY versions.env ...` resolves directly.

FROM mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04

# ---------------------------------------------------------------------------
# Pull pinned versions from the shared file. We `COPY` it into a build-stage
# location and `source` it in every RUN that needs the values; ARG/ENV alone
# don't survive across RUN boundaries in BuildKit, so we re-source per layer.
# ---------------------------------------------------------------------------
COPY versions.env /opt/versions.env

# ---------------------------------------------------------------------------
# Architecture detection. BuildKit auto-populates TARGETARCH; we resolve it
# into per-arch download-name fragments (Go's tarball, buf's release asset,
# vcpkg triplet) and persist them to /opt/buildargs.env so later RUN layers
# can `source` them — Dockerfile ARGs don't survive across RUN boundaries.
# ---------------------------------------------------------------------------
ARG TARGETARCH
RUN <<EOF
set -eux
case "${TARGETARCH}" in
  amd64) GO_ARCH=amd64; BUF_ARCH=x86_64;  TRIPLET=x64-linux   ;;
  arm64) GO_ARCH=arm64; BUF_ARCH=aarch64; TRIPLET=arm64-linux ;;
  *)     echo "unsupported TARGETARCH: ${TARGETARCH}"; exit 1 ;;
esac
mkdir -p /opt
printf 'GO_ARCH=%s\nBUF_ARCH=%s\nVCPKG_TRIPLET=%s\n' \
  "${GO_ARCH}" "${BUF_ARCH}" "${TRIPLET}" > /opt/buildargs.env
EOF

# ---------------------------------------------------------------------------
# Go — official tarball into /usr/local/go.
#
# PATH is set via ENV (not /etc/profile.d/) so non-interactive shells like
# the postCreateCommand and downstream RUNs see Go without sourcing profile.
# /home/vscode/go/bin lands `go install`-placed binaries on PATH automatically.
# ---------------------------------------------------------------------------
RUN <<EOF
set -eux
. /opt/versions.env
. /opt/buildargs.env
curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" \
  | tar -C /usr/local -xz
EOF
ENV PATH=/usr/local/go/bin:/home/vscode/go/bin:${PATH}

# ---------------------------------------------------------------------------
# buf — single-binary release on PATH.
# ---------------------------------------------------------------------------
RUN <<EOF
set -eux
. /opt/versions.env
. /opt/buildargs.env
curl -fsSL -o /usr/local/bin/buf \
  "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-Linux-${BUF_ARCH}"
chmod +x /usr/local/bin/buf
EOF

# ---------------------------------------------------------------------------
# vcpkg + protobuf via manifest mode.
#
# Pins (all from /opt/versions.env):
#   VCPKG_BASELINE_COMMIT — same commit testing-cpp.yml's VCPKG_COMMIT and
#       prepare.bat's VCPKG_BASELINE_COMMIT use. Bumping vcpkg = bump that
#       one line in versions.env.
#   PROTOBUF_VERSION    — defaults to versions.env's value (CI's primary).
#       Override at build time:
#         docker build --build-arg PROTOBUF_VERSION=3.21.12 ...
#       devcontainer.json wires this to the LOADER_PROTOBUF_VERSION host env var.
#
# Why manifest mode: classic-mode `vcpkg install --x-version=...` is silently
# a no-op; only manifest mode + overrides actually pin the port version. The
# post-install assertion catches the case where vcpkg's resolution still picks
# a different port revision than requested.
# ---------------------------------------------------------------------------
ARG PROTOBUF_VERSION=
ENV VCPKG_ROOT=/opt/vcpkg

# `_ARG_PROTOBUF_VERSION` captures the Dockerfile ARG via Dockerfile-level
# variable expansion BEFORE we source versions.env in the heredoc body
# (otherwise the file default would clobber the build-arg). Empty string
# means "no override; use file default".
RUN <<EOF
set -eux
_ARG_PROTOBUF_VERSION="${PROTOBUF_VERSION}"
. /opt/versions.env
. /opt/buildargs.env

# Prefer the build-arg when non-empty; otherwise the value just sourced
# from versions.env stays.
if [ -n "${_ARG_PROTOBUF_VERSION:-}" ]; then
  PROTOBUF_VERSION="${_ARG_PROTOBUF_VERSION}"
fi

# 1. Bring up vcpkg pinned to the baseline commit.
git clone https://github.com/microsoft/vcpkg.git "${VCPKG_ROOT}"
git -C "${VCPKG_ROOT}" checkout "${VCPKG_BASELINE_COMMIT}"
"${VCPKG_ROOT}/bootstrap-vcpkg.sh" -disableMetrics

# 2. Render a minimal manifest with builtin-baseline + the protobuf override.
mkdir -p /opt/vcpkg-manifest
cat > /opt/vcpkg-manifest/vcpkg.json <<MANIFEST
{
  "name": "loader-devcontainer",
  "version": "0.1.0",
  "dependencies": ["protobuf"],
  "overrides": [{ "name": "protobuf", "version": "${PROTOBUF_VERSION}" }],
  "builtin-baseline": "${VCPKG_BASELINE_COMMIT}"
}
MANIFEST

# 3. Manifest-mode install. Triplet comes from /opt/buildargs.env
#    (x64-linux on amd64, arm64-linux on arm64).
cd /opt/vcpkg-manifest
"${VCPKG_ROOT}/vcpkg" install \
    --triplet="${VCPKG_TRIPLET}" \
    --x-install-root=/opt/vcpkg-manifest/vcpkg_installed

# 4. Post-install assertion: vcpkg writes a per-port file whose name encodes
#    the resolved version. If the prefix doesn't match what we asked for,
#    fail loudly — silently producing a wrong-version image is the bug we
#    are explicitly defending against.
INFO_FILE=$(ls /opt/vcpkg-manifest/vcpkg_installed/vcpkg/info/protobuf_*_${VCPKG_TRIPLET}.list 2>/dev/null | head -n1)
case "$(basename "${INFO_FILE:-/missing}" 2>/dev/null)" in
  protobuf_${PROTOBUF_VERSION}*)
    ;;
  *)
    echo "ERROR: installed protobuf does not match requested version ${PROTOBUF_VERSION}."
    echo "       vcpkg installed-file marker: ${INFO_FILE:-<none>}"
    echo "       Bump VCPKG_BASELINE_COMMIT in .devcontainer/versions.env"
    echo "       (and prepare.bat + testing-cpp.yml will pick it up automatically)"
    echo "       to a commit that knows about the requested version."
    exit 1
    ;;
esac

# 5. Stable symlinks so ENV CMAKE_PREFIX_PATH (last layer) doesn't have to
#    care about the underlying triplet.
ln -s /opt/vcpkg-manifest/vcpkg_installed/${VCPKG_TRIPLET} /opt/vcpkg/active
ln -s /opt/vcpkg/active/tools/protobuf/protoc /usr/local/bin/protoc
EOF

# ---------------------------------------------------------------------------
# .NET SDK + Node.js — apt-based installs from the official Microsoft and
# NodeSource repositories. Versions are read from /opt/versions.env.
# apt-get clean + rm /var/lib/apt/lists at the end keeps the layer small.
# ---------------------------------------------------------------------------
RUN <<EOF
set -eux
. /opt/versions.env
curl -fsSL https://packages.microsoft.com/config/ubuntu/24.04/packages-microsoft-prod.deb \
    -o /tmp/ms.deb
dpkg -i /tmp/ms.deb
rm /tmp/ms.deb
curl -fsSL "https://deb.nodesource.com/setup_${NODE_VERSION}.x" | bash -
apt-get install -y --no-install-recommends \
    "dotnet-sdk-${DOTNET_VERSION}" \
    nodejs
apt-get clean
rm -rf /var/lib/apt/lists/*
EOF

# ---------------------------------------------------------------------------
# Final environment: with CMAKE_PREFIX_PATH pointing at vcpkg's installed
# tree, find_package(Protobuf CONFIG) resolves automatically — the existing
# README's "Dev at Linux → CMake (system protobuf)" recipe works as written
# inside the container.
# ---------------------------------------------------------------------------
ENV CMAKE_PREFIX_PATH=/opt/vcpkg/active

# Copy the post-create banner script into a stable on-image path so
# devcontainer.json can call it without needing the workspace mount.
COPY postcreate-banner.sh /usr/local/bin/loader-devcontainer-banner
RUN chmod +x /usr/local/bin/loader-devcontainer-banner
