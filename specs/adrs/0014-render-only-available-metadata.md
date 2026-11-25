# ADR-0014: Render only metadata that is actually available

- Status: Accepted
- Date: 2026-11-21

## Context

Not every protocol object delivers the same metadata. In particular, the hub
camera list contains no camera MAC address. Generic placeholders such as
"unknown", "not reported" or empty data fields create visual noise and suggest
information the gateway does not have.

## Decision

The interface renders a metadata field only when a concrete value from a
verified source exists. Missing fields are omitted entirely. No value is
guessed, derived or replaced by generic text.

Existing values are shown in full. Long identifiers such as the DID or the MAC
address wrap when needed instead of being cut off with an ellipsis.

Functional UI labels such as "Channel 2" are exempt: they serve operation and
are not claimed device metadata.

## Reasoning

- Higher information density and less noise
- A clear separation between knowledge and non-knowledge
- No misleading statements
- Technical users see only reliable data

## Consequences

- The camera MAC is not shown at the moment.
- The hub MAC appears only after a successful ARP lookup.
- Model, firmware, MCU version, name or motion state appear only with a valid
  protocol reply.
- Tiles can carry different numbers of metadata rows depending on the device.
- Metadata rows grow taller for long values because the value wraps instead of
  being cut off.

## Addendum 2026-11-24

The obligation to show values in full applies to metadata fields (drawer,
diagnostics, hub details); they wrap. Tile labels may truncate with an ellipsis
because the tile has a fixed area: the full name sits in the details drawer.
Functional labels ("Camera 2", "Base station") remain exempt as before.
