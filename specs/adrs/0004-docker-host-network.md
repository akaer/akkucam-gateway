# ADR-0004: Docker Hostnetwork für lokale Broadcasts

- Status: Accepted
- Datum: 2026-11-21

## Kontext

WAI und PPPP verwenden IPv4-Broadcasts auf UDP/7200 und UDP/32108. Docker-Bridge-Netze transportieren diese Broadcasts nicht zuverlässig in das physische LAN.

## Entscheidung

Docker Compose verwendet unter Linux `network_mode: host`.

## Begründung

- direkte und verifizierte Broadcast-Erreichbarkeit
- keine privilegierten Raw-Sockets erforderlich
- keine manuelle Portweiterleitung für dynamische UDP-Sitzungsports
- funktioniert auf dem vorhandenen NixOS-Dockerhost

## Betrachtete Alternativen

- Bridge plus UDP-Broadcast-Relay: zusätzliche Komponente und Konfiguration.
- macvlan/ipvlan: stärkere Isolation, aber komplexer und problematisch für Host↔Container-Kommunikation.
- `privileged`/Raw-Capture: unnötig und unsicherer.

## Konsequenzen

- geringere Netzwerkisolation des Containers.
- HTTP-Port bindet direkt am Host.
- primär Linux-spezifisches Compose-Deployment.
