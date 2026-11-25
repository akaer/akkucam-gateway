# ADR-0029: Design tokens and one icon set

- Status: Accepted
- Date: 2026-11-24

## Context

`web/style.css` had 25 font sizes, 19 radii, 9 shadows and 7 font weights for
exactly two shipped faces (400/700). The browser synthesized bold, so the same
headings looked different across platforms. "Inter" also sat in the body stack
although the font is never shipped. Icons came from three sources: Font Awesome
SVG (brand), custom inline SVGs (help, gear) and text glyphs (`◐ ⌄ ⋯ ▶ ■ ⛶ ▣ ×`)
that look different depending on the system font.

The brand stays Nord. Contrast was not part of the definition: the primary blue
was 3.50:1, hover 2.34:1, the code badge 1.99:1.

## Decision

### Color

`--accent` is darkened and split into four roles:

| Token | Light | Dark | Role |
|---|---|---|---|
| `--accent` | `#476487` | `#88C0D0` | Surface (button, brand square, active chip) |
| `--accent-hover` | `#3B5473` | `#9ACBDA` | Hover of the surface |
| `--accent-ink` | `#476487` | `#9ACBDA` | Text, links, status on surface |
| `--accent-text` | `#ECEFF4` | `#2E3440` | Text on `--accent` |
| `--danger-ink` | `#A04A54` | `#DCA0A6` | Error text |

`--accent-strong` is gone; it served as both hover surface and text color.
`--danger` remains for dots and borders.

Measured contrast (WCAG AA requires 4.5:1 for text): button 5.30:1 (was 3.50),
hover 6.74 (2.34), badge 4.52 (1.99), error text 4.79 (3.36), dark badge 4.90
(4.14).

### Typography

- Six steps: `--text-xs` 11 px, `--text-sm` 12, `--text-md` 13, `--text-base` 15,
  `--text-lg` 18, plus a `clamp()` for `h2`.
- Two weights: 400 and 700. The browser synthesizes other values, and they are not
  allowed.
- Inter Tight stays reserved for the brand name and `h2` (ADR-0023). `h3`, metric
  headings and body text use the system font.
- "Inter" leaves the body stack because the file is never shipped.

### Shape and depth

- Three radii: `--radius-sm` 10 px, `--radius-md` 14 px, `--radius-lg` 18 px;
  `--radius-pill` for pills.
- Four shadows: `--shadow-card`, `--shadow-popover`, `--shadow-modal`,
  `--shadow-thumb`.
- Two blur levels: top bar (18 px) and dialog backdrop (4 px). Tile chips on
  running videos are no longer blurred; that costs GPU time without adding
  information.
- One gutter token (`--gutter` 16 px) for the shell, `--canvas-gap` 12 px for the
  wall (8 px on mobile).

### Icons

- One inline sprite in `index.html` (`<symbol>`), 24 px grid, `currentColor`,
  sizes 14/18/28 px through `.icon-sm`, `.icon` and `.icon-lg`.
- No text glyphs as icons, no icon font, no CDN.
- The camera symbol remains the Font Awesome work (CC BY 4.0) and is defined in one
  place; favicon files stay separate.

## Reasoning

- Scales instead of single values make deviations visible and reviewable.
- Two real faces prevent bold from looking different across browsers.
- One icon set guarantees the same stroke width, alignment and theme color.
- Splitting `--accent` into surface and text is the prerequisite for reliable
  contrast in both themes.

## Alternatives considered

- **Dark button text on the old blue:** would keep the brand but visibly change the
  character of the primary button. Rejected.
- **Ship more Inter Tight faces (600/800):** more payload and license upkeep for
  intermediate steps the scale does not need. Rejected.
- **SVG sprite as a separate file:** external `<use>` works but adds a request and a
  second failure point. Rejected.
- **Icon font:** contradicts the local, license-bound single file and scales worse.
  Rejected.

## Consequences

- `web/style.css` uses tokens only; `TestStylesheetUsesOnlyShippedFontWeights`
  enforces 400/700.
- `TestDashboardUsesOneIconSet` forbids the old text glyphs in HTML and JS.
- The Playwright branding suite measures contrast and font weight in the browser
  for both themes.
- Tile labels still truncate with an ellipsis; the full name sits in the details
  drawer. ADR-0014 still governs metadata fields (complete, wrapped).
- Changes to color or scale need an update to this ADR and the contrast tests.

## Addendum 2026-11-24

- The six font steps are defined in `rem` (0.6875/0.75/0.8125/0.9375/1.125 rem); at
  a 16 px root the rendering stays identical, and it grows with a larger text scale
  (ADR-0030).
- Tile header and footer use `min-height` instead of fixed heights. The wall
  measures the actual chrome height (`measureComfortChrome`) instead of calculating
  with the 96 px constant.

## References

- ADR-0013: Responsive dashboard and theme
- ADR-0021: Bilingual UI
- ADR-0023: Inter Tight for brand name and headings
- ADR-0027: Live wall
- ADR-0030: Accessibility
- `web/style.css`, `web/index.html`, `web/locales/`
