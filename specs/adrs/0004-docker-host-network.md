# ADR-0004: Docker host networking for local broadcasts

- Status: Accepted
- Date: 2026-11-21

## Context

WAI and PPPP use IPv4 broadcasts on UDP/7200 and UDP/32108. Docker bridge
networks do not reliably carry these broadcasts into the physical LAN.

## Decision

Docker Compose uses `network_mode: host` on Linux.

## Reasoning

- Direct, verified broadcast reachability
- No privileged raw sockets needed
- No manual port forwarding for dynamic UDP session ports
- Works on the existing NixOS Docker host

## Alternatives considered

- **Bridge plus UDP broadcast relay:** an extra component and configuration.
- **macvlan/ipvlan:** stronger isolation, but more complex and problematic for
  host-to-container communication.
- **`privileged` or raw capture:** unnecessary and less secure.

## Consequences

- The container has less network isolation.
- The HTTP port binds directly on the host.
- The Compose deployment is primarily Linux-specific.
