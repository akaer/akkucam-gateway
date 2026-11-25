# ADR-0032: Semantic versioning and branch workflow

- Status: Accepted
- Date: 2026-11-25

## Context

Before the first commit there were no releases and no branch workflow. The gateway
ships as a container, and operation needs to show which build is running,
especially after a camera or protocol fix. A version string without a link to the
source revision is not enough.

## Decision

### Branch and pull request

- `main` is protected and always buildable. Changes run through `feature/<slug>` or
  `fix/<slug>` and a pull request. Documentation-only changes may use `docs/<slug>`.
- The pull request merges as a squash after green CI, then the branch is deleted.
  Direct commits to `main` are excluded, including version bumps. Template:
  `.github/pull_request_template.md`.

### Version

- Versioning follows Semantic Versioning 2.0.0. The major stays at `0` until the
  stable `1.0.0`; the first version is `0.1.0-beta.1`.
- Before `1.0.0`: `MINOR` for features and breaking changes, `PATCH` for
  backwards-compatible fixes. Prereleases carry `-beta.N`.
- The version lives only in `internal/version/version.go`. It is bumped in the pull
  request that makes the release. After the merge, `main` is tagged with
  `v<version>` (for example `v0.1.0-beta.1`). Tags are never moved or reused. Docker
  tags carry the version without the commit hash.

### Commit in the build

- `Dockerfile`, `Makefile`, `scripts/test.sh` and CI pass the Git commit as the
  build argument `COMMIT`; `-ldflags -X` writes it into the binary.
- Local `go run` builds read the VCS stamp from `debug.ReadBuildInfo()`.
- `GET /api/version` returns `version` and, when known, the seven-character
  `commit`. The help dialog shows both in a quiet line. If the commit is missing,
  the field is dropped; no placeholder is invented.

## Reasoning

- One PR per change keeps `main` buildable and makes CI evidence and review
  traceable.
- SemVer plus commit hash separates release number and exact source revision. The
  hash alone is unreadable for humans, the version alone is not unique.
- A single version source in Go avoids drift between file, tag and binary; the
  commit is injected at build time so the image stays reproducible and no `.git`
  has to land in the build context.
- The help dialog display follows ADR-0015: diagnostic information appears on
  request, not permanently in the top bar.

## Alternatives considered

- **Git tags only, no version constant:** requires `git describe` in the build and
  returns nothing in source archives or without Git.
- **Version through `ARG VERSION` in the Dockerfile:** moves the source into build
  scripts and Compose files; tag and binary can drift apart.
- **Version in `/api/health`:** mixes liveness with build metadata and forces
  monitoring endpoints to carry an extra contract field.
- **`debug.ReadBuildInfo()` alone:** does not work in the container because the
  build context contains no `.git`.
- **Version in the top bar:** permanent display without everyday value and extra
  competition with the live wall.

## Consequences

- Every change runs through a pull request; for single-person changes that costs an
  extra step.
- A release needs a version bump, a merge and a tag. Without a tag, `main` stays an
  intermediate state even when the version is already bumped.
- A plain `docker compose build` without `COMMIT` shows no hash. The documented
  paths `make up`, `make test` and CI set it.
- Branch protection and the required status checks are GitHub settings and not
  versioned as a file in the repository.

## References

- [Semantic Versioning 2.0.0](https://semver.org/)
- ADR-0015: Progressive disclosure
- ADR-0017: OpenAPI as the API contract
- ADR-0024: Help popup in the top bar
- `CONTRIBUTING.md`, `.github/pull_request_template.md`

## Addendum 2026-11-25

- `main` and the release tags `v*` are protected as GitHub rulesets: pull request,
  green checks, linear history, no force pushes, no tag changes. The rules live in
  GitHub, not in the repository.
- A tag `v<version>` starts `.github/workflows/release.yml`. The workflow checks
  that the tag and `internal/version/version.go` match and that the commit is on
  `main`, builds the image for `linux/amd64` and `linux/arm64`, pushes it to GHCR
  and creates a GitHub release with generated notes and the SBOM files.
- The image carries OCI labels with source, license, version and revision; that
  links the GHCR package to the repository.
- The first published release is `v0.1.0-beta.2`. The tag `v0.1.0-beta.1` stays
  unchanged but is not published as a release.
