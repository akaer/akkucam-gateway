# ADR-0027: Live-Wand mit Ansichten, Details-Drawer und mobiler Steuerleiste

- Status: Accepted
- Datum: 2026-11-23

## Kontext

Die bisherige Oberfläche zeigte pro geladener Basisstation einen eigenen Workspace
untereinander. Die Kamerakarten hatten je nach Anzahl unterschiedliche Größen, die
Setup-Seite blieb über den Streams sichtbar, und auf großen Displays blieb viel
Fläche ungenutzt. Für mehrere Hubs und für iPhone/iPad fehlte eine gemeinsame,
vorhersehbare Anordnung. Drei klickbare Prototypen haben die Varianten
geprüft; die Umsetzung folgt der gemeinsamen Live-Wand.

## Entscheidung

### Flow

- Solange keine Basisstation geladen ist, bleibt die Setup-Seite sichtbar
  (Discovery, Hub-Auswahl, Diagnose).
- Setup und Wand nutzen dieselbe Vollbreiten-Hülle: kompakter Hinweisbalken,
  Werkzeugleisten-Stil, 16-px-Ränder und kein zentrierter 1560-px-Container.
  Die Setup-Seite zeigt eine Setup-Leiste mit Discover/„Zurück zur Wand“, einen
  vollbreiten Empty State und darunter das Hub-Raster bzw. die Diagnose.
- Nach dem Laden der ersten Kanalliste ersetzt die Live-Wand die Setup-Seite.
- „Basisstation hinzufügen“ öffnet die Setup-Seite gezielt wieder. Der Flow ist
  gesperrt, solange Streams laufen; nach erfolgreichem Laden kehrt die Wand zurück.
- Diagnose und manueller Kanal-Fallback bleiben Teil der Setup-Seite.

### Gemeinsame Wand

- Alle geladenen Hubs erscheinen in **einer** Wand. Hub-Filterchips blenden Hubs
  aus, ohne Streams zu beenden.
- Ein Hub-Manager-Dialog zeigt alle geladenen Hubs mit vollem Modell/DID, Code-
  Status, Kamerazahl, Live-Zahl sowie Start/Stop, „Code entfernen“ je Hub und
  einem globalen „Alle Sicherheitscodes entfernen“.
- Globale Aktionen starten und beenden weiterhin alle Hubs gemeinsam (ADR-0026).

### Ansichten

- **Auto-Matrix:** Spalten und Zeilen werden aus Kameraanzahl und verfügbarer
  Fläche berechnet. Alle Kacheln sind gleich groß; ungerade Anzahlen ergeben
  sichtbare leere Rasterzellen. Reicht die Fläche nicht für lesbare Kacheln,
  scrollt die Wand.
- **Fokus:** eine große aktive Kachel und ein Streifen mit allen anderen Kameras
  (Desktop rechts, Mobile unten). Der Streifen wählt die aktive Kamera und zeigt
  für laufende Streams Miniatur-Livebilder, die aus den vorhandenen Videoelementen
  auf Canvas gezeichnet werden. Es wird keine zusätzliche HLS-Instanz gestartet.
- **Komfort:** gleiche Matrix wie Auto, aber mit fester Kopf- und Steuerzeile
  pro Kachel (Qualität, Bewegung, Start/Stop).
- Alle Ansichten beginnen oben unter der Wand-Leiste; horizontal bleibt das
  Raster zentriert.
- Die Auswahl wird als `localStorage['akkucam-wall-view']` gespeichert. Der
  Einstellungsdialog bietet „Auf Standard zurücksetzen“ (Auto) und entfernt den
  Eintrag wieder. Es werden keine Zugangsdaten gespeichert.

### Kachel und Video

- Eine Kachel zeigt Video oder Platzhalter, Kanalnummer, Namen, kurzes Hub-
  Kürzel und Status. Auto/Fokus nutzen ein Overlay; Komfort nutzt Kopf- und
  Fußzeile.
- Klick auf Video/Kachel öffnet den Details-Drawer (Desktop rechts, Mobile als
  Bottom-Sheet): Qualität, Bewegungserkennung, Start/Stop, Vollbild, vorhandene
  Gerätedaten und Live-Metriken.
