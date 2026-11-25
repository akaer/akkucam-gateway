# Test strategy

## Test levels

### Unit tests

Run inside the Docker build:

```bash
go test ./...
```

Currently covered:

- complete DID decoding from PPPP PUNCH and discovery
- malformed and truncated headers and length fields
- DRW media fragment parser
- fragment ordering including 16-bit wraparound
- app frame reassembly and resynchronization after corrupted bytes
- H.264 NAL detection
- security code bounds before network access
- stream status, defensive metadata copy and bitrate counter
- i18n catalogs: same keys, same placeholders, no empty values, all `data-i18n`
  keys from `index.html`, all keys used in `app.js` and all API error codes
- hub device info, formatted firmware and MCU versions, omitted unavailable fields

### Build test

```bash
docker compose build
```

The multi-stage build must work without a locally installed Go.

### API contract tests

The HTTP tests use fake dependencies and check the contract instead of real network
details:

- health schema and content type
- allowed methods and `405`/`Allow`
- unknown API paths as JSON 404 instead of a static file fallback
- channel and quality bounds
- security code not echoed in responses
- explicit rejection not retried; transport errors retried a limited number of times
- session control calls serialized
- status schema and idempotent stop
- `409 Conflict` for inventory and motion while streams are active
- explicit `enabled:false` in the motion contract
- stable error codes in `error` responses
- unknown errors as `internal_error` with HTTP 500 and a generic message
- OpenAPI contract covers every registered route

### Test layout

- Unit and integration tests sit next to the tested package. Unexported parsers and
  builders stay testable.
- API contract tests live in `internal/gateway`.
- Hardware tests use `//go:build hardware`.
- `tests/` is reserved for cross-package Go tests. The conventions are in
  `tests/README.md`.
- Browser E2E tests live in `web/e2e/` (Playwright) and run through
  `scripts/test-e2e.sh`.
- New test files follow the naming convention `<topic>_test.go` or
  `<topic>_integration_test.go`.

### Test structure (AAA)

Every test follows **Arrange/Act/Assert**. Comments mark the phases when they are
not obvious. One test checks **one behavior**; several assertions are allowed when
they describe the same result. Variants run as table subtests. Test names contain no
"And". Integration and hardware tests split into separate sessions even when that
costs runtime. Fuzz targets are the only exception. See ADR-0019.

### Deterministic protocol integration

`internal/ppcs/protocol_integration_test.go` starts an artificial UDP base station.
Its packets are built separately from the specification, not with the production
builders. Covered:

- exact LAN_SEARCH and byte-exact PUNCH echo
- P2P_RDY
- AES challenge and session key
- explicit rejection of a wrong code
- StartVideo and StartAudio with the requested channel and quality mode
- dynamic camera list over IO-Control 40962
- hub device info over IO-Control 5/6
- motion read-modify-write-verify over IO-Control 40964/40963
- battery request and battery reply
- H.264 Annex-B from the fake hub to the output writer
- DRW ACK and session behavior in the real UDP loop

### Fuzz seeds

Go fuzz targets ensure that header, PUNCH, media and app frame parsers do not panic
on arbitrary byte sequences. Seed cases run with every normal `go test`; longer fuzz
runs can be started separately.

### Browser E2E (Playwright)

`web/e2e/` holds Playwright tests. They run in the official image
`mcr.microsoft.com/playwright:v1.65.1-jammy` through `scripts/test-e2e.sh`; a local
Node installation is not needed. Every suite runs in Chromium and Firefox: project
names without a suffix are Chromium, `-firefox` is Firefox. The script starts the
container with `--network host`, checks `/api/health` first and passes the base URL
and optional hardware parameters as environment variables.

**I18n suite (deterministic, no device):**

- browser `de-DE` starts in German, `fr-FR` falls back to English
- `?lang=de|en` overrides detection
- language switch without reload (`performance.getEntriesByType('navigation').length === 1`)
- persistence across reload via `localStorage['akkucam-lang']`
- plurals follow the active language
- API error codes are translated
- `<html lang>` and `aria-label` follow the language
- hub actions change with the language ("Kameras neu laden", "Code entfernen")
- the help popup opens with the localized help text and closes via "Verstanden" /
  "Got it"

