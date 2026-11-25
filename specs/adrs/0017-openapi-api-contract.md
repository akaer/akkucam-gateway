# ADR-0017: OpenAPI as the machine-readable API contract

- Status: Accepted
- Date: 2026-11-21

## Context

The HTTP API was documented only as prose in `specs/api.md`. There was no
machine-readable contract, no interactive documentation and no automatic check
that the documentation matched the registered routes.

## Decision

- `internal/openapi/openapi.json` is the canonical OpenAPI 3.0 contract.
- `go:embed` embeds the file into the binary and serves it at
  `GET /api/openapi.json`.
- A locally bundled Swagger UI is served at `GET /api/docs`. No CDN or cloud
  resources are loaded.
- The Swagger UI assets live under `web/swagger/` and are served from the
  existing web root.
- A contract test compares the route table in `internal/gateway` with the paths
  and methods in the OpenAPI file. A mismatch fails the test.
- The file lives under `internal/openapi/` because `go:embed` can only embed
  files inside the package directory. `specs/api.md` remains the human
  companion and points to the contract.

## Reasoning

- A machine-readable contract for clients, generators and tests.
- No extra runtime dependency and no cloud access.
- The contract test stops documentation and implementation from drifting.
- A local Swagger UI makes the API explorable without extra tools.

## Alternatives considered

- **Only `specs/api.md`**: already present, but not machine-readable and not
  testable.
- **Swagger annotations in Go code**: adds toolchain dependencies and mixes
  documentation with handler code.
- **Load Swagger UI from a CDN**: violates the cloud-free rule.
- **OpenAPI file under `specs/`**: would break `go:embed` or require a copy.

## Consequences

- Every API change must update `openapi.json`.
- The spec uses a relative server URL so "Try it out" works from the current
  origin and does not point at `127.0.0.1` unconditionally.
- The Swagger UI assets are Apache-2.0; the license sits in
  `web/swagger/LICENSE`.
- `README.md` and `specs/api.md` link `/api/docs` and `/api/openapi.json`.
- The contract test runs with `go test ./...`.

## References

- ADR-0016: Stable API error codes
- `internal/openapi/openapi.go`
- `internal/gateway/openapi_test.go`
- `specs/api.md`
