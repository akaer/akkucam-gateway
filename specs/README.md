# Gateway-Spezifikationen

Dieser Ordner ist die technische und produktbezogene Source of Truth für den AkkuCam Gateway. Die Struktur ist von [OpenSpec](https://github.com/Fission-AI/OpenSpec) inspiriert, bleibt aber absichtlich ohne zusätzliche CLI oder Build-Abhängigkeit nutzbar.

## Dokumenttypen

### Aktuelle Spezifikationen

| Dokument | Inhalt |
|---|---|
| [product.md](product.md) | Ziel, Nutzerwert, Scope und Anforderungen |
| [architecture.md](architecture.md) | Systemkontext, Komponenten und Datenfluss |
| [protocol.md](protocol.md) | Verifizierter Stand des lokalen PPCS-Protokolls |
| [api.md](api.md) | HTTP-API und Laufzeitverhalten |
| [security.md](security.md) | Vertrauensgrenzen, Secrets und Restrisiken |
| [testing.md](testing.md) | Teststrategie und verifizierte Hardware |
| [roadmap.md](roadmap.md) | Offene Themen und geplante Entwicklung |
| [sources.md](sources.md) | Reverse-Engineering-Quellen und Drittsoftware |

Der maschinenlesbare API-Vertrag liegt bewusst außerhalb dieses Ordners in [`internal/openapi/openapi.json`](../internal/openapi/openapi.json), weil er per `go:embed` in die Anwendung eingebettet wird. `api.md` beschreibt denselben Vertrag für Menschen.

### Architecture Decision Records

Die Entscheidungen unter [adrs/](adrs/) erklären nicht nur **was** implementiert wurde, sondern auch **warum**, welche Alternativen betrachtet wurden und welche Konsequenzen akzeptiert werden.

ADRs sind nach Annahme historische Dokumente. Korrekturen und Ergänzungen werden als datierter Abschnitt `## Nachtrag` angehängt; die ursprüngliche Entscheidung bleibt lesbar. Wird eine Entscheidung ganz oder in wesentlichen Teilen ersetzt, erhält sie den Status `Superseded` oder `Partially superseded` mit Verweis auf das neue ADR. Der Status steht im Kopf des ADR und in der Statusspalte von [`adrs/README.md`](adrs/README.md).

## OpenSpec-Bezug

OpenSpec trennt dauerhafte Spezifikationen von geplanten Änderungen. Für dieses Projekt gilt analog:

```text
specs/
├── *.md                 # aktueller, implementierter Stand
├── adrs/                # dauerhaft nachvollziehbare Entscheidungen
└── changes/             # Änderungsartefakte
    ├── <change-name>/   # geplant, noch nicht umgesetzt
    │   ├── proposal.md
    │   ├── design.md
    │   └── tasks.md
    └── archive/<change-name>/  # umgesetzt und archiviert (Referenz)
```

Der Ordner `changes/` wird erst angelegt, wenn eine noch nicht implementierte Änderung konkret geplant wird. Nach der Umsetzung wandert der Change nach `changes/archive/<change-name>/`; dort bleibt er als Referenz erhalten, gilt aber nicht mehr als offener Vorschlag. Bereits implementierte Erkenntnisse werden nicht nachträglich als offene Changes dargestellt.

## Pflege

Eine Codeänderung muss die Spezifikation aktualisieren, wenn sie mindestens einen dieser Punkte verändert:

- öffentliches API oder UI-Verhalten
- Netzwerkprotokoll oder Paketformat
- Speicherung oder Verarbeitung von Zugangsdaten
- Deployment-, Container- oder Netzwerkmodell
- Medienformat, Qualität, Latenz oder Metadaten
- unterstützte Geräte oder Firmwarevarianten
- eine bestehende Architekturentscheidung
