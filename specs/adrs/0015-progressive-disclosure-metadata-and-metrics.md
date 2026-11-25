# ADR-0015: Metadaten und Live-Metriken standardmäßig zuklappbar

- Status: Partially superseded (ADR-0031)
- Datum: 2026-11-21
- Geändert: 2026-11-23 (Nachtrag, ADR-0027)

## Kontext

Das Dashboard soll auf mobilen Geräten und großen Displays schnell erfassbar sein. Permanentes Anzeigen von Geräte- und Protokolldetails (Firmware, MCU, MAC, Live-Metriken) erzeugt visuelles Rauschen und lenkt vom Hauptzweck ab: Streams starten/stoppen und Livevideo sehen.

Gleichzeitig sollen die technischen Informationen weiterhin verfügbar bleiben, ohne versteckt zu sein.

## Entscheidung

- Hub- und Kamera-Metadaten werden als `details/summary` umgesetzt.
- Standardzustand ist **zugeklappt**.
- Ein klarer Indikator (Chevron) signalisiert die Klickbarkeit.
- Live-Metriken (z. B. Bitrate, Akku, Laufzeit) werden nur innerhalb dieses Bereichs angezeigt.
- Live-Metriken werden nur gerendert/aktualisiert, wenn der Bereich aufgeklappt ist.

## Begründung

- bessere Lesbarkeit im Standardfall
- progressive disclosure: Details nur bei Bedarf
- reduziert Layoutsprünge und „Datenflimmern“ durch Statuspolling
- passt zum bestehenden Diagnose-Pattern

## Konsequenzen

- Nutzer sehen Live-Metriken erst nach dem Aufklappen.
- Der Gateway bleibt ADR-0014-konform: Es werden nur konkrete Werte angezeigt; fehlende Werte ergeben keine Platzhalterzeilen.

## Nachtrag 2026-11-23

Die Live-Wand (ADR-0027) behält die progressive Offenlegung bei. Live-Metriken
bleiben standardmäßig geschlossen und sind über den Details-Drawer sowie ein
explizites „Live-Daten“-Popover erreichbar. Das Popover enthält nur vorhandene
Werte; es bleibt bis zum Öffnen unsichtbar und verändert die Kachelhöhe nicht.

## Verweise

- ADR-0014: Nur tatsächlich verfügbare Metadaten anzeigen
- ADR-0013: responsives Dashboard und Theme

## Präzisierung 2026-11-24

Die Aussage „Live-Metriken werden nur gerendert/aktualisiert, wenn der Bereich aufgeklappt ist“ betrifft die Darstellung. Der Client fragt `/api/live/status` weiterhin sekündlich für jeden laufenden Stream ab, um Stream-Ende und Metriken zu erkennen; aktualisiert werden aber nur ein geöffneter Drawer und ein geöffnetes Popover (`updateStreamStatus` in `web/app.js`).

## Nachtrag 2026-11-24

Für die Live-Wand gilt die progressive Offenlegung der Live-Metriken nicht
mehr: Laufende Streams zeigen Bitrate, FPS und Akku dauerhaft als Overlay auf
der Kachel (ADR-0031). Der Details-Drawer bleibt der ausführliche Weg und
zeigt zusätzlich die Laufzeit. Auf Wunsch lässt sich das Overlay in den
Wand-Einstellungen abschalten; dann gilt wieder der ursprüngliche Zustand.

## Nachtrag 2026-11-24 (Setup-Seite)

Auf der Setup-Seite stehen die Metadaten einer geladenen Basisstation sofort
offen; das `details/summary` entfällt. Das Kartenraster zeigt nur, was nicht
schon im Kartenkopf steht (Firmware, MCU, MAC, SD-Werte); DID, IP und Modell
bleiben im Kopf. Nur die Diagnose bleibt weiterhin zuklappbar. Der Drawer
bleibt der progressive Weg für Kameradaten.
