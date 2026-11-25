#!/usr/bin/env sh
# Static analysis for the gateway, Go module and web assets.
#
# Every tool runs from a pinned container image; no local installation is
# needed on NixOS or in CI. Image variables below are the single source of
# truth for the versions used locally and in CI (ADR-0033).
#
# Usage:
#   ./scripts/lint.sh
set -eu

ROOT=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)

GO_IMAGE="${GO_IMAGE:-golang:1.27.4-bookworm}"
GOVULNCHECK_VERSION="${GOVULNCHECK_VERSION:-v1.10.0}"
GOLANGCI_IMAGE="${GOLANGCI_IMAGE:-golangci/golangci-lint:v2.15.0}"
SEMGREP_IMAGE="${SEMGREP_IMAGE:-semgrep/semgrep:1.187.0}"
SHELLCHECK_IMAGE="${SHELLCHECK_IMAGE:-koalaman/shellcheck:v0.11.0}"
HADOLINT_IMAGE="${HADOLINT_IMAGE:-hadolint/hadolint:v2.15.1}"
ACTIONLINT_IMAGE="${ACTIONLINT_IMAGE:-rhysd/actionlint:1.7.12}"

# Caches stay in the container; the repository is mounted read-only.
docker run --rm -e GOCACHE=/tmp/gocache -e GOMODCACHE=/tmp/gomod \
  -v "$ROOT:/src:ro" -w /src "$GO_IMAGE" \
  go run "golang.org/x/vuln/cmd/govulncheck@$GOVULNCHECK_VERSION" ./...

docker run --rm -e GOLANGCI_LINT_CACHE=/tmp/golangci-lint -e GOCACHE=/tmp/gocache \
  -v "$ROOT:/src:ro" -w /src "$GOLANGCI_IMAGE" \
  golangci-lint run --timeout 10m ./...

# Semgrep fetches the registry rulesets at scan time and sends no metrics.
docker run --rm -e HOME=/tmp -e SEMGREP_SEND_METRICS=off \
  -v "$ROOT:/src:ro" -w /src "$SEMGREP_IMAGE" \
  semgrep scan --error --metrics=off \
    --config p/golang \
    --config p/javascript \
    --config p/secrets \
    --config .semgrep/ \
    --exclude web/e2e/node_modules \
    --exclude sbom

docker run --rm -v "$ROOT:/src:ro" -w /src "$SHELLCHECK_IMAGE" scripts/*.sh

docker run --rm -v "$ROOT:/src:ro" -w /src "$HADOLINT_IMAGE" \
  hadolint --failure-threshold warning Dockerfile

docker run --rm -v "$ROOT:/src:ro" -w /src "$ACTIONLINT_IMAGE"
