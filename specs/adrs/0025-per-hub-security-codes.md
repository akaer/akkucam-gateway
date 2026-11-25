# ADR-0025: Security code per base station in browser RAM

- Status: Accepted
- Date: 2026-11-22
- Amends: [ADR-0012](0012-browser-session-credential.md)

## Context

ADR-0012 assumed a single base station and kept exactly one security code in a
JavaScript variable. As soon as two base stations with different codes sit on the
network, the interface unlocks both with the same code. The second base station
rejects the login (`unauthorized`), and the user cannot open it without reloading
the page.

## Decision

- Security codes live in browser RAM in a `Map`, addressed by the base station DID
  (`state.credentials`). The DID stays stable across discovery runs; when it is
  missing, the IP serves as the key.
- The code dialog names the base station (DID/model) so it is clear which code is
  expected. The input stays empty each time it opens.
- Each hub card shows a "Code saved" chip with a "Remove code" button as soon as a
  code for that DID sits in RAM. The chip appears only when a code exists.
- The button in diagnostics clears **all** codes (`state.credentials.clear()`).
- When the API answers with `unauthorized`, the code for the affected DID is
  discarded. The next attempt asks for the code again; there is no automatic retry.
- "Start all" and parallel channel actions share one prompt per DID (single
  flight), so several dialogs do not open at once.
- ADR-0012 still applies: no storage in `localStorage`, `sessionStorage`, cookies,
  IndexedDB or on the server. A reload empties the map.

## Reasoning

- **Security:** the code stays transient in browser RAM; the mapping to a DID is
  runtime information only.
- **Usability:** several base stations work without repeated reloads; codes entered
  once are reused per station.
- **Error resilience:** a rejected code does not block the station permanently
  because it is discarded after `unauthorized`.

## Consequences

- `web/app.js`: `state.credentials` replaces `state.securityCode`;
  `requestSecurityCode(hub)` with single flight; `forgetCredential`; `unauthorized`
  discards the code.
- `web/index.html`: hub name in the code dialog (`#code-dialog-hub`), chip per hub
  card, diagnostics button "Remove all security codes from browser storage".
- `web/style.css`: layout for `.hub-actions`, `.hub-credential`, `.dialog-hub`.
- `web/locales/{en,de}.json`: `dialog.hub`, `hub.codeSaved`, `hub.forgetCode`,
  `notice.forgotAllCodes`; `notice.forgotCode` and `dialog.hint` adjusted.
- No API, OpenAPI or SBOM change: the server still receives the code per request.
- New tests: `web/e2e/credentials.spec.ts` (separate codes, reuse, discard after
  401, remove code per hub).

## References

- ADR-0007: no credentials on disk
- ADR-0012: security code in browser RAM until reload
- ADR-0016: stable API error codes (`unauthorized`)
- `specs/security.md`, `specs/product.md`, `specs/testing.md`

## Addendum 2026-11-24

After ADR-0027 the global button "Remove all security codes from browser storage"
exists in two places: in the diagnostics area of the setup page (`#forget-code`)
and in the wall hub manager (`#forget-all-codes`). Both call the same handler
(`forgetAllCredentials`) and clear the whole map.

## Addendum 2026-11-24

In the interface the global button is called "Remove all security codes from
browser working memory" and needs two clicks: the first arms the action and
changes the label, the second removes the codes (ADR-0028). The wording "browser
storage" is gone because it could also mean `localStorage`.
