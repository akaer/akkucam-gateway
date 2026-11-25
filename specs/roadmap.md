# Roadmap

## Nächste sinnvolle Schritte

### Robustheit

- Session-Reconnect nach Funkabbruch
- Statusgrund und letzter Fehler über API ausgeben
- HLS-Startup und Kamera-Wake-up getrennt messen
- Paketverlusttests ergänzen
- Graceful Shutdown bei SIGTERM

### Sicherheit

- optionale HTTP-Authentifizierung
- dokumentiertes Reverse-Proxy-Beispiel für TLS
- Request-Größenlimits und explizite Timeouts für Body/Response

### Bedienung

- hochauflösende Standbilder bzw. ein Snapshot-Endpunkt (der Fokus-Streifen nutzt jetzt Live-Miniaturen aus den Videostreams)
- Zugriff auf die von der Basisstation erzeugten Aufzeichnungen (Bewegungserkennung aktiv, SD-Karte vorhanden)
- besserer Fehlerstatus bei leerem/abgeschaltetem Kanal
- weitere Sprachen über zusätzliche Kataloge in `web/locales/` (ADR-0021)

### Medien

- WebRTC als optionale Niedriglatenz-Alternative zu HLS
- RTSP-Ausgabe für Home Assistant, go2rtc oder NVRs
- SPS/PPS- und Auflösungsanzeige im Status

### Geräteunterstützung

- weitere Firmwarestände testen
- Verhalten der zweiten gefundenen Gerätefamilie (`WGAG`) getrennt untersuchen
- Kamera- und Hub-Fähigkeiten dynamisch abfragen

### Integrationen

