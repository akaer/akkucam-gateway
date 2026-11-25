# ADR-0027: Live wall with views, details drawer and mobile control bar

- Status: Accepted
- Date: 2026-11-23

## Context

The previous interface showed one workspace per loaded base station, stacked
below each other. Camera tiles had different sizes depending on the count, the
setup page stayed visible above the streams, and large displays left much space
unused. Several hubs and iPhone/iPad needed one shared, predictable layout. Three
clickable prototypes tested the variants; the implementation follows the shared
live wall.

## Decision

### Flow

- While no base station is loaded, the setup page stays visible (discovery, hub
  selection, diagnostics).
- Setup and wall share the same full-width shell: compact notice bar, toolbar
  style, 16 px gutters and no centered 1560 px container. The setup page shows a
  setup bar with discover and "Back to the wall", a full-width empty state and the
  hub grid or diagnostics below.
- After the first channel list loads, the live wall replaces the setup page.
- "Add base station" opens the setup page on purpose. The flow is locked while
  streams run; after a successful load the wall returns.
- Diagnostics and the manual channel fallback stay part of the setup page.

### Shared wall

- All loaded hubs appear in **one** wall. Hub filter chips hide hubs without
  stopping streams.
- A hub manager dialog shows all loaded hubs with the full model and DID, code
  status, camera count, live count, start and stop, "Remove code" per hub and a
  global "Remove all security codes".
- Global actions still start and stop all hubs together (ADR-0026).

### Views

- **Auto matrix:** columns and rows come from the camera count and the available
  area. All tiles are the same size; odd counts produce visible empty grid cells.
  When the area cannot hold readable tiles, the wall scrolls.
- **Focus:** one large active tile and a strip with all other cameras (desktop
  right, mobile bottom). The strip selects the active camera and shows miniature
  live frames for running streams, drawn from the existing video elements onto
  canvas. No extra HLS instance starts.
- **Comfort:** the same matrix as auto, with a fixed header and control row per
  tile (quality, motion, start and stop).
- All views start at the top below the wall bar; the grid stays centered
  horizontally.
- The choice is stored as `localStorage['akkucam-wall-view']`. The settings dialog
  offers "Reset to default" (auto) and removes the entry again. No credentials are
  stored.

### Tile and video

- A tile shows video or a placeholder, channel number, name, short hub label and
  status. Auto and focus use an overlay; comfort uses a header and footer.
- A click on the video or tile opens the details drawer (desktop right, mobile as
  a bottom sheet): quality, motion detection, start and stop, fullscreen, available
  device data and live metrics.
- The fullscreen button sits in the tile; native video controls appear in
  fullscreen, not in the small tile.
- Changing the view, filter or language **does not recreate video elements**.
  Each stream keeps its `<video>` element and HLS instance; invisible videos sit
  in the offscreen `#stream-holder`, visible ones attach to the new tile. Streams
  keep running.
- Tile frames and shadows follow the theme: light and soft in the light theme,
  subtle in the dark theme. Black stays reserved for the video area.

### Live data popover

- Only running streams show the "Live data" trigger.
- The popover is closed by default and shows only available values (bitrate, FPS,
  battery, uptime; ADR-0014). Channel and hub name are omitted because they already
  sit on the tile.
- A second click on the trigger, × or a click outside closes it. A click outside on
  a tile closes only the popover and opens no side panel; top bar and toolbar
  actions stay usable. Escape remains the keyboard fallback.
- Focus returns to the trigger when the popover closes through × or Escape.
- The tile height stays unchanged. Live metrics remain available in the drawer
  (ADR-0015).

### Mobile control bar

- Only for iPhone sizes: portrait `≤ 640 px`, landscape `≤ 1000 × 560 px`. iPad and
  desktop keep the full wall bar.
- Compact bar: filter pill, global start/stop toggle and "⋯".
- The bottom sheet holds appearance, hub filter, global actions, hub management
  and settings.
- Portrait uses one column and scrolls; landscape uses the best matrix.

### Further decisions

- The top bar keeps language, theme and help (help at the far right, ADR-0024).
  Wall settings sit in a gear in the wall bar or in the mobile sheet.
