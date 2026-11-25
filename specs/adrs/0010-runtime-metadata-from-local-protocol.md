# ADR-0010: Runtime metadata comes from the local protocol

- Status: Accepted
- Date: 2026-11-21

## Context

The vendor app shows variable stream bitrate and battery information per camera.
These values should be available without a cloud service.

## Decision

- Bitrate is computed from the H.264 payload bytes written to FFmpeg per second.
- Battery is read through IO-Control `40448` for the active channel.
- The read runs directly after stream start and then every 30 seconds.
- Status is served through `/api/live/status` and refreshed by the browser about
  once per second.

## Reasoning

- Matches the video data actually processed
- No platform-specific network interface measurement
- The battery format is verified from live captures and hardware replies
- No cloud or telemetry services needed

## Consequences

- The displayed bitrate excludes network overhead.
- The first bitrate can be `0` until the next one-second tick.
- Battery data can be missing until the first reply.
