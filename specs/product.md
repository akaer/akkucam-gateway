# Produktspezifikation

## Zweck

Der Gateway ermöglicht den lokalen Zugriff auf AkkuCams mit Basisstation aus einem normalen Webbrowser. Er ersetzt für Livevideo die Abhängigkeit von der Hersteller-App und verwendet für Discovery, Sitzungsaufbau, Authentifizierung und Medienübertragung ausschließlich das lokale LAN.

## Zielnutzer

- Besitzer einer AkkuCam PPIC90000/PPIC90010/PPIC90200 mit Basisstation
- technisch versierte Heimnetz- und Home-Automation-Nutzer
- Betreiber, die den Videostream nicht über Hersteller-Relayserver transportieren möchten

## Muss-Anforderungen

1. Der Gateway **MUSS** Basisstationen lokal erkennen.
2. Der Gateway **MUSS** eine direkte LAN-PPCS-Sitzung ohne Oguard-Verzeichnisserver aufbauen.
3. Der Sicherheitscode **DARF NICHT** persistent gespeichert oder geloggt werden. Bei mehreren Basisstationen **MUSS** der Browser je Basisstation (DID) einen eigenen Code im RAM halten, damit unterschiedliche Codes gleichzeitig funktionieren.
4. Das H.264-Video **MUSS** ohne qualitätsmindernde Neukodierung weitergereicht werden.
5. Die Weboberfläche **MUSS** Kanal 1 und Kanal 2 auswählen können.
6. Die Weboberfläche **MUSS** aktuelle Nutzdatenbitrate und Kamera-Akkustand anzeigen.
7. Der Container **DARF NICHT** privilegiert laufen oder `NET_RAW` benötigen.
8. Der Gateway **DARF NICHT** selbst Oguard- oder Hersteller-Cloudendpunkte kontaktieren.
9. Der Gateway **MUSS** die gekoppelte Kanalzahl dynamisch vom Hub lesen.
10. Unterschiedliche Kamerakanäle **MÜSSEN** gleichzeitig angezeigt werden können.
11. Metadaten **DÜRFEN NUR** angezeigt werden, wenn ein konkreter Wert aus Hub, Kamera oder lokaler Netzwerktabelle vorliegt. Vorhandene Werte **MÜSSEN** vollständig lesbar sein; die Oberfläche **DARF** sie nicht per Ellipse abschneiden, sondern bricht sie bei Bedarf um.
12. Die Bewegungserkennung eines gekoppelten Kanals **MUSS** lokal ein- und ausgeschaltet werden können.
13. Inventar- und Einstellungszugriffe **MÜSSEN** während aktiver Streams abgelehnt werden.
14. Das Containerimage **MUSS** einen Healthcheck gegen die lokale Gateway-API bereitstellen.
15. Die Weboberfläche **MUSS** Deutsch und Englisch unterstützen. Sie erkennt die Browsersprache, fällt sonst auf Englisch zurück und lässt sich ohne Reload umschalten. Die Wahl wird in LocalStorage gespeichert; der Sicherheitscode niemals.
16. Lizenzinformationen **MÜSSEN** ohne separate Lizenzseite über das Info-/Hilfe-Icon als scrollbarer Dialog erreichbar sein. Der Dialog listet die MIT-Projektlizenz und alle im Laufzeitimage enthaltenen Drittanbieter-Komponenten mit Lizenz; Betriebssystempakete dürfen als Distribution zusammengefasst werden (ADR-0022), die Einzelaufstellung liefert die SBOM. Die API-Dokumentation **MUSS** dort verlinkt sein. Ein Copyright-Text und ein Seiten-Footer sind nicht erforderlich.
17. Das Projekt **MUSS** eine Software-Stückliste (SBOM) in CycloneDX und SPDX bereitstellen, die aus dem Laufzeitimage erzeugt werden kann.
18. Markenname und die großen Schritt-Headings **MÜSSEN** die lokal gebündelte Schrift Inter Tight verwenden. Ein externer Schriftdienst oder CDN **DARF NICHT** verwendet werden. Der Hero-Textblock entfällt; der Discover-Button bleibt. In der Setup-Ansicht kann zusätzlich „Zurück zur Live-Wand“ erscheinen.
19. Ein Info-Icon ganz rechts im Topbar (rechts neben der Theme-Umschaltung) **MUSS** ein modales Hilfe-Popup mit dem Ablauf (Basisstation finden, gekoppelte Kameras öffnen, Livebild ohne Cloud-Dienst starten) in der aktiven Sprache öffnen. Das Popup **DARF** keine Daten speichern und **MUSS** ohne externe Ressourcen funktionieren.
20. Die Streams **aller** geladenen Basisstationen **MÜSSEN** gleichzeitig in einer gemeinsamen Live-Wand sichtbar sein. Beim Laden einer weiteren Basisstation **DARF** kein bestehender Stream ersetzt werden. Globale Aktionen **MÜSSEN** alle geladenen Hubs gemeinsam starten und beenden; jede Basisstation bleibt über den Hub-Manager zusätzlich einzeln steuerbar. Der Hub-Manager nennt Modell, DID, IP und – falls ermittelt – die MAC-Adresse.
21. Nach dem Laden der ersten Kanalliste **MUSS** die Live-Wand die Setup-Seite ersetzen. „Basisstation hinzufügen“ **MUSS** die Setup-Seite gezielt öffnen und während aktiver Streams gesperrt sein.
22. Die Wand **MUSS** die Ansichten Auto-Matrix, Fokus und Komfort anbieten. Auto und Komfort **MÜSSEN** gleich große Kacheln erzeugen; bei ungerader Anzahl **MUSS** eine sichtbare leere Rasterzelle entstehen. Bei zu vielen Kameras **DARF** die Wand scrollen. Alle Ansichten beginnen oben.
23. Ein Klick auf Video oder Kachel **MUSS** einen Details-Drawer mit Qualität, Bewegungserkennung, Start/Stop und vorhandenen Metadaten öffnen. Laufende Streams **MÜSSEN** Bitrate, FPS und Akku als Overlay unten links mit ausschließlich vorhandenen Werten anzeigen. Die Laufzeit steht im Drawer. Das Overlay **MUSS** in den Wand-Einstellungen abschaltbar sein (Standard: an).
24. Auf iPhone-Größen **MUSS** eine kompakte Wand-Leiste mit Filter, globalem Start/Stop und einem Menü-Sheet die volle Leiste ersetzen; iPad und Desktop behalten die volle Leiste. Die gewählte Ansicht und die Sichtbarkeit des Live-Overlays **MÜSSEN** in LocalStorage gespeichert und über die Einstellungen zurücksetzbar sein.
25. Ansichts-, Filter- und Sprachwechsel **DÜRFEN** laufende Streams nicht unterbrechen; das jeweilige Videoelement **MUSS** erhalten bleiben.

