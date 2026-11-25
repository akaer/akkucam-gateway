# ADR-0001: OpenSpec-inspirierte Dokumentation ohne CLI-Zwang

- Status: Accepted
- Datum: 2026-11-21

## Kontext

Das Projekt entstand iterativ aus Reverse Engineering. Entscheidungen, bestätigte Fakten und offene Annahmen dürfen nicht nur in Chatverläufen existieren. OpenSpec bietet einen sinnvollen Ansatz aus dauerhaften Specs und änderungsbezogenen Artefakten.

## Entscheidung

Das Projekt verwendet `specs/` mit dauerhaften Spezifikationen, nummerierten ADRs und bei Bedarf `specs/changes/<name>/`. Die Dokumente sind OpenSpec-inspiriert, benötigen aber keine installierte OpenSpec-CLI.

## Begründung

- Markdown bleibt unmittelbar lesbar
- keine zusätzliche Node-/Nix-/CLI-Abhängigkeit
- geeignet für den aktuellen kleinen Brownfield-Codebestand
- spätere Einführung des offiziellen OpenSpec-Workflows bleibt möglich

## Betrachtete Alternativen

- OpenSpec sofort installieren und initialisieren: erzeugt Agenten- und Workflowartefakte, die für den aktuellen Dokumentationsauftrag mehr Umfang als Nutzen bringen.
- nur ADRs: dokumentiert Entscheidungen, aber nicht den aktuellen Protokoll- und API-Vertrag.
- nur README: wird bei wachsendem Umfang unübersichtlich.

## Konsequenzen

- Maintainer müssen Specs manuell mit Codeänderungen synchron halten.
- künftige größere Features sollen vor Implementierung als Change mit Proposal, Design und Tasks angelegt werden.

## Nachtrag 2026-11-24: Nachträge und Archiv

Dieses ADR beschreibt die Dokumentationsstruktur. Zwei Regeln wurden präzisiert,
weil die Praxis sie bereits so handhabt:

- ADRs bleiben nach Annahme lesbar, dürfen aber datierte Abschnitte `## Nachtrag`
  erhalten. Ersetzt ein späteres ADR eine Entscheidung ganz oder in wesentlichen
  Teilen, wird der Status auf `Superseded` bzw. `Partially superseded` gesetzt und
  im Index geführt (`specs/adrs/README.md`).
- Umgesetzte Changes wandern von `specs/changes/<name>/` nach
  `specs/changes/archive/<name>/` und gelten dort als Referenz, nicht als offener
  Vorschlag.
