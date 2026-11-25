# ADR-0018: Test and code quality conventions

- Status: Accepted
- Date: 2026-11-21

## Context

The Go sources contained many unnamed numbers, few comments and scattered test
files with partly inconsistent names. Errors were not typed. The HTTP layer sat
in the `main` package and was therefore harder to test. A shared rule was missing
for where tests live and how to handle new errors, constants and API changes.

## Decision

- **No magic numbers.** Protocol offsets, timeouts, limits and buffer sizes get
  named constants in a package-wide `constants.go`.
- **Documentation.** Exported identifiers get Godoc. Byte layouts and protocol
  quirks are commented inline.
- **Typed errors.** Domain packages use sentinel errors and `%w`. The HTTP layer
  maps them through `internal/apierr`.
- **Tests stay in the package.** Unit and integration tests sit next to the code
  they test. `tests/` is reserved for cross-package end-to-end tests.
- **Build tags.** Hardware tests use `//go:build hardware`. Fake-hub integration
  tests run by default.
- **HTTP layer.** `internal/gateway` contains the API, validation, error mapping
  and route table. `cmd/gateway` is only the entry point.
- **API contract.** Every route must appear in the route table and in
  `internal/openapi/openapi.json`.
- **ADRs.** Structural or contract-relevant changes need a new or updated ADR.

## Reasoning

- Named constants make protocol code readable and reviewable.
- Typed errors decouple behavior from message texts.
- Tests next to the code can reach unexported parsers and builders.
- A thin `main` function and a dedicated API package improve testability and
  clarity.
- The contract test prevents outdated API documentation.

## Alternatives considered

- **All tests in `tests/`**: would make unexported functions unreachable or force
  shim exports.
- **Swagger annotations instead of a separate file**: extra toolchain, see
  ADR-0017.
- **Compare error texts**: brittle, see ADR-0016.

## Consequences

- New sources need a `constants.go` and Godoc when appropriate.
- New tests follow the naming convention in `tests/README.md`.
- New API routes need an entry in the route table and in `openapi.json`.
- `AGENTS.md` summarizes the rules for contributors.

## References

- ADR-0016: Stable API error codes
- ADR-0017: OpenAPI as the machine-readable API contract
- `tests/README.md`
- `AGENTS.md`