- Der Vollbild-Button sitzt in der Kachel; native Videosteuerung erscheint im
  Vollbild, nicht in der kleinen Kachel.
- Beim Wechsel von Ansicht, Filter oder Sprache werden **keine Videoelemente
  neu erzeugt**. Jeder Stream hält sein `<video>`-Element und seine HLS-Instanz;
  nicht sichtbare Videos liegen im offscreen `#stream-holder`, sichtbare werden
  an die neue Kachel gehängt. Streams laufen weiter.
- Kachelrahmen und Schatten sind themeabhängig: im Light Theme hell und weich,
  im Dark Theme dezent. Schwarz bleibt dem Videobereich vorbehalten.

### Live-Daten-Popover

- Nur laufende Streams zeigen den Trigger „Live-Daten“.
- Das Popover ist standardmäßig geschlossen und zeigt ausschließlich vorhandene
  Werte (Bitrate, FPS, Akku, Uptime; ADR-0014). Kanal- und Hub-Name entfallen,
  weil sie bereits auf der Kachel stehen.
- Erneuter Trigger-Klick, × oder Klick außerhalb schließt es. Ein Klick außerhalb
  auf eine Kachel schließt nur das Popover und öffnet keine Seitenleiste;
  Topbar-/Toolbar-Aktionen bleiben bedienbar. Escape bleibt Tastatur-Fallback.
- Der Fokus geht beim Schließen per ×/Escape zurück auf den Trigger.
- Die Kachelhöhe bleibt unverändert. Live-Metriken bleiben auch im Drawer
  verfügbar (ADR-0015).

### Mobile Steuerleiste

- Nur für iPhone-Größen: Hochformat `≤ 640 px`, Querformat `≤ 1000 × 560 px`.
  iPad und Desktop behalten die volle Wand-Leiste.
- Kompakte Leiste: Filter-Pill, globaler Start/Stop-Toggle und „⋯“.
- Das Bottom-Sheet enthält Darstellung, Hub-Filter, globale Aktionen,
  Hub-Verwaltung und Einstellungen.
- Hochformat nutzt eine Spalte und scrollt; Querformat nutzt die beste Matrix.

### Weitere Entscheidungen

- Die Topbar behält Sprache, Theme und Hilfe (Hilfe ganz rechts, ADR-0024).
  Die Wand-Einstellungen liegen in einem Zahnrad der Wand-Leiste bzw. im
  mobilen Sheet.
- Im Wandmodus und auf der Setup-Seite gibt es **keinen Seiten-Footer**.
  Lizenz- und API-Informationen öffnen sich über das Info-/Hilfe-Icon als
  scrollbarer Lizenzdialog. Die separate `licenses.html` und die Lizenzlinks im
  Einstellungsdialog entfallen; Copyright und MIT-Footerlink werden nicht mehr
  angezeigt. Der Dialog listet die MIT-Projektlizenz und alle Drittanbieter.
- Sicherheitscodes bleiben ausschließlich im Browser-RAM (ADR-0012/ADR-0025).
- Keine API-, OpenAPI- oder SBOM-Änderung. Die vorhandenen Live-Endpunkte
  liefern weiterhin pro DID und Kanal.

## Betrachtete Alternativen

- **Bestehende Workspace-Karten nur responsiver machen:** löst weder ungleiche
  Kachelgrößen noch den Platzverbrauch und lässt die Setup-Seite über den
  Streams stehen.
- **Tabs pro Hub:** würde wieder exklusiv umschalten; ADR-0026 verlangt parallele
  Sichtbarkeit. Filterchips und eine gemeinsame Wand erfüllen das ohne Verlust
  der Hub-Zuordnung.
- **Dauerhaft sichtbare Live-Metriken in jeder Kachel:** widerspricht ADR-0015
  und kostet Videofläche. Das Popover ist der explizite Einstieg.
- **Native Popover-API:** Light-Dismiss ließ den Folgeklick auf die Kachel die
  Seitenleiste öffnen. Ein eigenes Overlay mit kontrolliertem Klick-Handling ist
  auf Desktop und iPhone vorhersehbar.
- **Feste zwei Spalten im iPhone-Querformat:** zeigt bei fünf Kameras nur eine
  Reihe. Die Auto-Matrix nutzt die Höhe besser.

