# ADR-0003: Local operation only, no cloud fallback

- Status: Accepted
- Date: 2026-11-21

## Context

The vendor app contacts Oguard directory, push, firmware and relay services.
The main goal is local browser access without vendor dependency.

## Decision

The gateway implements only direct LAN discovery and the LAN PPCS session. There
is no automatic or optional cloud or relay fallback.

## Reasoning

- LAN_SEARCH/PUNCH/P2P_RDY works entirely on the local network.
- Video stays in the home network.
- An outage or shutdown of external servers does not affect the gateway.
- Fewer privacy and security risks.
- Less protocol and operational scope.

## Alternatives considered

- **Use Oguard only for rendezvous:** technically simpler for remote access, but
  contradicts the cloud-free goal.
- **Own external relay server:** high extra security and operational effort.

## Consequences

- Use only within the same broadcast or routing domain, or with an explicit
  broadcast relay.
- No remote access without a separate VPN into the home network.
- The base station itself can still open cloud connections; a firewall outside
  the gateway must handle that.
