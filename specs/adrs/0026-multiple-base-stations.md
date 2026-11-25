# ADR-0026: Mehrere Basisstationen gleichzeitig bedienen

- Status: Accepted
- Datum: 2026-11-22
- Partially superseded: [ADR-0027](0027-live-wall-layout.md)

## Kontext

Bisher hielt das Dashboard genau einen Workspace. Wurden die Kameras einer zweiten
Basisstation geladen, ersetzte deren Ansicht die erste. Wer zwei Hubs besitzt, konnte
damit nie beide gleichzeitig sehen, obwohl der Streammanager schon länger einen Stream
pro DID und Kanal zulässt (ADR-0011) und die API keine Obergrenze über DIDs hinweg
kennt. Der Wunsch ist, die Streams aller geladenen Basisstationen parallel abzuspielen.

## Entscheidung

- Jede geladene Basisstation erhält einen eigenen Workspace (`.workspace[data-hub-key]`)
  unterhalb von `#workspaces-section`. Das Laden einer weiteren Station ersetzt keinen
  bestehenden Workspace mehr; die Abschnitte bleiben untereinander sichtbar.
- Der Workspace-Kopf nennt die Basisstation (Modell · DID) und die Anzahl ihrer Kameras.
  Die Aktionen „Alle starten“/„Alle beenden“ bleiben pro Basisstation erhalten.
- Zusätzlich gibt es im Schritt-2-Kopf die globalen Aktionen „Alle Basisstationen
  starten“ und „Alle Basisstationen beenden“. Sie starten bzw. beenden alle Kanäle
  aller geladenen Hubs.
- Inventar- und Bewegungsaktionen sind gesperrt, solange **irgendein** Stream läuft –
  passend zur API, die bei aktiven Streams `409 conflict` liefert. Nach dem Stoppen
  werden Laden und Discovery wieder freigegeben.
- Sicherheitscodes bleiben je DID im Browser-RAM (ADR-0025). Die globalen Aktionen
  fragen fehlende Codes nacheinander ab (Single-Flight pro DID, ein Dialog zur Zeit).
- Sprachwechsel aktualisiert beide Workspaces in place, ohne Streams zu unterbrechen;
  Karten ohne Gerätenamen nutzen übersetzbare Fallback-Schlüssel.

## Betrachtete Alternativen

- **Reiter/Tabs pro Hub:** würde wieder exklusiv umschalten und widerspricht dem Ziel,
  alle Streams gleichzeitig zu sehen.
- **Eine gemeinsame Kamerawand ohne Hub-Gruppierung:** verliert die Zuordnung
  Kamera → Basisstation und damit die eindeutige Zuordnung von Code und Stream-ID.
- **Nur globale Aktionen:** verhindert, dass ein einzelner Hub gezielt gestartet wird.

## Konsequenzen

- `web/index.html`: statischer `#workspace`-Block ersetzt durch `#workspaces-section`
  (durch ADR-0027 weiter zur gemeinsamen Wand abgelöst)
  mit globalen Aktionen und `#workspaces`-Container.
- `web/app.js`: Workspace pro Hub, globale Start-/Stopp-Funktionen, Inventarsperre über
  alle Hubs, Prompt-Kette für mehrere Hubs, Fallback-Titel über i18n.
- `web/style.css`: Abstände und Scroll-Offset für gestapelte Workspaces.
- `web/locales/{en,de}.json`: `workspace.startAllHubs`, `workspace.stopAllHubs`,
  `notice.channelsLoaded` nennt den Hub.
- Tests: neue Playwright-Suite `workspaces.spec.ts` (zwei Hubs, Start/Stopp aller,
  Inventarsperre); Credentials- und Branding-Suite an die neue Struktur angepasst.
- Keine API-, OpenAPI- oder SBOM-Änderung: die API arbeitet bereits pro DID.

## Nachtrag 2026-11-23

ADR-0027 ersetzt die Workspace-Sektionen durch eine gemeinsame Live-Wand. Jeder
geladene Hub erscheint als Filterchip und im Hub-Manager, nicht mehr als eigener
`.workspace`-Abschnitt. Die Invarianten dieses ADR bleiben gültig:

- alle geladenen Hubs können gleichzeitig streamen,
- beim Laden eines weiteren Hubs wird kein bestehender Stream ersetzt,
- globale Aktionen starten/beenden alle Hubs, jeder Hub bleibt einzeln steuerbar,
- Inventar- und Bewegungsaktionen sind während aktiver Streams gesperrt,
- Codes bleiben je DID im Browser-RAM.

Die konkrete Wand-, Popover- und Mobile-Umsetzung steht in ADR-0027.

## Verweise

- ADR-0011: ein Stream pro DID und Kanal
- ADR-0025: Sicherheitscode je Basisstation
- `specs/product.md`, `specs/architecture.md`, `specs/testing.md`
