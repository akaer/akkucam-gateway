# ADR-0031: Live data overlay and visibility setting

- Status: Accepted
- Date: 2026-11-24

## Context

Live metrics sat behind a popover that only running streams opened with a "Live
data" button (ADR-0027). In practice people rarely found it, and the bitrate in
Mbit/s with one decimal was too coarse: for a payload stream between roughly 0.1
and 1.5 Mbit/s the display jumped between "0 Mbit/s" and "1.5 Mbit/s" and lost all
resolution.

At the same time every tile in the video area carried a pill with model and DID
label (`SMH-T1R · LYHFC`). With several cameras on the same base station it
repeated on every tile and competed with channel number and camera name.

ADR-0027 had explicitly rejected permanently visible metrics because they
contradict ADR-0015 and cost video area. This decision revises that deliberately:
one line of text costs little area and replaces an extra interaction step.

## Decision

### Tile content

- In **auto** and **focus**, the tile video shows only channel number, camera name
  and status. The hub label is gone, including from the focus strip.
- **Comfort** keeps the row with the hub label below the camera name because it sits
  outside the video area.
- The tile `aria-label` still names camera and base station.

### Overlay

- Running streams show a subtle overlay in the bottom left:
  `412 kbit/s · 25 fps · 87 %`.
- Only available values appear; if a value is missing, its segment is dropped. With
  no readings the overlay stays empty.
- Uptime stays in the details drawer only. The drawer still shows bitrate, FPS,
  battery (including charging state) and uptime.
- Bitrate appears everywhere in `kbit/s`, including above 1000 (`1480 kbit/s`).
  There is no unit switch.
- The overlay is `aria-hidden` and `pointer-events: none`: it is decoration over the
  video, and the accessible source remains the drawer. Clicks on the tile still open
  the details.
- Update: `updateStreamStatus` writes the text once per second; no markup is
  recreated.
- Look: 11 px, white text, tabular figures, 22 px high and therefore in line with
  the fullscreen button. Behind the text sits a 55 % black surface with a 6 px
  radius. It is necessary because camera scenes can be arbitrarily bright: on pure
  white the text still reaches 4.5:1, while a pure text shadow would fall below
  2:1. The surface stays limited to the text width and disappears while no values
  are present (`:empty`).

### Visibility

- The wall settings (gear, also in the mobile sheet) get a "Live data on tiles"
  switch. The default is **on**.
- Persistence as the fourth UI preference `localStorage['akkucam-wall-metrics']`.
  Only `off` is stored; a missing entry means on.
- "Reset to default" resets the view **and** the overlay.
- The switch takes effect immediately through a body class; no stream and no video
  element is touched.

## Reasoning

- kbit/s provides the resolution that Mbit/s with one decimal lacks.
- One line at the bottom edge is faster to grasp than a popover and hides no
  relevant picture content on 16:9 tiles; the small surface behind it guarantees
  contrast independent of the picture.
- The hub still appears in the drawer, the filter chip and the comfort view; inside
  the video it is pure repetition with one camera per hub.
- A switch preserves the choice that ADR-0015 wanted to protect without withholding
  metrics from everyone.

## Alternatives considered

- **Mbit/s with two decimals:** fixes the jumps but remains an awkward unit for a
  stream below 2 Mbit/s. Rejected.
- **Keep the popover and change only the unit:** does not fix discoverability.
  Rejected.
- **Overlay always on, no switch:** simpler, but ignores the preference for a calm
  video picture documented in ADR-0015. Rejected.
- **Switch per view:** three states for a property unrelated to layout. Rejected.
- **Overlay top right next to the status:** collides with the status pill and the
  picture content; bottom left is the calmest corner. Rejected.

## Consequences

- ADR-0015 applies to the wall only in part (status: `Partially superseded`); the
  drawer remains the progressive path.
- The popover section in ADR-0027 is replaced: no `#metrics-popover`, no trigger, no
  click guard, no focus handling.
- `web/AGENTS.md` allows four storage keys instead of three.
- New catalog keys: `camera.metric.kbpsValue`, `settings.liveData`,
  `settings.liveDataHint`; removed: `camera.liveDataClose`,
  `camera.metricsPending`, `camera.metric.mbpsValue`.
- Playwright: overlay content, click on the overlay opens the drawer, hub label only
  in comfort, switch including reset.
- `product.md`, `architecture.md` and `testing.md` are updated with it.

## References

- ADR-0014: Only available metadata
- ADR-0015: Progressive disclosure (partially replaced)
- ADR-0021: Bilingual UI
- ADR-0027: Live wall (popover section replaced)
- ADR-0028: Terminology
- ADR-0030: Accessibility
- `web/app.js`, `web/index.html`, `web/style.css`, `web/locales/`
