# ADR-0030: Barrierefreiheit der Oberfläche

- Status: Accepted
- Datum: 2026-11-24

## Kontext

Die Oberfläche respektierte bereits `prefers-reduced-motion`, hatte
Fokusringe und native `<dialog>`-Elemente mit Fokusführung. Ungeregelt waren
jedoch Kontrastziele, Trefferflächen, die Semantik der Kacheln und das
Verhalten auf Geräten mit Notch.

Zwei konkrete Fehler fielen dabei auf:

- Kacheln waren `<article tabindex="0">` mit eigenem Klick- und
  Tastaturhandler, enthielten aber verschachtelte Buttons. Ein `role="button"`
  wäre damit ungültig, ein generisches Element ohne Rolle ist für
  Screenreader nicht benannt.
- `viewport-fit=cover` war gesetzt, aber `env(safe-area-inset-*)` kam nirgends
  vor. Bottom-Sheet und Topbar konnten auf dem iPhone unter der Notch oder dem
  Home-Indicator liegen.

## Entscheidung

### Kontrast

- Text ≥ 4,5:1 gegen seine Fläche, auch für Placeholder- und Status-Text.
- Nicht-Text-Elemente (Fokusring, Rahmen, Icons) ≥ 3:1.
- Geprüft werden mindestens Primärbutton, Fehlertext und Status-Badge in
  beiden Themes; die Werte stehen in ADR-0029.

### Trefferflächen

- Icon-only-Buttons: 44 × 44 px (`--icon-button`), Dialog- und
  Popover-Schließen eingeschlossen.
- Inline-Textaktionen: mindestens 24 px Höhe plus Abstand (WCAG 2.5.8), weil
  44 px in Tabellen und Zeilen das Layout sprengen würden.
- Schalter bleiben Teil ihrer beschrifteten Zeile; die Zeile ist die
  Trefferfläche.

### Safe Area

- `viewport-fit=cover` bleibt; Topbar und mobile Sheets/Drawer erhalten
  `env(safe-area-inset-top)` bzw. `env(safe-area-inset-bottom)`.

### Semantik der Kachel

- Die Kachel ist `role="group"` mit `aria-label` aus Kameraname und
  Basisstation. Enter und Leertaste öffnen den Details-Drawer; Maus und Touch
  behalten die ganze Fläche als Ziel.
- Verschachtelte Buttons bleiben echte Buttons und damit einzeln fokussierbar.
  Eine `button`-Rolle für die ganze Kachel ist ausgeschlossen.

### Fokus und Tastatur

- `:focus-visible` mit 3 px Ring und Offset in beiden Themes.
- Das Live-Daten-Popover setzt den Fokus auf sein Schließen, gibt ihn bei ×
  und Escape an den Trigger zurück und schließt bei Klick außerhalb.
- Dialoge bleiben native `<dialog>`-Elemente.

### Bewegung

- Animationen laufen nur bei erlaubter Bewegung; `prefers-reduced-motion`
  deaktiviert Dialog-, Popover- und Schalter-Transitions sowie den
  Verbindungs-Puls.

## Begründung

- Ein Dashboard, das im Wohnzimmer auf dem Telefon bedient wird, braucht
  belastbare Ziele und Lesbarkeit bei Sonnenlicht.
- Die Gruppenrolle beschreibt die Kachel korrekt, ohne die enthaltenen
  Bedienelemente zu verschlucken.
- Safe-Area-Fehler fallen im Emulator nicht auf, auf dem Gerät sofort.

## Betrachtete Alternativen

- **Kachel als `role="button"`:** ungültig mit verschachtelten Buttons;
  Screenreader würden die inneren Aktionen nicht mehr ansagen. Verworfen.
- **Eigener „Details“-Button in jeder Kachel:** zusätzliche sichtbare
  Bedienung für eine Fläche, die bereits als Ganzes klickbar ist. Verworfen.
- **44 px für alle Textlinks:** würde Zeilen in Drawer und Manager
  auseinanderziehen; WCAG verlangt 24 px. Verworfen.
- **Axe/Lighthouse als Abhängigkeit:** neues Werkzeug und ADR-Pflicht für
  wenige Prüfungen; die Messungen sind mit `getComputedStyle` deterministisch
  nachstellbar. Verworfen.

## Konsequenzen

- `web/style.css` enthält `env(safe-area-inset-*)`; `TestStylesheetRespectsSafeAreaInsets`
  prüft das, weil Playwright Insets nicht simulieren kann.
- Die Wall-Suite prüft `role="group"`, `aria-label` und das Öffnen per
  Enter.
- Die Branding-Suite misst Kontrast und Schriftgewicht im Browser.
- Neue Bedienelemente brauchen eine Trefferfläche von 44 px (Icon) bzw.
  24 px (Inline-Text) und einen Kontrastnachweis.
- Screenshots mit Notch bleiben ein manueller Prüfpunkt auf echter Hardware.

## Nachtrag 2026-11-24

- **Textskalierung:** `<meta name="text-scale" content="scale">` und relative
  Typografie; der Playwright-Test prüft 200 % ohne Überlauf oder abgeschnittene
  Kachel-Leisten.
- **Leichte Verwerfung:** Dialoge, die per Klick außerhalb schließen dürfen,
  tragen `closedby="any"`. Ohne native Unterstützung (Safari) prüft ein
  Pointer-Fallback die Zeigerposition, nicht `event.target`. Code- und
  Hilfe-Dialog bleiben bei `closerequest`, damit eine Eingabe nicht verloren
  geht.
- `dialog:modal` und `::backdrop` begrenzen Scrollgesten auf die oberste
  Ebene.
- **Überschriften:** Setup und Wand sind die einzige `h1` der Seite;
  Dialogtitel bleiben konsistent `h2`.
- **Codefeld:** Sichtbarkeitsschalter als `<button type="button">` mit
  `aria-pressed`, statischem Label „Sicherheitscode anzeigen" und verstecktem
  Warntext; er startet bei jedem Öffnen maskiert. Schließen-Buttons nutzen
  `data-close` statt Inline-JavaScript.

## Nachtrag 2026-11-24 (Overlay statt Popover)

Das Live-Daten-Popover entfällt (ADR-0031). Seine Fokusregeln sind damit
gegenstandslos. Das Overlay auf der Kachel ist `aria-hidden="true"` und
`pointer-events: none`: Die Sekunden-Updates sollen Screenreader nicht
unterbrechen, und die zugängliche Quelle bleibt der Details-Drawer. Der
Schalter in den Wand-Einstellungen sitzt in einer beschrifteten Zeile; die
Zeile ist die Trefferfläche, nicht der 38 × 22 px große Schalter allein.

## Verweise

- ADR-0013: responsives Dashboard und Theme
- ADR-0021: zweisprachige UI
- ADR-0027: Live-Wand
- ADR-0029: Design-Tokens und Icons
- `web/style.css`, `web/app.js`, `web/index.html`, `web/e2e/`
- `specs/testing.md`
