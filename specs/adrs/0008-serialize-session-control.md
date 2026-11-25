# ADR-0008: Kontroll-Sessions werden serialisiert

- Status: Accepted
- Datum: 2026-11-21

## Kontext

Parallele Probe- und Authentifizierungsversuche führten zu sporadischen LAN-PUNCH-Timeouts. Die Basisstation und Broadcastantworten lassen sich bei konkurrierenden Sitzungen nicht eindeutig genug zuordnen.

## Entscheidung

Session-Probe, Authentifizierung und initialer Live-Sitzungsaufbau werden durch einen Prozess-Mutex serialisiert. Der laufende Medienstream arbeitet danach unabhängig.

## Begründung

- beseitigt beobachtete Race Conditions
- sehr geringe funktionale Einschränkung für eine Basisstation
- einfaches, gut nachvollziehbares Fehlermodell

## Betrachtete Alternativen

- Zuordnung nur nach dynamischem UDP-Port und DID: komplexer und weiterhin von Broadcasttiming abhängig.
- vollständig parallele Sessionmanager: erst bei mehreren unabhängigen Hubs relevant.

## Konsequenzen

- gleichzeitige neue Verbindungsversuche warten aufeinander.
- ein langsamer Timeout kann andere Kontrollaufrufe verzögern.
- bestehende Videodatenübertragung wird nicht blockiert.
