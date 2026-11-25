# ADR-0016: Stable API error codes and typed errors

- Status: Accepted
- Date: 2026-11-21

## Context

The HTTP API returned errors only as a free-form string in the `error` field.
Clients could distinguish errors only by HTTP status or by brittle text
comparison. Some internal errors reached the browser unchanged. Error codes were
missing entirely.

## Decision

- Every error response carries a stable `code` field next to the existing
  `error` field.
- The format stays flat and backward compatible: `{"error":"...","code":"..."}`.
- Codes are defined centrally in `internal/apierr`.
- Domain packages return typed sentinel errors (`errors.Is`) and wrap them with
  `%w`.
- The HTTP layer maps sentinel errors to a code and status. It does not compare
  error texts.
- Unknown errors return `internal_error` with HTTP 500 and a generic message.
  Details stay in the log.
- The security code appears in neither error messages nor logs.

## Reasoning

- Clients can handle errors reliably and independent of language.
- The API contract stays compatible for existing callers because `error` does
  not change.
- Typed errors stop behavior from changing when a message text is edited.
- Generic messages for unexpected errors reduce the risk of leaking internal
  details.

## Alternatives considered

- **Nested format** `{"error":{"code":"...","message":"..."}}`: more expressive,
  but breaks existing clients and the current dashboard logic.
- **HTTP status only:** too coarse; several causes share the same status.
- **Compare error texts:** already in use and rejected as brittle.

## Consequences

- `web/app.js` shows a translated description of the code
  (`error.code.<code>`) instead of the server message; unknown codes fall back to
  the server message.
- New error paths need a code in `internal/apierr` and a test.
- `specs/api.md` documents the format and the codes.
- `internal/openapi/openapi.json` describes the error schema.

## References

- ADR-0017: OpenAPI as the machine-readable API contract
- `internal/apierr/apierr.go`
- `specs/api.md`

## Addendum 2026-11-24

- The default branch of `writeError` in `internal/gateway/server.go` delegates to
  `apierr.Write`. Unknown errors therefore return `internal_error` with HTTP 500
  and a generic message as decided above; details go to the handler log. Before,
  they were reported incorrectly as `upstream_error`/502, so write failures in
  the HLS directory looked like device failures.
- `TestUnknownErrorUsesInternalError` in `internal/gateway/api_test.go` covers
  this path.
