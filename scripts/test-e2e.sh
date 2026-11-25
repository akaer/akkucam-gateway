#!/usr/bin/env sh
set -eu

ROOT=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
E2E="$ROOT/web/e2e"
IMAGE="${PLAYWRIGHT_IMAGE:-mcr.microsoft.com/playwright:v1.65.1-jammy}"
BASE_URL="${AKKUCAM_BASE_URL:-http://127.0.0.1:8080}"
USER_ID=$(id -u)
GROUP_ID=$(id -g)

if ! curl -fsS "$BASE_URL/api/health" >/dev/null 2>&1; then
  echo "gateway not reachable at $BASE_URL; start it with: docker compose up -d --wait" >&2
  exit 1
fi

if [ ! -x "$E2E/node_modules/.bin/playwright" ]; then
  docker run --rm --network host \
    --user "$USER_ID:$GROUP_ID" \
    -e HOME=/tmp \
    -v "$ROOT:/work" \
    -w /work/web/e2e \
    "$IMAGE" npm ci --no-audit --no-fund
fi

docker run --rm --network host \
  --user "$USER_ID:$GROUP_ID" \
  -e HOME=/tmp \
  -e AKKUCAM_BASE_URL="$BASE_URL" \
  -e AKKUCAM_TEST_SECURITY_CODE="${AKKUCAM_TEST_SECURITY_CODE:-}" \
  -e AKKUCAM_TEST_HUB_DID="${AKKUCAM_TEST_HUB_DID:-}" \
  -e AKKUCAM_TEST_LIVE_STREAM="${AKKUCAM_TEST_LIVE_STREAM:-}" \
  -v "$ROOT:/work" \
  -w /work/web/e2e \
  "$IMAGE" npx playwright test "$@"