**Security code suite (deterministic, mocked API):**

- two base stations with different codes unlock separately
- the same hub reuses the stored code without another dialog
- a rejected code (`unauthorized`) is discarded and requested again
- "Remove code" removes only the selected base station's code
- "Remove all security codes" needs two clicks; the first only arms
- the visibility toggle shows the entered code without changing it and starts masked
  after reopening

**Multiple base station suite (deterministic, mocked API):**

- two loaded hubs appear together in the live wall; no exclusive switching
- filter chips hide a hub without stopping its streams
- "Start all base stations" starts channels on every loaded hub
- inventory and discovery actions are locked while a stream runs
- "Stop all base stations" stops every stream
- hub metadata is visible without expanding; DID sits in the card header, MAC in the
  grid and in the hub manager, long values wrap instead of being cut off

**Live wall suite (deterministic, mocked API):**

- the setup page uses a full-width setup bar and an empty state; after loading the
  wall replaces the setup page, and no footer exists
- tiles start at the top edge and keep an empty grid cell for an odd camera count
- auto, focus and comfort switch without replacing the video element
- the focus strip shows canvas thumbnails for running streams
- the live data overlay shows KB/s, FPS and battery and lets the click through to
  the sidebar
- the hub short label is absent from the auto and focus video and present in comfort
- the "Live data on tiles" switch hides the overlay, stores only `off` and is reset
  by the reset action
- tiles are `role="group"` with `aria-label`; Enter opens the details drawer
- a backdrop click closes settings and drawer
- at 200 % text scale the tile bars and the page stay without overflow
- the details drawer opens on a tile click
- the selected view is stored in LocalStorage and reset through the settings
- the compact iPhone bar opens presentation, filters and settings
- a language switch updates the wall without interrupting the running stream

**Branding and license suite (deterministic, no device):**

- the brand mark renders the camera icon as inline SVG
- the favicon switches with the theme (`icon-light.svg` / `icon-dark.svg`)
- dashboard and wall have no footer and no copyright line
- the help popup opens the scrollable license dialog with the MIT project license and
  hls.js, Font Awesome, Inter Tight, Swagger UI, FFmpeg, Alpine and Go
- the license dialog is translated into German and links the API documentation
- brand name, setup and wall headings use Inter Tight; the brand name renders bold
- license files, font files and theme icons are served by the static file server
- primary button and status text reach 4.5:1 in both themes (measured in the browser)
- the theme switch returns to the system setting and deletes the LocalStorage entry
- all checked text uses only the shipped weights 400/700

**SBOM check:**

- `tests/sbom_test.go` validates the committed CycloneDX and SPDX files
- `internal/gateway/static_test.go` checks that every static asset referenced by the
  dashboard exists
- CI regenerates the SBOM with `scripts/sbom.sh` and uploads it as an artifact

**README comparison:**

- `tests/readme_test.go` checks the same section structure, the same relative link
  targets and the same code block types in `README.md` and `README_de.md`
- the translation itself stays a review task; the test finds missing or extra
  sections, links and commands

**Live suite (optional, real hardware):**

- discovery, security code and camera list against the real base station
- language switch with a loaded camera list without reload
- optional stream start and stop (`AKKUCAM_TEST_LIVE_STREAM=1`)
- no artifacts (trace, video, screenshot) in the live project so the security code
  never reaches disk

```bash
make test-e2e
# optional against real hardware:
AKKUCAM_TEST_SECURITY_CODE='<code>' AKKUCAM_TEST_LIVE_STREAM=1 ./scripts/test-e2e.sh --project=live
```

The security code is passed only as an environment variable. The live suite
deliberately writes no traces, videos or screenshots.

### Static analysis

`./scripts/lint.sh` (or `make lint`) checks Go code, browser sources, scripts,
Dockerfile and workflows in pinned container images; no local installation is
needed. Details and reasoning are in ADR-0033.

- `golangci-lint` with `default: standard`, `gosec`, `errorlint`, `noctx`,
  `bodyclose` and `misspell`; configuration in `.golangci.yml`
- `govulncheck` in the Go image of the build stage, so on the same standard library
  as the binary
