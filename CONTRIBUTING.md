# Contributing

## Branches and pull requests

`main` is protected and always buildable. Every change goes through a branch and
a pull request:

- `feature/<slug>` for new behavior
- `fix/<slug>` for bug fixes
- `docs/<slug>` for documentation-only changes

Keep branches focused on one change, rebase on `main` when it moved, and squash
merge the pull request. Delete the branch after the merge. Direct commits to
`main` are not allowed, including version bumps and documentation fixes.

CI must be green before merging: Go vet, unit and integration tests, race
detector, image build with SBOM, and the Playwright suites. UI changes need a
Playwright check in both browsers.

## Versioning

The project follows [Semantic Versioning 2.0.0](https://semver.org/) and starts
in the `0.y.z` beta range. Until `1.0.0`:

- `MINOR` for new features and breaking changes
- `PATCH` for backwards-compatible fixes
- prereleases carry a suffix such as `-beta.1`

`internal/version/version.go` holds the version. Bump it in the pull request
that makes the release, merge it, then tag `main` with `v<version>` (for example
`v0.1.0-beta.1`). A tag is never moved or reused. Docker tags use the version
without the commit hash.

The build carries the short Git commit: `make up`, `make test` and CI pass it
through `COMMIT` to the image, and local `go run` builds read the VCS stamp Go
records. `GET /api/version` and the help dialog show version and commit; a build
without a known commit omits the hash instead of inventing one.

## Tests

```bash
make test        # vet, unit, API and integration tests, race detector, image build
make test-e2e    # Playwright, Chromium and Firefox
```

Hardware checks are described in [`specs/testing.md`](specs/testing.md). Every
change keeps `specs/`, the OpenAPI document, both READMEs and the locale
catalogs in step, as described in [`AGENTS.md`](AGENTS.md).

## Secrets

Never commit or attach real security codes, DIDs, local IP addresses, capture
files or tokens. The gateway keeps the device security code in memory only.
