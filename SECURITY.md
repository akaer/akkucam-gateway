# Security policy

## Supported versions

The project is in the `0.2.x` release line. Only the latest release and the
current `main` branch receive fixes.

## Reporting a vulnerability

Do not open a public issue for a security problem. Use GitHub's private
vulnerability reporting: open the repository's **Security** tab and choose
**Report a vulnerability**. If that is not available, contact the maintainer
through GitHub ([@akaer](https://github.com/akaer)) to agree on a private
channel.

Include the affected version or commit, the base station model and firmware,
reproduction steps and the impact you expect. Never include real security
codes, DIDs, local IP addresses, packet captures or tokens. The gateway keeps
the device security code in memory only; a report does not need it.

## Scope

In scope: the gateway, the dashboard and the protocol implementation in this
repository. Out of scope: the base station firmware, the vendor app and
anything outside the gateway's own network path.

## Disclosure

Reports are handled on a best-effort basis; there is no service-level
agreement. Fixes are prepared privately and released together with an
advisory. Credit is given on request.
