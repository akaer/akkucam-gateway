# ADR-0003: Ausschließlich lokaler Betrieb ohne Cloud-Fallback

- Status: Accepted
- Datum: 2026-11-21

## Kontext

Die Hersteller-App kontaktiert Oguard-Verzeichnis-, Push-, Firmware- und Relaydienste. Das Hauptziel ist ein lokaler Browserzugriff ohne Herstellerabhängigkeit.

## Entscheidung

Der Gateway implementiert ausschließlich die direkte LAN-Discovery und LAN-PPCS-Sitzung. Es gibt keinen automatischen oder optionalen Cloud-/Relay-Fallback.

## Begründung

- LAN_SEARCH/PUNCH/P2P_RDY funktioniert vollständig lokal.
- Video bleibt im Heimnetz.
- Ausfall oder Abschaltung externer Server beeinträchtigt den Gateway nicht.
- Weniger Datenschutz- und Sicherheitsrisiken.
- Weniger Protokoll- und Betriebsumfang.

## Betrachtete Alternativen

- Oguard nur für Rendezvous verwenden: technisch einfacher für Fernzugriff, widerspricht aber dem Cloudfrei-Ziel.
- eigener externer Relayserver: hoher zusätzlicher Sicherheits- und Betriebsaufwand.

## Konsequenzen

- Nutzung nur im gleichen Broadcast-/Routingbereich beziehungsweise mit explizitem Broadcast-Relay.
- kein Fernzugriff ohne separates VPN ins Heimnetz.
- die Basisstation selbst kann weiterhin Cloudverbindungen aufbauen; das muss außerhalb des Gateways per Firewall geregelt werden.
