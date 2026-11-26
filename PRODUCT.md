# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

- Primary: technically minded home-lab users who run Docker on a home server, NAS or Raspberry Pi and want their AkkuCam base station in a browser on the local network. They tolerate setup work and value reproducibility, SBOM and headless operation.
- Also served: owners of an AkkuCam PPIC90000/PPIC90010/PPIC90200 with a base station who want to replace the vendor app for live viewing.
- Also served: privacy-conscious operators who refuse vendor relay or cloud transport.

When two audiences conflict, home-lab users win (decided 2026-11-26).

## Product Purpose

Watch AkkuCam cameras through their base station in a normal browser on the LAN,
without the vendor app and without a cloud service. The gateway discovers the
base station locally, opens the PPCS session directly to it and hands the H.264
stream to the dashboard as HLS; FFmpeg only remuxes.

Success (decided 2026-11-26): other owners get it running on documented hardware,
and the project grows through contributions, more device support and integrations
such as Home Assistant or RTSP, without leaving the LAN-only, cloud-free scope.

Behavioral requirements and non-goals are specified in `specs/product.md`; this
record is the product truth for design work, not a second specification.

## Positioning

The local path is the product: WAI discovery and the PPCS session are
reimplemented from observed behavior against the device's own security code, with
no Oguard directory, relay or firmware endpoint and no vendor library. One Go
process plus a static dashboard ship in one container that keeps no state.

## Operating Context

- Docker on a 64-bit Linux host (x86-64 or Raspberry Pi 4 arm64); `network_mode:
  host` is required because discovery uses IPv4 broadcast.
- Base station on the same IPv4 subnet; free TCP 8080, UDP 7200 and UDP 32108.
- The security code is entered per base station in the browser and held in RAM;
  it never reaches disk, logs or browser storage.
- Verified installation: `SMH-T1R`, firmware `1.0.0.12`, cameras `PPIC90000`,
  `PPIC90010`, `PPIC90200`. Other models and firmware are untested.
- No database and no persistent state; HLS segments live in `/tmp/akkucam-hls`.
- German and English UI; browser with Media Source Extensions (Chromium and
  Firefox tested).

## Capabilities and Constraints

- Live view of all loaded base stations on one wall; one stream per DID and
  channel, several channels in parallel.
- Quality modes Full HD, HD, SD and automatic; H.264 passthrough with `-c:v copy`.
- Motion detection per camera as a verified read-modify-write; inventory and
  motion changes are rejected while a stream runs.
- Bitrate, FPS, battery and uptime only when the device reported them; no
  placeholders for unknown values.
- Out of scope by decision: cloud or relay transport, remote access over the
  internet, audio and two-way talk, recording or playback in the gateway, user
  and permission management, server-side API localization.
- No third-party Go dependencies and no frontend framework or build step;
  vendored assets only; a new dependency needs an ADR.
- The image runs unprivileged, without `NET_RAW` and without a Docker socket.

## Brand Commitments

- Name: "AkkuCam Gateway"; the brand name and large step headings use the locally
  bundled Inter Tight. No CDN, no external font service.
- Independent project. ABUS and AkkuCam are trademarks of their owners; no
  affiliation with or support by the manufacturer is claimed, and no vendor code
  is included.
- MIT license with a license dialog reachable from the info icon; no page footer
  and no separate license page.
- UI voice: German in Du-form, no implementation disclaimers in user-facing copy
  (ADR-0028).

## Evidence on Hand

- `specs/product.md` (required behavior and non-goals), `specs/protocol.md`
  (verified protocol fields), `specs/security.md`, `specs/testing.md`
  (hardware runs), ADRs under `specs/adrs/`.
- `docs/screenshots/` — wall, details and iPhone views with test data only
  (synthetic DID and IP).
- `sbom/` — CycloneDX and SPDX bill of materials generated from the runtime image.

Absent and must not be invented: user numbers, testimonials, benchmarks, press,
pricing or third-party reviews.

## Product Principles

1. Cloud-free is structural, not a setting: LAN discovery and a direct PPCS
   session only, with no fallback.
2. Secrets never persist: the device code stays in browser RAM and transient Go
   memory, and out of logs, files, storage and artifacts.
3. Show only observed data; omit an unknown value instead of labeling it.
4. The picture stays as the device delivers it; no re-encoding.
5. Home-lab fit: one stateless container, host networking, pinned toolchain and a
   reproducible SBOM.

## Accessibility & Inclusion

- WCAG AA text contrast (at least 4.5:1), 44 × 44 px icon controls, inline text
  actions at least 24 px high, keyboard operation, layouts tested at a 32 px root
  (ADR-0030).
- Language switching between German and English happens without reload and
  without interrupting running streams.