## Konsequenzen

- `web/index.html`: Setup- und Wandbereich, Wand-Leiste, mobile Leiste,
  Details-Drawer, Einstellungen, Hub-Manager, Metrik-Popover, Lizenzdialog.
- `web/app.js`: Wand-Rendering, Matrixberechnung, Video-/HLS-Erhalt,
  Offscreen-Videohalter, Fokus-Miniaturen, Filter, Ansichten, Drawer, Popover,
  mobile Leiste, Hub-Manager.
- `web/style.css`: Wand-, Kachel-, Drawer-, Popover-, Lizenz- und Mobile-Stile
  inklusive themeabhängiger Kachelrahmen.
- `web/locales/{en,de}.json`: neue Schlüssel für Wand, Setup, Ansichten,
  Einstellungen, Hub-Manager, Live-Daten, Lizenzen und mobile Leiste.
- `web/e2e/wall.spec.ts`: neue deterministische Suite; bestehende Suiten an die
  Wand angepasst. Die frühere `licenses.html` und der Seiten-Footer entfallen.
- ADR-0013, ADR-0015, ADR-0022 und ADR-0026 werden durch diesen Beschluss ergänzt.

## Nachtrag 2026-11-24

Das Kachel-Overlay hat in `web/style.css` `z-index: 1`. Das `<video>` wird nach
dem Overlay in `.tile-media` eingehängt; ohne eigene Stapelebene legt Chromium
es beim Hit-Testing über die Overlay-Buttons, sodass „Live-Daten“ und Vollbild
nicht klickbar waren. Der Playwright-Test „closes live data on trigger and
outside without opening details“ deckt das ab.

## Nachtrag 2026-11-24

- Die Hülle nutzt `--gutter` (16 px). Der Wand-Canvas setzt `--canvas-gap`
  (12 px, mobil 8 px) ein, damit Kacheln und Ränder nicht doppelt zählen. Die
  frühere Formulierung „16-px-Ränder“ meinte die Hülle, nicht jede Innenkante.
- Der Layout-Status in der Wand-Leiste („3 × 2 · 4 Kameras (scrollt)“) ist
  Teil der Leiste: Er beschreibt Raster und Scrollbedarf, ist unter 980 px
  ausgeblendet und zählt nicht als Gerätemetadatum.
- Die Live-Daten zeigen die Laufzeit als formatierten Wert
  (`uptimeSeconds`), nicht als Serverzeichenkette (ADR-0021).

## Nachtrag 2026-11-24 (Live-Overlay)

- Das Hub-Kürzel entfällt im Video der Ansichten Auto und Fokus sowie im
  Fokus-Streifen. Komfort behält es unter dem Kameranamen.
- Den Live-Daten-Popover gibt es nicht mehr. Laufende Streams zeigen
  `412 kbit/s · 25 fps · 87 %` unten links als dezentes Overlay; die
  Laufzeit bleibt im Drawer (ADR-0031).
- Die Bitrate wird in `kbit/s` angezeigt, nicht in `Mbit/s`.
- Die Sichtbarkeit ist in den Wand-Einstellungen schaltbar
  (`akkucam-wall-metrics`, Standard an).

## Nachtrag 2026-11-24 (Leiste und Setup)

- Der Layout-Status („3 × 2 · 4 Kameras (scrollt)“) entfällt aus der
  Wand-Leiste. Raster und Scrollbedarf bleiben intern, sind aber kein
  Nutzertext mehr.
- Der Hub-Manager nennt zusätzlich zur IP die MAC-Adresse, sofern sie aus der
  Kanalliste oder der ARP-Tabelle bekannt ist.
- Die Setup-Karte zeigt die Hub-Metadaten offen statt hinter einem
  Aufklapper (ADR-0015).

## Verweise

- ADR-0011: Stream pro DID und Kanal
- ADR-0012/ADR-0025: Code nur im Browser-RAM
- ADR-0013: Responsive Dashboard und Theme
- ADR-0014: Nur vorhandene Metadaten
- ADR-0015: Progressive Disclosure
- ADR-0021: Zweisprachige UI
- ADR-0026: Mehrere Basisstationen
- `specs/product.md`, `specs/architecture.md`, `specs/testing.md`