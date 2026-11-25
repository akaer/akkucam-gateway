# ADR-0018: Test- und Codequalitätskonventionen

- Status: Accepted
- Datum: 2026-11-21

## Kontext

Die Go-Quellen enthielten viele unbenannte Zahlen, wenige Kommentare und verstreute Testdateien mit teils abweichenden Namen. Fehlermeldungen waren nicht typisiert. Die HTTP-Schicht lag im `main`-Paket und war dadurch schlechter testbar. Es fehlte eine gemeinsame Regel, wo Tests liegen und wie neue Fehler, Konstanten und API-Änderungen behandelt werden.

## Entscheidung

- **Keine Magic Numbers.** Protokoll-Offsets, Timeouts, Limits und Puffergrößen erhalten benannte Konstanten in einer paketweiten `constants.go`.
- **Dokumentation.** Exportierte Bezeichner erhalten Godoc. Byte-Layouts und Protokollbesonderheiten werden inline kommentiert.
- **Typisierte Fehler.** Fachpakete verwenden Sentinel-Fehler und `%w`. Die HTTP-Schicht mappt sie über `internal/apierr`.
- **Tests bleiben im Paket.** Unit- und Integrationstests liegen neben dem getesteten Code. `tests/` ist ausschließlich für paketübergreifende End-to-End-Tests.
- **Build-Tags.** Hardwaretests verwenden `//go:build hardware`. Fake-Hub-Integrationstests laufen standardmäßig mit.
- **HTTP-Schicht.** `internal/gateway` enthält API, Validierung, Fehler-Mapping und Routentabelle. `cmd/gateway` ist nur noch Einstiegspunkt.
- **API-Vertrag.** Jede Route muss in der Routentabelle und in `internal/openapi/openapi.json` stehen.
- **ADRs.** Strukturelle oder vertragsrelevante Änderungen benötigen ein neues oder aktualisiertes ADR.

## Begründung

- Benannte Konstanten machen Protokollcode lesbar und prüfbar.
- Typisierte Fehler entkoppeln Verhalten von Meldungstexten.
- Tests neben dem Code erlauben Zugriff auf unexportierte Parser und Builder.
- Eine schlanke `main`-Funktion und ein eigenes API-Paket verbessern Testbarkeit und Übersicht.
- Der Contract-Test verhindert veraltete API-Dokumentation.

## Betrachtete Alternativen

- **Alle Tests in `tests/`**: würde unexportierte Funktionen unzugänglich machen oder Shim-Exporte erzwingen.
- **Swagger-Annotationen statt separater Datei**: zusätzliche Toolchain, siehe ADR-0017.
- **Fehlertexte vergleichen**: brüchig, siehe ADR-0016.

## Konsequenzen

- Neue Quellen benötigen bei Bedarf eine `constants.go` und Godoc.
- Neue Tests folgen der Namenskonvention in `tests/README.md`.
- Neue API-Routen erfordern einen Eintrag in der Routentabelle und in `openapi.json`.
- `AGENTS.md` fasst die Regeln für Mitwirkende zusammen.

## Verweise

- ADR-0016: stabile API-Fehlercodes
- ADR-0017: OpenAPI als maschinenlesbarer API-Vertrag
- `tests/README.md`
- `AGENTS.md`
