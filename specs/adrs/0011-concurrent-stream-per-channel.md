# ADR-0011: Ein aktiver Stream pro DID und Kanal

- Status: Accepted
- Datum: 2026-11-21
- Ersetzt: [ADR-0009](0009-one-active-stream-per-did.md)

## Kontext

Eine SMH-T1R-Basisstation kann mehrere gekoppelte Akku-Kameras bedienen. Die reale Hardware meldet ihre Kanäle über IO-Control `40962` und kann für Kanal 1 und 2 gleichzeitig unabhängige lokale PPCS-Sitzungen liefern.

## Entscheidung

Der Streammanager adressiert Streams mit `<DID>-ch<Kanalnummer>`. Pro DID und Kanal ist genau ein Stream aktiv; unterschiedliche Kanäle derselben DID dürfen parallel laufen. Ein Neustart desselben Kanals ersetzt nur diesen Kanal.

## Begründung

- ermöglicht Dashboard- und Vergleichsansicht
- behält eindeutigen Lifecycle pro Kamerakanal
- verhindert weiterhin doppelte Sessions für denselben Kanal
- getrennte HLS-Pfade, Bitraten und Batteriewerte

## Konsequenzen

- „Alle starten“ baut die Kanalsitzungen kontrolliert nacheinander auf und lässt sie anschließend parallel laufen.
- Funk- und Akkuverbrauch steigen mit jedem aktiven Kanal.
- Stop kann pro Kanal oder gesammelt erfolgen.
- die Anzahl der Karten stammt aus der Hub-Kameraliste, nicht aus einer festen Annahme.

## Präzisierung 2026-11-24

Der Client startet die Kanäle eines Hubs parallel (`Promise.all` in `web/app.js`). Die geforderte kontrollierte Reihenfolge entsteht serverseitig: `POST /api/live/start` läuft unter demselben Prozess-Mutex wie Probe und Authentifizierung (ADR-0008). Die Sitzungsaufbauten laufen dadurch nacheinander, die Medienströme danach parallel.
