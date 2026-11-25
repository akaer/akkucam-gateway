# ADR-0013: Responsive dashboard with diagnostics and theme toggle

- Status: Accepted
- Date: 2026-11-21
- Changed: 2026-11-24 (addendum, ADR-0030)

## Context

The prototype interface mixed diagnostic actions with the normal operating flow,
used large displays poorly and was only partly mobile-friendly.

## Decision

The UI follows the flow discovery, select hub and enter code, dynamic stream
tiles, single and global stop. Diagnostics and the manual channel fallback sit in
a collapsible area. The layout uses a responsive tile matrix, native fullscreen
per tile and a light/dark theme.

On first visit the theme follows the system preference. Only the theme and
language choices are stored in localStorage, never the security code. ADR-0021
covers the language choice.

## Reasoning

- Clear primary actions without redundant probe or auth buttons
- Efficient use of large displays through parallel tiles
- Single-column, touch-friendly layout for the iPhone 14
- Technical details stay reachable for technical users
- Fullscreen uses browser and iOS features instead of a custom player

## Consequences

- hls.js stays bundled locally.
- UI state is volatile; discovery and the camera list reload after a page reload.
- The manual channel fallback appears only in the diagnostics area.

## Addendum 2026-11-23

ADR-0027 replaces the tile matrix and the permanently visible setup page with a
live wall. The wall takes over after the first load; auto matrix, focus and
comfort replace the earlier tile view. The wall bar, mobile compact bar, details
drawer and live-data popover are described there. Setup and wall share the same
full-width shell and the compact notice bar; a page footer no longer exists.
Theme and language persistence stay unchanged; only the wall view is added to
localStorage.

## Addendum 2026-11-24

- The "system" state is the **attribute-free** case: no `data-theme`, and the
  dark tokens apply through `@media (prefers-color-scheme: dark)`. Without
  JavaScript the page therefore follows the system preference.
- The two-state switch maps "system" and "opposite of what is currently
  rendered". When the target matches the system preference, the localStorage
  entry is **deleted** instead of stored; only that keeps the way back to the
  system setting.
- `<meta name="color-scheme">` stays in sync with the state (`light dark` in the
  system case), but remains a hint only; the CSS property `color-scheme` is
  authoritative.
- A `storage` event applies changes from other tabs.

## Addendum 2026-11-24 (fourth storage key)

Besides `akkucam-theme`, `akkucam-lang` and `akkucam-wall-view`, the interface
stores the visibility of the live overlay as `akkucam-wall-metrics` (`off` or
absent). It stays a pure UI preference; credentials and device data are still
not stored (ADR-0031).
