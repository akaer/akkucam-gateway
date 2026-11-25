# ADR-0008: Control sessions are serialized

- Status: Accepted
- Date: 2026-11-21

## Context

Parallel probe and authentication attempts caused sporadic LAN PUNCH timeouts.
The base station and broadcast replies cannot be mapped clearly enough when
sessions compete.

## Decision

A process mutex serializes session probe, authentication and the initial live
session setup. The running media stream then works independently.

## Reasoning

- Removes the observed race conditions
- Very small functional restriction for one base station
- A simple, traceable failure model

## Alternatives considered

- **Map by dynamic UDP port and DID only:** more complex and still dependent on
  broadcast timing.
- **Fully parallel session managers:** relevant only with several independent
  hubs.

## Consequences

- Concurrent new connection attempts wait for each other.
- A slow timeout can delay other control calls.
- Existing video transmission is not blocked.
