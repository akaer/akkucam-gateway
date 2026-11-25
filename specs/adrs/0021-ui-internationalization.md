# ADR-0021: Zweisprachige Browseroberfläche (Deutsch/Englisch)

- Status: Accepted
- Datum: 2026-11-22

## Kontext

Die Oberfläche war ausschließlich deutsch. Das Gateway ist aber auch für englischsprachige Nutzer und für Reviews interessant. Die API liefert bereits stabile Fehlercodes (ADR-0016), die sich clientseitig übersetzen lassen. Die UI ist eine statische Single-Page-App ohne Framework und ohne Build-Schritt.

## Entscheidung

Die Browseroberfläche wird clientseitig zweisprachig (Deutsch und Englisch):

- **Quellsprache Englisch.** `web/index.html` enthält englische Texte. Ohne JavaScript bleibt die Seite englisch.
- **Kataloge als JSON.** `web/locales/en.json` und `web/locales/de.json` werden vom vorhandenen statischen Dateiserver ausgeliefert. `web/i18n.js` lädt beide Kataloge und stellt `t()`, `setText()`, `applyTranslations()` und `setLanguage()` bereit.
- **Spracherkennung.** Reihenfolge: `?lang=de|en` in der URL, dann `localStorage['akkucam-lang']`, dann `navigator.languages` (erster Treffer für `de*`), sonst Englisch. Nur `de*` schaltet auf Deutsch.
- **Live-Umschaltung.** Ein Sprachschalter in der Topbar wechselt die Sprache ohne Reload. Laufende Streams werden nicht unterbrochen: Die Wand wird mit den neuen Texten neu gezeichnet, hängt aber vorhandene `<video>`-Elemente und HLS-Instanzen um, statt sie neu zu erzeugen (ADR‑0027).
- **Persistenz.** Die explizite Sprachwahl liegt in `localStorage['akkucam-lang']`. `?lang=` überschreibt, wird aber nicht gespeichert. Das Theme bleibt wie bisher in `localStorage['akkucam-theme']`.
- **API bleibt englisch.** `error` enthält weiterhin die englische Servermeldung. Die UI übersetzt bekannte `code`-Werte über `error.code.<code>` aus dem Katalog; unbekannte Codes zeigen die Servermeldung.
- **Keine Übersetzung von Gerätedaten.** DIDs, Modellnamen, IP-Adressen und vom Hub gelieferte Kanalnamen bleiben unverändert.
- **Barrierefreiheit.** `<html lang>` wird bei jedem Wechsel aktualisiert. Übersetzbare `title`- und `aria-label`-Attribute nutzen `data-i18n-title` und `data-i18n-aria-label`.
- **Swagger UI bleibt englisch.** Sie ist ein Entwicklerwerkzeug und wird als Ausnahme dokumentiert.
- **Playwright als E2E-Werkzeug.** `web/e2e/` enthält Browser-Tests für Spracherkennung, Live-Umschaltung, Persistenz, Fehlercodes und optional echte Hardware. Die Tests laufen im offiziellen Playwright-Image, nicht über eine Node-Toolchain im Produktionspfad.

## Begründung

- Die UI ist statisch; ein i18n-Framework wäre Overkill.
- JSON-Kataloge sind mit Go-Tests prüfbar (gleiche Schlüssel, gleiche Platzhalter, keine leeren Werte).
- Stabile Fehlercodes machen eine serverseitige Lokalisierung unnötig.
- Live-Umschaltung ohne Reload erhält laufende Streams und vermeidet einen Kamera-Wake-up.
- Englisch als Quell- und Fallbacksprache ist die kleinste gemeinsame Basis für unbekannte Browsersprachen.

## Betrachtete Alternativen

- **Serverseitige Lokalisierung:** verworfen. Der Server müsste Sprache, Kataloge und Templates kennen; die API bliebe trotzdem englisch.
- **i18n-Framework (z. B. i18next):** verworfen. Neue Laufzeitabhängigkeit und Build-Schritt für eine kleine Oberfläche.
- **Nur Deutsch mit englischer API:** verworfen. Die Anforderung ist ausdrücklich zweisprachig.
- **Sprachwechsel mit Reload:** verworfen. Unterbricht Streams und kostet einen erneuten Kamera-Wake-up.

## Konsequenzen

- Alle sichtbaren UI-Texte laufen über `web/i18n.js`; hartkodierte Texte in HTML oder JavaScript sind unzulässig. Ausnahmen sind Gerätedaten und Eigennamen.
- `de.json` und `en.json` müssen dieselben Schlüssel und Platzhalter haben. Go-Tests in `internal/gateway/i18n_test.go` prüfen das, inklusive aller `data-i18n`-Schlüssel aus `index.html`, der in `app.js` verwendeten Schlüssel und der API-Fehlercodes.
- Neue API-Fehlercodes benötigen einen Eintrag in beiden Katalogen.
- Neue Sprachen benötigen einen Katalog, einen Eintrag in `SUPPORTED` und einen Playwright-Test.
- Playwright-Tests liegen in `web/e2e/` und laufen über `scripts/test-e2e.sh`. Die Live-Suite ist optional und benötigt `AKKUCAM_TEST_SECURITY_CODE`; sie schreibt keine Artefakte, damit der Sicherheitscode nicht auf Platte landet.
- Der Sprachwechsel ist ein UI-Zustand; er ändert weder API-Vertrag noch Protokoll.

## Nachtrag 2026-11-24

- Katalogwerte sind vollständige Sätze. Satzteile werden nicht in `app.js`
  aneinandergehängt; Daten füllen Platzhalter (ADR-0028).
- Zahlen, Zeiten und Listen formatiert der Client mit `Intl`. Die API liefert
  `durationMs` und `uptimeSeconds` statt vorformatierter Zeichenketten, damit
  die Ausgabe der aktiven Sprache folgt.
- `TestLocaleValuesAreWholePhrases` ergänzt die Schlüssel- und
  Platzhalterprüfung um diese Regel.

## Verweise

- ADR-0012: Sicherheitscode nur im Browser-RAM
- ADR-0013: Responsive Dashboard und Theme
- ADR-0016: Stabile API-Fehlercodes
- `specs/product.md`, `specs/architecture.md`, `specs/api.md`, `specs/security.md`, `specs/testing.md`
- `web/i18n.js`, `web/locales/`, `web/e2e/`
