# ADR-0005: Native clean-room reimplementation of PPPP

- Status: Accepted
- Date: 2026-11-21

## Context

The vendor app ships `libPPCS_API.so` for Android/ARM only. The Docker host is
Linux/x86-64. The library is proprietary and tied to Android system libraries.

## Decision

The required local protocol subset is implemented natively in Go: discovery,
PUNCH, P2P_RDY, DRW, ACK, keepalive, authentication, IO-Control and media
reassembly.

## Reasoning

- Avoids Android and ARM emulation
- No proprietary binaries in the image
- Traceable and testable code
- Only the local protocol scope that is actually needed
- Live captures confirm the flow independently

## Alternatives considered

- **ARM library under QEMU/Bionic:** fragile and hard to distribute.
- **Foreign device as a bridge:** extra hardware and a continuing vendor
  dependency.
- **Full PPPP cloud implementation:** outside the local scope.

## Consequences

- The project owns retransmission, reassembly and compatibility.
- Further firmware variants may need protocol adjustments.
- Reverse-engineering provenance and third-party sources are documented.
