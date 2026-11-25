# AGENTS.md — Dashboard (`web/`)

Rules for the browser dashboard: static HTML/CSS/JS served by the gateway.
No framework, no bundler, no npm build step. `hls.js`, Font Awesome, Inter
Tight and Swagger UI are vendored under `web/` and delivered by the same
static file server. Globally relevant rules live in `../AGENTS.md`.

## i18n

- Every visible string goes through `web/i18n.js`; use `data-i18n`,
  `data-i18n-title`, `data-i18n-aria-label` or `setText()`/`t()`. No
  hardcoded user-facing text in `index.html` or `app.js`.
- English is the source and fallback language. German is the second catalog.
  Detection order: `?lang=de|en`, then `localStorage['akkucam-lang']`, then
  `navigator.languages` (`de*`), else English. `?lang` is not persisted.
- Language switching happens without reload and without interrupting streams.
- Every key exists in both `web/locales/en.json` and `web/locales/de.json`,
  with identical placeholders. New API error codes need `error.code.<code>`
  in both catalogs.
- Swagger UI stays English; it is developer tooling.
- Device data (DID, model, IP, hub channel names) is never translated.
- A catalog value is a whole sentence or label. Never concatenate fragments in
  `app.js` (`" · scrolls"`, `{suffix}`); pass data through placeholders and
  format lists, numbers and durations with `Intl` (ADR-0028).
