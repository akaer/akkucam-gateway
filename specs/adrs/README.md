# Architecture Decision Records

## Status

- `Proposed`: zur Diskussion
- `Accepted`: aktuell gültig
- `Deprecated`: noch vorhanden, aber nicht mehr empfohlen
- `Superseded`: durch ein späteres ADR ersetzt
- `Partially superseded`: in wesentlichen Teilen durch ein späteres ADR ersetzt; Ergänzungen stehen als datierter `Nachtrag` im Dokument
- `Rejected`: bewusst nicht umgesetzt

## Index

| ADR | Entscheidung | Status |
|---|---|---|
| [0001](0001-openspec-inspired-documentation.md) | OpenSpec-inspirierte Dokumentation ohne CLI-Zwang | Accepted |
| [0002](0002-use-go-for-the-native-gateway.md) | Go für den nativen Gateway | Accepted |
| [0003](0003-local-only-without-cloud-fallback.md) | ausschließlich lokaler Betrieb | Accepted |
| [0004](0004-docker-host-network.md) | Docker Hostnetwork für Broadcasts | Accepted |
| [0005](0005-clean-room-pppp-implementation.md) | native PPPP-Reimplementierung statt proprietärer Fremdbibliothek | Accepted |
| [0006](0006-h264-passthrough-and-hls.md) | H.264 Passthrough und HLS | Accepted |
| [0007](0007-credentials-in-memory-only.md) | Sicherheitscode nur im RAM | Superseded |
| [0008](0008-serialize-session-control.md) | Session-Kontrollpfade serialisieren | Accepted |
| [0009](0009-one-active-stream-per-did.md) | ein aktiver Stream pro Basisstation | Superseded |
| [0010](0010-runtime-metadata-from-local-protocol.md) | Bitrate und Akku lokal ermitteln | Accepted |
| [0011](0011-concurrent-stream-per-channel.md) | paralleler Stream pro DID und Kanal | Accepted |
| [0012](0012-browser-session-credential.md) | Code bis Reload im Browser-RAM | Accepted |
| [0013](0013-responsive-dashboard-and-theme.md) | responsives Dashboard und Theme | Accepted; geändert 2026-11-23 (ADR-0027) |
| [0014](0014-render-only-available-metadata.md) | nur verfügbare Metadaten anzeigen | Accepted |
| [0015](0015-progressive-disclosure-metadata-and-metrics.md) | Metadaten und Live-Metriken zuklappbar | Partially superseded (ADR-0031) |
| [0016](0016-stable-api-error-codes.md) | Stabile API-Fehlercodes und typisierte Fehler | Accepted |
| [0017](0017-openapi-api-contract.md) | OpenAPI als maschinenlesbarer API-Vertrag | Accepted |
| [0018](0018-test-and-code-quality-conventions.md) | Test- und Codequalitätskonventionen | Accepted |
| [0019](0019-test-structure-aaa-one-behavior.md) | Teststruktur: Arrange/Act/Assert und ein Verhalten pro Test | Accepted |
| [0020](0020-no-audio-support.md) | Kein Audio-Support | Accepted |
| [0021](0021-ui-internationalization.md) | Zweisprachige Browseroberfläche (Deutsch/Englisch) | Accepted; ergänzt 2026-11-24 (ADR-0028) |
| [0022](0022-app-icon-licenses-and-sbom.md) | App-Icon, Favicon, Lizenzhinweise und SBOM | Accepted; Lizenzhinweise ersetzt 2026-11-23 (ADR-0024/0027) |
| [0023](0023-inter-tight-brand-and-headings.md) | Inter Tight für Markenname und Schritt-Headings | Accepted; ergänzt 2026-11-24 (ADR-0029) |
| [0024](0024-help-popup.md) | Info-Icon und Hilfe-Popup im Topbar | Accepted; ergänzt 2026-11-23 (Lizenzdialog) |
| [0025](0025-per-hub-security-codes.md) | Sicherheitscode je Basisstation im Browser-RAM | Accepted |
| [0026](0026-multiple-base-stations.md) | Mehrere Basisstationen gleichzeitig bedienen | Partially superseded (ADR-0027) |
| [0027](0027-live-wall-layout.md) | Live-Wand, Ansichten, Drawer und mobile Steuerleiste | Accepted; geändert 2026-11-24 (ADR-0031) |
| [0028](0028-ui-copy-and-terminology.md) | UI-Texte, Terminologie und Fehlerwege | Accepted |
| [0029](0029-design-tokens-and-icons.md) | Design-Tokens und ein Icon-Satz | Accepted |
| [0030](0030-accessibility.md) | Barrierefreiheit der Oberfläche | Accepted |
| [0031](0031-live-metrics-overlay.md) | Live-Daten-Overlay und Sichtbarkeitseinstellung | Accepted |
| [0032](0032-versioning-and-branch-workflow.md) | Semantische Versionierung und Branch-Workflow | Accepted |
| [0033](0033-static-analysis-toolchain.md) | Statische Analyse im Toolchain-Container | Accepted |

## Vorlage

```markdown
# ADR-NNNN: Titel

- Status: Proposed
- Datum: YYYY-MM-DD

## Kontext
## Entscheidung
## Begründung
## Betrachtete Alternativen
## Konsequenzen
## Verweise
```
