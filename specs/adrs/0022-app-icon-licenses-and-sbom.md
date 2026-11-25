# ADR-0022: App-Icon, Favicon, Lizenzhinweise und SBOM

- Status: Accepted
- Datum: 2026-11-22
- Geändert: 2026-11-23 (Nachtrag, ADR-0024/ADR-0027)

> **Aktualisiert 2026-11-23:** Footer, Copyright und die separate Lizenzseite
> wurden durch den scrollbaren Lizenzdialog über das Hilfe-Icon ersetzt; siehe
> Nachtrag am Ende.
>
> Der Entscheidungstext unten beschreibt den Stand vom 2026-11-22. Die
> Abschnitte **Lizenzhinweise** und die zugehörigen Konsequenzen sind durch
> ADR-0024 und ADR-0027 ersetzt und nur noch historisch zu lesen. Maßgeblich
> ist der Nachtrag.

## Kontext

Das Dashboard zeigte bisher nur einen Textbuchstaben „A“ als Markenzeichen. Es gab kein Favicon, keine Übersicht der gebündelten Drittkomponenten und keine Software-Stückliste (SBOM). Für ein lokal betriebenes Produkt sind sichtbare Lizenzhinweise und eine maschinenlesbare SBOM jedoch Voraussetzung für Weitergabe und Betrieb.

Die Randbedingungen sind streng: keine Cloud, kein CDN, keine Persistenz von Zugangsdaten. Das Icon muss also lokal ausgeliefert werden und dem Hell-/Dunkel-Theme folgen. Die SBOM muss ohne externe SaaS-Dienste erzeugbar sein.

## Entscheidung

### Icon und Favicon

- Das Markenzeichen ist das Font-Awesome-Kamera-Icon („camera“, Solid) als lokal gebündeltes SVG: `web/icon-light.svg` und `web/icon-dark.svg`.
- Im Dashboard wird dasselbe Icon inline als SVG im `.brand-mark` gerendert. Die Farbe kommt über `currentColor` aus den Theme-Variablen (`--accent`, `--accent-text`); ein Theme-Wechsel ändert das Icon ohne Neuladen.
- Das Favicon ist dasselbe Kamera-Icon. Zwei Varianten (hell/dunkel) werden über `<link id="favicon" rel="icon">` eingebunden. Der Inline-Theme-Bootstrap im `<head>` und `applyTheme()` in `app.js` tauschen das `href` passend zum aktiven Theme.
- Es gibt keinen CDN-Verweis. Die Icon-Dateien liegen unter `web/` und werden vom selben statischen Dateiserver ausgeliefert wie das übrige Dashboard.
- Die Lizenz des Icons (Font Awesome Free, Icons unter CC BY 4.0) liegt als `web/FONTAWESOME-LICENSE.txt` bei.

### Lizenzhinweise (historischer Stand, ersetzt am 2026-11-23)

- ~~Der Footer enthält den Copyright-Hinweis „© André Raabe“ und einen Link „Lizenzen von Drittanbietern“ auf eine eigene, einfache Seite `web/licenses.html`.~~
- ~~Die Lizenzseite ist nicht aufklappbar.~~ Die Lizenztexte von hls.js, Font Awesome, Inter Tight und Swagger UI liegen als `web/HLS-LICENSE.txt`, `web/FONTAWESOME-LICENSE.txt`, `web/INTERTIGHT-LICENSE.txt` und `web/swagger/LICENSE` im Web-Root. Die Projektlizenz bleibt `web/LICENSE.txt` (MIT).
- **Aktuell (ADR-0024, ADR-0027):** Ein Info-Icon ganz rechts im Topbar öffnet das Hilfe-Popup. Dessen Knopf „Lizenzen“ öffnet den scrollbaren Lizenzdialog (`#licenses-dialog`). Er listet die MIT-Projektlizenz und die im Laufzeitimage enthaltenen Drittkomponenten: hls.js (Apache-2.0), Font Awesome Free (Icons: CC BY 4.0), Inter Tight (OFL-1.1), Swagger UI (Apache-2.0), FFmpeg (GPL-2.0-or-later), Alpine Linux (diverse) und Go (BSD-3-Clause).
- Es gibt **keinen** Seiten-Footer und **keine** `web/licenses.html`.
- Alle sichtbaren Texte laufen über die i18n-Kataloge (`help.*`, `licenses.*` und `footer.licenseFile` in `web/locales/en.json` und `web/locales/de.json`, ADR-0021).

