# ADR-0011: One active stream per DID and channel

- Status: Accepted
- Date: 2026-11-21
- Supersedes: [ADR-0009](0009-one-active-stream-per-did.md)

## Context

An SMH-T1R base station can serve several coupled battery cameras. The real
hardware reports its channels through IO-Control `40962` and can deliver
independent local PPCS sessions for channels 1 and 2 at the same time.

## Decision

The stream manager addresses streams as `<DID>-ch<channel number>`. Exactly one
stream is active per DID and channel; different channels of the same DID may run
in parallel. Restarting the same channel replaces only that channel.

## Reasoning

- Enables the dashboard and comparison view
- Keeps an unambiguous lifecycle per camera channel
- Still prevents duplicate sessions for the same channel
- Separate HLS paths, bitrates and battery values

## Consequences

- "Start all" establishes the channel sessions in controlled sequence and then
  lets them run in parallel.
- Radio and battery consumption grow with every active channel.
- Stop works per channel or as a batch.
- The number of tiles comes from the hub camera list, not from a fixed
  assumption.

## Clarification 2026-11-24

The client starts the channels of a hub in parallel (`Promise.all` in
`web/app.js`). The server enforces the controlled order: `POST /api/live/start`
runs under the same process mutex as probe and authentication (ADR-0008). The
session setups therefore run one after another, and the media streams then run
in parallel.
