# Design: Multi-Camera-Dashboard

## Backend

- neuer Endpoint `POST /api/channels`
- IO-Control `40962` liest Anzahl und 60-Byte-Kameraeinträge
- Stream-ID enthält DID und 1-basierte Kanalnummer
- Streammanager ersetzt nur identische DID/Kanal-Kombinationen
- Kontroll-Handshakes bleiben serialisiert, Medienloops laufen parallel

## Frontend

- Security-Code-Dialog, Code nur in JavaScript-Variable
- dynamische Karten aus Hubantwort
- „Alle starten“ startet Kanäle sequentiell; danach paralleler Betrieb
- Statuspolling pro Stream-ID
- HLS-Instanz pro Videokarte
- nativer Fullscreen beziehungsweise `webkitEnterFullscreen` auf iOS
- Diagnosebereich mit manuellem Kanal-Fallback
- CSS-Custom-Properties für Themes; Themepräferenz in LocalStorage
- Metadatenfelder werden wertabhängig gerendert; fehlende Werte erzeugen keine Platzhalterzeilen
- Metadaten und Live-Metriken sind standardmäßig zugeklappte `details/summary`-Bereiche (progressive disclosure)

## Fehlerverhalten

- schlägt die Kameraliste fehl, wird kein Kanal erfunden
- Diagnosebereich öffnet sich und bietet manuelle Slots 1–4
- Fehler eines Kanals beendet nicht die übrigen Streams