- `semgrep` with `p/golang`, `p/javascript`, `p/secrets` and the project rules in
  `.semgrep/gateway.yml`: no outgoing HTTP or TCP clients, no security code in
  `localStorage`/`sessionStorage`
- `shellcheck`, `hadolint`, `actionlint`

All tools block. Exceptions sit only in `.golangci.yml` and are justified there:
G115 in the protocol package, G304/G306 in tests. Every `//nolint` carries a reason
and a reference. Semgrep loads the registry rule sets at runtime and sends no
metrics.

Limits: Semgrep checks patterns, not data flows across function boundaries; the
storage rule does not recognize a generic key. The proof that security codes never
leave browser RAM stays with the Playwright security code tests. CVEs in the runtime
image are not covered; that image scan is noted as phase 2 in `specs/roadmap.md`.

### Hardware integration test

1. Start the container with host networking.
2. Call `/api/discover`.
3. Check model `SMH-T1R` and the complete `CMAG` DID.
4. Run a session probe.
5. Check authentication with a valid and an invalid code.
6. Start channel 1 in Full HD.
7. Check HLS playlist, resolution, bitrate and battery.
8. Stop the stream and check the UI status.
9. Start channel 2 and repeat the same checks.
10. Stop all streams, switch motion off and on per channel and check the read state.
11. Check the container status as `healthy` via `docker inspect`.

Automated run on Linux/NixOS:

```bash
export AKKUCAM_TEST_IP='<hub-ip>'
export AKKUCAM_TEST_DID='<hub-did>'
export AKKUCAM_TEST_SECURITY_CODE='<code>'
AKKUCAM_TEST_CHANNEL=0 ./scripts/test-hardware.sh
AKKUCAM_TEST_CHANNEL=1 ./scripts/test-hardware.sh
```

The code is passed only as the environment of the short-lived Go test container. It
must not land in shell history, CI or version control.

The regular gateway container must be stopped before the hardware test. The base
station can authenticate a competing second session without delivering video data,
which would produce a misleading test failure.

## Local test commands

On NixOS, Docker is enough; a local Go installation is not required:

```bash
make test           # vet, unit/API/integration, race detector, Docker build
make lint           # static analysis in pinned images (ADR-0033)
make test-e2e       # Playwright in Chromium and Firefox: i18n, branding, credentials, workspaces and wall suites; live suite only with AKKUCAM_TEST_SECURITY_CODE
make test-hardware  # optional, real hardware
./scripts/sbom.sh   # regenerate CycloneDX and SPDX SBOM from the image
```

## Requirement traceability

