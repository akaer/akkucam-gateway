# ADR-0002: Go für den nativen Gateway

- Status: Accepted
- Datum: 2026-11-21

## Kontext

Der Gateway benötigt UDP-Broadcasts, eine binäre Protokollzustandsmaschine, HTTP-API, Nebenläufigkeit und Prozesssteuerung für FFmpeg. Der Host läuft auf NixOS/x86-64; der Gateway soll als Dockercontainer portabel bleiben.

## Entscheidung

Der Gateway wird in Go implementiert. Er wird als statisches Linux-Binary in einer Multi-Stage-Docker-Buildstage erzeugt.

## Begründung

- Standardbibliothek deckt UDP, HTTP, AES und Prozesssteuerung ab.
- Goroutines passen zu Discovery, Sessionloop, Bitratenmessung und FFmpeg-Lifecycle.
- Statisches Binary vereinfacht das Runtime-Image.
- Die Datenrate von ein bis zwei Kamerastreams erfordert keine komplexere Optimierung.
- Schnelle Iteration ist beim Reverse Engineering wichtiger als maximale Low-Level-Kontrolle.

## Betrachtete Alternativen

### Rust

Vorteile bei speichersicherem Binärparsing und Zero-Copy. Zurückgestellt, weil der Engpass das Protokollverständnis und nicht CPU oder Speicher ist. Eine spätere isolierte Portierung von `internal/ppcs` bleibt möglich.

### Python/Node.js

Schnell für Prototypen, aber weniger attraktiv für einen kompakten langlebigen Netzwerkdaemon und ein minimales Runtime-Image.

### Fremdbibliothek als Bridge

Hätte eine proprietäre Herstellerbibliothek wiederverwendet, benötigt aber ARM-Emulation, mehr Privilegien und bringt Verteilungs- sowie Lizenzrisiken.

## Konsequenzen

- Host benötigt kein lokal installiertes Go.
- Go-Version und Build werden im Dockerfile festgelegt.
- Binärparser müssen weiterhin defensiv implementiert und getestet werden.
