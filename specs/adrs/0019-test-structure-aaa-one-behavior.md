# ADR-0019: Teststruktur: Arrange/Act/Assert und ein Verhalten pro Test

- Status: Accepted
- Datum: 2026-11-22

## Kontext

Viele Tests prüfen mehrere unabhängige Dinge in einer Funktion. Namen mit „And“ erschweren die Zuordnung eines Fehlschlags. Die drei Phasen Arrange, Act und Assert sind nicht überall sichtbar. Dadurch werden Tests länger, schwerer lesbar und bei Fehlschlägen unpräzise.

## Entscheidung

- Jeder Test folgt **Arrange/Act/Assert**. Die Phasen werden mit Kommentaren markiert, wenn sie nicht offensichtlich sind.
- **Ein Test prüft ein Verhalten.** Mehrere Assertions sind erlaubt, wenn sie dasselbe Ergebnis beschreiben (z. B. mehrere Felder eines geparsten Objekts).
- Varianten desselben Verhaltens laufen als **Tabellen-Subtests** mit eigenem AAA.
- Testnamen beschreiben das Verhalten und enthalten **kein „And“**.
- **Integrationstests** werden in getrennte Tests und Sessions aufgeteilt, wenn sie unabhängige Ergebnisse prüfen. Das gilt auch auf Kosten der Laufzeit.
- **Hardwaretests** werden ebenfalls aufgeteilt. Jede Session prüft genau ein Ergebnis.
- **Fuzz-Targets** sind die einzige Ausnahme: ein Target = ein Parser/Verhalten, ohne AAA-Kommentare.
- Helfer dürfen das Arrange kapseln, aber nicht Act oder Assert verdecken.

## Begründung

- Ein Fehlschlag zeigt sofort, welches Verhalten betroffen ist.
- Kleinere Tests sind leichter zu lesen, zu reviewen und zu ändern.
- Getrennte Sessions verhindern, dass ein Fehler in einem Teil die Beobachtung eines anderen Teils verdeckt.
- Die AAA-Kommentare machen die Testabsicht ohne Vorwissen sichtbar.

## Betrachtete Alternativen

- **Alles in einem Test belassen:** weniger Laufzeit, aber unklare Fehlschläge und stärkere Kopplung.
- **Nur bei neuen Tests anwenden:** bestehende Tests bleiben inkonsistent.
- **Streng eine Assertion pro Test:** würde zusammengehörige Vertragsfelder unnötig zersplittern.

## Konsequenzen

- Mehr Testfunktionen und teils längere Laufzeit, insbesondere bei Integration und Hardware.
- Gemeinsame Helfer für wiederkehrendes Arrange.
- `AGENTS.md`, `specs/testing.md` und `tests/README.md` dokumentieren die Konvention.
- Neue Tests folgen der Konvention; bestehende werden schrittweise angepasst.

## Verweise

- ADR-0018: Test- und Codequalitätskonventionen
- `specs/testing.md`
- `tests/README.md`
