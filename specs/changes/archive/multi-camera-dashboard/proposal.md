# Proposal: Multi-Camera-Dashboard

## Problem

Die Basisstation besitzt mehrere gekoppelte Kameras, die bisher nur über einen statischen Kanalwähler und ausschließlich nacheinander angezeigt werden konnten. Die Oberfläche enthielt redundante Diagnoseaktionen und skalierte weder auf große Bildschirme noch auf mobile Geräte optimal.

## Ziel

- gekoppelte Kanäle dynamisch vom Hub laden
- mehrere Kanäle gleichzeitig streamen
- Einzel- und Sammelaktionen
- Fullscreen pro Karte
- responsive Dashboarddarstellung
- Light/Dark-Theme
- technische Details ohne Belastung des Hauptflows
- vorhandene Geräte-/Firmware-Metadaten ohne generische Platzhalter
- Bewegungserkennung je Kamera schalten
- konfliktträchtige Inventar-/Einstellungsaktionen bei laufenden Streams sperren

## Erfolgskriterien

- reale Hub-Kameraliste liefert zwei Karten mit Gerätenamen
- beide HLS-Streams laufen parallel und besitzen eigene Statusmetrik
- kein Code wird persistent gespeichert
- Oberfläche ist ab 320 px Breite nutzbar und verwendet große Displays mit mehreren Spalten
- Kamera-MAC und andere nicht gelieferte Werte erzeugen keine UI-Felder
- Motion-Änderungen verwenden vollständiges Read-Modify-Write und werden verifiziert

## Status

Implemented.
