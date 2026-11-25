#!/usr/bin/env sh
# Generate Software Bill of Materials (SBOM) for the gateway container image.
#
# Outputs (both machine-readable):
#   sbom/akkucam-gateway.cyclonedx.json  CycloneDX 1.x JSON
#   sbom/akkucam-gateway.spdx.json       SPDX 2.3 JSON
#
# Usage:
#   ./scripts/sbom.sh [image]
#
# The image defaults to $AKKUCAM_IMAGE or "akkucam-gateway:sbom". If the image
# does not exist locally, the script builds it from this repository. The scan
# runs in the pinned Syft container; no Syft installation is required on the
# host. The image is exported with "docker save" first, so the scan container
# never needs access to the Docker socket.
set -eu

ROOT=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
IMAGE="${1:-${AKKUCAM_IMAGE:-akkucam-gateway:sbom}}"
SYFT_IMAGE="${SYFT_IMAGE:-anchore/syft:v1.57.2}"
OUT="${SBOM_DIR:-$ROOT/sbom}"
NAME="${SBOM_NAME:-akkucam-gateway}"
VERSION="${SBOM_VERSION:-$(git -C "$ROOT" rev-parse --short HEAD 2>/dev/null || echo dev)}"

if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  echo "Image $IMAGE not found locally; building it from $ROOT" >&2
  docker build -t "$IMAGE" "$ROOT"
fi

mkdir -p "$OUT"

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT INT TERM
mkdir -p "$WORK/home" "$WORK/cache" "$WORK/tmp"

# Export the image so Syft does not need the Docker socket. The archive is
# removed together with the temporary directory.
docker save "$IMAGE" -o "$WORK/image.tar"

docker run --rm \
  --user "$(id -u):$(id -g)" \
  -e HOME=/work/home \
  -e XDG_CACHE_HOME=/work/cache \
  -e TMPDIR=/work/tmp \
  -v "$WORK:/work" \
  -v "$OUT:/out" \
  "$SYFT_IMAGE" \
  "docker-archive:/work/image.tar" \
  --source-name "$NAME" \
  --source-supplier "André Raabe" \
  --source-version "$VERSION" \
  -o "cyclonedx-json=/out/$NAME.cyclonedx.json" \
  -o "spdx-json=/out/$NAME.spdx.json"

echo "Wrote $OUT/$NAME.cyclonedx.json" >&2
echo "Wrote $OUT/$NAME.spdx.json" >&2
