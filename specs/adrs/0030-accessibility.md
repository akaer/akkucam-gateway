# ADR-0030: Accessibility of the interface

- Status: Accepted
- Date: 2026-11-24

## Context

The interface already respected `prefers-reduced-motion`, had focus rings and
native `<dialog>` elements with focus handling. Contrast targets, hit areas, tile
semantics and behavior on devices with a notch were unregulated.

Two concrete defects stood out:

- Tiles were `<article tabindex="0">` with their own click and keyboard handler but
  contained nested buttons. A `role="button"` would be invalid, and a generic
  element without a role is unnamed for screen readers.
- `viewport-fit=cover` was set, but `env(safe-area-inset-*)` appeared nowhere. The
  bottom sheet and top bar could sit under the notch or the home indicator on the
  iPhone.

## Decision

### Contrast

- Text at least 4.5:1 against its surface, including placeholder and status text.
- Non-text elements (focus ring, borders, icons) at least 3:1.
- At minimum the primary button, error text and status badge are checked in both
  themes; the values are in ADR-0029.

### Hit areas

- Icon-only buttons: 44 × 44 px (`--icon-button`), including dialog and popover
  close buttons.
- Inline text actions: at least 24 px high plus spacing (WCAG 2.5.8), because 44 px
  would break the layout in tables and rows.
- Switches remain part of their labeled row; the row is the hit area.

### Safe area

- `viewport-fit=cover` stays; the top bar and mobile sheets and drawers get
  `env(safe-area-inset-top)` and `env(safe-area-inset-bottom)`.

### Tile semantics

- The tile is `role="group"` with an `aria-label` built from camera name and base
  station. Enter and Space open the details drawer; mouse and touch keep the whole
  area as the target.
- Nested buttons remain real buttons and therefore individually focusable. A
  `button` role for the whole tile is ruled out.

### Focus and keyboard

- `:focus-visible` with a 3 px ring and offset in both themes.
- The live data popover sets focus to its close button, returns focus to the trigger
  on × and Escape and closes on an outside click.
- Dialogs remain native `<dialog>` elements.

### Motion

- Animations run only when motion is allowed; `prefers-reduced-motion` disables
  dialog, popover and switch transitions and the connection pulse.

## Reasoning

- A dashboard used on a phone in the living room needs reliable targets and
  readability in sunlight.
- The group role describes the tile correctly without swallowing the controls
  inside it.
- Safe-area defects do not show in the emulator but appear immediately on the
  device.

## Alternatives considered

- **Tile as `role="button"`:** invalid with nested buttons; screen readers would no
  longer announce the inner actions. Rejected.
- **Separate "Details" button in every tile:** extra visible control for an area
  that is already clickable as a whole. Rejected.
- **44 px for all text links:** would stretch rows in the drawer and manager; WCAG
  requires 24 px. Rejected.
- **Axe or Lighthouse as a dependency:** a new tool and an ADR obligation for a few
  checks; `getComputedStyle` reproduces the measurements deterministically.
  Rejected.

## Consequences

- `web/style.css` contains `env(safe-area-inset-*)`;
  `TestStylesheetRespectsSafeAreaInsets` checks it because Playwright cannot
  simulate insets.
- The wall suite checks `role="group"`, `aria-label` and opening with Enter.
- The branding suite measures contrast and font weight in the browser.
- New controls need a hit area of 44 px (icon) or 24 px (inline text) and a contrast
  proof.
- Screenshots with a notch remain a manual check on real hardware.

## Addendum 2026-11-24

- **Text scaling:** `<meta name="text-scale" content="scale">` and relative
  typography; the Playwright test checks 200 % without overflow or clipped tile
  bars.
- **Light dismissal:** dialogs that may close on an outside click carry
  `closedby="any"`. Without native support (Safari), a pointer fallback checks the
  pointer position, not `event.target`. The code and help dialogs keep
  `closerequest` so input is not lost.
- `dialog:modal` and `::backdrop` limit scroll gestures to the top layer.
- **Headings:** setup and wall are the page's only `h1`; dialog titles stay
  consistently `h2`.
- **Code field:** the visibility toggle is a `<button type="button">` with
  `aria-pressed`, a static label "Show security code" and hidden warning text; it
  starts masked every time it opens. Close buttons use `data-close` instead of
  inline JavaScript.

## Addendum 2026-11-24 (overlay instead of popover)

The live data popover is gone (ADR-0031). Its focus rules no longer apply. The
overlay on the tile is `aria-hidden="true"` and `pointer-events: none`: the
per-second updates should not interrupt screen readers, and the accessible source
remains the details drawer. The switch in the wall settings sits in a labeled row;
the row is the hit area, not the 38 × 22 px switch alone.

## References

- ADR-0013: Responsive dashboard and theme
- ADR-0021: Bilingual UI
- ADR-0027: Live wall
- ADR-0029: Design tokens and icons
- `web/style.css`, `web/app.js`, `web/index.html`, `web/e2e/`
- `specs/testing.md`