- **HomeKit (Apple Home) über Home Assistant oder Homebridge.** Kein nativer HAP-Support im Gateway. Voraussetzung ist die RTSP-Ausgabe; ein Standbild kann aus dem RTSP-Stream erzeugt werden. Details siehe [HomeKit (Apple Home)](#homekit-apple-home).

## HomeKit (Apple Home)

Geprüft, nicht umgesetzt. Der Gateway bleibt HomeKit-frei; die Einbindung erfolgt über eine Brücke. Empfohlen sind die Home-Assistant-HomeKit-Bridge oder Homebridge mit `camera-ffmpeg`. Kameras laufen im Accessory-Mode, ein Accessoire pro Kanal. Ein eigener Snapshot-Endpunkt ist optional, weil beide Brücken ein Frame aus dem RTSP-Stream ziehen können.

**Nicht-Ziele:** HomeKit Secure Video, iCloud-Aufnahmen, Zwei-Wege-Audio/Gegensprechen, MFi-Zertifizierung und „Works with Apple Home“-Badge, nativer HAP-Support im Gateway.

**Grenzen:** H.264-Passthrough ist möglich, muss aber mit echter Hardware geprüft werden. Bewegungs-Benachrichtigungen setzen eine echte Ereignisquelle voraus; der Gateway kennt derzeit nur den Ein-/Aus-Zustand der Bewegungserkennung. Aufzeichnungen bleiben auf der Basisstation-SD und erscheinen nicht in Apple Home.

### User Stories

- **US-HK-01 – Kamera im Apple Home sehen:** Als Apple-Home-Nutzer mit Home Assistant möchte ich jede Gateway-Kamera als eigenes Accessoire sehen, damit ich sie im Home-App live öffnen kann. *Akzeptanz:* pro Kanal ein Accessoire; Live-View startet; H.264 möglichst ohne Transcoding; Kachel zeigt ein Standbild.
- **US-HK-02 – Zwei Kanäle getrennt:** Als Nutzer mit zwei gekoppelten Kameras möchte ich beide Kanäle als getrennte Accessoires, damit ich sie unabhängig öffnen kann. *Akzeptanz:* zwei Accessoires; parallele Streams; keine gegenseitige Blockade.
- **US-HK-03 – Standbild in der Kachel:** Als Nutzer möchte ich ein aktuelles Standbild sehen, damit die Kachel nicht schwarz bleibt. *Akzeptanz:* Snapshot aus RTSP oder eigenem Endpunkt; JPEG; Aktualisierung im von Apple vorgegebenen Rhythmus.
- **US-HK-04 – Siri und Szenen:** Als Nutzer möchte ich die Kamera per Siri oder Home-Szene öffnen. *Akzeptanz:* Accessoire ist in Szenen wählbar; Siri-Aufruf funktioniert, soweit Apple das unterstützt.
- **US-HK-05 – Kopplung überlebt Neustarts:** Als Nutzer möchte ich nach einem Gateway-Neustart nicht neu koppeln. *Akzeptanz:* Pairing-Daten liegen bei HA/Homebridge; Gateway-Neustart erfordert kein Re-Pairing.
- **US-HK-06 – Kein Audio, kein Audiochaos:** Als Nutzer möchte ich in HomeKit bewusst kein Audio, damit bei mehreren Streams kein Audiochaos entsteht. *Akzeptanz:* `support_audio: false` bzw. `audio: false`; keine Audio-Steuerelemente; ADR-0020 bleibt gültig.
- **US-HK-07 – Datenschutz bleibt lokal:** Als datenschutzorientierter Nutzer möchte ich, dass die Einbindung lokal bleibt und keine Cloud-Aufnahmen entstehen. *Akzeptanz:* HKSV deaktiviert; keine iCloud-Aufnahmen; Dokumentation weist darauf hin.
- **US-HK-08 – Aufzeichnungen klar abgegrenzt:** Als Nutzer möchte ich verstehen, dass Aufnahmen auf der Basisstation-SD liegen und nicht in Apple Home. *Akzeptanz:* Doku und Feature-Status; HomeKit zeigt keine Aufnahmen; Zugriff auf SD-Aufzeichnungen bleibt separater Roadmap-Punkt.
- **US-HK-09 – Homebridge-Vorlage:** Als Homebridge-Nutzer möchte ich eine fertige Config-Vorlage, damit ich die Kamera in wenigen Minuten einbinde. *Akzeptanz:* Dokumentation mit RTSP-URL, optional Snapshot-URL, `audio: false`, `vcodec: copy`, `maxStreams`.
- **US-HK-10 – Home-Assistant-Anleitung:** Als Home-Assistant-Nutzer möchte ich eine Anleitung für die HomeKit-Bridge im Accessory-Mode. *Akzeptanz:* Dokumentation mit `mode: accessory`, `support_audio: false`, `video_codec: copy`, `stream_count`, ein Accessoire pro Kamera.
- **US-HK-11 – Optionale Integration:** Als Nutzer ohne Apple-Umgebung möchte ich, dass der Gateway unverändert weiterläuft. *Akzeptanz:* keine neue Pflichtabhängigkeit im Container; Healthcheck unverändert; HomeKit-Doku separat.
- **US-HK-12 – Bewegungs-Benachrichtigungen (optional):** Als Nutzer möchte ich optional Bewegungs-Benachrichtigungen in Apple Home. *Akzeptanz:* nur anbieten, wenn eine echte Ereignisquelle existiert; `linked_motion_sensor` (HA) oder Motion-Accessoire (Homebridge); sonst ausdrücklich nicht zusagen.
- **US-HK-13 – Kopplung ohne HA/Homebridge (später):** Als Nutzer ohne Home Assistant oder Homebridge möchte ich die Kamera direkt mit Apple Home koppeln. *Akzeptanz:* HAP-Pairing, mDNS, SRTP, RTP, Snapshot; umfassende Tests; rechtliche Klärung.
- **US-HK-14 – Pairing sicher speichern (später):** Als Nutzer möchte ich, dass die HomeKit-Kopplung Neustarts überlebt, ohne dass Zugangsdaten in Logs landen. *Akzeptanz:* Pairing-Daten persistent und geschützt; kein Secret im Log; eigenes ADR.

## Bewusst zurückgestellt

- Internet-Relay und Fernzugriff
- Firmwareupdates über den Gateway
- Admin-Einstellungen und Passwortänderungen
- Cloud-Push-Nachrichten

Diese Funktionen erweitern Angriffsfläche und Scope erheblich und benötigen jeweils einen eigenen Änderungsvorschlag beziehungsweise ADR.
