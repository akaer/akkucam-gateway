# Teststrategie

## Teststufen

### Unit-Tests

Werden im Docker-Build ausgeführt:

```bash
go test ./...
```

Aktuell abgedeckt:

- vollständige DID-Dekodierung aus PPPP-PUNCH und Discovery
- malformed/truncated Header und Längenfelder
- DRW-Medienfragmentparser
- Fragmentreihenfolge einschließlich 16-Bit-Wraparound
- App-Frame-Reassembly und Resynchronisierung nach beschädigten Bytes
- H.264-NAL-Erkennung
- Sicherheitscode-Grenzen vor Netzwerkzugriff
- Streamstatus, defensive Metadatenkopie und Bitratenzähler
- i18n-Kataloge: gleiche Schlüssel, gleiche Platzhalter, keine leeren Werte, alle `data-i18n`-Schlüssel aus `index.html`, alle in `app.js` verwendeten Schlüssel und alle API-Fehlercodes
- Hub-Geräteinfo, formatierte Firmware-/MCU-Versionen und ausgelassene nicht verfügbare Felder

### Build-Test

```bash
docker compose build
```

Der Multi-Stage-Build muss ohne lokal installiertes Go funktionieren.

### API-Contract-Tests

Die HTTP-Tests verwenden Fake-Abhängigkeiten und prüfen den Vertrag statt realer Netzwerkdetails:

- Health-Schema und Content-Type
- erlaubte Methoden sowie `405`/`Allow`
- unbekannte API-Pfade als JSON-404 statt Static-File-Fallback
- Kanal- und Qualitätsgrenzen
- Sicherheitscode wird nicht in Antworten gespiegelt
- explizite Ablehnung wird nicht wiederholt; Transportfehler werden begrenzt wiederholt
- Session-Kontrollaufrufe sind serialisiert
- Statusschema und idempotentes Stoppen
- `409 Conflict` für Inventar und Motion während aktiver Streams
- explizites `enabled:false` im Motion-Vertrag
- stabile Fehlercodes in `error`-Antworten
- unbekannte Fehler als `internal_error` mit HTTP 500 und generischer Meldung
- OpenAPI-Vertrag deckt jede registrierte Route ab

### Testlayout

- Unit- und Integrationstests liegen neben dem getesteten Paket. Dadurch bleiben unexportierte Parser und Builder testbar.
- API-Contract-Tests liegen in `internal/gateway`.
- Hardwaretests verwenden `//go:build hardware`.
- `tests/` ist ausschließlich für paketübergreifende Go-Tests vorgesehen. Die Konventionen stehen in `tests/README.md`.
- Browser-E2E-Tests liegen in `web/e2e/` (Playwright) und werden über `scripts/test-e2e.sh` ausgeführt.
- Neue Testdateien folgen der Namenskonvention `<thema>_test.go` bzw. `<thema>_integration_test.go`.

### Teststruktur (AAA)

Jeder Test folgt **Arrange/Act/Assert**. Die Phasen werden mit Kommentaren markiert, wenn sie nicht offensichtlich sind. Ein Test prüft **ein Verhalten**; mehrere Assertions sind erlaubt, wenn sie dasselbe Ergebnis beschreiben. Varianten laufen als Tabellen-Subtests. Testnamen enthalten kein „And“. Integrationstests und Hardwaretests werden in getrennte Sessions aufgeteilt, auch wenn das Laufzeit kostet. Fuzz-Targets sind die einzige Ausnahme. Siehe ADR-0019.

### Deterministische Protokollintegration

`internal/ppcs/protocol_integration_test.go` startet eine künstliche UDP-Basisstation. Deren Pakete werden separat nach der Spezifikation konstruiert, nicht mit den Produktionsbuildern. Geprüft werden:

- exaktes LAN_SEARCH und bytegenaues PUNCH-Echo
- P2P_RDY
- AES-Challenge und Session-Key
- explizite Ablehnung eines falschen Codes
- StartVideo und StartAudio mit gewünschtem Kanal/Qualitätsmodus
- dynamische Kameraliste über IO-Control 40962
- Hub-Geräteinfo über IO-Control 5/6
- Motion Read-Modify-Write-Verify über IO-Control 40964/40963
- Batterieanfrage und Batterieantwort
- H.264 Annex-B vom Fake-Hub bis zum Output-Writer
- DRW-ACK-/Sessionverhalten im realen UDP-Loop

