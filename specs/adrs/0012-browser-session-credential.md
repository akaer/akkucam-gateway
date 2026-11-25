# ADR-0012: Sicherheitscode bis zum Seiten-Reload im Browser-RAM

- Status: Accepted
- Datum: 2026-11-21
- Ersetzt: [ADR-0007](0007-credentials-in-memory-only.md)
- Ergänzt durch: [ADR-0025](0025-per-hub-security-codes.md) (ein Code je Basisstation)

## Kontext

Mehrere Kanäle erfordern mehrere authentifizierte PPCS-Sitzungen. Eine erneute Codeeingabe pro Kanal macht „Alle starten“ und Kanalwechsel unpraktisch.

## Entscheidung

Der Sicherheitscode wird je Basisstation beim Laden ihrer Kameraliste abgefragt und ausschließlich in einer JavaScript-Variable gehalten. Er bleibt bis Seiten-Reload, Tab-Schließen oder explizitem „Code entfernen“ verfügbar. Es erfolgt keine Speicherung in LocalStorage, SessionStorage, Cookie oder Serverdatei. Mit mehreren Basisstationen hält der Browser einen Code pro DID (ADR-0025).

## Begründung

- ein Codeprompt pro Browsersitzung
- parallele Starts und Kanalwechsel bleiben komfortabel
- Reload entfernt das Secret zuverlässig aus dem Anwendungszustand
- keine persistente Browserablage

## Konsequenzen

- JavaScript im Seitenkontext kann auf den Code zugreifen; deshalb dürfen keine externen Laufzeitskripte eingebunden werden.
- laufende Streams bleiben nach „Code entfernen“ aktiv, neue Starts erfordern erneute Eingabe.
- der Server erhält den Code weiterhin nur je Sessionaufbau und persistiert ihn nicht.

## Nachtrag 2026-11-24

Das Eingabefeld für den Code trägt `autocomplete="off"`. Passwortmanager und
Browser-Autofill sollen den Gerätecode nicht dauerhaft speichern; das wäre
eine Persistenz, die diese Entscheidung ausschließt (ADR-0028). Die
Oberflächentexte nennen den Speicherort außerdem „Arbeitsspeicher“, nicht
„Browserspeicher“, weil letzteres auch `localStorage` bezeichnet.
