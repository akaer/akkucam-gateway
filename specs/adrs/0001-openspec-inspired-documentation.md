# ADR-0001: OpenSpec-inspired documentation without a CLI requirement

- Status: Accepted
- Date: 2026-11-21

## Context

The project grew out of reverse engineering. Its behavior and contracts would
otherwise live only in chat history. OpenSpec offers a structure for current
specifications and change proposals, but adds a toolchain.

## Decision

The repository keeps its specifications in `specs/` as plain Markdown. The layout
follows OpenSpec: current documents at the top level, architecture decisions in
`adrs/`, and change proposals in `changes/`. No OpenSpec CLI, schema or generator
is required.

## Reasoning

- Plain Markdown needs no installation and no build step.
- The structure stays readable in any editor and on GitHub.
- Decisions and current behavior sit next to each other.
- A later move to OpenSpec tooling stays possible because the layout already
  matches.

## Alternatives considered

- **OpenSpec CLI and schema:** adds a dependency and a validation step for a
  single-person project.
- **Wiki or issue tracker only:** separates behavior from the code and from the
  decisions.
- **Code comments only:** carries no product or protocol context.

## Consequences

- Every behavior change updates the matching document.
- Structural or contract changes add an ADR.
- `specs/README.md` explains the document types and maintenance rules.

## Addendum 2026-11-24

- ADRs are historical once accepted. Corrections go into a dated `## Addendum`
  section; the original decision stays visible.
- A replaced decision gets the status `Superseded` or `Partially superseded` and a
  reference to the replacing ADR.
- Implemented changes move from `changes/` to `changes/archive/`.
