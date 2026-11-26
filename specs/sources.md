# Sources and provenance

## Primary sources

- base station firmware `1.0.0.12` available locally
- live interoperability tests against own hardware
- local protocol analysis of network traffic

Real DIDs, IP addresses, push tokens, security codes and camera images are not part
of the specification and must not enter public artifacts.

## Public references

- [magicus/pppp-dissector](https://github.com/magicus/pppp-dissector), MIT: PPPP
  message types and research documentation
- [nosoop-onlyslop/p2pcam](https://github.com/nosoop-onlyslop/p2pcam), Unlicense:
  independent Go reimplementation of a related Uniden/PPPP variant
- [DavidVentura/cam-reverse](https://github.com/DavidVentura/cam-reverse):
  independent PPPP session and streaming research
- [Fission-AI/OpenSpec](https://github.com/Fission-AI/OpenSpec), MIT: inspiration
  for the spec-driven document structure
- [Inter Tight](https://fonts.google.com/specimen/Inter+Tight), SIL Open Font
  License 1.1: font for brand name and step headings

The device-specific fields, battery commands, quality modes and tested runtime
behavior were validated against real hardware.

## Bundled third-party software

- `web/hls.min.js`, hls.js 1.5.18; license in `web/HLS-LICENSE.txt` (Apache-2.0)
- `web/icon-light.svg`, `web/icon-dark.svg`, camera icon from Font Awesome Free
  6.7.2 (solid); license in `web/FONTAWESOME-LICENSE.txt` (icons: CC BY 4.0)
- `web/fonts/inter-tight-latin.woff2`, `web/fonts/inter-tight-bold-latin.woff2`,
  Inter Tight (regular 400 and bold 700) from Google Fonts; license in
  `web/INTERTIGHT-LICENSE.txt` (SIL Open Font License 1.1)
- `web/swagger/*`, Swagger UI; license in `web/swagger/LICENSE` (Apache-2.0)
- FFmpeg from Alpine Linux is installed in the runtime image; license
  GPL-2.0-or-later (package `ffmpeg` in Alpine 3.24). The image ships the
  unmodified package; corresponding source is the Alpine `aports` package at
  <https://gitlab.alpinelinux.org/alpine/aports/-/tree/master/main/ffmpeg>
- Alpine Linux base system in the runtime image; package licenses in the SBOM
- Go standard library in the statically linked gateway binary; license BSD-3-Clause

The complete package list of the runtime image sits in the SBOM as CycloneDX and
SPDX under `sbom/`. `./scripts/sbom.sh` generates it from the built image. The static
web assets (hls.js, Font Awesome, Swagger UI) have no package-manager entry and are
therefore also documented in `sbom/README.md` and in the license dialog.
