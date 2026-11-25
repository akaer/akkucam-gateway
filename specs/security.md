# Security and privacy

## Trust model

The current gateway targets a trusted home network. Anyone who can reach the HTTP
API can trigger discovery and session attempts. Before exposing it across VLAN or
internet boundaries, HTTP authentication or an authenticating reverse proxy is
mandatory.

## Security code

- The browser asks for it per base station when the camera list loads.
- It stays in a JavaScript variable until reload, tab close or explicit removal,
  addressed by the base station DID (one map, one code per hub).
- The browser sends it to the Go process for further channel starts, inventory
  requests and motion changes; the process holds it only for the matching session
  setup.
- A code the base station rejects (`unauthorized`) is discarded in the browser so
  the next attempt asks again.
- It is not stored in localStorage, sessionStorage, cookies, files, databases, HLS
  paths or logs.
- Theme preference and language choice may live in localStorage; they contain no
  credentials.
- The handler removes it from the request struct after processing.
- Go strings cannot be reliably wiped from memory; "RAM only" therefore means no
  cryptographically guaranteed memory erasure.
- HTTP is unencrypted on the LAN; an attacker who can capture traffic can see the
  code.

For higher requirements, use TLS through a local reverse proxy.

## Cloud freedom

The gateway sends no packets to the observed Oguard directory, relay, wake-up or
firmware endpoints. The local flow uses UDP/7200 and UDP/32108 plus the dynamic hub
port.

This statement covers the gateway only. The unchanged base station can still open
its own internet connections. Complete network isolation needs firewall rules on
the router, for example blocking outbound UDP/32100 and further internet access from
the hub IP. Such rules are deliberately not part of the container.

## Media

- Video payload is recognizable as H.264 in the local legacy protocol and is not
  end-to-end encrypted.
- Any participant with a sufficient LAN capture position can potentially observe
  video data.
- HLS is served unencrypted over HTTP.
- HLS files are temporary, but reachable by every HTTP client while the gateway
  runs.

## Container

Positive properties:

- the process runs as an unprivileged user
- no `privileged: true`
- no `CAP_NET_RAW`
- no Docker socket mounts
- static Go binary
- no persistent secret configuration
- the HTTP health check contains no device identifier or credentials

Accepted risks:

- `network_mode: host` removes Docker network isolation
- HTTP port 8080 binds to all host interfaces by default
- FFmpeg processes untrusted camera bitstreams
- Alpine and FFmpeg updates must be applied regularly

## Static analysis

`make lint` checks Go code, browser sources, scripts, Dockerfile and workflows in
pinned container images (ADR-0033). Blocking checks are `gosec`, `govulncheck`,
Semgrep with the project rules for cloud freedom and secret handling, plus
Shellcheck, Hadolint and Actionlint. The scan does not cover CVEs in the runtime
image; that image scan is noted as phase 2 in `specs/roadmap.md`.

## Recommended hardening

1. Run the gateway only in a trusted VLAN.
2. Limit port 8080 to local source networks through the host firewall.
3. Put a reverse proxy with TLS and authentication in front.
4. Block internet access from the base station when push and cloud features are
   dispensable.
5. Rebuild the container image regularly.
6. Keep real security codes out of issues, logs and version control.

## Known legacy protocol risks

- AES-128-ECB
- zero-padded security code as the AES key
- proprietary XOR obfuscation instead of modern transport encryption
- broadcast discovery without authentication
- device DID visible on the LAN

The gateway reproduces these mechanisms only for interoperability and does not treat
them as modern security controls.
