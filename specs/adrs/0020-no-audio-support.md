# ADR-0020: No audio support

- Status: Accepted
- Date: 2026-11-22

## Context

Audio reception would be technically possible. The verified session sequence
sends `StartVideo` and `StartAudio`; the hub can then deliver audio data next to
video data. At the same time the dashboard shows several coupled cameras in
parallel. With several streams running at once, a listener can no longer tell
which camera an audio stream belongs to. Audio becomes chaos instead of
information. Intercom is not planned. Browser autoplay rules and extra bandwidth
and CPU also argue against audio.

## Decision

- The gateway supports **no audio**: no audio reception, no browser audio output
  and no intercom.
- Incoming audio data is discarded. There are no audio controls in the interface.
- `StartAudio` stays part of the session setup for now because the verified
  sequence includes `StartVideo` + `StartAudio`. The command serves protocol
  compatibility only, not audio. Whether `StartVideo` alone is enough is a later
  hardware check.

## Reasoning

- With several parallel streams, audio cannot be assigned clearly to one camera.
  Playback would confuse users.
- Live video is the product purpose; audio adds no value for this scenario.
- Audio increases bandwidth, CPU load and interface complexity.
- Consistent audio support for a single stream only would be a special rule
  without a clear benefit.

## Alternatives considered

- **Full audio support:** rejected because several streams create audio chaos.
- **Audio only with exactly one active stream:** rejected because the behavior
  jumps with the number of streams and is hard to explain.
- **Intercom:** rejected, outside the product scope.
- **WebRTC audio:** rejected, extra signaling without product value.

## Consequences

- Audio disappears from the roadmap.
- `specs/product.md` lists audio reception, audio output and intercom as
  non-goals.
- `specs/protocol.md` still documents `StartAudio` but notes that the gateway
  does not evaluate it.
- The tests stay unchanged; the fake hub still expects `StartAudio`.
- Recordings are not part of the gateway. The base station records on its own
  when motion detection is active for the channel and an SD card is present.
  Access to those recordings is a roadmap item.

## References

- [ADR-0006](0006-h264-passthrough-and-hls.md): H.264 passthrough and HLS
- [ADR-0011](0011-concurrent-stream-per-channel.md): one stream per DID and
  channel in parallel
- `specs/protocol.md`
- `specs/product.md`
- `specs/roadmap.md`
