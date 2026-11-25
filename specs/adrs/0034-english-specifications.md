# ADR-0034: English specifications

- Status: Accepted
- Date: 2026-11-25

## Context

The source of truth in `specs/` was written in German. The code, the English
README, `CONTRIBUTING.md`, `SECURITY.md`, the issue templates and the commit
messages are English. A reader who wants to check a behavior against the spec
has to switch languages mid-task. Reviewers outside the German-speaking world
cannot follow the decisions at all.

ADR-0001 fixes the document structure and treats ADRs as historical records.
Changing the language must not change a decision, a date or a status.

## Decision

- `specs/` is written in English, starting with this ADR. The existing
  documents are translated in place.
- A translation keeps the structure, the ADR numbers, the dates, the status
  values and the meaning. It does not add, remove or reorder arguments.
- The German originals stay in the Git history. No second copy is maintained.
- `README_de.md` stays German. It is the German user-facing README and not part
  of the specification set.
- German UI quotes inside an ADR stay German. The surrounding prose becomes
  English.
- Headings follow one vocabulary: `Context`, `Decision`, `Reasoning`,
  `Alternatives considered`, `Consequences`, `References`, `Addendum`,
  `Clarification`.
- Normative keywords keep their RFC 2119 meaning: `MUSS` becomes `MUST`,
  `DARF NICHT` becomes `MUST NOT`, `SOLL` becomes `SHOULD`.

## Reasoning

- One language for the source of truth matches the code and the contribution
  documents. A reader no longer switches languages to verify a behavior.
- English reaches more reviewers and matches the language of the upstream
  projects the gateway depends on.
- Translating in place avoids two documents that drift apart. The Git history
  keeps the German text for anyone who needs it.
- The ADR structure and the decisions survive a translation. Only the wording
  changes.

## Alternatives considered

- **Keep German:** the mixed language stays, and non-German reviewers cannot
  follow the contracts.
- **English twin documents next to the German originals:** doubles the
  maintenance and invites drift.
- **Translate only the ADRs:** the top-level specs would stay German while the
  ADRs that reference them become English.
- **Machine translation without review:** the protocol and security documents
  carry normative meaning; wrong wording changes the contract.

## Consequences

- One large documentation diff. Grouped commits keep it reviewable.
- German-speaking maintainers read and write the specs in English from now on.
- Internal anchors that point at translated headings change. The one internal
  anchor in `specs/roadmap.md` targets a language-neutral heading and stays.
- The README language note changes in both READMEs so the pair stays in step.

## References

- ADR-0001: OpenSpec-inspired documentation
- ADR-0021: Bilingual browser interface
- `AGENTS.md`, `specs/README.md`, `README.md`, `README_de.md`
