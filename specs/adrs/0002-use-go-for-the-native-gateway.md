# ADR-0002: Go for the native gateway

- Status: Accepted
- Date: 2026-11-21

## Context

The gateway needs UDP broadcasts, a binary protocol state machine, an HTTP API,
concurrency and process control for FFmpeg. The host runs NixOS/x86-64, and the
gateway should stay portable as a Docker container.

## Decision

The gateway is implemented in Go. A multi-stage Docker build produces a static
Linux binary.

## Reasoning

- The standard library covers UDP, HTTP, AES and process control.
- Goroutines fit discovery, the session loop, bitrate measurement and the
  FFmpeg lifecycle.
- A static binary keeps the runtime image small.
- One or two camera streams do not need heavier optimization.
- Fast iteration matters more during reverse engineering than maximum
  low-level control.

## Alternatives considered

### Rust

Rust offers memory-safe binary parsing and zero-copy. Deferred because the
bottleneck is protocol understanding, not CPU or memory. An isolated port of
`internal/ppcs` later remains possible.

### Python/Node.js

Fast for prototypes, but less attractive for a compact, long-running network
daemon and a minimal runtime image.

### Vendor library as a bridge

Would reuse a proprietary vendor library, but needs ARM emulation, more
privileges and carries distribution and license risks.

## Consequences

- The host needs no locally installed Go.
- The Dockerfile fixes the Go version and the build.
- Binary parsers still need defensive implementation and tests.
