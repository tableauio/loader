# syntax=docker/dockerfile:1.7
# tableauio/loader devcontainer
#
# Single-stage, multi-arch (amd64 + arm64) image bringing the full
# C++/Go/.NET/Node toolchain plus protobuf 6.33.4 (via vcpkg) at the
# exact versions CI uses. See docs/superpowers/specs/2026-05-29-devcontainer-design.md.

FROM mcr.microsoft.com/devcontainers/cpp:1-ubuntu-24.04

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
# Go 1.24.0 — official tarball into /usr/local/go.
#
# PATH is set via ENV (not /etc/profile.d/) so non-interactive shells like
# the postCreateCommand and downstream RUNs see Go without sourcing profile.
# /home/vscode/go/bin lands `go install`-placed binaries on PATH automatically.
# ---------------------------------------------------------------------------
ARG GO_VERSION=1.24.0
RUN <<EOF
set -eux
. /opt/buildargs.env
curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz" \
  | tar -C /usr/local -xz
EOF
ENV PATH=/usr/local/go/bin:/home/vscode/go/bin:${PATH}

# ---------------------------------------------------------------------------
# buf 1.67.0 — single-binary release on PATH.
# ---------------------------------------------------------------------------
ARG BUF_VERSION=1.67.0
RUN <<EOF
set -eux
. /opt/buildargs.env
curl -fsSL -o /usr/local/bin/buf \
  "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-Linux-${BUF_ARCH}"
chmod +x /usr/local/bin/buf
EOF

# ---------------------------------------------------------------------------
# vcpkg + protobuf via manifest mode.
#
# Pins:
#   VCPKG_BASELINE_COMMIT — same commit testing-cpp.yml's VCPKG_COMMIT and
#       prepare.bat's VCPKG_BASELINE_COMMIT use. Bumping vcpkg = bump all three.
#   PROTOBUF_VERSION    — defaults to the modern 6.33.4 line (CI's primary).
#       Override at build time:
#         docker build --build-arg PROTOBUF_VERSION=3.21.12 ...
#       devcontainer.json wires this to the LOADER_PROTOBUF_VERSION host env var.
#
# Why manifest mode: classic-mode `vcpkg install --x-version=...` is silently
# a no-op; only manifest mode + overrides actually pin the port version. The
# post-install assertion catches the case where vcpkg's resolution still picks
# a different port revision than requested.
# ---------------------------------------------------------------------------
ARG VCPKG_BASELINE_COMMIT=dc8d75cfc3281b8e2a4ed8ee4163c891190df932
ARG PROTOBUF_VERSION=6.33.4
ENV VCPKG_ROOT=/opt/vcpkg

RUN <<EOF
set -eux
. /opt/buildargs.env

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
    echo "       Bump VCPKG_BASELINE_COMMIT (in this Dockerfile, prepare.bat,"
    echo "       and testing-cpp.yml) to a commit that knows about the requested version."
    exit 1
    ;;
esac

# 5. Stable symlinks so ENV CMAKE_PREFIX_PATH (last layer) doesn't have to
#    care about the underlying triplet.
ln -s /opt/vcpkg-manifest/vcpkg_installed/${VCPKG_TRIPLET} /opt/vcpkg/active
ln -s /opt/vcpkg/active/tools/protobuf/protoc /usr/local/bin/protoc
EOF

# ---------------------------------------------------------------------------
# .NET SDK 8.0 + Node.js 20 LTS — apt-based installs from the official
# Microsoft and NodeSource repositories. apt-get clean + rm /var/lib/apt/lists
# at the end keeps the layer small.
# ---------------------------------------------------------------------------
RUN <<EOF
set -eux
curl -fsSL https://packages.microsoft.com/config/ubuntu/24.04/packages-microsoft-prod.deb \
    -o /tmp/ms.deb
dpkg -i /tmp/ms.deb
rm /tmp/ms.deb
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y --no-install-recommends \
    dotnet-sdk-8.0 \
    nodejs
apt-get clean
rm -rf /var/lib/apt/lists/*
EOF