| Specification requirement | Automated proof |
|---|---|
| local discovery | discovery unit tests; hardware test |
| direct LAN session setup | `TestLocalSessionAuthenticationIntegration` |
| code not in API response | `TestLiveStartContractAndSecretIsNotEchoed`, `TestAuthenticationContractDoesNotEchoSecret` |
| H.264 unchanged to output | `TestLiveProtocolVideoFromSpec`, `TestCountingWriterPassesCompletePayload` |
| channel 1/2 and quality mode | fake hub validates bytes 0 and 4; hardware test per channel |
| dynamic channel count | `TestParseCameraListFromSpecification`, `TestQueryChannelsIntegrationFromSpec`, API contract test |
| parallel channels | `TestDifferentChannelsRunConcurrently`; real dual-stream test |
| bitrate | `TestCountingWriterCountsSuccessfullyWrittenBytesOnly`, `TestMeterBitrateReportsOneSecondDelta` |
| battery of the active channel | `TestLiveProtocolBatteryFromSpec`; hardware test |
| no competing control sessions | `TestSessionControlRequestsAreSerialized` |
| inventory/motion only without a stream | `TestInventoryRejectsWhileStreamsAreActive`, `TestMotionRejectsWhileStreamsAreActive` |
| motion read-modify-write-verify | `TestSetMotionReadModifyWriteVerifyIntegration` |
| no invented metadata | `TestUnavailableCameraMetadataIsOmitted`; ADR-0014 |
| metadata fully visible | Playwright workspaces suite (DID/MAC without ellipsis) |
| stable error codes | `TestErrorResponsesCarryStableCodes`, `TestAuthenticationRejectionCarriesUnauthorizedCode`, `TestProbeTimeoutCarriesUpstreamTimeoutCode`, `internal/apierr` |
| OpenAPI contract | `TestOpenAPIDocumentsEveryRoute`, `TestOpenAPIHasNoUnknownRoutes`, `TestSpecIsValidJSON` |
| API contract | `internal/gateway/api_test.go` |
| bilingual UI (DE/EN) | `TestLocalesHaveSameKeys`, `TestLocalePlaceholdersMatch`, `TestLocaleValuesAreWholePhrases`, `TestIndexHTMLKeysExistInLocales`, `TestAppJSKeysExistInLocales`; Playwright i18n suite |
| language switch without stream interruption | Playwright wall suite (video marker stays); Playwright live suite (navigation stays at 1) |
| error codes translated | `TestAPICodesHaveTranslations`; Playwright test "API error codes are translated" |
| help popup in the topbar | Playwright i18n suite ("help popup") |
| security code per base station | Playwright credentials suite (separate codes, reuse, 401, remove code) |
| several base stations at once | Playwright workspaces and wall suites (shared wall, filters, start/stop all hubs, inventory lock) |
| live wall: equal tiles and empty grid cell | Playwright wall suite (tile sizes, `.tile-empty`) |
| setup: full width, setup bar and empty state | Playwright wall suite (widths, empty state, hub grid) |
| live wall: views and video retention | Playwright wall suite (auto/focus/comfort, video marker) |
| live wall: focus thumbnails | Playwright wall suite (canvas thumbnails for running streams) |
| live wall: theme-aware tile frames | Playwright branding and wall suites plus screenshot review |
| live data overlay | Playwright wall suite (content, click opens drawer, switch, reset) |
| iPhone wall bar | Playwright wall suite (compact bar, menu sheet, settings) |
| wall view persisted and resettable | Playwright wall suite (`akkucam-wall-view`) |
| app icon and favicon | `TestBrandMarkRendersCameraIcon`, `TestFaviconUsesLocalThemeIcons`; Playwright branding suite |
| license versions without drift | `TestLicenseVersionsMatchArtifacts` (hls.js bundle, SBOM, Dockerfile) |
| typography (Inter Tight: brand name, setup and wall headings) | Playwright branding suite; `TestStylesheetUsesOnlyShippedFontWeights` |
| design tokens and icon set | `TestDashboardUsesOneIconSet`; Playwright branding suite (contrast, weights) |
| accessibility (hit areas, safe area, tile semantics) | `TestStylesheetRespectsSafeAreaInsets`; Playwright wall suite (role, keyboard) |
| text scaling (200 %) | `TestStylesheetScalesWithRootFontSize`; Playwright wall suite |
| theme follows the system setting | Playwright branding suite (theme cycle) |
| no inline JavaScript in markup | `TestDashboardHasNoInlineEventHandlers` |
| code visibility | Playwright credentials suite |
| machine values instead of server text (`durationMs`, `uptimeSeconds`) | API contract tests, OpenAPI comparison, Playwright wall suite |
| license notices (help popup + scrollable dialog) | `TestDashboardReferencesExistingStaticAssets`; Playwright branding suite |
| SBOM in CycloneDX and SPDX | `TestCycloneDXSBOMIsValid`, `TestSPDXSBOMIsValid`, `TestSBOMsListRuntimeComponents` |
| no data races | `go test -race -count=1 ./...` |

## Bugs found by tests so far

1. **API routing:** A disallowed `GET /api/live/start` fell into the static file
   server and returned 404 instead of 405. Known API paths now have explicit method
   limits and `Allow` headers.
2. **App frame resynchronization:** A corrupted prefix with a plausible large length
   could block the assembler permanently. The parser now also validates the specified
   stream type 1 to 4 and keeps searching afterwards.
3. **Hardware test diagnosis:** Competing sessions first appeared only as a video
   timeout. The test now distinguishes battery-before-video, early session end and a
   real video timeout; the script refuses to run while the gateway container is
   active.
4. **Auth retry:** The retry logic checked the error text for `rejected`. It now uses
   `errors.Is` with sentinel errors from `internal/ppcs`.
