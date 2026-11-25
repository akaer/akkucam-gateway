# Product specification

## Purpose

The gateway lets people watch AkkuCams with a base station in a normal web
browser, on the local network. It replaces the vendor app for live video and uses
only the local LAN for discovery, session setup, authentication and media
transfer.

## Target users

- owners of an AkkuCam PPIC90000/PPIC90010/PPIC90200 with a base station
- technically minded home network and home automation users
- operators who do not want the video stream carried over vendor relay servers

## Required behavior

1. The gateway **MUST** discover base stations locally.
2. The gateway **MUST** open a direct LAN PPCS session without the Oguard
   directory server.
3. The security code **MUST NOT** be stored persistently or logged. With several
   base stations, the browser **MUST** keep a separate code per base station (DID)
   in RAM so that different codes work at the same time.
4. The gateway **MUST** pass H.264 video through without quality-reducing
   re-encoding.
5. The web interface **MUST** be able to select channel 1 and channel 2.
6. The web interface **MUST** show the current payload bitrate and the camera
   battery level.
7. The container **MUST NOT** run privileged or require `NET_RAW`.
8. The gateway **MUST NOT** contact Oguard or vendor cloud endpoints.
9. The gateway **MUST** read the coupled channel count from the hub dynamically.
10. Different camera channels **MUST** be viewable at the same time.
11. Metadata **MUST** appear only when a concrete value exists from the hub, the
    camera or the local network table. Existing values **MUST** be fully
    readable; the interface **MUST NOT** truncate them with an ellipsis and
    **MUST** wrap them when needed.
12. Motion detection of a coupled channel **MUST** be switchable locally.
13. Inventory and settings requests **MUST** be rejected while streams are
    active.
14. The container image **MUST** provide a health check against the local gateway
    API.
15. The web interface **MUST** support German and English. It detects the browser
    language, falls back to English and switches without a reload. The choice is
    stored in localStorage; the security code never is.
16. License information **MUST** be reachable without a separate license page
    through the info/help icon as a scrollable dialog. The dialog lists the MIT
    project license and every third-party component in the runtime image with its
    license; operating system packages **MAY** be grouped as a distribution
    (ADR-0022), and the SBOM provides the itemized list. The API documentation
    **MUST** be linked there. A copyright line and a page footer are not
    required.
17. The project **MUST** provide a software bill of materials (SBOM) in CycloneDX
    and SPDX that can be generated from the runtime image.
18. The brand name and the large step headings **MUST** use the locally bundled
    Inter Tight font. An external font service or CDN **MUST NOT** be used. The
    hero text block is removed; the discover button stays. The setup view **MAY**
    also show "Back to the live wall".
19. An info icon at the far right of the top bar (right of the theme toggle)
    **MUST** open a modal help popup with the flow (find a base station, open the
    coupled cameras, start the live view without a cloud service) in the active
    language. The popup **MUST NOT** store data and **MUST** work without
    external resources.
20. The streams of **all** loaded base stations **MUST** be visible at the same
    time in one shared live wall. Loading another base station **MUST NOT**
    replace an existing stream. Global actions **MUST** start and stop all loaded
    hubs together; each base station stays individually controllable through the
    hub manager. The hub manager names model, DID, IP and, when determined, the
    MAC address.
21. After the first channel list loads, the live wall **MUST** replace the setup
    page. "Add base station" **MUST** open the setup page on purpose and **MUST**
    be blocked while streams are active.
22. The wall **MUST** offer the views auto matrix, focus and comfort. Auto and
    comfort **MUST** produce equal tile sizes; an odd count **MUST** leave a
    visible empty grid cell. The wall **MAY** scroll when there are too many
    cameras. All views start at the top.
23. A click on the video or tile **MUST** open a details drawer with quality,
    motion detection, start/stop and available metadata. Running streams **MUST**
    show bitrate, FPS and battery as an overlay in the bottom left with available
    values only. Uptime sits in the drawer. The overlay **MUST** be switchable in
    the wall settings (default: on).
24. On iPhone sizes, a compact wall bar with filter, global start/stop and a menu
    sheet **MUST** replace the full bar; iPad and desktop keep the full bar. The
    selected view and the visibility of the live overlay **MUST** be stored in
    localStorage and **MUST** be resettable through the settings.
25. View, filter and language changes **MUST NOT** interrupt running streams; the
    video element **MUST** survive the change.

## Current feature set

- WAI and PPCS LAN discovery
- full DID and model detection
- PUNCH/P2P_RDY session setup
- AES challenge authentication
- camera wake-up by starting the live view
- H.264 Annex-B reception, reassembly and HLS output
- Full HD, auto, HD and SD as requested quality modes
- dynamic hub camera list with names, models and available firmware and MCU
  versions
- locally read hub metadata and an optional hub MAC from the ARP table
- motion detection per camera as a verified read-modify-write
- parallel display of several coupled cameras
- brand name and step headings in locally bundled Inter Tight; hero text block
  removed
- info icon in the top bar opens a bilingual help popup with the flow
- security codes per base station in browser RAM, with "Remove code" per hub and
  "Remove all security codes from browser memory" (two clicks) globally
- shared live wall with hub filter chips and hub manager; global start/stop
  actions for all hubs, plus per hub
- full-width setup bar with empty state, hub grid, always visible hub metadata and
  diagnostics; no centered 1560 px container
- auto matrix, focus and comfort with equal tile sizes, empty grid cells and
  top-aligned layout
- theme-aware tile frames (light and soft in the light theme, subtle in the dark
  theme)
- details drawer per camera (right on desktop, bottom sheet on mobile) with
  quality, motion, start/stop and metadata
- live data overlay for running streams (KB/s, FPS, battery); available values
  only, switchable in the settings
- live thumbnails in the focus strip from the existing video elements
- compact iPhone wall bar with filter pill, global start/stop and menu sheet
- license dialog through the help popup; no page footer, no separate license page
- selected wall view and overlay visibility in localStorage, with "Reset to
  default"
- bitrate of the H.264 payload stream in the details drawer and the live data
  overlay (KB/s)
- battery capacity in the live data overlay; charging state additionally in the
  details drawer
- start, status and stop over the HTTP API

## Non-goals of the current state

- remote access over the internet
- Oguard relay or cloud integration
- user and permission management in the gateway
- recording or persistent storage of videos or snapshots in the gateway; the base
  station records on its own when motion detection is active for the channel and
  an SD card is present
- audio reception, audio output, intercom or event archive (see
  [ADR-0020](adrs/0020-no-audio-support.md))
- a complete replacement for every app setting
- simultaneous output of both camera channels in one PPCS session; each channel
  deliberately uses its own session
- server-side localization of the API; `error` stays English and the UI
  translates stable error codes

Access to the recordings the base station produces is not part of the current
state. It is a roadmap item.

## Quality definition

The gateway does not change the picture quality the device delivers. `Full HD`
requests device quality mode `1`. In the verified installation this mode delivers
H.264 Main Profile at 1920×1080 and 25 fps. Bitrate and picture quality still
depend on the scene, the radio link, the camera firmware and the camera's internal
encoding.
