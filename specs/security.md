# Sicherheit und Datenschutz

## Vertrauensmodell

Der aktuelle Gateway ist für ein vertrauenswürdiges Heimnetz vorgesehen. Jede Person, die die HTTP-API erreichen kann, kann Discovery und Sessionversuche auslösen. Vor einer Veröffentlichung über VLAN- oder Internetgrenzen ist zwingend HTTP-Authentifizierung beziehungsweise ein authentifizierender Reverse Proxy erforderlich.

## Sicherheitscode

- wird vom Browser je Basisstation beim Laden der Kameraliste abgefragt
- bleibt bis Reload, Tab-Schließen oder explizitem Entfernen in einer JavaScript-Variable, adressiert über die DID der Basisstation (eine Map, ein Code pro Hub)
- wird für weitere Kanalstarts, Inventarabfragen und Motion-Änderungen an den Go-Prozess gesendet und dort nur für den jeweiligen Sitzungsaufbau gehalten
- ein von der Basisstation abgelehnter Code (`unauthorized`) wird im Browser verworfen, damit der nächste Versuch erneut fragt
- wird nicht in LocalStorage, SessionStorage, Cookies, Dateien, Datenbanken, HLS-Pfaden oder Logs gespeichert
- Themepräferenz und Sprachwahl dürfen in LocalStorage liegen; sie enthalten keine Zugangsdaten
- wird nach Handlerverarbeitung aus dem Request-Struct entfernt
- Go-Strings können nicht garantiert aktiv aus dem Speicher gelöscht werden; „nur im RAM“ bedeutet daher keine kryptografisch garantierte Speicherlöschung
- HTTP ist im LAN unverschlüsselt; ein Angreifer mit Mitschnittmöglichkeit kann den Code sehen

Für höhere Anforderungen soll TLS über einen lokalen Reverse Proxy eingesetzt werden.

## Cloudfreiheit

Der Gateway sendet keine Pakete an die beobachteten Oguard-Verzeichnis-, Relay-, Wake-up- oder Firmwareendpunkte. Der lokale Ablauf nutzt UDP/7200 und UDP/32108 sowie den dynamischen Hub-Port.

Diese Aussage gilt nur für den Gateway. Die unveränderte Basisstation kann unabhängig davon weiterhin eigene Internetverbindungen aufbauen. Vollständige Netztrennung erfordert Firewallregeln am Router, beispielsweise ausgehendes UDP/32100 und weiteren Internetzugriff der Hub-IP zu blockieren. Solche Regeln sind bewusst nicht Bestandteil des Containers.

## Medien

- Videonutzdaten sind im lokalen Altprotokoll als H.264 erkennbar und nicht Ende-zu-Ende verschlüsselt.
- Jeder Teilnehmer mit ausreichender LAN-Mitschnittposition kann potenziell Videodaten beobachten.
- HLS wird unverschlüsselt über HTTP ausgeliefert.
- HLS-Dateien sind temporär, aber während des Betriebs für alle HTTP-Clients erreichbar.

## Container

Positive Eigenschaften:

- Prozess läuft als unprivilegierter Benutzer
- kein `privileged: true`
- kein `CAP_NET_RAW`
- keine Docker-Socket-Mounts
- statisches Go-Binary
- keine persistente Secret-Konfiguration
- HTTP-Healthcheck enthält keine Gerätekennung oder Zugangsdaten

Akzeptierte Risiken:

- `network_mode: host` hebt Docker-Netzwerkisolation auf
- HTTP-Port 8080 bindet standardmäßig an alle Hostinterfaces
- FFmpeg verarbeitet nicht vertrauenswürdige Kamerabitstreams
- Alpine-/FFmpeg-Updates müssen regelmäßig eingespielt werden

## Empfohlene Härtung

1. Gateway nur in vertrauenswürdigem VLAN betreiben.
2. Port 8080 per Host-Firewall auf lokale Quellnetze begrenzen.
3. Reverse Proxy mit TLS und Authentifizierung vorschalten.
4. Internetzugriff der Basisstation blockieren, wenn Push-/Cloudfunktionen entbehrlich sind.
5. Containerimage regelmäßig neu bauen.
6. Keine realen Sicherheitscodes in Issues, Logs oder Versionsverwaltung aufnehmen.

## Bekannte Altprotokollrisiken

- AES-128-ECB
- nullaufgefüllter Sicherheitscode als AES-Key
- proprietäre XOR-Obfuskation statt moderner Transportverschlüsselung
- Broadcast-Discovery ohne Authentifizierung
- Geräte-DID im LAN sichtbar

Der Gateway reproduziert diese Mechanismen nur zur Interoperabilität und betrachtet sie nicht als moderne Sicherheitskontrollen.
