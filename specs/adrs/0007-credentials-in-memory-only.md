# ADR-0007: Security code in memory only

- Status: Superseded by [ADR-0012](0012-browser-session-credential.md)
- Date: 2026-11-21

## Context

The device security code is required for AES challenge authentication. The
vendor app stores comparable values with weak obfuscation only.

## Decision

The gateway does not persist the security code. The browser asks for it during
authentication or stream start and sends it in the request. After the session
is established, only the derived session key is used.

## Reasoning

- No secret in Compose, the image, a database or a configuration file
- Smaller impact from a file access on the host
- A simple and transparent security boundary

## Alternatives considered

- **Environment variable:** often ends up in process and container metadata.
- **Configuration file:** convenient, but needs secret management and file
  permissions.
- **Encrypted database:** only moves the key problem.

## Consequences

- The user must enter the code for every new stream.
- A future convenient channel switch should keep the code in RAM at most for
  an active browser session.
- HTTP does not protect the request from capture on the LAN; TLS remains a
  recommended hardening step.
