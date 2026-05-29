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
