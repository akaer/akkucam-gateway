# ADR-0017: OpenAPI als maschinenlesbarer API-Vertrag

- Status: Accepted
- Datum: 2026-11-21

## Kontext

Die HTTP-API war nur in `specs/api.md` als Prosa dokumentiert. Es gab keinen maschinenlesbaren Vertrag, keine interaktive Dokumentation und keine automatische Prüfung, ob die Dokumentation zu den tatsächlich registrierten Routen passt.

## Entscheidung

- `internal/openapi/openapi.json` ist der kanonische OpenAPI-3.0-Vertrag.
- Die Datei wird per `go:embed` in die Binärdatei eingebettet und unter `GET /api/openapi.json` ausgeliefert.
- Eine lokal gebündelte Swagger UI wird unter `GET /api/docs` bereitgestellt. Es werden keine CDN- oder Cloud-Ressourcen geladen.
- Die Swagger-UI-Assets liegen unter `web/swagger/` und werden mit der bestehenden Web-Wurzel ausgeliefert.
- Ein Contract-Test vergleicht die Routentabelle in `internal/gateway` mit den Pfaden und Methoden der OpenAPI-Datei. Abweichungen lassen den Test fehlschlagen.
- Die Datei liegt bewusst unter `internal/openapi/`, weil `go:embed` nur Dateien innerhalb des Paketverzeichnisses einbetten kann. `specs/api.md` bleibt die menschliche Ergänzung und verweist auf den Vertrag.

## Begründung

- Maschinenlesbarer Vertrag für Clients, Generatoren und Tests.
- Keine zusätzliche Laufzeitabhängigkeit und kein Cloud-Zugriff.
- Der Contract-Test verhindert, dass Doku und Implementierung auseinanderlaufen.
- Eine lokale Swagger UI macht die API ohne zusätzliche Werkzeuge erkundbar.

## Betrachtete Alternativen

- **Nur `specs/api.md`**: bereits vorhanden, aber nicht maschinenlesbar und nicht testbar.
- **Swagger-Annotationen im Go-Code**: erzeugt zusätzliche Toolchain-Abhängigkeiten und vermischt Dokumentation mit Handler-Code.
- **Swagger UI vom CDN laden**: verstößt gegen die Cloud-Free-Regel.
- **OpenAPI-Datei unter `specs/`**: würde `go:embed` verletzen oder eine Kopie erfordern.

## Konsequenzen

- Jede API-Änderung muss `openapi.json` aktualisieren.
- Die Spec verwendet eine relative Server-URL, damit „Try it out“ vom jeweiligen Origin arbeitet und nicht fest auf `127.0.0.1` zeigt.
- Die Swagger-UI-Assets unterliegen Apache-2.0; die Lizenz liegt unter `web/swagger/LICENSE`.
- `README.md` und `specs/api.md` verlinken `/api/docs` und `/api/openapi.json`.
- Der Contract-Test läuft mit `go test ./...`.

## Verweise

- ADR-0016: stabile API-Fehlercodes
- `internal/openapi/openapi.go`
- `internal/gateway/openapi_test.go`
- `specs/api.md`