## Aktueller Funktionsumfang

- WAI- und PPCS-LAN-Discovery
- vollständige DID- und Modellerkennung
- PUNCH/P2P_RDY-Sitzungsaufbau
- AES-Challenge-Authentifizierung
- Kamera-Wake-up durch Start des Liveviews
- H.264 Annex-B Empfang, Reassembly und HLS-Ausgabe
- Full HD, Auto, HD und SD als anforderbare Qualitätsmodi
- dynamische Hub-Kameraliste mit Namen, Modellen und vorhandenen Firmware-/MCU-Versionen
- lokal gelesene Hub-Metadaten und optionale Hub-MAC aus der ARP-Tabelle
- Bewegungserkennung pro Kamera als verifiziertes Read-Modify-Write
- parallele Anzeige mehrerer gekoppelter Kameras
- Markenname und Schritt-Headings in lokal gebündelter Inter Tight; Hero-Textblock entfernt
- Info-Icon im Topbar öffnet ein zweisprachiges Hilfe-Popup mit dem Ablauf
- Sicherheitscodes je Basisstation im Browser-RAM, mit „Code entfernen“ pro Hub und „Alle Sicherheitscodes aus dem Browser-Arbeitsspeicher entfernen“ (zwei Klicks) global
- gemeinsame Live-Wand mit Hub-Filterchips und Hub-Manager; globale Start-/Stopp-Aktionen für alle Hubs, zusätzlich pro Hub
- Vollbreite Setup-Leiste mit Empty State, Hub-Raster, immer sichtbaren Hub-Metadaten und Diagnose; kein zentrierter 1560-px-Container
- Auto-Matrix, Fokus und Komfort mit gleich großen Kacheln, leeren Rasterzellen und Ausrichtung am oberen Rand
- themeabhängige Kachelrahmen (hell/weich im Light Theme, dezent im Dark Theme)
- Details-Drawer pro Kamera (Desktop rechts, Mobile als Bottom-Sheet) mit Qualität, Bewegung, Start/Stop und Metadaten
- Live-Daten-Overlay für laufende Streams (kbit/s, FPS, Akku); nur vorhandene Werte, in den Einstellungen abschaltbar
- Live-Miniaturbilder im Fokus-Streifen aus den vorhandenen Videoelementen
- kompakte iPhone-Wand-Leiste mit Filter-Pill, globalem Start/Stop und Menü-Sheet
- Lizenzdialog über das Hilfe-Popup; kein Seiten-Footer, keine separate Lizenzseite
- gewählte Wand-Ansicht und Overlay-Sichtbarkeit in LocalStorage, mit „Auf Standard zurücksetzen“
- Bitrate des H.264-Nutzdatenstroms im Details-Drawer und Live-Daten-Overlay (kbit/s)
- Akkukapazität im Live-Daten-Overlay; Ladezustand zusätzlich im Details-Drawer
- Start, Status und Stop über HTTP-API

## Nicht-Ziele des aktuellen Stands

- Fernzugriff über das Internet
- Oguard-Relay- oder Cloudintegration
- Benutzer- und Rechteverwaltung im Gateway
- Aufzeichnung oder persistente Speicherung von Videos/Snapshots im Gateway; die Basisstation zeichnet selbst auf, sofern für den Kanal die Bewegungserkennung aktiv und eine SD-Karte vorhanden ist
- Audioempfang, Audioausgabe, Gegensprechen oder Ereignisarchiv (siehe [ADR-0020](adrs/0020-no-audio-support.md))
- vollständiger Ersatz aller App-Einstellungen
- gleichzeitige Ausgabe beider Kamerakanäle in einer einzelnen PPCS-Sitzung; jeder Kanal verwendet bewusst eine eigene Sitzung
- serverseitige Lokalisierung der API; `error` bleibt englisch, die UI übersetzt stabile Fehlercodes

Der Zugriff auf die von der Basisstation erzeugten Aufzeichnungen ist nicht Teil des aktuellen Stands. Er ist als Roadmap-Punkt vorgesehen.

## Qualitätsdefinition

Der Gateway verändert die vom Gerät gelieferte Bildqualität nicht. `Full HD` fordert den Gerätequalitätsmodus `1` an. In der verifizierten Installation liefert dieser Modus H.264 Main Profile mit 1920×1080 bei 25 fps. Bitrate und Bildqualität bleiben von Szene, Funkverbindung, Kamera-Firmware und kamerainterner Kodierung abhängig.
