# ADR-0006: H.264-Passthrough und HLS für Browserwiedergabe

- Status: Accepted
- Datum: 2026-11-21

## Kontext

Die Kamera liefert H.264 Annex-B. Browser können diesen Elementarstream nicht direkt über eine normale HTTP-URL abspielen.

## Entscheidung

Der Gateway schreibt H.264 unverändert an FFmpeg. FFmpeg setzt PTS/DTS und remuxt mit `-c:v copy` nach lokalem HLS/MPEG-TS. Im Browser wird ein lokal gebündeltes hls.js verwendet.

## Begründung

- keine Qualitätsverluste durch Neukodierung
- geringe CPU-Last
- breite Browserunterstützung
- einfache HTTP-Auslieferung
- vorhandene FFmpeg-Paketierung

Der `setts`-Bitstreamfilter setzt monotone Zeitstempel, weil besonders Kanal 2 teilweise keine verwertbaren PTS/DTS liefert.

## Betrachtete Alternativen

- WebRTC: niedrigere Latenz, aber deutlich komplexere Signalisierung und Paketierung.
- MJPEG: hohe Bandbreite und notwendige Dekodierung/Neukodierung.
- RTSP: gut für NVRs, aber nicht direkt browserfähig.
- Transkodiertes HLS: unnötiger Qualitätsverlust und CPU-Bedarf.

## Konsequenzen

- mehrere Sekunden Latenz durch HLS-Segmentierung.
- FFmpeg wird Teil der Runtime und Angriffsfläche.
- HLS-Dateien sind temporär und werden fortlaufend gelöscht.
