# AGENTS.md — AkkuCam Local Gateway

Fully local, cloud-free Go gateway for AkkuCam base stations (for example
`SMH-T1R`) with a dependency-free browser dashboard. The gateway speaks the
LAN discovery and PPCS protocol directly, remuxes H.264 to HLS with FFmpeg and
serves the dashboard from the same binary.

Sources of truth: `specs/` for behavior and contracts (`specs/README.md` is the
index, `specs/adrs/README.md` the ADR index with status), `web/AGENTS.md` for
frontend rules, `tests/README.md` for test layout.

## Hard rules

1. **Cloud-free.** Only LAN discovery and LAN PPCS sessions. No Oguard,
   directory, push, relay or firmware endpoints, no optional fallback.
2. **No secret persistence.** The device security code stays in browser RAM
   (per DID) and transiently in Go handler memory. Never write it to disk,
   logs, specs, ADRs, fixtures or browser storage.
3. **Only verified data.** Omit metadata the gateway did not observe. No
   placeholders such as "unknown" or "not available".
4. **Host networking.** Docker Compose uses `network_mode: host` because WAI
   and PPCS rely on IPv4 broadcasts (ADR-0004).
5. **Clean room.** The protocol subset is reimplemented from observed behavior.
   No proprietary vendor library, no copied vendor code (ADR-0005).
6. **Dependencies need an ADR.** The Go module has no third-party
   dependencies; the frontend has no framework and no build step. Adding
   either requires an ADR plus license, `specs/sources.md` and SBOM updates.
7. **Vendored assets only.** No CDN, no Google Fonts. Every bundled component
   needs its license file in `web/`, an entry in the license dialog, an entry
   in `specs/sources.md` and a refreshed `sbom/` snapshot (ADR-0022/0023).
8. **PR-only main once a remote exists.** Every change runs through
   `feature/<slug>` or `fix/<slug>` and a pull request; `main` stays buildable.
   While no `origin` is configured, a local merge into `main` is allowed.
   Releases bump `internal/version/version.go` and tag `v<version>`
   (ADR-0032).

## Runtime and configuration

- HTTP on `:8080` (`AKKUCAM_LISTEN`). Static files from `AKKUCAM_WEB_ROOT`,
  default `/app/web` (exists only inside the image).
- UDP `7200` (WAI discovery) and UDP `32108` (PPCS). HLS segments are written
  to `/tmp/akkucam-hls`; there is no database and no persistent state.
- The image runs as an unprivileged user; FFmpeg is the only extra runtime tool.

## Local workflow

- Dashboard only (no rebuild): `AKKUCAM_WEB_ROOT=$PWD/web go run ./cmd/gateway`.
  Web assets are served from disk, not embedded; `web/` is copied into the
  image, so a container needs `docker compose build` after UI changes.
- Full stack: `make up`, then `curl -s http://127.0.0.1:8080/api/health`.
  `make up` passes the Git commit into the image; check it with
  `curl -s http://127.0.0.1:8080/api/version`.
- A 404 after a code change usually means an old container still owns `:8080`
  (`docker compose ps` shows image and creation time). Recreate it; a plain
  `docker build` does not replace a running container.
- `docker build` runs `go test ./...` inside the build stage, and
  `scripts/test.sh` ends with a `docker build`. A red test therefore fails the
  image build as well.

## Commands

| Purpose | Command |
|---|---|
| Go checks (vet, tests, race, image build) | `make test` or `./scripts/test.sh` |
| Static analysis (Go, JS, shell, Dockerfile, workflows) | `make lint` or `./scripts/lint.sh` |
| Browser tests (Chromium + Firefox, all projects) | `make test-e2e` or `./scripts/test-e2e.sh` |
| Live stream against real hardware | `AKKUCAM_TEST_SECURITY_CODE='<code>' AKKUCAM_TEST_LIVE_STREAM=1 ./scripts/test-e2e.sh --project=live` |
| Protocol test against a real hub | `AKKUCAM_TEST_IP=… AKKUCAM_TEST_DID=… AKKUCAM_TEST_SECURITY_CODE=… ./scripts/test-hardware.sh` |
| SBOM refresh | `make sbom` or `./scripts/sbom.sh` |
| Running version | `curl -s http://127.0.0.1:8080/api/version` |

