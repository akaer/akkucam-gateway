# ADR-0033: Static analysis in the toolchain container

- Status: Accepted
- Date: 2026-11-25

## Context

Until now `go vet`, unit and integration tests, fuzzing seeds, Playwright and the
SBOM secured quality. Missing were a check for unchecked errors, injection sinks,
CVEs in the Go standard library and an analysis of the browser sources;
`web/app.js` sees no Go tool. Installing tools directly on NixOS would leave the
Docker workflow and create a second, drifting toolchain in CI.

A trial run on the existing state produced 21 Go findings, 0 Semgrep findings, 2
Shellcheck warnings, 1 Hadolint warning, and 32 reachable CVEs in the Go standard
library because Go 1.23 has received no fixes since August 2025.

## Decision

- `./scripts/lint.sh` (target `make lint`) runs every analysis in a pinned
  container image. The image variables in the script are the only version source
  for local and CI; there is no Nix package and no CI-specific installer.
- **golangci-lint** v2.15.0, `default: standard` plus `gosec`, `errorlint`, `noctx`,
  `bodyclose`, `misspell`; configuration in `.golangci.yml`.
- **govulncheck** v1.10.0 in the Go image of the build stage (Go 1.27.4) and
  therefore on the same standard library as the binary.
- **semgrep** 1.187.0 with `p/golang`, `p/javascript`, `p/secrets` and custom rules
  in `.semgrep/gateway.yml`: no outbound HTTP clients (ADR-0003), no outbound TCP,
  no security code in `localStorage`/`sessionStorage` (ADR-0007, ADR-0012).
- **shellcheck**, **hadolint**, **actionlint** for scripts, Dockerfile and
  workflows.
- Every tool blocks. Exceptions live only in `.golangci.yml` and are justified
  there: G115 in the protocol package `internal/ppcs`, G304/G306 in tests.
- Every `//nolint` carries a reason and a reference. Currently there is exactly one:
  G706 in the request logger because the path is masked with `%q` and `log` does
  not interpret escapes.
- Semgrep fetches the registry rulesets at runtime and sends no metrics
  (`--metrics=off`, `SEMGREP_SEND_METRICS=off`). The custom rules live in the repo.
- The toolchain bump to Go 1.27 belongs to the same step because an analysis gate
  on an EOL toolchain only produces CVE noise.

## Reasoning

- Container pinning keeps the NixOS host, local runs and CI identical and matches
  the existing pattern of `test.sh` and `sbom.sh`.
- Semgrep covers `web/*.js` without Node, npm or a build step.
- govulncheck checks reachable code instead of package versions only; it does not
  replace the SBOM but adds the question "do we call this?".
- The custom Semgrep rules make the hard rules testable that previously existed only
  in prose.
- Analysis tools do not land in the image; `go.mod` stays without third-party
  modules, and SBOM and license dialog do not change.

## Alternatives considered

- **`go vet` and staticcheck only**: no security rules, no frontend.
- **gosec and staticcheck separately**: more scripts for the same coverage.
- **ESLint** for the frontend: npm plus a build step, violates the frontend rule.
- **CodeQL**: better taint analysis, but tied to GitHub Actions and to GHAS for
  private repositories; there is no remote yet.
- **nix-shell or Flake dev shell**: the nixpkgs version drifts, CI would need a
  second setup, and Semgrep is not version-identical there.
- **Leave govulncheck on Go 1.23**: 32 reachable standard library CVEs, including
  hits through `net/http.ListenAndServe`, the core of the gateway.

## Consequences

- New commands: `make lint`, `./scripts/lint.sh`; new CI job `lint`.
- Semgrep rules change independently of the repository. A new finding is fixed or
  narrowly excluded, never by turning off the gate.
- Lint needs network access for image pulls and registry rules.
- On a Go bump, the Go image, govulncheck and golangci-lint must match;
  golangci-lint must be built with at least the module Go version.
- No effect on the runtime image, SBOM or license dialog.
- Phase 2, open and noted in `specs/roadmap.md`: container image scan with Trivy on
  the built image, complementing the SBOM.

## References

- ADR-0003: Local operation without cloud fallback
- ADR-0006: H.264 passthrough and HLS
- ADR-0007: Credentials in memory only
- ADR-0012: Security code in browser RAM
- ADR-0018: Test and code quality conventions
- `specs/testing.md`, `specs/security.md`, `.golangci.yml`, `.semgrep/gateway.yml`