### Fuzz-Seeds

Go-Fuzz-Targets stellen sicher, dass Header-, PUNCH-, Medien- und App-Frame-Parser bei beliebigen Bytefolgen nicht paniken. Die Seed-Cases laufen mit jedem normalen `go test`; längere Fuzzläufe können separat gestartet werden.

### Browser-E2E (Playwright)

`web/e2e/` enthält Playwright-Tests. Sie laufen im offiziellen Image `mcr.microsoft.com/playwright:v1.49.1-jammy` über `scripts/test-e2e.sh`; eine lokale Node-Installation ist nicht nötig. Jede Suite läuft in Chromium und Firefox: Projektnamen ohne Suffix sind Chromium, `-firefox` ist Firefox. Das Skript startet den Container mit `--network host`, prüft vorher `/api/health` und übergibt Basis-URL sowie optionale Hardware-Parameter als Umgebungsvariablen.

**I18n-Suite (deterministisch, kein Gerät):**

- Browser `de-DE` startet auf Deutsch, `fr-FR` fällt auf Englisch zurück
- `?lang=de|en` überschreibt die Erkennung
- Sprachwechsel ohne Reload (`performance.getEntriesByType('navigation').length === 1`)
- Persistenz über Reload via `localStorage['akkucam-lang']`
- Pluralformen folgen der aktiven Sprache
- API-Fehlercodes werden übersetzt
- `<html lang>` und `aria-label` folgen der Sprache
- Hub-Aktionen wechseln mit der Sprache („Kameras neu laden“, „Code entfernen“)
- das Hilfe-Popup öffnet mit dem lokalisierten Hilfetext und schließt über „Verstanden“/„Got it“

**Sicherheitscode-Suite (deterministisch, gemockte API):**

- zwei Basisstationen mit unterschiedlichen Codes werden getrennt entsperrt
- derselbe Hub verwendet den gespeicherten Code ohne erneuten Dialog
- ein abgelehnter Code (`unauthorized`) wird verworfen und erneut abgefragt
- „Code entfernen“ entfernt nur den Code der gewählten Basisstation
- „Alle Sicherheitscodes entfernen“ benötigt zwei Klicks; der erste armiert nur
- der Sichtbarkeitsschalter zeigt den eingegebenen Code, ohne ihn zu verändern, und startet nach erneutem Öffnen maskiert

**Mehrbasisstationen-Suite (deterministisch, gemockte API):**

- zwei geladene Hubs erscheinen gemeinsam in der Live-Wand; keine exklusive Umschaltung
- Filterchips blenden einen Hub aus, ohne seine Streams zu beenden
- „Alle Basisstationen starten“ startet Kanäle auf allen geladenen Hubs
- Inventar- und Discovery-Aktionen sind gesperrt, solange ein Stream läuft
- „Alle Basisstationen beenden“ beendet alle Streams
- Hub-Metadaten sind ohne Aufklappen sichtbar; DID steht im Kartenkopf, MAC im Raster und im Hub-Manager, lange Werte brechen um statt abgeschnitten zu werden

**Live-Wand-Suite (deterministisch, gemockte API):**

- die Setup-Seite nutzt eine vollbreite Setup-Leiste und einen Empty State; nach dem Laden ersetzt die Wand die Setup-Seite, ein Footer existiert nicht
- Kacheln beginnen am oberen Rand und behalten bei ungerader Kameraanzahl eine leere Rasterzelle
- Auto, Fokus und Komfort wechseln, ohne das Videoelement zu ersetzen
- der Fokus-Streifen zeigt Canvas-Miniaturen für laufende Streams
- das Live-Daten-Overlay zeigt kbit/s, FPS und Akku und lässt den Klick zur Seitenleiste durch
- das Hub-Kürzel fehlt im Video von Auto und Fokus und steht in Komfort
- der Schalter „Live-Daten auf Kacheln“ blendet das Overlay aus, speichert nur `off` und wird vom Reset zurückgesetzt
- Kacheln sind `role="group"` mit `aria-label`; Enter öffnet den Details-Drawer
- ein Klick auf den Backdrop schließt Einstellungen und Drawer
- bei 200 % Textskalierung bleiben Kachel-Leisten und Seite ohne Überlauf
- der Details-Drawer öffnet per Kachelklick
- die gewählte Ansicht wird in LocalStorage gespeichert und über die Einstellungen zurückgesetzt
- die kompakte iPhone-Leiste öffnet Darstellung, Filter und Einstellungen
- ein Sprachwechsel aktualisiert die Wand, ohne den laufenden Stream zu unterbrechen