- German is Du-form. Progress messages use the state passive ("Kameras werden
  geladen …"), not an imperative aimed at the user.
- No implementation disclaimers in UI copy (`localStorage`, "matrix",
  "fallback"). Explain the action or the benefit.
- Destructive actions need two clicks with a changing label; errors that can
  be retried offer "Retry"/"Erneut versuchen" (ADR-0028).
- Use the glossary from ADR-0028 (base station, camera, channel, live view,
  stream, details, live data, memory). "Browserspeicher" is banned because it
  also means `localStorage`.

## Design system

- `web/style.css` is token-driven (ADR-0029): six type steps (`--text-xs`
  … `--text-lg` plus a `clamp()` for `h2`), weights 400/700 only, three radii,
  four shadows, two blur levels, `--gutter` and `--canvas-gap`. Add a step
  before adding a number.
- Inter Tight is shipped in 400 and 700 only. Other `font-weight` values are
  rejected by `TestStylesheetUsesOnlyShippedFontWeights`.
- Icons come from the inline sprite at the top of `index.html`
  (`#icon-camera`, `#icon-close`, …) and are sized with `.icon`,
  `.icon-sm`, `.icon-lg`. Text glyphs (`◐`, `⌄`, `▶`, `⛶`, …) are forbidden.
- Text contrast is at least 4.5:1 against its surface; `--accent` is the
  surface colour, `--accent-ink` the text colour. Do not merge them.
- Icon-only controls are 44 × 44 px, inline text actions at least 24 px high.
  Mobile sheets respect `env(safe-area-inset-*)` (ADR-0030).
- Type sizes are `rem` (never `px`); the page opts into `text-scale`.
  Test layouts at a 32 px root before calling them done.
- Dialogs that may close on an outside click use `closedby="any"`; never
  re-implement the backdrop check with `event.target`. No inline `onclick`
  handlers in `index.html`.
- The security code field starts masked; the toggle is a `<button>` with
  `aria-pressed` and a visually hidden warning. It must not change the value.
- License entries in the dialog carry `data-component`; every displayed
  version must be verifiable by `TestLicenseVersionsMatchArtifacts`
  (hls.js bundle, SBOM, Dockerfile). Do not add a version without a source.

## Storage

Only four keys are allowed, and only as UI preferences:

- `akkucam-theme`
- `akkucam-lang`
- `akkucam-wall-view`
- `akkucam-wall-metrics` (`off`; absent means on)

Never store credentials or any device data. Codes live in the in-memory
`state.credentials` map keyed by DID (ADR-0012/ADR-0025).

## Wall invariants

- A running stream keeps its `<video>` element and HLS instance for its whole
  lifetime. View, filter and language changes re-render tiles around the
  existing videos; they never recreate a stream.
- Invisible videos stay in the offscreen `#stream-holder`; the focus strip
  paints miniature frames from those videos onto canvases. No second HLS
  instance.
- Auto and Comfort use equal tile sizes and visible empty cells for odd
  camera counts. Focus shows one large tile plus a strip. All views start at
  the top and scroll when needed.
- Filter chips hide hubs without stopping their streams. The hub manager and
  the global actions start/stop hubs; loading another hub never replaces
  existing streams.
- Setup is replaced by the wall after the first channel list loads. "Add base
  station" reopens setup and is blocked while any stream runs. Discovery,
  inventory and motion actions are disabled while any stream is active.
- Setup hub cards show the metadata openly, without `details/summary`. The
  grid carries only what the header does not already say (firmware, MCU, MAC,
  SD values); DID, IP and model stay in the header. The hub manager lists IP
  and MAC when known.

## Tiles, drawer, metrics

- Clicking a tile opens the details drawer (right on desktop, bottom sheet on
  mobile) with quality, motion, start/stop, fullscreen, available device data
  and live metrics.
- Tiles are `role="group"` with an `aria-label` of camera and hub; Enter and
  Space open the drawer. Never give the tile itself a `button` role, because
  it contains real buttons.
- Auto and Focus show only channel number, name and status inside the video.
  The hub short label lives in Comfort's header and in the filter chips; do
  not put it back into the video overlay.
- Running streams show `KB/s · fps · battery %` as a permanent overlay in
  the tile's bottom-left corner. It is `aria-hidden`, `pointer-events: none`
  and updated via `textContent` once per second. The drawer keeps the full
  metrics including uptime. `akkucam-wall-metrics=off` hides the overlay
  (setting in the wall dialog, default on).
- The overlay keeps a 55 % black scrim behind the text: camera scenes can be
  pure white, where text shadows alone drop below 4.5:1. Do not remove it
  without replacing the contrast guarantee.
- Tile frames use theme-aware borders and shadows: light and soft in the light
  theme, subtle in the dark theme. Black is reserved for the video area.
- The fullscreen button sits in the tile; native controls appear only in
  fullscreen.

## Layout and branding

- Setup and wall share the full-width shell: compact notice bar, toolbar-like
  control strip, 16 px gutters. No centered max-width container.
- The compact wall bar is for iPhone sizes only (portrait `<= 640px`,
  landscape `<= 1000 x 560px`). iPad and desktop keep the full toolbar.
- Brand mark and favicon use the vendored Font Awesome camera icon
  (`icon-light.svg`, `icon-dark.svg`); keep both variants in sync. The brand
  name and `h2` headings use Inter Tight via `--font-heading`.
- There is no page footer and no `licenses.html`. The info icon at the far
  right of the topbar opens the localized help dialog, which links to the
  scrollable license dialog (project MIT license plus every third-party
  component).

## Verification

Playwright lives in `web/e2e/` and runs through `../scripts/test-e2e.sh`.
Every suite runs in Chromium (short project name) and Firefox (`-firefox`
suffix):

| Project | Covers |
|---|---|
| `i18n` | detection, live switching, error codes, help dialog |
| `branding` | icons, favicon, typography, license dialog, German catalog |
| `credentials` | per-hub codes, reuse, discard on `unauthorized`, forget |
| `workspaces` | multiple hubs, global start/stop, inventory lock |
| `wall` | setup/wall flow, views, tile sizes, video retention, focus thumbnails, metrics overlay, settings, iPhone bar |
| `live` | real hardware; skipped without `AKKUCAM_TEST_SECURITY_CODE`, no artifacts |

Add or extend a project when changing user-visible behavior. Never commit
`web/e2e/test-results/`, `web/e2e/playwright-report/` or `.auth/`; they are
ignored and may contain credentials.
