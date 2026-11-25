# Gateway specifications

This folder is the technical and product source of truth for the AkkuCam
Gateway. The structure follows [OpenSpec](https://github.com/Fission-AI/OpenSpec)
but stays usable without an extra CLI or build dependency.

## Document types

### Current specifications

| Document | Content |
|---|---|
| [product.md](product.md) | Goal, user value, scope and requirements |
| [architecture.md](architecture.md) | System context, components and data flow |
| [protocol.md](protocol.md) | Verified state of the local PPCS protocol |
| [api.md](api.md) | HTTP API and runtime behavior |
| [security.md](security.md) | Trust boundaries, secrets and residual risks |
| [testing.md](testing.md) | Test strategy and verified hardware |
| [roadmap.md](roadmap.md) | Open topics and planned development |
| [sources.md](sources.md) | Reverse-engineering sources and third-party software |

The machine-readable API contract lives outside this folder in
[`internal/openapi/openapi.json`](../internal/openapi/openapi.json) because
`go:embed` embeds it into the application. `api.md` describes the same contract
for humans.

### Architecture Decision Records

The decisions under [adrs/](adrs/) explain **what** was implemented, **why**,
which alternatives were considered and which consequences were accepted.

ADRs are historical documents once accepted. Corrections and additions go into
a dated `## Addendum` section so the original decision stays readable. When a
later ADR replaces a decision in whole or in part, the original gets the status
`Superseded` or `Partially superseded` and a reference to the new ADR. The
status appears in the ADR header and in the status column of
[`adrs/README.md`](adrs/README.md).

## OpenSpec relationship

OpenSpec separates permanent specifications from planned changes. This project
follows the same split:

```text
specs/
├── *.md                 # current, implemented state
├── adrs/                # permanent, traceable decisions
└── changes/             # change artifacts
    ├── <change-name>/   # planned, not implemented yet
    │   ├── proposal.md
    │   ├── design.md
    │   └── tasks.md
    └── archive/<change-name>/  # implemented and archived (reference)
```

The `changes/` folder appears only when a concrete, not yet implemented change
is planned. After implementation the change moves to
`changes/archive/<change-name>/`, where it stays as a reference and no longer
counts as an open proposal. Implemented findings are not presented afterwards
as open changes.

## Maintenance

A code change must update the specification when it changes at least one of
these points:

- public API or UI behavior
- network protocol or packet format
- storage or processing of credentials
- deployment, container or network model
- media format, quality, latency or metadata
- supported devices or firmware variants
- an existing architecture decision
