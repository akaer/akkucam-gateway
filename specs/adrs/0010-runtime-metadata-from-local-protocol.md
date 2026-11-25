# ADR-0010: Laufzeitmetadaten werden lokal ermittelt

- Status: Accepted
- Datum: 2026-11-21

## Kontext

Die Hersteller-App zeigt variable Streambitrate und Akkuinformationen je Kamera. Diese Werte sollen ohne Cloud verfügbar sein.

## Entscheidung

- Bitrate wird aus den pro Sekunde an FFmpeg geschriebenen H.264-Nutzdatenbytes berechnet.
- Batterie wird über IO-Control `40448` für den aktiven Kanal abgefragt.
- Abfrage erfolgt direkt nach Streamstart und danach alle 30 Sekunden.
- Status wird über `/api/live/status` geliefert und vom Browser ungefähr sekündlich aktualisiert.

## Begründung

- entspricht den tatsächlich verarbeiteten Videodaten
- keine plattformspezifische Netzwerkschnittstellenmessung
- Batterieformat ist aus Live-Mitschnitt und Hardwareantwort verifiziert
- keine Cloud- oder Telemetriedienste notwendig

## Konsequenzen

- angezeigte Bitrate enthält keinen Netzwerkoverhead.
- erste Bitrate kann bis zum nächsten Sekundentakt `0` sein.
- Batterie kann bis zur ersten Antwort fehlen.
