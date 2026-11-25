# Quellen und Provenienz

## Primärquellen

- lokal vorliegende Basisstationsfirmware `1.0.0.12`
- Live-Interoperabilitätstests gegen die eigene Hardware
- lokale Protokollanalyse des Netzwerkverkehrs

Reale DIDs, IP-Adressen, Push-Tokens, Sicherheitscodes und Kamerabilder sind keine Spezifikationsbestandteile und dürfen nicht in öffentliche Artefakte übernommen werden.

## Öffentliche Referenzen

- [magicus/pppp-dissector](https://github.com/magicus/pppp-dissector), MIT — PPPP-Nachrichtentypen und Forschungsdokumentation
- [nosoop-onlyslop/p2pcam](https://github.com/nosoop-onlyslop/p2pcam), Unlicense — unabhängige Go-Reimplementierung einer verwandten Uniden/PPPP-Variante
- [DavidVentura/cam-reverse](https://github.com/DavidVentura/cam-reverse) — unabhängige PPPP-Session- und Streamingforschung
- [Fission-AI/OpenSpec](https://github.com/Fission-AI/OpenSpec), MIT — Inspiration für spec-driven Dokumentstruktur
- [Inter Tight](https://fonts.google.com/specimen/Inter+Tight), SIL Open Font License 1.1 — Schrift für Markenname und Schritt-Headings

Die gerätespezifischen Felder, Batteriekommandos, Qualitätsmodi und das getestete Laufzeitverhalten wurden gegen reale Hardware validiert.

## Gebündelte Drittsoftware

- `web/hls.min.js`, hls.js 1.5.18; Lizenz in `web/HLS-LICENSE.txt` (Apache-2.0)
- `web/icon-light.svg`, `web/icon-dark.svg`, Kamera-Icon aus Font Awesome Free 6.7.2 (Solid); Lizenz in `web/FONTAWESOME-LICENSE.txt` (Icons: CC BY 4.0)
- `web/fonts/inter-tight-latin.woff2`, `web/fonts/inter-tight-bold-latin.woff2`, Inter Tight (Regular 400 und Bold 700) von Google Fonts; Lizenz in `web/INTERTIGHT-LICENSE.txt` (SIL Open Font License 1.1)
- `web/swagger/*`, Swagger UI; Lizenz in `web/swagger/LICENSE` (Apache-2.0)
- FFmpeg aus Alpine Linux wird im Runtime-Image installiert; Lizenz GPL-2.0-or-later (Paket `ffmpeg` in Alpine 3.21)
- Alpine-Linux-Basissystem im Runtime-Image; Paketlizenzen siehe SBOM
- Go-Standardbibliothek im statisch gelinkten Gateway-Binary; Lizenz BSD-3-Clause

Die vollständige Paketliste des Laufzeitimages liegt als SBOM in CycloneDX und SPDX unter `sbom/`. Sie wird mit `./scripts/sbom.sh` aus dem gebauten Image erzeugt. Die statischen Web-Assets (hls.js, Font Awesome, Swagger UI) haben keinen Paketmanager-Eintrag und sind deshalb zusätzlich in `sbom/README.md` und im Lizenzdialog dokumentiert.
