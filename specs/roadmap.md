# Roadmap

## Next sensible steps

### Robustness

- session reconnect after a radio drop
- expose status reason and last error over the API
- measure HLS startup and camera wake-up separately
- add packet loss tests
- graceful shutdown on SIGTERM

### Security

- optional HTTP authentication
- documented reverse proxy example for TLS
- request size limits and explicit timeouts for body and response
- container image scan with Trivy on the built image, complementing the SBOM
  (phase 2 from ADR-0033)

### Usability

- high-resolution stills or a snapshot endpoint (the focus strip now uses live
  thumbnails from the video streams)
- access to recordings the base station produces (motion detection active, SD card
  present)
- better error status for an empty or disabled channel
- more languages through additional catalogs in `web/locales/` (ADR-0021)

### Media

- WebRTC as an optional low-latency alternative to HLS
- RTSP output for Home Assistant, go2rtc or NVRs
- SPS/PPS and resolution display in the status

### Device support

- test more firmware versions
- investigate the behavior of the second device family found (`WGAG`) separately
- query camera and hub capabilities dynamically

### Integrations

- **HomeKit (Apple Home) through Home Assistant or Homebridge.** No native HAP
  support in the gateway. RTSP output is a prerequisite; a still image can be
  generated from the RTSP stream. Details in
  [HomeKit (Apple Home)](#homekit-apple-home).

## HomeKit (Apple Home)

Reviewed, not implemented. The gateway stays HomeKit-free; integration goes through
a bridge. The Home Assistant HomeKit bridge or Homebridge with `camera-ffmpeg` are
recommended. Cameras run in accessory mode, one accessory per channel. A dedicated
snapshot endpoint is optional because both bridges can pull a frame from the RTSP
stream.

**Non-goals:** HomeKit Secure Video, iCloud recordings, two-way audio or intercom,
MFi certification and the "Works with Apple Home" badge, native HAP support in the
gateway.

**Limits:** H.264 passthrough is possible but must be tested with real hardware.
Motion notifications need a real event source; the gateway currently knows only the
on/off state of motion detection. Recordings stay on the base station SD card and do
not appear in Apple Home.

### User stories

- **US-HK-01: See the camera in Apple Home.** As an Apple Home user with Home
  Assistant, I want each gateway camera as its own accessory so I can open it live in
  the Home app. *Acceptance:* one accessory per channel; live view starts; H.264
  without transcoding where possible; the tile shows a still image.
- **US-HK-02: Two channels separately.** As a user with two coupled cameras, I want
  both channels as separate accessories so I can open them independently.
  *Acceptance:* two accessories; parallel streams; no mutual blocking.
- **US-HK-03: Still image in the tile.** As a user, I want a current still image so
  the tile does not stay black. *Acceptance:* snapshot from RTSP or a dedicated
  endpoint; JPEG; refresh at the rhythm Apple prescribes.
- **US-HK-04: Siri and scenes.** As a user, I want to open the camera through Siri or
  a Home scene. *Acceptance:* the accessory is selectable in scenes; the Siri call
  works as far as Apple supports it.
- **US-HK-05: Pairing survives restarts.** As a user, I do not want to pair again
  after a gateway restart. *Acceptance:* pairing data lives in HA or Homebridge; a
  gateway restart needs no re-pairing.
- **US-HK-06: No audio, no audio chaos.** As a user, I deliberately want no audio in
  HomeKit so several streams do not create audio chaos. *Acceptance:*
  `support_audio: false` or `audio: false`; no audio controls; ADR-0020 stays valid.
- **US-HK-07: Privacy stays local.** As a privacy-minded user, I want the integration
  to stay local and produce no cloud recordings. *Acceptance:* HKSV disabled; no
  iCloud recordings; the documentation says so.
- **US-HK-08: Recordings clearly separated.** As a user, I want to understand that
  recordings sit on the base station SD card and not in Apple Home. *Acceptance:*
  docs and feature status; HomeKit shows no recordings; access to SD recordings stays
  a separate roadmap item.
- **US-HK-09: Homebridge template.** As a Homebridge user, I want a ready-made config
  template so I can add the camera in a few minutes. *Acceptance:* documentation with
  RTSP URL, optional snapshot URL, `audio: false`, `vcodec: copy`, `maxStreams`.
- **US-HK-10: Home Assistant guide.** As a Home Assistant user, I want a guide for the
  HomeKit bridge in accessory mode. *Acceptance:* documentation with `mode:
  accessory`, `support_audio: false`, `video_codec: copy`, `stream_count`, one
  accessory per camera.
- **US-HK-11: Optional integration.** As a user without an Apple environment, I want
  the gateway to keep running unchanged. *Acceptance:* no new mandatory dependency in
  the container; health check unchanged; HomeKit docs separate.
- **US-HK-12: Motion notifications (optional).** As a user, I want optional motion
  notifications in Apple Home. *Acceptance:* offer them only when a real event source
  exists; `linked_motion_sensor` (HA) or a motion accessory (Homebridge); otherwise do
  not promise them.
- **US-HK-13: Pairing without HA or Homebridge (later).** As a user without Home
  Assistant or Homebridge, I want to pair the camera directly with Apple Home.
  *Acceptance:* HAP pairing, mDNS, SRTP, RTP, snapshot; extensive tests; legal
  clarification.
- **US-HK-14: Store pairing securely (later).** As a user, I want the HomeKit pairing
  to survive restarts without credentials reaching logs. *Acceptance:* pairing data
  persistent and protected; no secret in the log; its own ADR.

## Deliberately deferred

- internet relay and remote access
- firmware updates through the gateway
- admin settings and password changes
- cloud push messages

These features expand the attack surface and scope considerably and each needs its
own change proposal or ADR.
