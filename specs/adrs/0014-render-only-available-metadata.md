# ADR-0014: Nur tatsächlich verfügbare Metadaten anzeigen

- Status: Accepted
- Datum: 2026-11-21

## Kontext

Nicht jedes Protokollobjekt liefert dieselben Metadaten. Insbesondere enthält die Hub-Kameraliste keine Kamera-MAC-Adresse. Generische Platzhalter wie „unbekannt“, „nicht gemeldet“ oder leere Datenfelder erzeugen visuelles Rauschen und suggerieren einen Informationsgehalt, den der Gateway nicht besitzt.

## Entscheidung

Die Oberfläche rendert ein Metadatenfeld nur, wenn ein konkreter Wert aus einer verifizierten Quelle vorhanden ist. Fehlende Felder werden vollständig ausgelassen. Es werden keine Werte geraten, abgeleitet oder durch generische Texte ersetzt.

Vorhandene Werte werden vollständig angezeigt. Lange Bezeichner wie DID oder MAC-Adresse brechen bei Bedarf um, statt per Ellipse abgeschnitten zu werden.

Funktionale UI-Bezeichnungen wie „Kanal 2“ sind davon ausgenommen: Sie dienen der Bedienung und sind keine behaupteten Gerätemetadaten.

## Begründung

- höhere Informationsdichte und weniger Rauschen
- klare Trennung zwischen Wissen und Nichtwissen
- keine irreführenden Aussagen
- technisch affine Nutzer sehen nur belastbare Daten

## Konsequenzen

- Kamera-MAC wird aktuell nicht angezeigt.
- Hub-MAC wird nur bei erfolgreichem ARP-Lookup angezeigt.
- Modell, Firmware, MCU-Version, Name oder Motion-State erscheinen jeweils nur bei gültiger Protokollantwort.
- Karten können je nach Gerät unterschiedlich viele Metadatenzeilen besitzen.
- Metadatenzeilen werden bei langen Werten höher, weil der Wert umbricht statt abgeschnitten zu werden.

## Nachtrag 2026-11-24

Die Pflicht, Werte vollständig zu zeigen, gilt für Metadatenfelder (Drawer,
Diagnose, Hub-Details); sie brechen um. Kachel-Beschriftungen dürfen dagegen
per Ellipse kürzen, weil die Kachel eine feste Fläche hat: Der vollständige
Name steht im Details-Drawer. Funktionale Beschriftungen („Kamera 2“,
„Basisstation“) bleiben wie bisher von der Regel ausgenommen.
