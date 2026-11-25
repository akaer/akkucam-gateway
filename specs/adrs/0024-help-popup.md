# ADR-0024: Info-Icon und Hilfe-Popup im Topbar

- Status: Accepted
- Datum: 2026-11-22
- Geändert: 2026-11-23 (Nachtrag, Lizenzdialog)

## Kontext

Der Einstieg in das Dashboard war nach dem Entfernen des Hero-Textblocks (ADR-0023) sehr knapp. Neue Nutzer sehen nur den Discover-Button und den Hinweis „Ready. Discover a base station to begin.“ Der Ablauf – Basisstation finden, gekoppelte Kameras öffnen, Livebild ohne Cloud-Dienst starten – war nirgends erklärt.

Die Erklärung soll erreichbar sein, ohne den Einstieg wieder mit Text zu füllen. Sie muss zweisprachig sein, dem Theme folgen und ohne externe Ressourcen funktionieren.

## Entscheidung

- Im Topbar sitzt ganz rechts, rechts neben der Theme-Umschaltung, ein kleines Info-Icon (`.icon-button` mit Inline-SVG). Das Icon ist ein eigener, lokal gezeichneter SVG-Pfad; es kommt kein CDN und keine zusätzliche Icon-Bibliothek hinzu.
- Ein Klick öffnet ein modales `<dialog>`-Popup (`#help-dialog`) mit dem Ablauf in der aktiven Sprache:
  - Deutsch: „Basisstation finden, gekoppelte Kameras öffnen und Livebild ohne Cloud-Dienst starten.“
  - Englisch: „Find a base station, open its cameras and start a live view without any cloud service.“
- Das Popup nutzt dieselbe Dialog-Basis wie der Sicherheitscode-Dialog (`.modal-dialog`), schließt über „Verstanden“/„Got it“, das × und Escape. Es enthält keine Formulardaten und speichert nichts.
- Die Texte liegen in `web/locales/en.json` und `web/locales/de.json` unter `help.*`. Der Sprachwechsel aktualisiert das Popup über den bestehenden `i18n:change`-Pfad.
- Das Icon ist per `aria-label`/`title` benannt; der Dialog ist über `aria-labelledby` an seine Überschrift gebunden.

## Begründung

- **Cloudfrei und leichtgewichtig:** Ein Inline-SVG und ein natives `<dialog>` benötigen keine Bibliothek, keinen Build-Schritt und keine Netzwerkzugriffe.
- **Zweisprachig ohne Sonderweg:** Die Hilfe nutzt denselben Katalog wie die übrige Oberfläche. Ein Sprachwechsel wirkt sofort, ohne Reload und ohne laufende Streams zu unterbrechen.
- **Kein neuer Einstiegstext:** Der Hero-Bereich bleibt aufgeräumt; die Erklärung erscheint nur auf Wunsch.
- **Konsistenz:** Das Popup teilt sich Styling und Schließverhalten mit dem bereits vorhandenen Sicherheitscode-Dialog.
- **Barrierefreiheit:** Ein echtes `<dialog>` liefert Fokusführung und Escape-Verhalten; `aria-label` und `aria-labelledby` benennen Icon und Dialog.

## Betrachtete Alternativen

- **Hilfetext dauerhaft im Hero anzeigen:** Widerspricht der Entscheidung, den Hero-Textblock zu entfernen. Verworfen.
- **`<details>`-Aufklapper im Topbar:** Weniger auffällig, kein modal Fokus und keine Escape-Behandlung. Verworfen.
- **Tooltip am Discover-Button:** Zeigt den Ablauf nur bruchstückhaft und ist auf Touch-Geräten unzuverlässig. Verworfen.
- **Icon-Bibliothek nachladen:** Zusätzliche Abhängigkeit und Lizenzpflege für ein einzelnes Icon. Verworfen.

## Konsequenzen

- `web/index.html` enthält den Help-Button im Topbar und den `#help-dialog`; `web/app.js` verdrahtet den Öffnen-Klick.
- `web/style.css` führt die gemeinsame Dialog-Basis als `.modal-dialog`; der Sicherheitscode-Dialog nutzt sie ebenfalls.
- Neue i18n-Schlüssel: `help.open`, `help.eyebrow`, `help.title`, `help.text`, `help.dismiss` in beiden Katalogen.
- Neue Tests: `web/e2e/i18n.spec.ts` prüft den lokalisierten Hilfetext und das Schließen über „Got it“. `TestIndexHTMLKeysExistInLocales` deckt die neuen Schlüssel ab.
- Keine neuen statischen Dateien, keine Lizenzänderung, keine SBOM-Änderung.

## Nachtrag 2026-11-23

Das Hilfe-Popup enthält zusätzlich einen Button „Lizenzen“ (`#open-licenses`).
Er schließt die Hilfe und öffnet den scrollbaren Lizenzdialog; damit sind die
Lizenzinformationen ohne separaten Footer und ohne eigene `licenses.html`
erreichbar. Der Dialog nutzt dieselbe Dialog-Basis, listet MIT und alle
Drittanbieter und verlinkt die API-Dokumentation. Neue Schlüssel:
`help.wallText`, `licenses.open`.

## Verweise

- ADR-0003: lokal, ohne Cloud-Fallback
- ADR-0021: zweisprachige UI
- ADR-0023: Inter Tight für Markenname und Schritt-Headings
- `web/index.html`, `web/app.js`, `web/style.css`, `web/locales/{en,de}.json`
- `specs/product.md`, `specs/architecture.md`, `specs/testing.md`
