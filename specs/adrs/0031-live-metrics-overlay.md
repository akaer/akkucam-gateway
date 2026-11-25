# ADR-0031: Live-Daten-Overlay und Sichtbarkeitseinstellung

- Status: Accepted
- Datum: 2026-11-24

## Kontext

Die Live-Metriken lagen hinter einem Popover, das nur laufende Streams über
einen Knopf „Live-Daten“ öffneten (ADR-0027). In der Praxis wurde es kaum
gefunden, und die Bitrate in Mbit/s mit einer Nachkommastelle war zu grob: Bei
einem Nutzdatenstrom zwischen etwa 0,1 und 1,5 Mbit/s sprang die Anzeige
zwischen „0 Mbit/s“ und „1,5 Mbit/s“ und verlor jede Auflösung.

Gleichzeitig stand auf jeder Kachel im Videobereich ein Pill mit Modell und
DID-Kürzel (`SMH-T1R · LYHFC`). Bei mehreren Kameras derselben Basisstation
wiederholte er sich auf jeder Kachel und konkurrierte mit Kanalnummer und
Kameranamen.

ADR-0027 hatte dauerhaft sichtbare Metriken ausdrücklich verworfen, weil sie
ADR-0015 widersprechen und Videofläche kosten. Diese Entscheidung wird hier
bewusst revidiert: Eine einzelne Textzeile kostet wenig Fläche und ersetzt
einen zusätzlichen Bedienschritt.

## Entscheidung

### Kachelinhalt

- In **Auto** und **Fokus** zeigt die Kachel im Video nur noch Kanalnummer,
  Kameranamen und Status. Das Hub-Kürzel entfällt, auch im Fokus-Streifen.
- **Komfort** behält die Zeile mit dem Hub-Kürzel unter dem Kameranamen, weil
  sie außerhalb des Videobereichs liegt.
- Das `aria-label` der Kachel nennt weiterhin Kamera und Basisstation.

### Overlay

- Laufende Streams zeigen unten links ein dezentes Overlay:
  `412 kbit/s · 25 fps · 87 %`.
- Angezeigt werden nur vorhandene Werte; fehlt ein Wert, entfällt sein
  Segment. Ohne Messwerte bleibt das Overlay leer.
- Die Laufzeit bleibt ausschließlich im Details-Drawer. Der Drawer zeigt
  weiterhin Bitrate, FPS, Akku (inklusive Ladezustand) und Laufzeit.
- Die Bitrate wird überall in `kbit/s` dargestellt, auch oberhalb von 1000
  (`1480 kbit/s`). Es gibt keinen Einheitenwechsel.
- Das Overlay ist `aria-hidden` und `pointer-events: none`: Es ist Dekoration
  über dem Video, die zugängliche Quelle bleibt der Drawer. Klicks auf die
  Kachel öffnen weiterhin die Details.
- Aktualisierung: `updateStreamStatus` schreibt einmal pro Sekunde den Text;
  es wird kein Markup neu erzeugt.
- Optik: 11 px, weiße Schrift, tabellarische Ziffern, 22 px hoch und damit
  auf einer Linie mit dem Vollbild-Knopf. Hinter dem Text liegt eine zu 55 %
  schwarze Fläche mit 6 px Radius. Sie ist nötig, weil Kameraszenen beliebig
  hell sein können: Auf reinem Weiß ergibt das noch 4,5:1, mit reinem
  Textschatten wären es unter 2:1. Die Fläche bleibt auf die Textbreite
  begrenzt und verschwindet, solange keine Werte vorliegen (`:empty`).

### Sichtbarkeit

- Die Wand-Einstellungen (Zahnrad, auch im mobilen Sheet) bekommen einen
  Schalter „Live-Daten auf Kacheln“. Standard ist **an**.
- Persistenz als vierte UI-Präferenz `localStorage['akkucam-wall-metrics']`.
  Gespeichert wird nur `off`; ein fehlender Eintrag bedeutet an.
- „Auf Standard zurücksetzen“ setzt Ansicht **und** Overlay zurück.
- Der Schalter wirkt sofort über eine Body-Klasse; es wird kein Stream und
  kein Videoelement angefasst.

## Begründung

- kbit/s liefert die Auflösung, die Mbit/s mit einer Nachkommastelle fehlt.
- Eine Zeile am unteren Rand ist schneller erfassbar als ein Popover und
  verdeckt bei 16:9-Kacheln keinen relevanten Bildinhalt; die kleine Fläche
  dahinter stellt den Kontrast unabhängig vom Bildinhalt sicher.
- Der Hub steht weiterhin im Drawer, im Filterchip und in der Komfort-Ansicht;
  innerhalb des Videos ist er bei einer Kamera pro Hub reine Wiederholung.
- Ein Schalter erhält die Wahlfreiheit, die ADR-0015 schützen wollte, ohne
  allen Nutzern die Metriken vorzuenthalten.

## Betrachtete Alternativen

- **Mbit/s mit zwei Nachkommastellen:** löst die Sprünge, bleibt aber eine
  unhandliche Einheit für einen Strom unter 2 Mbit/s. Verworfen.
- **Popover behalten und nur die Einheit ändern:** behebt das Auffinden nicht.
  Verworfen.
- **Overlay immer an, ohne Schalter:** einfacher, nimmt aber die in ADR-0015
  dokumentierte Präferenz für ein ruhiges Videobild ernst. Verworfen.
- **Schalter je Ansicht:** drei Zustände für eine Eigenschaft, die nichts mit
  dem Layout zu tun hat. Verworfen.
- **Overlay oben rechts neben dem Status:** kollidiert mit dem Status-Pill und
  dem Bildinhalt; unten links ist die ruhigste Ecke. Verworfen.

## Konsequenzen

- ADR-0015 gilt für die Wand nur noch eingeschränkt (Status: `Partially
  superseded`); der Drawer bleibt der progressive Weg.
- Der Popover-Abschnitt in ADR-0027 ist ersetzt: kein `#metrics-popover`, kein
  Trigger, kein Klick-Guard, kein Fokus-Handling.
- `web/AGENTS.md` erlaubt vier Storage-Keys statt drei.
- Neue Katalogschlüssel: `camera.metric.kbpsValue`, `settings.liveData`,
  `settings.liveDataHint`; entfernt: `camera.liveDataClose`,
  `camera.metricsPending`, `camera.metric.mbpsValue`.
- Playwright: Overlay-Inhalt, Klick auf das Overlay öffnet den Drawer,
  Hub-Kürzel nur in Komfort, Schalter inklusive Reset.
- `product.md`, `architecture.md` und `testing.md` werden mitgezogen.

## Verweise

- ADR-0014: nur verfügbare Metadaten
- ADR-0015: progressive Offenlegung (teilweise ersetzt)
- ADR-0021: zweisprachige UI
- ADR-0027: Live-Wand (Popover-Abschnitt ersetzt)
- ADR-0028: Terminologie
- ADR-0030: Barrierefreiheit
- `web/app.js`, `web/index.html`, `web/style.css`, `web/locales/`
