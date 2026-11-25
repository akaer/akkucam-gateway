# ADR-0025: Sicherheitscode je Basisstation im Browser-RAM

- Status: Accepted
- Datum: 2026-11-22
- Ergänzt: [ADR-0012](0012-browser-session-credential.md)

## Kontext

ADR-0012 ging von einer einzelnen Basisstation aus und hielt genau einen Sicherheitscode in einer JavaScript-Variablen. Sobald zwei Basisstationen mit unterschiedlichen Codes im Netz stehen, entsperrt die Oberfläche beide mit demselben Code. Die zweite Basisstation lehnt die Anmeldung ab (`unauthorized`), und der Nutzer kann sie ohne Neuladen der Seite nicht mehr öffnen.

## Entscheidung

- Sicherheitscodes liegen im Browser-RAM in einer `Map`, adressiert über die DID der Basisstation (`state.credentials`). Die DID ist über Discovery-Läufe stabil; fehlt sie, dient die IP als Schlüssel.
- Der Code-Dialog nennt die Basisstation (DID/Modell), damit klar ist, welcher Code erwartet wird. Das Eingabefeld bleibt bei jedem Öffnen leer.
- Jede Hub-Karte zeigt einen Chip „Code gespeichert“ mit einem „Code entfernen“-Knopf, sobald für diese DID ein Code im RAM liegt. Der Chip erscheint nur bei vorhandenem Code.
- Der Knopf in den Diagnosen löscht **alle** Codes (`state.credentials.clear()`).
- Antwortet die API mit `unauthorized`, wird der Code der betroffenen DID verworfen. Der nächste Versuch fragt den Code erneut ab; es gibt keinen automatischen Wiederholungsversuch.
- „Alle starten“ und parallele Kanalaktionen teilen sich einen Prompt pro DID (Single-Flight), damit nicht mehrere Dialoge aufgehen.
- Es bleibt bei ADR-0012: keine Speicherung in `localStorage`, `sessionStorage`, Cookies, IndexedDB oder auf dem Server. Ein Reload leert die Map.

## Begründung

- **Sicherheit:** Der Code bleibt flüchtig im Browser-RAM; die Zuordnung zu einer DID ist reine Laufzeitinformation.
- **Bedienbarkeit:** Mehrere Basisstationen funktionieren ohne wiederholtes Neuladen; einmal eingegebene Codes werden je Station wiederverwendet.
- **Fehlerrobustheit:** Ein abgelehnter Code blockiert die Station nicht dauerhaft, weil er nach `unauthorized` verworfen wird.

## Konsequenzen

- `web/app.js`: `state.credentials` ersetzt `state.securityCode`; `requestSecurityCode(hub)` mit Single-Flight; `forgetCredential`; `unauthorized` verwirft den Code.
- `web/index.html`: Hub-Name im Code-Dialog (`#code-dialog-hub`), Chip je Hub-Karte, Diagnose-Knopf „Alle Sicherheitscodes aus dem Browserspeicher entfernen“.
- `web/style.css`: Layout für `.hub-actions`, `.hub-credential`, `.dialog-hub`.
- `web/locales/{en,de}.json`: `dialog.hub`, `hub.codeSaved`, `hub.forgetCode`, `notice.forgotAllCodes`; `notice.forgotCode` und `dialog.hint` angepasst.
- Keine API-, OpenAPI- oder SBOM-Änderung: der Server erhält den Code weiterhin pro Request.
- Neue Tests: `web/e2e/credentials.spec.ts` (getrennte Codes, Wiederverwendung, Verwerfen nach 401, Code entfernen je Hub).

## Verweise

- ADR-0007: keine Zugangsdaten auf Platte
- ADR-0012: Sicherheitscode bis zum Reload im Browser-RAM
- ADR-0016: stabile API-Fehlercodes (`unauthorized`)
- `specs/security.md`, `specs/product.md`, `specs/testing.md`

## Nachtrag 2026-11-24

Der globale Knopf „Alle Sicherheitscodes aus dem Browserspeicher entfernen“
existiert nach ADR-0027 an zwei Stellen: im Diagnosebereich der Setup-Seite
(`#forget-code`) und im Hub-Manager der Wand (`#forget-all-codes`). Beide rufen
denselben Handler (`forgetAllCredentials`) auf und löschen die gesamte Map.

## Nachtrag 2026-11-24

Der globale Knopf heißt in der Oberfläche „Alle Sicherheitscodes aus dem
Browser-Arbeitsspeicher entfernen“ und benötigt zwei Klicks: Der erste armiert
die Aktion und ändert die Beschriftung, der zweite entfernt die Codes
(ADR-0028). Die Formulierung „Browserspeicher“ entfällt, weil sie auch
`localStorage` bezeichnen könnte.