### SBOM

- Die SBOM wird mit [Syft](https://github.com/anchore/syft) aus dem gebauten Laufzeitimage erzeugt, nicht aus dem Quellbaum. Damit enthält sie die tatsächlich ausgelieferten Betriebssystempakete inklusive FFmpeg und Alpine.
- Erzeugt werden beide Formate: CycloneDX (JSON) und SPDX (JSON).
- `scripts/sbom.sh` pinnt das Syft-Image (`SYFT_IMAGE`, Standard `anchore/syft:v1.57.2`), exportiert das Image mit `docker save` und scannt das Archiv. Der Docker-Socket wird nicht benötigt.
- Die erzeugten Dateien liegen als Snapshot unter `sbom/akkucam-gateway.cyclonedx.json` und `sbom/akkucam-gateway.spdx.json`. `sbom/README.md` beschreibt Geltungsbereich und Regeneration.
- CI erzeugt die SBOM bei jedem Lauf neu und lädt sie als Artefakt hoch. Der committete Snapshot dient als lesbare Referenz, nicht als automatisch aktualisierte Datei.
- Die im Image gebündelten Web-Assets (hls.js, Font Awesome, Inter Tight, Swagger UI) sind keine Paketmanager-Abhängigkeiten und erscheinen daher nicht automatisch in der SBOM. Sie werden im Lizenzdialog und in `sbom/README.md` explizit dokumentiert.

## Begründung

- **Cloudfrei:** Ein lokal gebündeltes SVG vermeidet CDN-Aufrufe und zusätzliche Netzabhängigkeiten. Ein SVG ist zudem kleiner als eine Icon-Schrift und scharf in jeder Größe.
- **Theme-Konsistenz:** `currentColor` und Theme-Variablen halten das Icon automatisch synchron mit dem Dashboard. Das Favicon folgt demselben Theme über die beiden Varianten.
- **Lizenzklarheit:** Die Footer-Rubrik macht die Weitergabe transparenter, ohne die Oberfläche zu überladen. Die Lizenztexte liegen im selben Web-Root und sind offline erreichbar.
- **Reproduzierbarkeit:** Eine SBOM aus dem Image beschreibt das ausgelieferte Artefakt, nicht nur den Quellbaum. Zwei Formate decken unterschiedliche Abnehmer ab (CycloneDX für Security-Tools, SPDX für Compliance).
- **Keine neue Laufzeitabhängigkeit:** Syft läuft nur beim Erzeugen der SBOM in einem Container. Das Gateway selbst bleibt unverändert.

## Betrachtete Alternativen

- **Font-Awesome-Webfont bündeln:** Mehr Payload, zusätzliche Font-Datei, komplizierteres Theme-Handling. Verworfen.
- **Zuklappbare Lizenzliste im Footer:** Würde den Footer mit einer zweiten Aufklapp-Ebene belasten. Eine eigene, einfache Lizenzseite ist übersichtlicher und direkt verlinkbar. Verworfen. *(Am 2026-11-23 durch den Lizenzdialog im Hilfe-Popup nach ADR-0024/ADR-0027 überholt.)*
- **Icon als PNG/Favicon-Paket:** Kein Theme-Support ohne mehrere Rastervarianten. Verworfen.
- **CDN für Icon oder Font:** Verstößt gegen die Cloudfrei-Vorgabe. Verworfen.
- **Nur `go version -m` oder `go mod`-Graph:** Erfasst keine Betriebssystempakete und keine gebündelten Web-Assets. Verworfen.
- **SBOM nur in CycloneDX:** SPDX wird für Lizenz-Compliance häufig verlangt. Beide Formate sind mit Syft ohne Zusatzaufwand erzeugbar.
- **SBOM im Docker-Build erzeugen:** Der Build müsste Syft enthalten und würde langsamer. Verworfen; die SBOM bleibt ein separater Schritt.

## Konsequenzen

- Neue Dateien: `web/icon-light.svg`, `web/icon-dark.svg`, `web/FONTAWESOME-LICENSE.txt`, `scripts/sbom.sh`, `sbom/README.md`, `sbom/akkucam-gateway.cyclonedx.json`, `sbom/akkucam-gateway.spdx.json`. (`web/licenses.html` entfiel am 2026-11-23.)
- Neue i18n-Schlüssel: `help.open`, `help.eyebrow`, `help.title`, `help.text`, `help.dismiss`, `help.wallText`, `licenses.open`, `licenses.eyebrow`, `licenses.heading`, `licenses.intro`, `licenses.projectLicense`, `licenses.thirdPartyHeading`, `footer.apiDocs`, `footer.licenseFile`, `footer.license.various`.
- Neue Tests: `internal/gateway/static_test.go` prüft Icon- und Asset-Verweise; `tests/sbom_test.go` prüft beide SBOM-Formate; `web/e2e/branding.spec.ts` prüft Icon, Favicon, Marken-Typografie, Hilfe- und Lizenzdialog sowie die deutsche Übersetzung.
- Wird eine Abhängigkeit im Image geändert, muss die SBOM mit `./scripts/sbom.sh` neu erzeugt und der Snapshot aktualisiert werden. CI baut das Image und erzeugt zusätzlich bei jedem Lauf eine frische SBOM als Artefakt (`.github/workflows/test.yml`).
- Der Lizenzdialog listet nur Komponenten, die im Image enthalten sind. Reine Entwicklungsabhängigkeiten (z. B. Playwright) erscheinen dort nicht.

- Das Icon ist ein Font-Awesome-Werk; bei Änderungen am Icon muss die Attribution in `web/FONTAWESOME-LICENSE.txt` erhalten bleiben.

## Nachtrag 2026-11-23

Footer, Copyright-Zeile, MIT-Footerlink und die separate `licenses.html` entfallen.
Die Lizenzinformationen öffnen sich über das Info-/Hilfe-Icon als scrollbarer
Lizenzdialog. Er listet die MIT-Projektlizenz und alle Drittkomponenten; die
API-Dokumentation ist dort verlinkt. Der Einstellungsdialog enthält keine
Lizenzlinks mehr. Die Lizenztexte bleiben lokal im Web-Root.

## Nachtrag 2026-11-24

Der Lizenzdialog nennt sieben Gruppen: das Projekt (MIT), die gebündelten
Web-Assets (hls.js, Font Awesome, Inter Tight, Swagger UI), FFmpeg, Alpine
Linux und Go. Die SBOM führt dagegen rund 125 einzelne Alpine-Pakete. Beides
bleibt gültig: `product.md` verlangt die Nennung aller Drittanbieter-
Komponenten, der Dialog fasst die Betriebssystempakete unter „Alpine Linux“
zusammen. Eine vollständige Paketliste bleibt der SBOM vorbehalten.

## Verweise

- ADR-0003: lokal, ohne Cloud-Fallback
- ADR-0013: responsives Dashboard und Theme
- ADR-0021: zweisprachige UI
- `specs/sources.md`, `specs/architecture.md`, `specs/testing.md`
- `scripts/sbom.sh`, `sbom/README.md`

## Nachtrag 2026-11-24

Die im Dialog angezeigten Versionen werden gegen die ausgelieferten Artefakte
geprüft (`TestLicenseVersionsMatchArtifacts`): hls.js gegen
`web/hls.min.js`, FFmpeg, Alpine und Go gegen die committete SBOM sowie die
Base-Images im `Dockerfile`. Wer eine Version ändert, muss die SBOM neu
erzeugen (`make sbom`); sonst schlägt der Test fehl.

Die Font-Awesome-Version entfällt im Dialog. Für sie gibt es keine
maschinenlesbare Quelle (weder Lizenzdatei noch SVG), und eine ungeprüfte
Zahl ist schlechter als keine. Die Lizenz (CC BY 4.0) bleibt genannt.
