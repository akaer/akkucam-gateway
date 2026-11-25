# ADR-0021: Bilingual browser interface (German/English)

- Status: Accepted
- Date: 2026-11-22

## Context

The interface was German only. The gateway is also interesting for English
speakers and for reviews. The API already returns stable error codes (ADR-0016)
that the client can translate. The UI is a static single-page app without a
framework and without a build step.

## Decision

The browser interface becomes bilingual on the client (German and English):

- **Source language English.** `web/index.html` contains English text. Without
  JavaScript the page stays English.
- **Catalogs as JSON.** `web/locales/en.json` and `web/locales/de.json` are served
  by the existing static file server. `web/i18n.js` loads both catalogs and
  provides `t()`, `setText()`, `applyTranslations()` and `setLanguage()`.
- **Language detection.** Order: `?lang=de|en` in the URL, then
  `localStorage['akkucam-lang']`, then `navigator.languages` (first `de*` match),
  otherwise English. Only `de*` switches to German.
- **Live switching.** A language switch in the top bar changes the language
  without a reload. Running streams are not interrupted: the wall redraws with
  the new text but reattaches existing `<video>` elements and HLS instances
  instead of recreating them (ADR-0027).
- **Persistence.** The explicit language choice lives in
  `localStorage['akkucam-lang']`. `?lang=` overrides it but is not stored. The
  theme stays in `localStorage['akkucam-theme']` as before.
- **API stays English.** `error` still carries the English server message. The UI
  translates known `code` values through `error.code.<code>` from the catalog;
  unknown codes show the server message.
- **No translation of device data.** DIDs, model names, IP addresses and channel
  names from the hub stay unchanged.
- **Accessibility.** `<html lang>` updates on every switch. Translatable `title`
  and `aria-label` attributes use `data-i18n-title` and `data-i18n-aria-label`.
- **Swagger UI stays English.** It is developer tooling and documented as an
  exception.
- **Playwright as the E2E tool.** `web/e2e/` contains browser tests for language
  detection, live switching, persistence, error codes and optionally real
  hardware. The tests run in the official Playwright image, not through a Node
  toolchain in the production path.

## Reasoning

- The UI is static; an i18n framework would be overkill.
- JSON catalogs are checkable with Go tests (same keys, same placeholders, no
  empty values).
- Stable error codes make server-side localization unnecessary.
- Live switching without a reload keeps running streams and avoids a camera
  wake-up.
- English as the source and fallback language is the smallest common base for
  unknown browser languages.

## Alternatives considered

- **Server-side localization:** rejected. The server would need to know language,
  catalogs and templates, and the API would still be English.
- **i18n framework (for example i18next):** rejected. A new runtime dependency and
  build step for a small interface.
- **German only with an English API:** rejected. The requirement is explicitly
  bilingual.
- **Language switch with a reload:** rejected. It interrupts streams and costs
  another camera wake-up.

## Consequences

- All visible UI text goes through `web/i18n.js`; hardcoded text in HTML or
  JavaScript is not allowed. Device data and proper names are exceptions.
- `de.json` and `en.json` must have the same keys and placeholders. Go tests in
  `internal/gateway/i18n_test.go` check that, including every `data-i18n` key in
  `index.html`, the keys used in `app.js` and the API error codes.
- New API error codes need an entry in both catalogs.
- New languages need a catalog, an entry in `SUPPORTED` and a Playwright test.
- Playwright tests live in `web/e2e/` and run through `scripts/test-e2e.sh`. The
  live suite is optional and needs `AKKUCAM_TEST_SECURITY_CODE`; it writes no
  artifacts so the security code never reaches disk.
- The language switch is UI state; it changes neither the API contract nor the
  protocol.

## Addendum 2026-11-24

- Catalog values are whole sentences. `app.js` does not concatenate fragments;
  data fills placeholders (ADR-0028).
- The client formats numbers, times and lists with `Intl`. The API returns
  `durationMs` and `uptimeSeconds` instead of preformatted strings so the output
  follows the active language.
- `TestLocaleValuesAreWholePhrases` extends the key and placeholder check with
  this rule.

## References

- ADR-0012: Security code in browser RAM only
- ADR-0013: Responsive dashboard and theme
- ADR-0016: Stable API error codes
- `specs/product.md`, `specs/architecture.md`, `specs/api.md`,
  `specs/security.md`, `specs/testing.md`
- `web/i18n.js`, `web/locales/`, `web/e2e/`
