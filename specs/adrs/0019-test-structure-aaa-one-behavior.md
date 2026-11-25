# ADR-0019: Test structure: Arrange/Act/Assert and one behavior per test

- Status: Accepted
- Date: 2026-11-22

## Context

Many tests checked several independent things in one function. Names containing
"And" made it hard to attribute a failure. The three phases Arrange, Act and
Assert were not visible everywhere. Tests grew longer, harder to read and
imprecise on failure.

## Decision

- Every test follows **Arrange/Act/Assert**. Comments mark the phases when they
  are not obvious.
- **One test checks one behavior.** Several assertions are allowed when they
  describe the same result (for example several fields of a parsed object).
- Variants of the same behavior run as **table subtests** with their own AAA.
- Test names describe the behavior and contain **no "And"**.
- **Integration tests** split into separate tests and sessions when they check
  independent results. That applies even at the cost of runtime.
- **Hardware tests** split as well. Every session checks exactly one result.
- **Fuzz targets** are the only exception: one target equals one parser or
  behavior, without AAA comments.
- Helpers may encapsulate Arrange but must not hide Act or Assert.

## Reasoning

- A failure shows immediately which behavior is affected.
- Smaller tests are easier to read, review and change.
- Separate sessions stop a failure in one part from hiding the observation of
  another part.
- The AAA comments make the test intent visible without prior knowledge.

## Alternatives considered

- **Keep everything in one test:** less runtime, but unclear failures and tighter
  coupling.
- **Apply only to new tests:** existing tests stay inconsistent.
- **Strictly one assertion per test:** would split related contract fields
  without need.

## Consequences

- More test functions and partly longer runtime, especially for integration and
  hardware.
- Shared helpers for recurring Arrange code.
- `AGENTS.md`, `specs/testing.md` and `tests/README.md` document the convention.
- New tests follow the convention; existing tests are adjusted step by step.

## References

- ADR-0018: Test and code quality conventions
- `specs/testing.md`
- `tests/README.md`
