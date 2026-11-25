# Test layout

Go tests live next to the package they exercise. This is intentional:

- in-package tests can access unexported protocol builders and parsers,
- `go test ./...` discovers them without extra configuration,
- `go vet` and the race detector work on the same package graph.

## Naming

| File | Purpose |
|---|---|
| `<source>_test.go` | Unit tests for the matching source file |
| `protocol_integration_test.go` | Deterministic fake-hub protocol integration |
| `protocol_hardware_test.go` | Real hardware test, `//go:build hardware` |
| `protocol_fuzz_test.go` | Fuzz seeds and targets |
| `api_test.go` | HTTP contract tests in `internal/gateway` |

## Arrange/Act/Assert

Every test follows three phases:

```go
func TestThing(t *testing.T) {
	// Arrange
	input := ...

	// Act
	result, err := Do(input)

	// Assert
	if err != nil { ... }
	if result != want { ... }
}
```

Mark the phases with comments when they are not obvious. One test verifies one behavior. Multiple assertions are fine when they describe the same result. Variants use table-driven subtests. Avoid `And` in test names. Split integration and hardware tests into separate sessions per behavior; only fuzz targets are exempt. See ADR-0019.

## Directories

- `internal/*` — unit and integration tests for that package.
- `internal/gateway` — HTTP contract tests and the OpenAPI route contract.
- `tests/` — only cross-package Go tests. No unit tests here. Repository checks that read files outside the Go packages (`sbom_test.go`, `license_test.go`, `readme_test.go`) live here too.
- `web/e2e/` — Playwright browser tests. Run them with `./scripts/test-e2e.sh`.

## Running

```sh
./scripts/test.sh
```

Hardware tests require `AKKUCAM_TEST_*` environment variables and are excluded from the default build.
