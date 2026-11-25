# ADR-0028: UI copy, terminology and error paths

- Status: Accepted
- Date: 2026-11-24

## Context

The interface was functionally bilingual but linguistically undecided: catalog
values were glued into sentences in `app.js` (`" · scrolls"`, `{suffix}`), three
registers alternated ("Load cameras", "Loading cameras …", "Waiting for the video
stream"), and "browser storage" meant both transient RAM and `localStorage`.
Implementation explanations replaced actions ("Reset removes the entry from
localStorage"), errors offered no retry, and one destructive action deleted all
security codes without asking.

## Decision

### Whole sentences

- A catalog value is a whole sentence or a whole label. Fragments are not
  concatenated in JavaScript.
- Data may fill placeholders (`{count}`, `{channels}`, `{hub}`). The client formats
  lists with `Intl.ListFormat` and numbers and times with `Intl.NumberFormat`.
- The API delivers machine values (`durationMs`, `uptimeSeconds`), not preformatted
  strings.

### Terminology (DE/EN)

| Term | Meaning | Avoid |
|---|---|---|
| Basisstation / base station | the device on the network | "Hub" in visible text |
| Kamera / camera | the coupled channel | "Kanal" for the device |
| Kanal / channel | the number of an input | "Kamera" for the number |
| Livebild / live view | the visible picture | "Stream" for the picture |
| Stream / stream | the transport (start/stop) | none |
| Details / details | the drawer | "Drawer", "Seitenleiste" in German |
| Live-Daten / live data | the overlay on the tile | "Metriken" |
| Arbeitsspeicher / memory | transient browser RAM | "Browserspeicher" (ambiguous) |

### Register and tone

- German uses the informal "Du" throughout.
- Progress messages use the state passive or an infinitive ("Kameras werden geladen
  …", "Videostream wird vorbereitet …"), not an imperative aimed at the user.
- No implementation disclaimers: no `localStorage`, no "matrix", no "fallback", no
  "optional" hints. Explain the action or the benefit, not the code.
- Error messages do not end in a dead end. When a retry is possible, the message
  offers "Erneut versuchen".

### Destructive actions

- "Remove all security codes" needs two clicks: the first arms the action and
  changes the label ("Remove all codes for real?"), the second executes it. After
  four seconds the button falls back.
- No modal confirmation dialog; the second click is the confirmation.

### Device code

- The origin and storage of the security code on the base station are not part of
  the product. The interface does not explain them; the code is entered, checked
  and held in browser RAM only as ADR-0012 describes.

### Codes and password managers

- The code field carries `autocomplete="off"` so no password manager persists the
  device code (ADR-0012, ADR-0025).

## Reasoning

- One whole sentence per key is the only form that translates without guessing word
  order.
- A glossary prevents the same area from getting five names ("Wand", "Live-Wand",
  "Wandsteuerung", "Wand-Einstellungen", "Wand-Einrichtung").
- Machine values from the API keep formatting and language in one place.
- The two-click confirmation protects against data loss without opening another
  dialog.

## Alternatives considered

- **Keep fragments in the catalog and assemble them in JS:** wrong across languages
  (word order, brackets, list separators) and invisible to the catalog tests.
  Rejected.
- **Build time and number formats in the Go server:** would bind the API to one
  language and contradict ADR-0021. Rejected.
- **Formal "Sie" in German:** inconsistent with the shipped "Du" form and too
  distant for a home network device. Rejected.
- **Modal confirmation dialog before deletion:** an extra layer for a reversible
  action (codes can be entered again). Rejected.

## Consequences

- `web/locales/{en,de}.json` no longer holds fragment values;
  `TestLocaleValuesAreWholePhrases` checks that.
- `web/app.js` formats lists, milliseconds and uptimes; the API delivers
  `durationMs` and `uptimeSeconds` (OpenAPI, `specs/api.md`).
- The hub manager and diagnostics share the same two-click mechanism; the
  credentials Playwright suite covers it.
- Locked actions (motion, discovery, "Add base station") carry a `title` with the
  reason.
- New copy follows the glossary; deviations need a change to this ADR.

## References

- ADR-0012/ADR-0025: Code in browser RAM only
- ADR-0014: Only available metadata
- ADR-0021: Bilingual UI
- ADR-0029: Design tokens
- `web/locales/`, `web/app.js`, `web/i18n.js`
- `specs/product.md`, `specs/api.md`, `specs/testing.md`
