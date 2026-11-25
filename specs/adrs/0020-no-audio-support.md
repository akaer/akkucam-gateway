# ADR-0020: Kein Audio-Support

- Status: Accepted
- Datum: 2026-11-22

## Kontext

Audioempfang wäre technisch möglich. Die verifizierte Sitzungssequenz sendet `StartVideo` und `StartAudio`; der Hub kann danach neben Videodaten auch Audiodaten liefern. Gleichzeitig zeigt das Dashboard mehrere gekoppelte Kameras parallel. Bei mehreren gleichzeitig laufenden Streams ist nicht mehr erkennbar, zu welcher Kamera ein Audiostrom gehört. Audio wird damit zum Chaos statt zur Information. Ein Gegensprech-Feature ist nicht geplant. Browser-Autoplay-Regeln und zusätzliche Bandbreite/CPU sprechen ebenfalls gegen Audio.

## Entscheidung

- Der Gateway unterstützt **kein Audio**: keinen Audioempfang, keine Browser-Audioausgabe und kein Gegensprechen.
- Eingehende Audiodaten werden verworfen. Es gibt keine Audio-Steuerelemente in der Oberfläche.
- `StartAudio` bleibt vorerst Teil des Sitzungsaufbaus, weil die verifizierte Sequenz `StartVideo` + `StartAudio` umfasst. Der Befehl dient ausschließlich der Protokollkompatibilität und nicht der Audiofunktion. Ob `StartVideo` allein genügt, ist eine spätere Hardwareprüfung.

## Begründung

- Bei mehreren parallelen Streams ist Audio keiner Kamera eindeutig zuzuordnen. Die Wiedergabe wäre für Nutzer verwirrend.
- Livevideo ist der Produktzweck; Audio liefert für dieses Szenario keinen Mehrwert.
- Audio erhöht Bandbreite, CPU-Last und Komplexität in der Oberfläche.
- Ein konsistenter Audio-Support nur für einen einzelnen Stream wäre eine Sonderregel ohne klaren Nutzen.

## Betrachtete Alternativen

- **Vollständiger Audio-Support:** verworfen, weil mehrere Streams Audiochaos erzeugen.
- **Audio nur bei genau einem aktiven Stream:** verworfen, weil das Verhalten je nach Anzahl der Streams springt und schwer erklärbar ist.
- **Gegensprechen:** verworfen, nicht Teil des Produktumfangs.
- **WebRTC-Audio:** verworfen, zusätzliche Signalisierung ohne Produktnutzen.

## Konsequenzen

- Audio verschwindet aus der Roadmap.
- `specs/product.md` führt Audioempfang, Audioausgabe und Gegensprechen als Nicht-Ziele.
- `specs/protocol.md` dokumentiert weiterhin `StartAudio`, weist aber auf die fehlende Auswertung hin.
- Die Tests bleiben unverändert; der Fake-Hub erwartet `StartAudio` weiterhin.
- Aufzeichnungen sind nicht Teil des Gateways. Die Basisstation zeichnet selbst auf, sofern für den Kanal die Bewegungserkennung aktiv und eine SD-Karte vorhanden ist. Ein Zugriff auf diese Aufzeichnungen ist als Roadmap-Punkt vorgesehen.

## Verweise

- [ADR-0006](0006-h264-passthrough-and-hls.md): H.264-Passthrough und HLS
- [ADR-0011](0011-concurrent-stream-per-channel.md): paralleler Stream pro DID und Kanal
- `specs/protocol.md`
- `specs/product.md`
- `specs/roadmap.md`