**Branding- und Lizenz-Suite (deterministisch, kein Gerät):**

- Markenzeichen rendert das Kamera-Icon als Inline-SVG
- Favicon wechselt mit dem Theme (`icon-light.svg` / `icon-dark.svg`)
- Dashboard und Wand haben keinen Footer und keine Copyright-Zeile
- das Hilfe-Popup öffnet den scrollbaren Lizenzdialog mit MIT-Projektlizenz und hls.js, Font Awesome, Inter Tight, Swagger UI, FFmpeg, Alpine und Go
- der Lizenzdialog ist auf Deutsch übersetzt und verlinkt die API-Dokumentation
- Markenname, Setup- und Wand-Headings nutzen Inter Tight; der Markenname wird fett dargestellt
- Lizenzdateien, Schriftdateien und Theme-Icons werden vom statischen Dateiserver ausgeliefert
- Primärbutton und Status-Text erreichen in beiden Themes 4,5:1 (gemessen im Browser)
- der Theme-Schalter kehrt zur Systemeinstellung zurück und löscht dabei den LocalStorage-Eintrag
- alle geprüften Textstellen nutzen ausschließlich die gelieferten Gewichte 400/700

**SBOM-Prüfung:**

- `tests/sbom_test.go` validiert die committeten CycloneDX- und SPDX-Dateien
- `internal/gateway/static_test.go` prüft, dass alle im Dashboard referenzierten statischen Assets existieren
- CI erzeugt die SBOM mit `scripts/sbom.sh` neu und lädt sie als Artefakt hoch

**README-Abgleich:**

- `tests/readme_test.go` prüft in `README.md` und `README_de.md` die gleiche Abschnittsstruktur, die gleichen relativen Linkziele und die gleichen Codeblock-Typen
- die Übersetzung selbst bleibt Review-Aufgabe; der Test findet fehlende oder zusätzliche Abschnitte, Links und Befehle

**Live-Suite (optional, echte Hardware):**

- Discovery, Sicherheitscode und Kameraliste gegen die reale Basisstation
- Sprachwechsel bei geladener Kameraliste ohne Reload
- optional Streamstart und -stopp (`AKKUCAM_TEST_LIVE_STREAM=1`)
- keine Artefakte (Trace, Video, Screenshot) im Live-Projekt, damit der Sicherheitscode nicht auf Platte landet

```bash
make test-e2e
# optional gegen echte Hardware:
AKKUCAM_TEST_SECURITY_CODE='<code>' AKKUCAM_TEST_LIVE_STREAM=1 ./scripts/test-e2e.sh --project=live
```

Der Sicherheitscode wird ausschließlich als Umgebungsvariable übergeben. Die Live-Suite schreibt bewusst keine Traces, Videos oder Screenshots.

### Statische Analyse

`./scripts/lint.sh` (beziehungsweise `make lint`) prüft Go-Code, Browserquellen, Skripte, Dockerfile und Workflows in gepinnten Container-Images; eine lokale Installation ist nicht nötig. Details und Begründungen stehen in ADR-0033.

- `golangci-lint` mit `default: standard`, `gosec`, `errorlint`, `noctx`, `bodyclose` und `misspell`; Konfiguration in `.golangci.yml`
- `govulncheck` im Go-Image der Build-Stage, also auf derselben Standardbibliothek wie das Binary
- `semgrep` mit `p/golang`, `p/javascript`, `p/secrets` und den Projektregeln in `.semgrep/gateway.yml`: keine ausgehenden HTTP-/TCP-Clients, kein Sicherheitscode in `localStorage`/`sessionStorage`
- `shellcheck`, `hadolint`, `actionlint`

