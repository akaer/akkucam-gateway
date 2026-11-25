# ADR-0023: Inter Tight für Markenname und Schritt-Headings

- Status: Accepted
- Datum: 2026-11-22

## Kontext

Die Oberfläche nutzte bisher dieselbe Systemschrift für alle Texte. Der Markenname „AkkuCam“ im Topbar und die großen Schritt-Headings („Basisstation auswählen“ und „X Kameras“) sollen sich typografisch stärker abheben. Gleichzeitig war der Hero-Textblock („Your cameras. At a glance.“ plus Erklärtext) zu prominent und wurde entfernt.

Die Randbedingungen bleiben: keine Cloud, kein CDN. Die Schrift muss lokal ausgeliefert werden, und ihre Lizenz muss sichtbar dokumentiert sein.

## Entscheidung

- Inter Tight wird als lokal gebündelte Web-Schrift unter `web/fonts/` ausgeliefert:
  - `web/fonts/inter-tight-latin.woff2` (Regular 400)
  - `web/fonts/inter-tight-bold-latin.woff2` (Bold 700)
- `web/style.css` bindet beide Schnitte per `@font-face` ein. `font-display: swap` sorgt für lesbaren Text während des Ladens. Es gibt keinen Google-Fonts- oder sonstigen CDN-Aufruf.
- Der Markenname im Topbar (`.brand`) nutzt Inter Tight. Der Name „AkkuCam“ wird durch das vorhandene `<strong>` fett (700) dargestellt; die Tagline bleibt in der normalen Schriftstärke.
- Auch die großen Schritt-Headings (`h2`, insbesondere „Basisstation auswählen“ und „X Kameras“) nutzen Inter Tight. Die Schriftfamilie ist über `--font-heading` zentral definiert.
- Der Hero-Textblock (Eyebrow, Headline und Erklärtext) wird entfernt. Der Discover-Button bleibt als einzige Aktion im Hero-Bereich erhalten.
- Die Schrift steht unter der SIL Open Font License 1.1. Der Lizenztext liegt als `web/INTERTIGHT-LICENSE.txt` bei und ist im Lizenzdialog verlinkt.
- Inter Tight erscheint im Lizenzdialog, in `specs/sources.md` und in `sbom/README.md`. Da die Schrift keinen Paketmanager-Eintrag hat, erscheint sie nicht automatisch in der SBOM.

## Begründung

- **Cloudfrei:** Lokale WOFF2-Dateien vermeiden externe Abhängigkeiten und funktionieren offline.
- **Lizenzklarheit:** Die OFL erlaubt die lokale Einbindung; der Lizenztext wird mitgeliefert und ist im Dashboard erreichbar.
- **Geringe Payload:** Zwei statische Schnitte reichen für Markenname und Schritt-Headings. Eine variable Schrift wäre flexibler, aber für diesen Einsatz unnötig größer.
- **Gezielter Einsatz:** Markenname und Schritt-Headings wechseln die Schrift. Die übrige Oberfläche bleibt bei der Systemschrift, was die Lesbarkeit und Konsistenz erhält.
- **Weniger Ablenkung:** Der entfernte Hero-Textblock macht den Einstieg direkter; der Discover-Button bleibt als klare Handlungsaufforderung.

## Betrachtete Alternativen

- **Google-Fonts-CDN einbinden:** Verstößt gegen die Cloudfrei-Vorgabe. Verworfen.
- **Variable Font ausliefern:** Größere Datei und mehr Komplexität für zwei benötigte Schnitte. Verworfen.
- **Hero-Headline behalten und nur die Schrift ändern:** Wurde zugunsten des aufgeräumten Einstiegs verworfen.
- **Markenname in der Systemschrift belassen:** Der gewünschte visuelle Akzent ginge verloren. Verworfen.

## Konsequenzen

- Neue Dateien: `web/fonts/inter-tight-latin.woff2`, `web/fonts/inter-tight-bold-latin.woff2`, `web/INTERTIGHT-LICENSE.txt`.
- `web/style.css` enthält zwei `@font-face`-Blöcke, die Variable `--font-heading` und die Typografie für `.brand` und `h2`.
- Der Hero-Textblock und die zugehörigen `hero.*`-i18n-Schlüssel entfallen. Der Discover-Button bleibt.
- Der Lizenzdialog, `specs/sources.md` und `sbom/README.md` führen Inter Tight mit OFL-1.1 auf.
- Neue Tests: `web/e2e/branding.spec.ts` prüft, dass Markenname und Schritt-Headings Inter Tight verwenden, die Schriftdateien und der Lizenztext ausgeliefert werden.
- Bei einem Wechsel der Schrift oder weiterer Schnitte müssen Lizenzdialog, `specs/sources.md` und die SBOM-Dokumentation aktualisiert werden.

## Nachtrag 2026-11-24

- Es bleibt bei zwei Schnitten (400/700). Andere Gewichte synthetisiert der
  Browser und sind in `style.css` unzulässig (ADR-0029).
- Der Geltungsbereich ist `h2` und der Markenblock. `h3`, Metrik-Überschriften
  und Fließtext nutzen die Systemschrift; die frühere Ausweitung auf alle
  `h1–h3` war ein Umsetzungsfehler.
- „Inter“ ist aus dem Body-Stack entfernt, weil die Datei nicht ausgeliefert
  wird.

## Verweise

- ADR-0003: lokal, ohne Cloud-Fallback
- ADR-0021: zweisprachige UI
- ADR-0022: gebündelte Web-Assets, Lizenzhinweise und SBOM
- `specs/sources.md`, `specs/architecture.md`, `specs/testing.md`
- `web/style.css`, `web/INTERTIGHT-LICENSE.txt`
