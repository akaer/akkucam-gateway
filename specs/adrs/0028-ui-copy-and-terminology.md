# ADR-0028: UI-Texte, Terminologie und Fehlerwege

- Status: Accepted
- Datum: 2026-11-24

## Kontext

Die Oberfläche war funktional zweisprachig, aber sprachlich unentschieden:
Katalogwerte wurden in `app.js` zu Sätzen zusammengeklebt
(`" · scrolls"`, `{suffix}`), drei Register wechselten sich ab
(„Kameras laden“, „Lade Kameras …“, „Warte auf den Videostream“), und
„Browserspeicher“ bezeichnete sowohl den flüchtigen RAM als auch
`localStorage`. Dazu kamen Erklärungen der Implementierung statt einer
Handlung („Reset removes the entry from localStorage“), Fehler ohne
Wiederholungsmöglichkeit und eine destruktive Aktion, die ohne Rückfrage alle
Sicherheitscodes löschte.

## Entscheidung

### Vollständige Sätze

- Ein Katalogwert ist ein vollständiger Satz oder eine vollständige
  Beschriftung. Satzteile werden nicht in JavaScript aneinandergehängt.
- Daten dürfen Platzhalter füllen (`{count}`, `{channels}`, `{hub}`).
  Listen formatiert der Client mit `Intl.ListFormat`, Zahlen und Zeiten mit
  `Intl.NumberFormat`.
- Die API liefert Maschinenwerte (`durationMs`, `uptimeSeconds`), keine
  vorformatierten Zeichenketten.

### Terminologie (DE/EN)

| Begriff | Bedeutung | Verboten |
|---|---|---|
| Basisstation / base station | das Gerät im Netz | „Hub“ in sichtbaren Texten |
| Kamera / camera | der gekoppelte Kanal | „Kanal“ für das Gerät |
| Kanal / channel | die Nummer eines Eingangs | „Kamera“ für die Nummer |
| Livebild / live view | das sichtbare Bild | „Stream“ für das Bild |
| Stream / stream | der Transport (Start/Stop) | — |
| Details / details | der Drawer | „Drawer“, „Seitenleiste“ in DE |
| Live-Daten / live data | das Overlay auf der Kachel | „Metriken“ |
| Arbeitsspeicher / memory | flüchtiger Browser-RAM | „Browserspeicher“ (mehrdeutig) |

### Register und Ton

- Deutsch durchgehend Du-Form.
- Fortschrittsmeldungen stehen im Zustandspassiv oder als Infinitiv
  („Kameras werden geladen …“, „Videostream wird vorbereitet …“), nicht als
  Aufforderung an den Nutzer.
- Keine Implementierungs-Disclaimer: kein `localStorage`, keine „Matrix“,
  kein „Fallback“, keine „Optional“-Hinweise. Erklärt wird die Handlung oder
  der Nutzen, nicht der Code.
- Fehlermeldungen enden nicht in einer Sackgasse: Wenn eine Wiederholung
  möglich ist, bietet die Meldung „Erneut versuchen“ an.

### Destruktive Aktionen

- „Alle Sicherheitscodes entfernen“ benötigt zwei Klicks: Der erste armiert
  die Aktion und ändert die Beschriftung („Wirklich alle Codes entfernen?“),
  der zweite führt sie aus. Nach vier Sekunden fällt der Knopf zurück.
- Kein modaler Bestätigungsdialog; der zweite Klick ist die Bestätigung.

### Gerätecode

- Herkunft und Speicherung des Sicherheitscodes auf der Basisstation sind
  nicht Teil des Produkts. Die Oberfläche erklärt sie nicht; der Code wird
  eingegeben, geprüft und wie in ADR-0012 beschrieben nur im Browser-RAM
  gehalten.

### Codes und Passwortmanager

- Das Code-Feld trägt `autocomplete="off"`, damit kein Passwortmanager den
  Gerätecode persistiert (ADR-0012, ADR-0025).

## Begründung

- Ein vollständiger Satz pro Schlüssel ist die einzige Form, die sich
  übersetzen lässt, ohne Wortstellung zu raten.
- Ein Glossar verhindert, dass dieselbe Fläche fünf Namen bekommt
  („Wand“, „Live-Wand“, „Wandsteuerung“, „Wand-Einstellungen“,
  „Wand-Einrichtung“).
- Maschinenwerte aus der API halten Formatierung und Sprache an einer Stelle.
- Die Zwei-Klick-Bestätigung schützt vor Datenverlust, ohne einen weiteren
  Dialog zu öffnen.

## Betrachtete Alternativen

- **Fragmente im Katalog belassen und in JS zusammensetzen:** sprachabhängig
  falsch (Wortstellung, Klammern, Aufzählungszeichen) und von den
  Katalogtests nicht erfassbar. Verworfen.
- **Zeit- und Zahlformate im Go-Server erzeugen:** würde die API an eine
  Sprache binden und ADR-0021 widersprechen. Verworfen.
- **Höflichkeitsform („Sie“) im Deutschen:** inkonsistent mit der bereits
  ausgelieferten Du-Form und für ein Heimnetzwerk-Gerät zu distanziert.
  Verworfen.
- **Modaler Bestätigungsdialog vor dem Löschen:** zusätzliche Ebene für eine
  reversible Handlung (Codes lassen sich neu eingeben). Verworfen.

## Konsequenzen

- `web/locales/{en,de}.json` enthält keine Fragmentwerte mehr;
  `TestLocaleValuesAreWholePhrases` prüft das.
- `web/app.js` formatiert Listen, Millisekunden und Laufzeiten; die API
  liefert `durationMs` und `uptimeSeconds` (OpenAPI, `specs/api.md`).
- Der Hub-Manager und die Diagnose teilen sich denselben
  Zwei-Klick-Mechanismus; der Playwright-Test der Credentials-Suite deckt ihn
  ab.
- Gesperrte Aktionen (Motion, Discovery, „Basisstation hinzufügen“) tragen
  einen `title` mit dem Grund.
- Neue Texte werden gegen das Glossar geschrieben; Abweichungen benötigen
  eine Änderung an diesem ADR.

## Verweise

- ADR-0012/ADR-0025: Code nur im Browser-RAM
- ADR-0014: nur verfügbare Metadaten
- ADR-0021: zweisprachige UI
- ADR-0029: Design-Tokens
- `web/locales/`, `web/app.js`, `web/i18n.js`
- `specs/product.md`, `specs/api.md`, `specs/testing.md`