Alle Werkzeuge blockieren. Ausnahmen stehen nur in `.golangci.yml` und sind dort begründet: G115 im Protokollpaket, G304/G306 in Tests. Jedes `//nolint` trägt Grund und Verweis. Semgrep lädt die Registry-Regelsets zur Laufzeit und sendet keine Metriken.

Grenzen: Semgrep prüft Muster, keine Datenflüsse über Funktionsgrenzen; die Storage-Regel erkennt einen generischen Schlüssel nicht. Der Nachweis, dass Sicherheitscodes den Browser-RAM nicht verlassen, bleibt bei den Playwright-Sicherheitscode-Tests. CVEs im Laufzeit-Image deckt die Analyse nicht ab; dieser Bildscan ist als Phase 2 in `specs/roadmap.md` notiert.

### Hardware-Integrationstest

1. Container mit Hostnetwork starten.
2. `/api/discover` aufrufen.
3. Modell `SMH-T1R` und vollständige `CMAG`-DID prüfen.
4. Session-Probe durchführen.
5. Authentifizierung mit gültigem und ungültigem Code prüfen.
6. Kanal 1 in Full HD starten.
7. HLS-Playlist, Auflösung, Bitrate und Akku prüfen.
8. Stream stoppen und UI-Status prüfen.
9. Kanal 2 starten und dieselben Prüfungen wiederholen.
10. Alle Streams stoppen, Motion je Kanal aus- und wieder einschalten und den gelesenen Zustand prüfen.
11. Containerstatus über `docker inspect` als `healthy` prüfen.

Automatisierter Lauf auf Linux/NixOS:

```bash
export AKKUCAM_TEST_IP='<hub-ip>'
export AKKUCAM_TEST_DID='<hub-did>'
export AKKUCAM_TEST_SECURITY_CODE='<code>'
AKKUCAM_TEST_CHANNEL=0 ./scripts/test-hardware.sh
AKKUCAM_TEST_CHANNEL=1 ./scripts/test-hardware.sh
```

Der Code wird nur als Environment des kurzlebigen Go-Testcontainers übergeben. Er darf nicht in Shell-History, CI oder Versionsverwaltung landen.

Vor dem Hardwaretest muss der reguläre Gatewaycontainer gestoppt sein. Die Basisstation kann eine konkurrierende zweite Sitzung authentifizieren, ohne ihr Videodaten zu liefern; dies würde einen irreführenden Testfehler erzeugen.

## Lokale Testbefehle

Auf NixOS genügt Docker; eine lokale Go-Installation ist nicht erforderlich:

```bash
make test           # vet, Unit/API/Integration, Race Detector, Docker-Build
make lint           # statische Analyse in gepinnten Images (ADR-0033)
make test-e2e       # Playwright in Chromium und Firefox: i18n-, Branding-, Credentials-, Workspaces- und Wall-Suite; Live-Suite nur mit AKKUCAM_TEST_SECURITY_CODE
make test-hardware  # optional, echte Hardware
./scripts/sbom.sh   # CycloneDX- und SPDX-SBOM aus dem Image neu erzeugen
```

## Anforderungs-Traceability

