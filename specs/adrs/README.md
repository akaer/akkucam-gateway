# Architecture Decision Records

## Status

- `Proposed`: under discussion
- `Accepted`: currently valid
- `Deprecated`: still present but no longer recommended
- `Superseded`: replaced by a later ADR
- `Partially superseded`: replaced in essential parts by a later ADR; additions appear in a dated `Addendum` in the document
- `Rejected`: deliberately not implemented

## Index

| ADR | Decision | Status |
|---|---|---|
| [0001](0001-openspec-inspired-documentation.md) | OpenSpec-inspired documentation without a CLI requirement | Accepted |
| [0002](0002-use-go-for-the-native-gateway.md) | Go for the native gateway | Accepted |
| [0003](0003-local-only-without-cloud-fallback.md) | Local operation only | Accepted |
| [0004](0004-docker-host-network.md) | Docker host networking for broadcasts | Accepted |
| [0005](0005-clean-room-pppp-implementation.md) | Native PPPP reimplementation instead of a proprietary third-party library | Accepted |
| [0006](0006-h264-passthrough-and-hls.md) | H.264 passthrough | Accepted |
| [0007](0007-credentials-in-memory-only.md) | Security code in memory only | Superseded |
| [0008](0008-serialize-session-control.md) | Serialize session control paths | Accepted |
| [0009](0009-one-active-stream-per-did.md) | One active stream per base station | Superseded |
| [0010](0010-runtime-metadata-from-local-protocol.md) | Read bitrate and battery locally | Accepted |
| [0011](0011-concurrent-stream-per-channel.md) | One stream per DID and channel in parallel | Accepted |
| [0012](0012-browser-session-credential.md) | Security code in browser RAM until reload | Accepted |
| [0013](0013-responsive-dashboard-and-theme.md) | Responsive dashboard and theme | Accepted; changed 2026-11-23 (ADR-0030) |
| [0014](0014-render-only-available-metadata.md) | Render only available metadata | Accepted |
| [0015](0015-progressive-disclosure-metadata-and-metrics.md) | Collapsible metadata and live metrics | Partially superseded (ADR-0031) |
| [0016](0016-stable-api-error-codes.md) | Stable API error codes and errors | Accepted |
| [0017](0017-openapi-api-contract.md) | OpenAPI as the machine-readable API contract | Accepted |
| [0018](0018-test-and-code-quality-conventions.md) | Test and code quality conventions | Accepted |
| [0019](0019-test-structure-aaa-one-behavior.md) | Test structure: Arrange, Act, Assert and one behavior per test | Accepted |
| [0020](0020-no-audio-support.md) | No audio support | Accepted |
| [0021](0021-ui-internationalization.md) | Bilingual browser interface (German/English) | Accepted; amended 2026-11-24 (ADR-0028) |
| [0022](0022-app-icon-licenses-and-sbom.md) | App icon, favicon, license notices and SBOM | Accepted; license notices replaced 2026-11-23 (ADR-0024/0027) |
| [0023](0023-inter-tight-brand-and-headings.md) | Inter Tight for the brand name and step headings | Accepted; amended 2026-11-24 (ADR-0029) |
| [0024](0024-help-popup.md) | Info icon and help popup in the top bar | Accepted; amended 2026-11-23 (license dialog) |
| [0025](0025-per-hub-security-codes.md) | Security code per base station | Accepted |
| [0026](0026-multiple-base-stations.md) | Serve several base stations at once | Partially superseded (ADR-0027) |
| [0027](0027-live-wall-layout.md) | Live wall, views, details drawer and control bar | Accepted; changed 2026-11-24 (ADR-0031) |
| [0028](0028-ui-copy-and-terminology.md) | UI copy, terminology and error paths | Accepted |
| [0029](0029-design-tokens-and-icons.md) | Design tokens and one icon set | Accepted |
| [0030](0030-accessibility.md) | Accessibility of the interface | Accepted |
| [0031](0031-live-metrics-overlay.md) | Live data overlay and visibility setting | Accepted |
| [0032](0032-versioning-and-branch-workflow.md) | Semantic versioning and branch workflow | Accepted |
| [0033](0033-static-analysis-toolchain.md) | Static analysis in the toolchain container | Accepted |
| [0034](0034-english-specifications.md) | English specifications | Accepted |

## Template

```markdown
# ADR-NNNN: Title

- Status: Proposed
- Date: YYYY-MM-DD

## Context
## Decision
## Reasoning
## Alternatives considered
## Consequences
## References
```
