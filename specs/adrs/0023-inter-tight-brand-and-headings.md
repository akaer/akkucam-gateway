# ADR-0023: Inter Tight for the brand name and step headings

- Status: Accepted
- Date: 2026-11-22

## Context

The interface used the same system font for all text. The brand name "AkkuCam" in
the top bar and the large step headings ("Select a base station" and "X cameras")
should stand out more typographically. At the same time, the hero text block
("Your cameras. At a glance." plus explanatory text) was too prominent and was
removed.

The constraints stay: no cloud, no CDN. The font must ship locally and its
license must be documented visibly.

## Decision

- Inter Tight ships as a locally bundled web font under `web/fonts/`:
  - `web/fonts/inter-tight-latin.woff2` (regular 400)
  - `web/fonts/inter-tight-bold-latin.woff2` (bold 700)
- `web/style.css` loads both cuts through `@font-face`. `font-display: swap`
  keeps text readable while the font loads. There is no Google Fonts or other CDN
  call.
- The brand name in the top bar (`.brand`) uses Inter Tight. The name "AkkuCam"
  renders bold (700) through the existing `<strong>`; the tagline stays at normal
  weight.
- The large step headings (`h2`, especially "Select a base station" and
  "X cameras") also use Inter Tight. The font family is defined centrally through
  `--font-heading`.
- The hero text block (eyebrow, headline and explanatory text) is removed. The
  discover button stays as the only action in the hero area.
- The font is under the SIL Open Font License 1.1. The license text ships as
  `web/INTERTIGHT-LICENSE.txt` and is linked in the license dialog.
- Inter Tight appears in the license dialog, in `specs/sources.md` and in
  `sbom/README.md`. The font has no package-manager entry, so it does not appear
  in the SBOM automatically.

## Reasoning

- **Cloud-free:** local WOFF2 files avoid external dependencies and work offline.
- **License clarity:** the OFL allows local embedding; the license text ships
  along and is reachable in the dashboard.
- **Small payload:** two static cuts are enough for the brand name and step
  headings. A variable font would be more flexible but unnecessarily larger for
  this use.
- **Targeted use:** the brand name and step headings switch fonts. The rest of the
  interface keeps the system font, which preserves readability and consistency.
- **Less distraction:** the removed hero text block makes the entry point more
  direct; the discover button stays as a clear call to action.

## Alternatives considered

- **Load from the Google Fonts CDN:** violates the cloud-free rule. Rejected.
- **Ship a variable font:** larger file and more complexity for two needed cuts.
  Rejected.
- **Keep the hero headline and change only the font:** rejected in favor of the
  cleaner entry point.
- **Leave the brand name in the system font:** the intended visual accent would be
  lost. Rejected.

## Consequences

- New files: `web/fonts/inter-tight-latin.woff2`,
  `web/fonts/inter-tight-bold-latin.woff2`, `web/INTERTIGHT-LICENSE.txt`.
- `web/style.css` contains two `@font-face` blocks, the `--font-heading` variable
  and the typography for `.brand` and `h2`.
- The hero text block and its `hero.*` i18n keys disappear. The discover button
  stays.
- The license dialog, `specs/sources.md` and `sbom/README.md` list Inter Tight
  with OFL-1.1.
- New tests: `web/e2e/branding.spec.ts` checks that the brand name and step
  headings use Inter Tight and that the font files and license text are served.
- Changing the font or adding cuts requires updating the license dialog,
  `specs/sources.md` and the SBOM documentation.

## Addendum 2026-11-24

- It stays at two cuts (400/700). The browser synthesizes other weights and
  `style.css` rejects them (ADR-0029).
- The scope is `h2` and the brand block. `h3`, metric headings and body text use
  the system font; the earlier expansion to all `h1–h3` was an implementation
  mistake.
- "Inter" is removed from the body stack because the file does not ship.

## References

- ADR-0003: local, no cloud fallback
- ADR-0021: bilingual UI
- ADR-0022: bundled web assets, license notices and SBOM
- `specs/sources.md`, `specs/architecture.md`, `specs/testing.md`
- `web/style.css`, `web/INTERTIGHT-LICENSE.txt`
