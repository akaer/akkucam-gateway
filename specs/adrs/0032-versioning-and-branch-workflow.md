# ADR-0032: Semantische Versionierung und Branch-Workflow

- Status: Accepted
- Datum: 2026-11-25

## Kontext

Bis zum ersten Commit gab es weder Releases noch einen Branch-Workflow. Das
Gateway wird als Container ausgeliefert, und im Betrieb muss erkennbar sein,
welcher Stand läuft – besonders nach einem Kamera- oder Protokollfix. Ein
Versionsstring ohne Bezug zum Quellstand reicht dafür nicht.

## Entscheidung

### Branch und Pull Request

- `main` ist geschützt und immer baubar. Änderungen laufen über
  `feature/<slug>` oder `fix/<slug>` und einen Pull Request. Reine
  Dokumentationsänderungen dürfen `docs/<slug>` nutzen.
- Der Pull Request wird nach grüner CI als Squash gemergt, der Branch danach
  gelöscht. Direkte Commits auf `main` sind ausgeschlossen, auch für
  Versionsbumps. Vorlage: `.github/pull_request_template.md`.

### Version

- Versionierung folgt Semantic Versioning 2.0.0. Der Major bleibt bis zum
  stabilen `1.0.0` bei `0`; die erste Version ist `0.1.0-beta.1`.
- Vor `1.0.0`: `MINOR` für Features und Breaking Changes, `PATCH` für
  abwärtskompatible Fixes. Vorabversionen tragen `-beta.N`.
- Die Version steht ausschließlich in `internal/version/version.go`. Sie wird
  in dem Pull Request erhöht, der das Release ausmacht. Nach dem Merge wird
  `main` mit `v<version>` getaggt (zum Beispiel `v0.1.0-beta.1`). Tags werden
  nicht verschoben oder wiederverwendet. Docker-Tags tragen die Version ohne
  Commit-Hash.

### Commit im Build

- `Dockerfile`, `Makefile`, `scripts/test.sh` und CI übergeben den Git-Commit
  als Build-Argument `COMMIT`; `-ldflags -X` schreibt ihn ins Binary.
- Lokale `go run`-Builds lesen den VCS-Stempel aus `debug.ReadBuildInfo()`.
- `GET /api/version` liefert `version` und, sofern bekannt, den siebenstelligen
  `commit`. Der Hilfe-Dialog zeigt beides in einer stillen Zeile. Fehlt der
  Commit, entfällt das Feld; es wird kein Platzhalter erfunden.

## Begründung

- Ein PR pro Änderung hält `main` baubar und macht CI-Nachweise und Review
  nachvollziehbar.
- SemVer Plus Commit-Hash trennt Release-Nummer und exakten Quellstand. Der
  Hash allein ist für Menschen unlesbar, die Version allein nicht eindeutig.
- Eine einzige Versionsquelle in Go vermeidet Drift zwischen Datei, Tag und
  Binary; der Commit wird beim Build injiziert, damit das Image reproduzierbar
  bleibt und kein `.git` im Build-Kontext landen muss.
- Die Anzeige im Hilfe-Dialog folgt ADR-0015: Diagnoseinformation erscheint auf
  Nachfrage, nicht dauerhaft in der Topbar.

## Betrachtete Alternativen

- **Nur Git-Tags ohne Versionskonstante:** erfordert `git describe` im Build und
  liefert in Quellarchiven oder ohne Git keinen Wert.
- **Version über `ARG VERSION` im Dockerfile:** verlagert die Quelle in
  Build-Skripte und Compose-Dateien; Tag und Binary können auseinanderlaufen.
- **Version im `/api/health`:** vermischt Liveness mit Build-Metadaten und
  zwingt Monitoring-Endpunkte zu einem zusätzlichen Vertragsfeld.
- **`debug.ReadBuildInfo()` allein:** funktioniert nicht im Container, weil der
  Build-Kontext kein `.git` enthält.
- **Version in der Topbar:** dauerhafte Anzeige ohne Nutzwert im Alltag und
  zusätzliche Konkurrenz zur Live-Wand.

## Konsequenzen

- Jede Änderung durchläuft einen Pull Request; das kostet bei Ein-Personen-
  Änderungen einen zusätzlichen Schritt.
- Ein Release benötigt einen Versionsbump, einen Merge und einen Tag. Ohne Tag
  bleibt `main` ein Zwischenstand, auch wenn die Version bereits erhöht ist.
- Ein roher `docker compose build` ohne `COMMIT` zeigt keinen Hash. Die
  dokumentierten Wege `make up`, `make test` und CI setzen ihn.
- Branch Protection und die Pflicht-Statuschecks sind GitHub-Einstellungen und
  nicht als Datei im Repository versioniert.

## Verweise

- [Semantic Versioning 2.0.0](https://semver.org/)
- ADR-0015: Progressive Disclosure
- ADR-0017: OpenAPI als API-Vertrag
- ADR-0024: Hilfe-Popup im Topbar
- `CONTRIBUTING.md`, `.github/pull_request_template.md`

## Nachtrag 2026-11-25

- `main` und die Release-Tags `v*` sind als GitHub-Rulesets geschützt: Pull
  Request, grüne Checks, lineare History, keine Force-Pushes, keine
  Tag-Änderungen. Die Regeln liegen in GitHub, nicht im Repository.
- Ein Tag `v<version>` startet `.github/workflows/release.yml`. Der Workflow
  prüft, dass Tag und `internal/version/version.go` übereinstimmen und der
  Commit auf `main` liegt, baut das Image für `linux/amd64` und `linux/arm64`,
  schiebt es nach GHCR und legt einen GitHub Release mit generierten Notes und
  den SBOM-Dateien an.
- Das Image trägt OCI-Labels mit Quelle, Lizenz, Version und Revision; das
  verlinkt das GHCR-Paket mit dem Repository.
- Das erste veröffentlichte Release ist `v0.1.0-beta.2`. Der Tag
  `v0.1.0-beta.1` bleibt unverändert, wird aber nicht als Release
  veröffentlicht.
