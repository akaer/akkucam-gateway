# ADR-0013: Responsive Dashboard mit Diagnosebereich und Theme-Toggle

- Status: Accepted
- Datum: 2026-11-21
- Geändert: 2026-11-24 (Nachtrag, ADR-0030)

## Kontext

Die prototypische Oberfläche mischte Diagnoseaktionen und normalen Bedienfluss, nutzte große Displays schlecht und war nur begrenzt mobilfreundlich.

## Entscheidung

Die UI folgt dem Flow Discovery → Hub auswählen/Code eingeben → dynamische Streamkarten → Einzel-/Gesamtstop. Diagnoseinformationen und manueller Kanal-Fallback liegen in einem einklappbaren Bereich. Das Layout verwendet eine responsive Kartenmatrix, pro Karte nativen Fullscreen und ein Light/Dark-Theme.

Das Theme folgt beim ersten Besuch der Systempräferenz. Nur die Theme- und Sprachwahl wird in LocalStorage gespeichert; niemals der Sicherheitscode. Die Sprachwahl ist in ADR-0021 geregelt.

## Begründung

- klare Primäraktionen ohne redundante Probe-/Auth-Buttons
- effiziente Nutzung großer Displays durch parallele Karten
- einspaltige, touchfreundliche Darstellung für iPhone 14
- technische Details bleiben für affine Nutzer erreichbar
- Fullscreen nutzt Browser-/iOS-Funktionen statt eigener Playerimplementierung

## Konsequenzen

- hls.js bleibt lokal gebündelt.
- UI-Zustand ist flüchtig; Discovery und Kanalliste werden nach Reload neu geladen.
- manueller Kanal-Fallback ist bewusst nur im Diagnosebereich sichtbar.

## Nachtrag 2026-11-23

ADR-0027 ersetzt die Kartenmatrix und die dauerhaft sichtbare Setup-Seite durch
eine Live-Wand. Nach dem ersten Laden übernimmt die Wand; Auto-Matrix, Fokus und
Komfort lösen die bisherige Kartenansicht ab. Die Wand-Leiste, mobile
Kompaktleiste, Details-Drawer und Live-Daten-Popover sind dort beschrieben.
Setup und Wand nutzen dieselbe Vollbreiten-Hülle und den kompakten Hinweisbalken;
ein Seiten-Footer existiert nicht mehr. Theme- und Sprachpersistenz bleiben
unverändert; zusätzlich wird nur die Wand-Ansicht in LocalStorage gespeichert.

## Nachtrag 2026-11-24

- Der Zustand „System" ist der **attributlose** Fall: kein `data-theme`, die
  dunklen Tokens greifen über `@media (prefers-color-scheme: dark)`. Ohne
  JavaScript folgt die Seite damit der Systempräferenz.
- Der Zwei-Zustands-Schalter bildet „System" und „Gegenteil des aktuell
  Gerenderten" ab. Entspricht das Ziel der Systempräferenz, wird der
  LocalStorage-Eintrag **gelöscht** statt gespeichert; nur so ist der Rückweg
  zur Systemeinstellung möglich.
- `<meta name="color-scheme">` wird mit dem Zustand synchron gehalten
  (`light dark` im Systemfall), bleibt aber nur ein Hinweis; maßgeblich ist
  die CSS-Eigenschaft `color-scheme`.
- Ein `storage`-Event übernimmt Änderungen aus anderen Tabs.

## Nachtrag 2026-11-24 (Vierter Storage-Key)

Neben `akkucam-theme`, `akkucam-lang` und `akkucam-wall-view` speichert die
Oberfläche die Sichtbarkeit des Live-Overlays als
`akkucam-wall-metrics` (`off` oder nicht gesetzt). Es bleibt bei reinen
UI-Präferenzen; Zugangsdaten und Gerätedaten werden weiterhin nicht gespeichert
(ADR-0031).
