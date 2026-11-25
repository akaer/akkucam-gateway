# ADR-0009: Ein aktiver Stream pro Basisstations-DID

- Status: Superseded by [ADR-0011](0011-concurrent-stream-per-channel.md)
- Datum: 2026-11-21

## Kontext

Eine Basisstation kann mehrere Kamerakanäle haben. Parallele Sessions erhöhen Funk-, Akku- und Protokollbelastung. Der aktuelle Hauptanwendungsfall ist ein Browserplayer mit Kanalumschaltung.

## Entscheidung

Der Streammanager erlaubt genau einen aktiven Stream pro DID. Ein neuer Start stoppt und ersetzt den vorhandenen Stream. Der gewählte Kameraindex bestimmt Kanal 1 oder 2.

## Begründung

- verhindert verwaiste FFmpeg- und PPCS-Prozesse
- entspricht dem aktuellen UI-Modell
- reduziert Last auf Basisstation und Akku-Kameras
- eindeutige HLS-URL pro DID

## Betrachtete Alternativen

- paralleler Stream pro DID und Kanal: sinnvoll für NVR/Quad-View, derzeit nicht benötigt.
- Kanalwechsel in bestehender Session: geringere Umschaltzeit, erfordert aber zusätzliche Stop/Change-Zustandslogik.

## Konsequenzen

- Kanalwechsel baut die Session neu auf und fragt den Sicherheitscode erneut ab.
- gleichzeitige Anzeige beider Kameras ist aktuell nicht möglich.
