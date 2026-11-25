# ADR-0033: Statische Analyse im Toolchain-Container

- Status: Accepted
- Datum: 2026-11-25

## Kontext

Bis hierhin sicherten `go vet`, Unit- und Integrationstests, Fuzzing-Seeds, Playwright und die SBOM die Qualität. Es fehlte eine Prüfung auf ungeprüfte Fehler, Injection-Senken, CVEs in der Go-Standardbibliothek und eine Analyse der Browserquellen; `web/app.js` sieht kein Go-Werkzeug. Werkzeuge direkt auf NixOS zu installieren würde den Docker-Workflow verlassen und eine zweite, driftende Toolchain in der CI erzeugen.

Der Probelauf mit dem bestehenden Stand ergab: 21 Go-Findings, 0 Semgrep-Findings, 2 Shellcheck-Warnungen, 1 Hadolint-Warnung – und 32 erreichbare CVEs in der Go-Standardbibliothek, weil Go 1.23 seit August 2025 keine Fixes mehr erhält.

## Entscheidung

- `./scripts/lint.sh` (Ziel `make lint`) führt jede Analyse in einem gepinnten Container-Image aus. Die Image-Variablen im Skript sind die einzige Versionsquelle für lokal und CI; es gibt kein Nix-Paket und keinen CI-eigenen Installer.
- **golangci-lint** v2.15.0, `default: standard` plus `gosec`, `errorlint`, `noctx`, `bodyclose`, `misspell`; Konfiguration in `.golangci.yml`.
- **govulncheck** v1.10.0 im Go-Image der Build-Stage (Go 1.27.4) und damit auf derselben Standardbibliothek wie das Binary.
- **semgrep** 1.187.0 mit `p/golang`, `p/javascript`, `p/secrets` und eigenen Regeln in `.semgrep/gateway.yml`: keine ausgehenden HTTP-Clients (ADR-0003), kein ausgehendes TCP, kein Sicherheitscode in `localStorage`/`sessionStorage` (ADR-0007, ADR-0012).
- **shellcheck**, **hadolint**, **actionlint** für Skripte, Dockerfile und Workflows.
- Alle Werkzeuge laufen blockierend. Ausnahmen stehen ausschließlich in `.golangci.yml` und sind dort begründet: G115 im Protokollpaket `internal/ppcs`, G304/G306 in Tests.
- Jedes `//nolint` trägt Grund und Verweis. Aktuell existiert genau eines: G706 im Request-Logger, weil der Pfad mit `%q` maskiert wird und `log` keine Escapes interpretiert.
- Semgrep lädt die Registry-Regelsets zur Laufzeit und sendet keine Metriken (`--metrics=off`, `SEMGREP_SEND_METRICS=off`). Die eigenen Regeln liegen im Repo.
- Der Toolchain-Bump auf Go 1.27 gehört zum selben Schritt, weil das Analyse-Gate auf einem EOL-Toolchain nur CVE-Rauschen erzeugt.

## Begründung

- Container-Pinning hält NixOS-Host, lokale Läufe und CI identisch und passt zum bestehenden Muster von `test.sh` und `sbom.sh`.
- Semgrep deckt `web/*.js` ohne Node, npm oder Build-Schritt ab.
- govulncheck prüft erreichbaren Code statt nur Paketversionen und ersetzt damit kein SBOM, ergänzt es aber um die Frage „rufen wir das an?“.
- Die eigenen Semgrep-Regeln machen die Hard Rules prüfbar, die bisher nur in Prosa standen.
- Analysewerkzeuge landen nicht im Image; `go.mod` bleibt ohne Fremdmodule, SBOM und Lizenzdialog ändern sich nicht.

## Betrachtete Alternativen

- **Nur `go vet` und staticcheck**: keine Sicherheitsregeln, kein Frontend.
- **gosec und staticcheck einzeln**: mehr Skripte für dieselbe Abdeckung.
- **ESLint** für das Frontend: npm plus Build-Schritt, verstößt gegen die Frontend-Regel.
- **CodeQL**: bessere Taint-Analyse, aber an GitHub Actions und für private Repos an GHAS gebunden; es gibt noch kein Remote.
- **nix-shell oder Flake-DevShell**: nixpkgs-Version driftet, CI bräuchte ein zweites Setup, und Semgrep ist dort nicht versionsgleich.
- **govulncheck auf Go 1.23 belassen**: 32 erreichbare Standardbibliothek-CVEs, davon Treffer über `net/http.ListenAndServe` – also den Kern des Gateways.

## Konsequenzen

- Neue Befehle: `make lint`, `./scripts/lint.sh`; neuer CI-Job `lint`.
- Semgrep-Regeln ändern sich unabhängig vom Repo. Ein neues Finding wird behoben oder eng ausgeschlossen, nie das Gate abgeschaltet.
- Lint braucht Netz für Image-Pulls und Registry-Regeln.
- Bei einem Go-Bump müssen Go-Image, govulncheck und golangci-lint zusammenpassen; golangci-lint muss mit mindestens der Modul-Go-Version gebaut sein.
- Kein Einfluss auf Laufzeit-Image, SBOM oder Lizenzdialog.
- Phase 2, offen und in `specs/roadmap.md` notiert: Container-Image-Scan mit Trivy auf dem gebauten Image, ergänzend zur SBOM.

## Verweise

- ADR-0003: lokal ohne Cloud-Fallback
- ADR-0006: H.264-Passthrough und HLS
- ADR-0007: Zugangsdaten nur im Speicher
- ADR-0012: Sicherheitscode im Browser-RAM
- ADR-0018: Test- und Codequalitätskonventionen
- `specs/testing.md`, `specs/security.md`, `.golangci.yml`, `.semgrep/gateway.yml`
