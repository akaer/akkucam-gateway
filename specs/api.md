# HTTP-API

## Allgemeines

- Standardadresse: `http://<gateway-host>:8080`
- Format: JSON, UTF-8
- aktueller Stand: keine HTTP-Authentifizierung
- Einsatzbereich: vertrauenswürdiges lokales Netz

## Fehlerformat

Fehlerantworten sind flach und abwärtskompatibel:

```json
{"error":"valid ip, did and securityCode are required","code":"validation_failed"}
```

`error` enthält eine menschenlesbare Meldung. `code` ist stabil und für Clients gedacht. Unbekannte Fehler werden als `internal_error` mit HTTP 500 und einer generischen Meldung ausgegeben; Details bleiben im Log.

Die API bleibt englisch. Die Weboberfläche übersetzt bekannte Codes über `error.code.<code>` aus `web/locales/*.json` (ADR-0021); andere Clients können denselben Code-Mechanismus nutzen. Neue Codes müssen in beiden Katalogen ergänzt werden.

| Code | Typischer HTTP-Status | Bedeutung |
|---|---:|---|
| `invalid_json` | 400 | Request-Body ist kein gültiges JSON |
| `validation_failed` | 400 | Pflichtfeld fehlt oder Wert liegt außerhalb des erlaubten Bereichs |
| `method_not_allowed` | 405 | Pfad existiert, Methode nicht |
| `not_found` | 404 | API-Pfad existiert nicht |
| `conflict` | 409 | Aktion ist während eines aktiven Streams gesperrt |
| `unauthorized` | 401 | Sicherheitscode wurde von der Basisstation abgelehnt |
| `upstream_timeout` | 504 | Basisstation antwortet nicht rechtzeitig |
| `upstream_error` | 502 | Basisstation meldet einen Fehler oder die Sitzung bricht ab |
| `internal_error` | 500 | Unerwarteter interner Fehler |

## Maschinenlesbarer Vertrag

- OpenAPI-Dokument: `GET /api/openapi.json`
- Swagger UI: `GET /api/docs`

Die Swagger UI wird lokal ausgeliefert und lädt keine CDN-Ressourcen. Der Vertrag liegt in `internal/openapi/openapi.json`; ein Contract-Test prüft, dass Routentabelle und Dokument übereinstimmen. Die Spec verwendet die relative Server-URL `/`, damit „Try it out“ die Requests an denselben Origin schickt – auch wenn das Dashboard über eine LAN-IP oder einen Reverse Proxy geöffnet wird. Eine feste `127.0.0.1`-URL würde im Browser auf den Client zeigen und `NetworkError` auslösen.

## `GET /api/health`

Beispielantwort:

```json
{"status":"ok","mode":"local-only","live":"hls"}
```

## `GET /api/version`

Liefert die semantische Version und den kurzen Git-Commit des laufenden
Binaries. Fehlt der Commit im Build, entfällt das Feld.

```json
{"version":"0.1.0-beta.2","commit":"bfe827d"}
```

## `POST /api/discover`

Startet WAI- und PPPP-LAN-Discovery. Request-Body ist leer.

```json
{
  "devices": [
    {
      "ip": "192.0.2.10",
      "did": "CMAG-000001-ABCDE",
      "model": "SMH-T1R",
      "ppcsPort": 17500,
      "mac": "02:00:00:00:00:10",
      "local": true
    }
  ],
  "durationMs": 4001
}
```

Der PPCS-Port ist dynamisch und darf nicht dauerhaft gespeichert werden. `mac` fehlt, wenn die Adresse nicht aus der lokalen ARP-Tabelle ermittelt werden konnte.

## `POST /api/session/probe`

Öffnet eine lokale Sitzung bis `P2P_RDY` und schließt sie sofort wieder.

```json
{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE"}
```

Der Endpunkt sendet keine Zugangsdaten oder Kamerakommandos.

## `POST /api/session/authenticate`

Prüft den Sicherheitscode und schließt die Sitzung nach erfolgreicher Authentifizierung.

```json
{
  "ip":"192.0.2.10",
  "did":"CMAG-000001-ABCDE",
  "securityCode":"example-code"
}
```

