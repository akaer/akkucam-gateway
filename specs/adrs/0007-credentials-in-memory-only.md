# ADR-0007: Sicherheitscode nur im Arbeitsspeicher

- Status: Superseded by [ADR-0012](0012-browser-session-credential.md)
- Datum: 2026-11-21

## Kontext

Der Geräte-Sicherheitscode ist für die AES-Challenge-Authentifizierung erforderlich. Die Hersteller-App speichert vergleichbare Werte nur schwach verschleiert.

## Entscheidung

Der Gateway persistiert den Sicherheitscode nicht. Der Browser fragt ihn bei Authentifizierung oder Streamstart ab und sendet ihn im Request. Nach Sitzungsaufbau wird nur der abgeleitete Session-Key verwendet.

## Begründung

- kein Secret in Compose, Image, Datenbank oder Konfigurationsdatei
- geringere Folgen eines Dateizugriffs auf den Host
- einfache und transparente Sicherheitsgrenze

## Betrachtete Alternativen

- Environment-Variable: landet häufig in Prozess- und Container-Metadaten.
- Konfigurationsdatei: bequem, erfordert aber Secret-Management und Dateirechte.
- verschlüsselte Datenbank: Schlüsselproblem wird nur verschoben.

## Konsequenzen

- Nutzer muss den Code bei jedem neuen Stream eingeben.
- zukünftiger komfortabler Kanalwechsel sollte den Code höchstens innerhalb einer aktiven Browsersitzung im RAM halten.
- HTTP schützt den Request im LAN nicht vor Mitschnitt; TLS bleibt eine empfohlene Härtung.