| Spezifikationsanforderung | Automatisierter Nachweis |
|---|---|
| lokale Discovery | Discovery-Unit-Tests; Hardwaretest |
| direkter LAN-Sitzungsaufbau | `TestLocalSessionAuthenticationIntegration` |
| Code nicht in API-Antwort | `TestLiveStartContractAndSecretIsNotEchoed`, `TestAuthenticationContractDoesNotEchoSecret` |
| H.264 unverändert bis Output | `TestLiveProtocolVideoFromSpec`, `TestCountingWriterPassesCompletePayload` |
| Kanal 1/2 und Qualitätsmodus | Fake-Hub validiert Bytes 0 und 4; Hardwaretest je Kanal |
| dynamische Kanalzahl | `TestParseCameraListFromSpecification`, `TestQueryChannelsIntegrationFromSpec`, API-Contract-Test |
| parallele Kanäle | `TestDifferentChannelsRunConcurrently`; realer Dual-Stream-Test |
| Bitrate | `TestCountingWriterCountsSuccessfullyWrittenBytesOnly`, `TestMeterBitrateReportsOneSecondDelta` |
| Akku des aktiven Kanals | `TestLiveProtocolBatteryFromSpec`; Hardwaretest |
| keine konkurrierenden Kontrollsessions | `TestSessionControlRequestsAreSerialized` |
| Inventar/Motion nur ohne Stream | `TestInventoryRejectsWhileStreamsAreActive`, `TestMotionRejectsWhileStreamsAreActive` |
| Motion Read-Modify-Write-Verify | `TestSetMotionReadModifyWriteVerifyIntegration` |
| keine erfundenen Metadaten | `TestUnavailableCameraMetadataIsOmitted`; ADR-0014 |
| Metadaten vollständig sichtbar | Playwright-Workspaces-Suite (DID/MAC ohne Ellipse) |
| stabile Fehlercodes | `TestErrorResponsesCarryStableCodes`, `TestAuthenticationRejectionCarriesUnauthorizedCode`, `TestProbeTimeoutCarriesUpstreamTimeoutCode`, `internal/apierr` |
| OpenAPI-Vertrag | `TestOpenAPIDocumentsEveryRoute`, `TestOpenAPIHasNoUnknownRoutes`, `TestSpecIsValidJSON` |
| API-Vertrag | `internal/gateway/api_test.go` |
| zweisprachige UI (DE/EN) | `TestLocalesHaveSameKeys`, `TestLocalePlaceholdersMatch`, `TestLocaleValuesAreWholePhrases`, `TestIndexHTMLKeysExistInLocales`, `TestAppJSKeysExistInLocales`; Playwright-I18n-Suite |
| Sprachwechsel ohne Stream-Abbruch | Playwright-Wall-Suite (Video-Marker bleibt erhalten); Playwright-Live-Suite (Navigation bleibt bei 1) |
| Fehlercodes übersetzt | `TestAPICodesHaveTranslations`; Playwright-Test „API error codes are translated“ |
| Hilfe-Popup im Topbar | Playwright-I18n-Suite („help popup“) |
| Sicherheitscode je Basisstation | Playwright-Credentials-Suite (getrennte Codes, Wiederverwendung, 401, Code entfernen) |
| Mehrere Basisstationen gleichzeitig | Playwright-Workspaces- und Wall-Suite (gemeinsame Wand, Filter, Start/Stopp aller Hubs, Inventarsperre) |
| Live-Wand: gleiche Kacheln und leere Rasterzelle | Playwright-Wall-Suite (Kachelgrößen, `.tile-empty`) |
| Setup: Vollbreite, Setup-Leiste und Empty State | Playwright-Wall-Suite (Breiten, Empty State, Hub-Raster) |
| Live-Wand: Ansichten und Videoerhalt | Playwright-Wall-Suite (Auto/Fokus/Komfort, Video-Marker) |
| Live-Wand: Fokus-Miniaturen | Playwright-Wall-Suite (Canvas-Miniaturen für laufende Streams) |
| Live-Wand: themeabhängige Kachelrahmen | Playwright-Branding-/Wall-Suite und Screenshot-Review |
| Live-Daten-Overlay | Playwright-Wall-Suite (Inhalt, Klick öffnet Drawer, Schalter, Reset) |
| iPhone-Wand-Leiste | Playwright-Wall-Suite (kompakte Leiste, Menü-Sheet, Einstellungen) |
| Wand-Ansicht persistiert und rücksetzbar | Playwright-Wall-Suite (`akkucam-wall-view`) |
| App-Icon und Favicon | `TestBrandMarkRendersCameraIcon`, `TestFaviconUsesLocalThemeIcons`; Playwright-Branding-Suite |
| Lizenzversionen ohne Drift | `TestLicenseVersionsMatchArtifacts` (hls.js-Bundle, SBOM, Dockerfile) |
| Typografie (Inter Tight: Markenname, Setup- und Wand-Headings) | Playwright-Branding-Suite; `TestStylesheetUsesOnlyShippedFontWeights` |
| Design-Tokens und Icon-Satz | `TestDashboardUsesOneIconSet`; Playwright-Branding-Suite (Kontrast, Gewichte) |
| Barrierefreiheit (Trefferflächen, Safe Area, Kachel-Semantik) | `TestStylesheetRespectsSafeAreaInsets`; Playwright-Wall-Suite (Rolle, Tastatur) |
| Textskalierung (200 %) | `TestStylesheetScalesWithRootFontSize`; Playwright-Wall-Suite |
| Theme folgt der Systemeinstellung | Playwright-Branding-Suite (Theme-Zyklus) |
| Kein Inline-JavaScript im Markup | `TestDashboardHasNoInlineEventHandlers` |
| Code-Sichtbarkeit | Playwright-Credentials-Suite |
| Maschinenwerte statt Servertexte (`durationMs`, `uptimeSeconds`) | API-Contract-Tests, OpenAPI-Abgleich, Playwright-Wall-Suite |
| Lizenzhinweise (Hilfe-Popup + scrollbarer Dialog) | `TestDashboardReferencesExistingStaticAssets`; Playwright-Branding-Suite |
| SBOM in CycloneDX und SPDX | `TestCycloneDXSBOMIsValid`, `TestSPDXSBOMIsValid`, `TestSBOMsListRuntimeComponents` |
| keine Datenrennen | `go test -race -count=1 ./...` |

