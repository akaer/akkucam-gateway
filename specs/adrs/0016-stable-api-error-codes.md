# ADR-0016: Stabile API-Fehlercodes und typisierte Fehler

- Status: Accepted
- Datum: 2026-11-21

## Kontext

Die HTTP-API lieferte Fehler bisher ausschließlich als freie Zeichenkette im Feld `error`. Clients konnten Fehler daher nur über den HTTP-Status oder über brüchige Textvergleiche unterscheiden. Interne Fehler wurden teils unverändert an den Browser weitergegeben. Fehlercodes fehlten vollständig.

## Entscheidung

- Jede Fehlerantwort enthält zusätzlich zum bisherigen Feld `error` ein stabiles Feld `code`.
- Das Format bleibt flach und abwärtskompatibel: `{"error":"...","code":"..."}`.
- Codes werden zentral in `internal/apierr` definiert.
- Fachpakete liefern typisierte Sentinel-Fehler (`errors.Is`) und wrappen sie mit `%w`.
- Die HTTP-Schicht mappt Sentinel-Fehler auf Code und Status. Sie vergleicht keine Fehlertexte.
- Unbekannte Fehler werden als `internal_error` mit HTTP 500 und einer generischen Meldung ausgegeben. Details bleiben im Log.
- Der Sicherheitscode erscheint weder in Fehlermeldungen noch in Logs.

## Begründung

- Clients können Fehler stabil und sprachunabhängig behandeln.
- Der API-Vertrag bleibt für bestehende Aufrufer kompatibel, weil `error` unverändert bleibt.
- Typisierte Fehler verhindern, dass sich Verhalten ändert, sobald ein Meldungstext angepasst wird.
- Generische Meldungen für unerwartete Fehler reduzieren das Risiko, interne Details preiszugeben.

## Betrachtete Alternativen

- **Verschachteltes Format** `{"error":{"code":"...","message":"..."}}`: ausdrucksstärker, aber bricht bestehende Clients und die aktuelle Dashboard-Logik.
- **Nur HTTP-Status**: zu grob, mehrere Ursachen teilen sich denselben Status.
- **Fehlertexte vergleichen**: bereits vorhanden und als brüchig verworfen.

## Konsequenzen

- `web/app.js` zeigt eine übersetzte Beschreibung des Codes (`error.code.<code>`) statt der Servermeldung; unbekannte Codes fallen auf die Servermeldung zurück.
- Neue Fehlerpfade benötigen einen Code in `internal/apierr` und einen Test.
- `specs/api.md` dokumentiert das Format und die Codes.
- `internal/openapi/openapi.json` beschreibt das Fehlerschema.

## Verweise

- ADR-0017: OpenAPI als maschinenlesbarer API-Vertrag
- `internal/apierr/apierr.go`
- `specs/api.md`

## Nachtrag 2026-11-24

- Der Default-Zweig von `writeError` in `internal/gateway/server.go` delegiert an `apierr.Write`. Unbekannte Fehler liefern damit wie oben festgelegt `internal_error` mit HTTP 500 und einer generischen Meldung; Details stehen im Log des Handlers. Zuvor wurden sie fälschlich als `upstream_error`/502 gemeldet, sodass etwa Schreibfehler im HLS-Verzeichnis wie Gerätefehler aussahen.
- `TestUnknownErrorUsesInternalError` in `internal/gateway/api_test.go` sichert diesen Pfad ab.