- In wall mode and on the setup page there is **no page footer**. License and API
  information opens through the info/help icon as a scrollable license dialog. The
  separate `licenses.html` and the license links in the settings dialog are gone;
  copyright and the MIT footer link are no longer shown. The dialog lists the MIT
  project license and all third-party components.
- Security codes stay exclusively in browser RAM (ADR-0012/ADR-0025).
- No API, OpenAPI or SBOM change. The existing live endpoints still deliver per DID
  and channel.

## Alternatives considered

- **Make the existing workspace cards only more responsive:** solves neither
  unequal tile sizes nor the space use and leaves the setup page above the streams.
- **Tabs per hub:** would switch exclusively again; ADR-0026 requires parallel
  visibility. Filter chips and one shared wall meet that without losing the hub
  mapping.
- **Permanently visible live metrics in every tile:** contradicts ADR-0015 and
  costs video area. The popover is the explicit entry point.
- **Native popover API:** light dismiss let the following click on the tile open
  the side panel. A custom overlay with controlled click handling is predictable on
  desktop and iPhone.
- **Fixed two columns in iPhone landscape:** shows only one row with five cameras.
  The auto matrix uses the height better.

## Consequences

- `web/index.html`: setup and wall area, wall bar, mobile bar, details drawer,
  settings, hub manager, metric popover, license dialog.
- `web/app.js`: wall rendering, matrix calculation, video and HLS retention,
  offscreen video holder, focus miniatures, filters, views, drawer, popover, mobile
  bar, hub manager.
- `web/style.css`: wall, tile, drawer, popover, license and mobile styles,
  including theme-aware tile frames.
- `web/locales/{en,de}.json`: new keys for wall, setup, views, settings, hub
  manager, live data, licenses and mobile bar.
- `web/e2e/wall.spec.ts`: new deterministic suite; existing suites adjusted to the
  wall. The former `licenses.html` and the page footer are gone.
- ADR-0013, ADR-0015, ADR-0022 and ADR-0026 are amended by this decision.

## Addendum 2026-11-24

The tile overlay has `z-index: 1` in `web/style.css`. The `<video>` is attached
after the overlay in `.tile-media`; without its own stacking context, Chromium
places it above the overlay buttons during hit testing, so "Live data" and
fullscreen were not clickable. The Playwright test "closes live data on trigger and
outside without opening details" covers this.

## Addendum 2026-11-24

- The shell uses `--gutter` (16 px). The wall canvas uses `--canvas-gap` (12 px,
  8 px on mobile) so tiles and gutters do not count twice. The earlier wording
  "16 px gutters" meant the shell, not every inner edge.
- The layout status in the wall bar ("3 × 2 · 4 cameras (scrolls)") is part of the
  bar: it describes the grid and scroll need, hides below 980 px and does not count
  as device metadata.
- Live data shows uptime as a formatted value (`uptimeSeconds`), not as a server
  string (ADR-0021).

## Addendum 2026-11-24 (live overlay)

- The hub label is gone from the video in auto and focus views and from the focus
  strip. Comfort keeps it below the camera name.
- The live data popover is gone. Running streams show `412 kbit/s · 25 fps · 87 %`
  in the bottom left as a subtle overlay; uptime stays in the drawer (ADR-0031).
- Bitrate appears in `kbit/s`, not `Mbit/s`.
- Visibility is switchable in the wall settings (`akkucam-wall-metrics`, default
  on).

## Addendum 2026-11-24 (bar and setup)

- The layout status ("3 × 2 · 4 cameras (scrolls)") is gone from the wall bar. Grid
  and scroll need stay internal but are no longer user text.
- The hub manager also names the MAC address when it is known from the channel list
  or the ARP table.
- The setup card shows hub metadata openly instead of behind a disclosure
  (ADR-0015).

## References

- ADR-0011: Stream per DID and channel
- ADR-0012/ADR-0025: Code in browser RAM only
- ADR-0013: Responsive dashboard and theme
- ADR-0014: Only available metadata
- ADR-0015: Progressive disclosure
- ADR-0021: Bilingual UI
- ADR-0026: Several base stations
- `specs/product.md`, `specs/architecture.md`, `specs/testing.md`