Der Code wird weder in Antworten noch Logs ausgegeben. Transport-Timeouts werden bis zu dreimal wiederholt; eine explizite Ablehnung nicht.

## `POST /api/live/start`

Startet oder ersetzt den aktiven Stream derselben DID und desselben Kanals.

```json
{
  "ip":"192.0.2.10",
  "did":"CMAG-000001-ABCDE",
  "securityCode":"example-code",
  "cameraIndex":0,
  "quality":1
}
```

Erfolgsantwort:

```json
{
  "id":"CMAG-000001-ABCDE-ch1",
  "playlistUrl":"/hls/CMAG-000001-ABCDE-ch1/index.m3u8"
}
```

Der Endpunkt bestätigt Session und FFmpeg-Start. Bis zur ersten HLS-Playlist können wegen Kamera-Wake-up weitere Sekunden vergehen.

Die Stream-ID enthält die Kanalnummer, beispielsweise `CMAG-000001-ABCDE-ch2`. Dadurch können mehrere Kanäle derselben DID gleichzeitig laufen.

## `POST /api/channels`

Authentifiziert lokal und liest die gekoppelte Kameraliste der Basisstation.

```json
{"ip":"192.0.2.10","did":"CMAG-000001-ABCDE","securityCode":"example-code"}
```

```json
{
  "did":"CMAG-000001-ABCDE",
  "hub": {
    "model":"SMH-T1R",
    "vendor":"ExampleVendor",
    "firmware":"1.0.0.12",
    "mcuVersion":"1.2.3",
    "mac":"02:00:00:00:00:10",
    "sdCardStatus":"normal",
    "sdTotalBytes":32000000000,
    "sdFreeBytes":8000000000
  },
  "channels":[
    {
      "index":0,
      "name":"Einfahrt",
      "model":"SAPP-TAS",
      "fwVersion":"0.0.36.0",
      "mcuVersion":"0.5.2.0",
      "fwVersionRaw":9216,
      "mcuVersionRaw":328192,
      "updateable":false,
      "motionEnabled":true,
      "pirMode":1,
      "pirSensitivityDay":65
    }
  ]
}
```

Nicht verfügbare Metadatenfelder fehlen vollständig. Eine Kamera-MAC ist nicht Teil der bekannten Hub-Antworten und wird daher nicht ausgegeben. Während eines aktiven Streams antwortet der Endpunkt mit `409 Conflict`.

## `POST /api/motion`

Ändert die PIR-Bewegungserkennung eines gekoppelten Kanals. Der Gateway liest zunächst den vollständigen Parameterblock, ändert nur `pirMode` und verifiziert die Einstellung anschließend durch erneutes Lesen.

```json
{
  "ip":"192.0.2.10",
  "did":"CMAG-000001-ABCDE",
  "securityCode":"example-code",
  "channel":0,
  "enabled":false
}
```

```json
{"index":0,"motionEnabled":false,"pirMode":0,"pirSensitivityDay":65}
```

`enabled` muss auch für `false` explizit vorhanden sein. Während eines aktiven Streams antwortet der Endpunkt mit `409 Conflict`.

## `GET /api/live/status?id=<id>`

```json
{
  "id":"CMAG-000001-ABCDE-ch2",
  "running":true,
  "channel":1,
  "bitrateBytesPerSecond":105469,
  "framesPerSecond":25,
  "battery":{"capacity":87,"charging":false},
  "uptimeSeconds":42
}
```

`channel` ist nullbasiert. `battery` fehlt, solange noch keine gültige Antwort vorliegt. `framesPerSecond` wird aus der eingehenden H.264-Payloadrate pro Sekunde abgeleitet und fehlt, solange noch keine Messung vorliegt.

## `POST /api/live/stop`

```json
{"id":"CMAG-000001-ABCDE-ch2"}
```

Beendet PPCS-Sitzung und FFmpeg-Prozess. Mehrfaches Stoppen ist idempotent.

## HLS

```text
GET /hls/<id>/index.m3u8
GET /hls/<id>/indexN.ts
```

HLS-Dateien sind flüchtige Laufzeitdaten und keine Aufzeichnungsschnittstelle.
