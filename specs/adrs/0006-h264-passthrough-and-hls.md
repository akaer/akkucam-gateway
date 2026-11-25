# ADR-0006: H.264 passthrough and HLS for browser playback

- Status: Accepted
- Date: 2026-11-21

## Context

The camera delivers H.264 Annex-B. Browsers cannot play this elementary stream
directly from a normal HTTP URL.

## Decision

The gateway writes H.264 unchanged to FFmpeg. FFmpeg sets PTS/DTS and remuxes
with `-c:v copy` into local HLS/MPEG-TS. The browser uses a locally bundled
hls.js.

## Reasoning

- No quality loss from re-encoding
- Low CPU load
- Broad browser support
- Simple HTTP delivery
- FFmpeg already handles the packaging

The `setts` bitstream filter sets monotonic timestamps because channel 2 in
particular sometimes delivers no usable PTS/DTS.

## Alternatives considered

- **WebRTC:** lower latency, but far more complex signaling and packaging.
- **MJPEG:** high bandwidth and forced decoding or re-encoding.
- **RTSP:** good for NVRs, but not directly browser-capable.
- **Transcoded HLS:** unnecessary quality loss and CPU demand.

## Consequences

- HLS segmentation adds several seconds of latency.
- FFmpeg becomes part of the runtime and the attack surface.
- HLS files are temporary and deleted continuously.
