# ADR-0029: Design-Tokens und ein Icon-Satz

- Status: Accepted
- Datum: 2026-11-24

## Kontext

`web/style.css` hatte 25 Schriftgrößen, 19 Radien, 9 Schatten und 7
Schriftgewichte bei genau zwei ausgelieferten Schnitten (400/700). Der
Browser synthetisierte Fettungen, dadurch wirkten dieselben Überschriften je
nach Plattform unterschiedlich. Nebenbei stand „Inter“ im Body-Stack, obwohl
die Schrift nie ausgeliefert wird. Icons kamen aus drei Quellen: Font-Awesome-
SVG (Marke), eigene Inline-SVGs (Hilfe, Zahnrad) und Textglyphen
(`◐ ⌄ ⋯ ▶ ■ ⛶ ▣ ×`), die je nach Systemfont unterschiedlich aussehen.

Die Marke bleibt Nord. Kontrast war jedoch nicht Teil der Definition: Das
Primärblau lag bei 3,50:1, der Hover bei 2,34:1, das Code-Badge bei 1,99:1.

## Entscheidung

### Farbe

`--accent` wird abgedunkelt und in vier Rollen getrennt:

| Token | Light | Dark | Rolle |
|---|---|---|---|
| `--accent` | `#476487` | `#88C0D0` | Fläche (Button, Markenquadrat, aktiver Chip) |
| `--accent-hover` | `#3B5473` | `#9ACBDA` | Hover der Fläche |
| `--accent-ink` | `#476487` | `#9ACBDA` | Text, Links, Status auf Surface |
| `--accent-text` | `#ECEFF4` | `#2E3440` | Text auf `--accent` |
| `--danger-ink` | `#A04A54` | `#DCA0A6` | Fehlertext |

`--accent-strong` entfällt; es diente gleichzeitig als Hover-Fläche und als
Textfarbe. `--danger` bleibt für Punkte und Rahmen.

Gemessene Kontraste (WCAG AA verlangt 4,5:1 für Text): Button 5,30:1
(vorher 3,50), Hover 6,74 (2,34), Badge 4,52 (1,99), Fehlertext 4,79 (3,36),
Dark-Badge 4,90 (4,14).

### Typografie

- Sechs Stufen: `--text-xs` 11 px, `--text-sm` 12, `--text-md` 13,
  `--text-base` 15, `--text-lg` 18, plus `clamp()` für `h2`.
- Zwei Gewichte: 400 und 700. Andere Werte erzeugt der Browser synthetisch und
  sind unzulässig.
- Inter Tight bleibt Markenname und `h2` vorbehalten (ADR-0023). `h3`,
  Metrik-Überschriften und Fließtext nutzen die Systemschrift.
- „Inter“ verschwindet aus dem Body-Stack, weil die Datei nicht ausgeliefert
  wird.

### Form und Tiefe

- Drei Radien: `--radius-sm` 10 px, `--radius-md` 14 px, `--radius-lg` 18 px;
  `--radius-pill` für Pillen.
- Vier Schatten: `--shadow-card`, `--shadow-popover`, `--shadow-modal`,
  `--shadow-thumb`.
- Zwei Blur-Ebenen: Topbar (18 px) und Dialog-Backdrop (4 px). Kachel-Chips
  auf laufenden Videos sind nicht mehr unscharf; das kostet GPU-Zeit ohne
  Informationsgewinn.
- Ein Gutter-Token (`--gutter` 16 px) für die Hülle, `--canvas-gap` 12 px für
  die Wand (mobil 8 px).

### Icons

- Ein Inline-Sprite in `index.html` (`<symbol>`), 24-px-Raster,
  `currentColor`, Größen 14/18/28 px über `.icon-sm`, `.icon` und `.icon-lg`.
- Keine Textglyphen als Icons, kein Icon-Font, kein CDN.
- Das Kamera-Symbol bleibt das Font-Awesome-Werk (CC BY 4.0) und wird an
  einer Stelle definiert; Favicon-Dateien bleiben separat.

## Begründung

- Skalen statt Einzelwerte machen Abweichungen sichtbar und Review-fähig.
- Zwei echte Schnitte verhindern, dass Fettung je nach Browser unterschiedlich
  aussieht.
- Ein Icon-Satz garantiert dieselbe Strichstärke, Ausrichtung und
  Theme-Farbe.
- Die Aufteilung von `--accent` in Fläche und Text ist die Voraussetzung für
  belastbare Kontraste in beiden Themes.

## Betrachtete Alternativen

- **Button-Text dunkel auf dem alten Blau:** hätte die Marke erhalten, aber
  den Charakter des Primärbuttons sichtbar verändert. Verworfen.
- **Weitere Inter-Tight-Schnitte nachliefern (600/800):** größere Payload und
  Lizenzpflege für Zwischenstufen, die die Skala nicht braucht. Verworfen.
- **SVG-Sprite als separate Datei:** externes `<use>` funktioniert, erzeugt
  aber einen zusätzlichen Request und eine zweite Fehlerquelle. Verworfen.
- **Icon-Font:** widerspricht der lokalen, lizenzpflichtigen Einzeldatei und
  skaliert schlechter. Verworfen.

## Konsequenzen

- `web/style.css` verwendet nur noch Tokens; `TestStylesheetUsesOnlyShippedFontWeights`
  erzwingt 400/700.
- `TestDashboardUsesOneIconSet` verbietet die alten Textglyphen in HTML und JS.
- Die Playwright-Branding-Suite misst Kontrast und Schriftgewicht im Browser
  für beide Themes.
- Kachel-Beschriftungen kürzen weiterhin per Ellipse; der vollständige Name
  steht im Details-Drawer. Für Metadatenfelder bleibt ADR-0014 maßgeblich
  (vollständig, umgebrochen).
- Änderungen an Farbe oder Skala benötigen eine Aktualisierung dieses ADR und
  der Kontrasttests.

## Nachtrag 2026-11-24

- Die sechs Schriftstufen sind in `rem` definiert (0.6875/0.75/0.8125/
  0.9375/1.125 rem); bei 16 px Root bleibt die Darstellung identisch, bei
  erhöhter Textskala wächst sie mit (ADR-0030).
- Kachel-Kopf und -Fuß nutzen `min-height` statt fester Höhen. Die Wand misst
  die tatsächliche Chrome-Höhe (`measureComfortChrome`) statt mit der
  Konstante 96 px zu rechnen.

## Verweise

- ADR-0013: responsives Dashboard und Theme
- ADR-0021: zweisprachige UI
- ADR-0023: Inter Tight für Markenname und Headings
- ADR-0027: Live-Wand
- ADR-0030: Barrierefreiheit
- `web/style.css`, `web/index.html`, `web/locales/`