5. **Stale container:** After a code change, a container from the old image still ran
   on port 8080 (`gateway-akkucam-gateway:latest`, previous day 22:50). The new image
   was correct but not deployed; every API endpoint answered with 404. `docker compose
   build` and `docker compose up -d` fix that. Before debugging 404s, always check
   `docker compose ps` and the image date.
6. **Notice without the `notice-text` class:** `renderNotice()` looked for
   `.notice-text`, and the markup did not have the class. Every status and error
   message stayed on the initial text. Playwright exposed the bug; the class is now in
   the markup.
7. **Missing `setText`:** `app.js` called `setText()` without defining the function.
   Every channel start, status change and credential display threw a `ReferenceError`.
   The function is now defined as a wrapper around `i18n.setText`.
8. **One code for all base stations:** `state.securityCode` held exactly one code.
   With two hubs and different codes, the second hub was unlocked with the first hub's
   code and rejected with `unauthorized`. Codes now sit in a map per DID (ADR-0025);
   the credentials suite covers the case.
9. **Exclusive workspace:** The dashboard held exactly one workspace. Loading the
   cameras of a second base station made the first view disappear, although the stream
   manager allows several DIDs in parallel. Every loaded hub now has its own section,
   and global actions control all of them together (ADR-0026); the workspaces suite
   covers the case. The sections were later replaced by the shared live wall
   (ADR-0027).
10. **Raw key in the hub card:** `hub.reloadChannels` was missing from both catalogs,
    so the button showed the key instead of "Kameras neu laden".
    `TestAppJSKeysExistInLocales` now also checks keys in dynamic `data-i18n`
    expressions; the i18n suite covers "Kameras neu laden" and "Code entfernen".
11. **Truncated metadata:** DID and MAC were cut with an ellipsis in the hub details
    table. Values now wrap; the workspaces suite checks the full display.
12. **Focus video on a hidden grid tile:** Switching to focus left the old grid DOM
    hidden in the document. The video mapping then found the hidden tile instead of
    the visible focus tile. The inactive view is now emptied; the wall suite checks
    video count and markers.
13. **Popover click opened the drawer:** The native popover closed on light dismiss,
    the following click landed on the tile and opened the sidebar. A custom overlay
    with controlled click handling then closed only the popover (historical; the
    popover was replaced by ADR-0031).

## Current coverage

Measured 2026-11-25:

| Package | Statement coverage |
|---|---:|
| `cmd/gateway` | 0.0 % (no tests) |
| `internal/apierr` | 52.6 % |
| `internal/discovery` | 14.6 % |
| `internal/gateway` | 80.7 % |
| `internal/openapi` | 16.7 % |
| `internal/ppcs` | 75.4 % |
| `internal/stream` | 79.3 % |
| `internal/version` | 46.7 % |
| total | 69.0 % |

The low discovery coverage comes mostly from OS and socket paths. Packet parsers and
merging are tested deterministically; the real broadcast path is part of the hardware
test. Coverage is a diagnostic value, not an end in itself.

### Protocol regression

When framing or reassembly changes, use stored, locally protected PCAPs. PCAPs and
extracted camera images do not belong in version control by default because of their
sensitive content.

## Verified environment

- host: Linux/NixOS, x86-64
- host: Raspberry Pi 4 Model B Rev 1.4, 64-bit
- container: Alpine Linux, Go 1.27.4 build stage
- Docker Engine with `network_mode: host`
- base station: `SMH-T1R`, firmware `1.0.0.12`
- two coupled battery cameras on channels 1 and 2
- browser playback through locally bundled hls.js
- verified stream: H.264 Main, 1920×1080, 25 fps
- observed variable H.264 payload bitrate: roughly 30 to 180 KB/s, scene dependent

## Acceptance criteria

- no gateway traffic to cloud servers
- first stream start typically within 25 seconds
- channels 1 and 2 each deliver a current picture
- channels 1 and 2 can stay active at the same time and have different stream IDs
- bitrate updates about once per second
- battery reply belongs to the selected channel
- stop ends FFmpeg and resets the UI
- a new start replaces the existing stream without orphaned processes