## Bisher durch Tests gefundene Fehler

1. **API-Routing:** Ein nicht erlaubtes `GET /api/live/start` fiel in den statischen FileServer und lieferte 404 statt 405. Bekannte API-Pfade haben nun explizite Methodengrenzen und `Allow`-Header.
2. **App-Frame-Resynchronisierung:** Ein beschädigter Prefix mit plausibler Großlänge konnte den Assembler dauerhaft blockieren. Der Parser validiert nun zusätzlich den spezifizierten Streamtyp 1–4 und sucht danach weiter.
3. **Hardwaretest-Diagnose:** Konkurrierende Sessions wurden zunächst nur als Video-Timeout sichtbar. Der Test unterscheidet jetzt Batterie-vor-Video, vorzeitiges Sessionende und echten Video-Timeout; das Skript verweigert den Lauf bei aktivem Gatewaycontainer.
4. **Auth-Retry:** Die Wiederholungslogik prüfte den Fehlertext auf `rejected`. Sie verwendet jetzt `errors.Is` mit Sentinel-Fehlern aus `internal/ppcs`.
5. **Veralteter Container:** Nach einer Codeänderung lief auf Port 8080 noch ein Container des alten Images (`gateway-akkucam-gateway:latest`, Vortag 22:50). Das neue Image war korrekt, aber nicht deployt; alle API-Endpunkte antworteten mit 404. `docker compose build` und `docker compose up -d` beheben das. Vor dem Debugging von 404ern immer `docker compose ps` und das Image-Datum prüfen.
6. **Notice ohne `notice-text`-Klasse:** `renderNotice()` suchte `.notice-text`, das Markup hatte die Klasse nicht. Dadurch blieben alle Status- und Fehlermeldungen auf dem Starttext stehen. Playwright hat den Fehler aufgedeckt; die Klasse ist jetzt im Markup.
7. **Fehlendes `setText`:** `app.js` rief `setText()` auf, ohne die Funktion zu definieren. Jeder Kanalstart, jede Statusänderung und jede Credential-Anzeige warf dadurch einen `ReferenceError`. Die Funktion ist jetzt als Wrapper um `i18n.setText` definiert.
8. **Ein Code für alle Basisstationen:** `state.securityCode` hielt genau einen Code. Bei zwei Hubs mit unterschiedlichen Codes wurde der zweite Hub mit dem Code des ersten entsperrt und mit `unauthorized` abgelehnt. Codes liegen jetzt je DID in einer Map (ADR-0025); die Credentials-Suite deckt den Fall ab.
9. **Exklusiver Workspace:** Das Dashboard hielt genau einen Workspace. Beim Laden der Kameras einer zweiten Basisstation verschwand die Ansicht der ersten, obwohl der Streammanager mehrere DIDs parallel erlaubt. Jeder geladene Hub hat jetzt einen eigenen Abschnitt, globale Aktionen steuern alle gemeinsam (ADR-0026); die Workspaces-Suite deckt den Fall ab. Die Abschnitte wurden später durch die gemeinsame Live-Wand ersetzt (ADR-0027).
10. **Rohschlüssel in der Hub-Karte:** `hub.reloadChannels` fehlte in beiden Katalogen; der Button zeigte deshalb den Schlüssel statt „Kameras neu laden“. `TestAppJSKeysExistInLocales` prüft jetzt auch Schlüssel in dynamischen `data-i18n`-Ausdrücken; die I18n-Suite deckt „Kameras neu laden“ und „Code entfernen“ ab.
11. **Abgeschnittene Metadaten:** DID und MAC wurden in der Hub-Details-Tabelle per Ellipse gekürzt. Werte brechen jetzt um; die Workspaces-Suite prüft die vollständige Darstellung.
12. **Fokus-Video an versteckter Grid-Kachel:** Beim Wechsel auf Fokus blieb das alte Grid-DOM verborgen im Dokument. Die Videozuordnung fand dadurch die versteckte Kachel statt der sichtbaren Fokus-Kachel. Die jeweils inaktive Ansicht wird jetzt geleert; die Wall-Suite prüft Videoanzahl und Marker.
13. **Popover-Klick öffnete den Drawer:** Das native Popover schloss per Light-Dismiss, der Folgeklick landete auf der Kachel und öffnete die Seitenleiste. Ein eigenes Overlay mit kontrolliertem Klick-Handling schloss damals nur das Popover (historisch; das Popover wurde durch ADR-0031 ersetzt).

