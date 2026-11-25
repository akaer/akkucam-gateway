# ADR-0012: Security code in browser RAM until page reload

- Status: Accepted
- Date: 2026-11-21
- Supersedes: [ADR-0007](0007-credentials-in-memory-only.md)
- Amended by: [ADR-0025](0025-per-hub-security-codes.md) (one code per base station)

## Context

Several channels need several authenticated PPCS sessions. Entering the code
again for every channel makes "Start all" and channel switching impractical.

## Decision

The browser asks for the security code per base station when it loads that
station's camera list and keeps it only in a JavaScript variable. It stays
available until page reload, tab close or an explicit "Remove code". It is never
written to localStorage, sessionStorage, a cookie or a server file. With several
base stations, the browser keeps one code per DID (ADR-0025).

## Reasoning

- One code prompt per browser session
- Parallel starts and channel switches stay convenient
- Reload removes the secret from application state reliably
- No persistent browser storage

## Consequences

- JavaScript in the page context can reach the code, so no external runtime
  scripts may be included.
- Running streams stay active after "Remove code"; new starts need the code
  again.
- The server still receives the code only per session setup and does not persist
  it.

## Addendum 2026-11-24

The code input carries `autocomplete="off"`. Password managers and browser
autofill should not store the device code permanently; that would be the kind of
persistence this decision rules out (ADR-0028). The UI copy also calls the
storage location "working memory" instead of "browser storage", because the
latter also means `localStorage`.
