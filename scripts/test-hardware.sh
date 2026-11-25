#!/usr/bin/env sh
set -eu

: "${AKKUCAM_TEST_IP:?set AKKUCAM_TEST_IP}"
: "${AKKUCAM_TEST_DID:?set AKKUCAM_TEST_DID}"
: "${AKKUCAM_TEST_SECURITY_CODE:?set AKKUCAM_TEST_SECURITY_CODE}"

if docker ps --format '{{.Names}}' | grep -qx 'akkucam-gateway'; then
  echo "error: stop the running gateway first: docker compose stop" >&2
  echo "the basis station may authenticate a competing session without sending it video" >&2
  exit 2
fi

docker run --rm --network host \
  -e AKKUCAM_TEST_IP \
  -e AKKUCAM_TEST_DID \
  -e AKKUCAM_TEST_SECURITY_CODE \
  -e AKKUCAM_TEST_CHANNEL="${AKKUCAM_TEST_CHANNEL:-0}" \
  -v "$PWD:/src" -w /src \
  golang:1.27.4-alpine \
  go test -tags=hardware -count=1 -v ./internal/ppcs -run TestHardware