## Aktuelle Abdeckung

Letzte lokale Messung:

| Paket | Statement Coverage |
|---|---:|
| `cmd/gateway` | 0,0 % (keine Tests) |
| `internal/apierr` | 52,6 % |
| `internal/discovery` | 14,9 % |
| `internal/gateway` | 80,1 % |
| `internal/openapi` | 16,7 % |
| `internal/ppcs` | 75,5 % |
| `internal/stream` | 78,4 % |
| gesamt | 69,4 % |

Die niedrige Discovery-Abdeckung entsteht überwiegend durch OS-/Socketpfade. Paketparser und Zusammenführung werden deterministisch getestet; die echte Broadcaststrecke ist Teil des Hardwaretests. Coverage ist ein Diagnosewert, kein Selbstzweck.

### Protokollregression

Bei Änderungen an Framing oder Reassembly sollen gespeicherte, lokal geschützte PCAPs verwendet werden. PCAPs und extrahierte Kamerabilder gehören wegen ihres sensitiven Inhalts nicht standardmäßig in die Versionsverwaltung.

## Verifizierte Umgebung

- Host: Linux/NixOS, x86-64
- Container: Alpine Linux, Go 1.27.4 Buildstage
- Docker Engine mit `network_mode: host`
- Basisstation: `SMH-T1R`, Firmware `1.0.0.12`
- zwei gekoppelte Akku-Kameras auf Kanal 1 und 2
- Browserwiedergabe über lokal gebündeltes hls.js
- verifizierter Stream: H.264 Main, 1920×1080, 25 fps
- beobachtete variable H.264-Nutzdatenbitrate: ungefähr 30–180 KB/s, szenenabhängig

## Akzeptanzkriterien

- kein Gateway-Traffic zu Cloudservern
- erster Streamstart typischerweise innerhalb von 25 Sekunden
- Kanal 1 und 2 liefern jeweils ein aktuelles Bild
- Kanal 1 und 2 können gleichzeitig aktiv bleiben und besitzen unterschiedliche Stream-IDs
- Bitrate aktualisiert sich ungefähr sekündlich
- Batterieantwort gehört zum gewählten Kanal
- Stop beendet FFmpeg und setzt die UI zurück
- erneuter Start ersetzt den vorhandenen Stream ohne verwaiste Prozesse