The browser script needs a reachable gateway and installs the Playwright
dependencies inside the pinned image on first use. Every suite runs twice:
projects without a suffix use Chromium, `*-firefox` projects use Firefox. Pass
`--project=<suite>` or `--project=<suite>-firefox` to run one browser; the Live
suite receives the security code and disables traces, videos and screenshots so
nothing reaches disk. Hardware tests require a stopped gateway and use
`//go:build hardware`.

## Recipes

| Change | Touch | Verify |
|---|---|---|
| New API endpoint | route table in `internal/gateway`, `internal/openapi/openapi.json`, `internal/apierr` code if new, `web/locales/{en,de}.json` | `make test` (contract test compares both directions) |
| New error path | sentinel error with `%w`, mapping in `internal/gateway`, code in `internal/apierr`, both catalogs | `make test` (`TestErrorResponsesCarryStableCodes`, i18n tests) |
| New UI text | `web/index.html` / `web/app.js` through `web/i18n.js`, key in both catalogs | `make test` (i18n key tests) |
| Static analysis rule or finding | `.golangci.yml` / `.semgrep/gateway.yml`; `//nolint` only with reason and ADR-0033 | `make lint` |
| UI behavior or layout | see `web/AGENTS.md` | `make test-e2e` |
| Protocol detail | `internal/ppcs` with `constants.go`, `specs/protocol.md`, ADR | `make test`, optionally hardware |
| Image dependency | ADR, license dialog, `specs/sources.md`, `./scripts/sbom.sh` | `make test`, committed SBOM diff |
| Release | `internal/version/version.go`, tag `v<version>` | `make test`, `GET /api/version` carries the tag's version and commit |

## Definition of done

1. `specs/` reflects the new behavior; structural or contract changes add or
   amend an ADR below `specs/adrs/`.
2. `internal/openapi/openapi.json` matches every route; the contract test
   passes.
3. Tests are extended at the right layer: unit, API contract, fake-hub
   integration or Playwright.
4. `make lint` and `make test` pass, including `go vet`, the race detector and
   the static analysis gate.
5. User-visible changes are covered by Playwright; language switching must not
   interrupt running streams.
6. Image dependencies changed: regenerate `sbom/` and update license dialog
   and `specs/sources.md`.
7. `README.md` (English) and `README_de.md` (German) carry the same content.
   Update both in the same change and keep sections, commands, ports,
   environment variables, links and screenshots in step with the code.
   `tests/readme_test.go` checks the structure; the wording is reviewed by hand.
8. The change went through a `feature/` or `fix/` branch and a pull request,
   or was merged into `main` locally while no remote exists. A release bumps
   the version in `internal/version/version.go` and tags `main` with
   `v<version>` (ADR-0032).

## Pointers

- `specs/README.md` — document types and maintenance rules
- `CONTRIBUTING.md` — branch, pull request, commit message and release workflow
- `specs/adrs/README.md` — ADR index with status
- `specs/testing.md` — test strategy, hardware runs
- `tests/README.md` — Go test layout and naming
- `sbom/README.md` — SBOM scope and regeneration

## Protocol notes (verified)

- Hub device info: IO-Control `5` request, `6` response.
- Coupled camera list: `40962`.
- Camera parameters: `40964` GET (256 bytes), `40963` SET (write all 256).
- Motion detection uses `pirMode` (`0` off, `1` on) and must be
  read-modify-write-verify.
- Stream IDs are `<DID>-ch<N>`, one active stream per DID and channel
  (ADR-0011).
