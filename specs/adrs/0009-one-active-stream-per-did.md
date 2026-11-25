# ADR-0009: One active stream per base station DID

- Status: Superseded by [ADR-0011](0011-concurrent-stream-per-channel.md)
- Date: 2026-11-21

## Context

A base station can have several camera channels. Parallel sessions increase
radio, battery and protocol load. The main use case was a browser player with
channel switching.

## Decision

The stream manager allows exactly one active stream per DID. A new start stops
and replaces the existing stream. The selected camera index determines channel 1
or 2.

## Reasoning

- Prevents orphaned FFmpeg and PPCS processes
- Matches the UI model at the time
- Reduces load on the base station and the battery cameras
- One unambiguous HLS URL per DID

## Alternatives considered

- **Parallel stream per DID and channel:** useful for NVR or quad view, not
  needed at the time.
- **Channel switch within an existing session:** shorter switch time, but needs
  extra stop/change state logic.

## Consequences

- A channel switch rebuilds the session and asks for the security code again.
- Showing both cameras at once was not possible at the time.
