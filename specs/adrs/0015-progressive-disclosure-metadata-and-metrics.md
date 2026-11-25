# ADR-0015: Metadata and live metrics collapsed by default

- Status: Partially superseded (ADR-0031)
- Date: 2026-11-21
- Changed: 2026-11-23 (addendum, ADR-0027)

## Context

The dashboard should be quick to grasp on mobile devices and large displays.
Permanently showing device and protocol details (firmware, MCU, MAC, live
metrics) creates visual noise and distracts from the main purpose: start and
stop streams and watch live video.

At the same time, the technical information should stay available without being
hidden.

## Decision

- Hub and camera metadata use `details/summary`.
- The default state is **collapsed**.
- A clear indicator (chevron) signals that the element is clickable.
- Live metrics (for example bitrate, battery, uptime) appear only inside this
  area.
- Live metrics are rendered and updated only while the area is open.

## Reasoning

- Better readability in the default case
- Progressive disclosure: details only on demand
- Fewer layout shifts and less "data flicker" from status polling
- Matches the existing diagnostics pattern

## Consequences

- Users see live metrics only after opening the area.
- The gateway stays ADR-0014-compliant: only concrete values appear; missing
  values produce no placeholder rows.

## Addendum 2026-11-23

The live wall (ADR-0027) keeps progressive disclosure. Live metrics stay closed
by default and are reachable through the details drawer and an explicit
"Live data" popover. The popover contains only available values; it stays
invisible until opened and does not change the tile height.

## References

- ADR-0014: Render only metadata that is actually available
- ADR-0013: Responsive dashboard and theme

## Clarification 2026-11-24

The statement "live metrics are rendered and updated only while the area is
open" concerns the display. The client still polls `/api/live/status` once per
second for every running stream to detect stream end and metrics; only an open
drawer and an open popover update (`updateStreamStatus` in `web/app.js`).

## Addendum 2026-11-24

Progressive disclosure of live metrics no longer applies to the live wall:
running streams show bitrate, FPS and battery permanently as an overlay on the
tile (ADR-0031). The details drawer remains the detailed path and also shows
uptime. The overlay can be turned off in the wall settings; the original state
then applies again.

## Addendum 2026-11-24 (setup page)

On the setup page the metadata of a loaded base station is open immediately; the
`details/summary` element is gone. The tile grid shows only what the tile header
does not already say (firmware, MCU, MAC, SD values); DID, IP and model stay in
the header. Only diagnostics remains collapsible. The drawer remains the
progressive path for camera data.
