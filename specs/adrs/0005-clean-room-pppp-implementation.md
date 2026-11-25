# ADR-0005: Native Clean-Room-Reimplementierung von PPPP

- Status: Accepted
- Datum: 2026-11-21

## Kontext

Die Hersteller-App enthält `libPPCS_API.so` nur für Android/ARM. Der Dockerhost ist Linux/x86-64. Die Bibliothek ist proprietär und an Android-Systembibliotheken gebunden.

## Entscheidung

Die benötigte lokale Teilmenge des Protokolls wird nativ in Go implementiert: Discovery, PUNCH, P2P_RDY, DRW, ACK, Keepalive, Authentifizierung, IO-Control und Medien-Reassembly.

## Begründung

- vermeidet Android- und ARM-Emulation
- keine proprietären Binärdateien im Image
- nachvollziehbarer und testbarer Code
- nur tatsächlich benötigter lokaler Protokollumfang
- Live-Mitschnitte bestätigen den Ablauf unabhängig

## Betrachtete Alternativen

- ARM-Bibliothek unter QEMU/Bionic: fragil und schwer zu verteilen.
- Fremdgerät als Bridge: zusätzliche Hardware und weiter bestehende Herstellerabhängigkeit.
- vollständige PPPP-Cloudimplementierung: außerhalb des lokalen Scopes.

## Konsequenzen

- Verantwortung für Retransmission, Reassembly und Kompatibilität liegt beim Projekt.
- weitere Firmwarevarianten können Protokollanpassungen benötigen.
- Reverse-Engineering-Provenienz und Drittquellen werden dokumentiert.
