# ADR-0024: Info icon and help popup in the top bar

- Status: Accepted
- Date: 2026-11-22
- Changed: 2026-11-23 (addendum, license dialog)

## Context

After the hero text block was removed (ADR-0023), the entry point into the
dashboard was very bare. New users saw only the discover button and the notice
"Ready. Discover a base station to begin." The flow (find a base station, open
its coupled cameras, start a live view without a cloud service) was explained
nowhere.

The explanation should be reachable without filling the entry point with text
again. It must be bilingual, follow the theme and work without external
resources.

## Decision

- A small info icon (`.icon-button` with inline SVG) sits at the far right of the
  top bar, right of the theme toggle. The icon is a locally drawn SVG path; no CDN
  and no extra icon library.
- A click opens a modal `<dialog>` popup (`#help-dialog`) with the flow in the
  active language:
  - German: "Basisstation finden, gekoppelte Kameras öffnen und Livebild ohne
    Cloud-Dienst starten."
  - English: "Find a base station, open its cameras and start a live view without
    any cloud service."
- The popup uses the same dialog base as the security code dialog
  (`.modal-dialog`) and closes through "Got it", the × and Escape. It carries no
  form data and stores nothing.
- The text lives in `web/locales/en.json` and `web/locales/de.json` under `help.*`.
  A language switch updates the popup through the existing `i18n:change` path.
- The icon is named through `aria-label`/`title`; the dialog binds to its heading
  through `aria-labelledby`.

## Reasoning

- **Cloud-free and lightweight:** an inline SVG and a native `<dialog>` need no
  library, no build step and no network access.
- **Bilingual without a special path:** the help uses the same catalog as the rest
  of the interface. A language switch takes effect immediately, without a reload
  and without interrupting running streams.
- **No new entry text:** the hero area stays clean; the explanation appears only
  on demand.
- **Consistency:** the popup shares styling and close behavior with the existing
  security code dialog.
- **Accessibility:** a real `<dialog>` provides focus handling and Escape
  behavior; `aria-label` and `aria-labelledby` name the icon and the dialog.

## Alternatives considered

- **Show help text permanently in the hero:** contradicts the decision to remove
  the hero text block. Rejected.
- **`<details>` disclosure in the top bar:** less noticeable, no modal focus and
  no Escape handling. Rejected.
- **Tooltip on the discover button:** shows the flow only in fragments and is
  unreliable on touch devices. Rejected.
- **Load an icon library:** extra dependency and license maintenance for a single
  icon. Rejected.

## Consequences

- `web/index.html` contains the help button in the top bar and the `#help-dialog`;
  `web/app.js` wires the open click.
- `web/style.css` provides the shared dialog base as `.modal-dialog`; the security
  code dialog uses it too.
- New i18n keys: `help.open`, `help.eyebrow`, `help.title`, `help.text`,
  `help.dismiss` in both catalogs.
- New tests: `web/e2e/i18n.spec.ts` checks the localized help text and closing
  through "Got it". `TestIndexHTMLKeysExistInLocales` covers the new keys.
- No new static files, no license change, no SBOM change.

## Addendum 2026-11-23

The help popup also carries a "Licenses" button (`#open-licenses`). It closes the
help and opens the scrollable license dialog, so license information is reachable
without a separate footer and without a `licenses.html` page. The dialog uses the
same dialog base, lists MIT and all third-party components and links the API
documentation. New keys: `help.wallText`, `licenses.open`.

## References

- ADR-0003: local, no cloud fallback
- ADR-0021: bilingual UI
- ADR-0023: Inter Tight for the brand name and step headings
- `web/index.html`, `web/app.js`, `web/style.css`, `web/locales/{en,de}.json`
- `specs/product.md`, `specs/architecture.md`, `specs/testing.md`
