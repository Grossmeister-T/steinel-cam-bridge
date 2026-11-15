# GMT-Fork der Steinel CAM Bridge

Stand: 15.11.2026, Version `1.5.2-beta.11-gmt.4`, Zweig `gmt-beta11`.

Basis ist der Entwicklerzweig `origin/feature/l620-sofia-download` (Commit `b1bb559`), nicht `main`. Der Entwickler pflegt den laufenden Code auf diesem Nebenzweig; `main` bekommt nur die Versionsnummern und Doku der Beta-Veröffentlichungen und steht im Code noch etwa auf beta.5. Ein erster Bau auf `main` (beta.10-gmt.1) hatte deshalb Korrekturen aus beta.6 bis beta.11 verloren, sichtbar an der Dämmerungsschwelle (Rohwert 210 statt 10 lx).

Ziel des Forks: schnelles Bild am Dashboard und lückenlose Aufnahmen auf dem NAS mit der L 625 CAM SC. Alles Übrige bleibt wie im Original.

## Was anders ist

| Thema | Original (beta.11) | Fork |
|---|---|---|
| Neuer Zuschauer | wartet bis zu 2 s auf das nächste Schlüsselbild | bekommt sofort das aktuelle Schlüsselbild samt Folgebildern (`instant_start`) |
| Lampe liefert kein Bild | Strom reißt ab, Abnehmer melden Fehler und verbinden neu | letztes Schlüsselbild läuft 1×/s weiter, höchstens `hold_last_frame_seconds` (Standard 300) |
| Neue Kamerasitzung | neue RTP-Zählung mitten im Strom | Strom bleibt durchgehend (feste SSRC, fortlaufende Nummern und Zeitstempel), Weitergabe erst ab einem Schlüsselbild |
| Auflösung | wird einmal angefragt, nie geprüft; Anzeige = Wunschwert | echte Bildgröße wird aus dem Videostrom gelesen, bei Abweichung erneut angefragt (10 s, 30 s, 60 s, dann 3× alle 5 min); Anzeige = Istwert |
| SDP | SPS/PPS nur, wenn der Treiber sie setzt | SPS/PPS werden aus dem Strom übernommen |
| Ton | `aac` oder `pcmu` | zusätzlich `none`: reiner Videostrom, kein Mikrofon, kein Gegensprechen |
| Grundlicht | Helligkeit und Dauer als Zahl | zusätzlich Schalter „Grundlicht“ wie in der App: aus = 0 %, ein = zuletzt eingestellte Helligkeit (nach einem Neustart der Bridge 10 %, bis die Lampe einen Wert meldet) |
| Bewegungssensor | geht bei jeder Befehlsantwort mit „alarm“ im Text an (z. B. Sirene) | Antworten auf eigene Befehle (`resp`) zählen nicht als Bewegung |
| Werte der Kamera | Geräteinfo wird nie abgefragt (Firmware leer) | `get_device_info` einmal pro Minute: Sensoren WLAN-Signal (%), SD-Karte frei (GB), SD-Karte Status, Firmware |
| Zeitstempel im Bild | fest oben links | Schalter „Zeitstempel im Bild“ und Auswahl „Zeitstempel Position“ (vier Ecken) |
| Lautsprecher | nicht einstellbar | Zahl „Lautsprecher Lautstärke“ (0–100 %) |
| Rahmen im Bild | Kamera zeichnet Erkennungsrahmen ins Video (rot = Person, grün = Bewegungsbereich), in der App nicht abschaltbar | Auswahl „Kamera-Erkennung“: „KI-Erkennung“ (`mode: "ai"`, Auslieferung) oder „Aus“ (`mode: "off"`, keine Rahmen) |
| Diagnose | – | Option `debug_api`: `POST /api/debug/command` schickt einzelne Befehle an die Kamera, Antworten ohne Zugangsdaten |

## Befehle der Kamera (L 625, Firmware 0.1.2.31)

Ermittelt am 15.11.2026 über die Diagnose-Schnittstelle. Alle gehen als JSON über den DataChannel, Aufbau wie in `AGENTS_L625.md`.

| Befehl | Inhalt |
|---|---|
| `get_device_info` | Firmware, KI- und MCU-Version, Hersteller `JSW`, Modell `LC-W5`, WLAN (`RSSI` in %, Netzname), SD-Karte (frei/gesamt in MB, Status), `time.OSD`, `volume`, `detect`, `record`, `video` |
| `get_time_setting` / `set_time_setting` | `OSD`: 0 = aus, 1 = oben links, 2 = unten links, 3 = oben rechts, 4 = unten rechts (am Bild geprüft; 5 verhält sich wie 4). Setzen mit `action: "OSD"`. `format` 0/1/2 wird gespeichert, ändert das Bild aber nicht. |
| `get_volume_setting` / `set_volume_setting` | `volume` 0–100 |
| `get_detect_setting` / `set_detect_setting` | `mode` (Auslieferung `"ai"`), `od_enable`. Setzen mit `action: "mode"`. Die Firmware speichert jeden Text ungeprüft, auch erfundene Werte. |
| `get_video_setting` | Hauptstrom H.265 2K 2560 kbit/s, `sub1` H.264 1080p 2048 kbit/s (Strom der Bridge), `sub2` MJPEG 360p; `environment`, `image`, `lightmode`, `lightenc` |
| `get_record_setting` | `auto_rec`, `mode`, `record_time` |
| `get_notification_setting` | `enable` |
| `get_schedule_setting` | `enable`, `list` |

Keine Antwort auf: `get_ai_setting`, `get_osd_setting`, `get_alarm_setting`, `get_motion_setting`, `get_md_setting`, `get_light_setting`, `get_image_setting`, `get_privacy_setting`, `get_track_setting`, `get_area_setting`, `get_zone_setting`, `get_detect_area`, `get_audio_setting`, `get_system_setting`, `get_event_setting`, `get_pir_setting`, `get_od_setting`, `get_human_setting`, `get_speaker_setting`, `get_sd_setting`, `get_wifi_info`, `get_network_setting`, `get_siren_setting`, `get_ota_setting`, `get_led_setting`, `get_night_setting`, `get_status`.

Aus dem Original übernommen, nicht mehr vom Fork: Nachlaufzeit, Grundlicht-Helligkeit, Grundlicht-Dauer (0–12 h), Icon-Korrektur, Umrechnung der Dämmerungsschwelle, Wiederholpausen nach einem Abriss.

## Neue Optionen

| Option | Standard | Bedeutung |
|---|---|---|
| `instant_start` | `true` | Sofortbild für neue RTSP-Zuschauer |
| `hold_last_frame_seconds` | `300` | Standbild bei Aussetzern, 0 = aus (nur L 625) |
| `audio_codec` | `aac` | neu: `none` |
| `debug_api` | `false` | Diagnose-Schnittstelle; ohne `"write": true` nur `get_…`-Befehle. Nach Gebrauch wieder ausschalten, sie ist im LAN ohne Passwort erreichbar, solange kein Bridge-Zugangsschutz gesetzt ist. |

## Geprüft

Am PC:

- Unit-Tests für die neue Logik (`pkg/h264x`, `pkg/rtsp`, `pkg/mqtt`), darunter zwei Tests mit einem echten RTSP-Client.
- Simulierte Kamera (`tools/stream-sim`, 1080p, 15 Bilder/s, Schlüsselbild alle 2 s) gegen ffmpeg:
  - erstes Bild mit Sofortbild nach 0,7–1,9 s, ohne nach 3,3–4,0 s, je 8 Versuche, keine Dekoderfehler;
  - drei simulierte Ausfälle von je 8 s mit neuer RTP-Zählung: Dekodieren und Aufnehmen (`-c copy`) laufen ohne Fehler durch; ohne Standbild meldet ffmpeg bei jedem Ausfall einen Lesefehler.
- Bau für den Pi (linux/arm64) samt `go vet` über alle Pakete.

An der echten Lampe (15.11.2026, lokale HA-App auf dem Pi):

- Kopplung, Videostrom, MQTT; HA-Kamera liefert nach App-Wechsel und App-Update ohne Neuladen ein Bild; Surveillance Station zeigt ein aktuelles Bild.
- ffprobe: H.264 Baseline 1920×1080, 15 Bilder/s; erstes Paket ist ein Schlüsselbild.
- Erstes Bild für einen reinen Videoabnehmer (ffmpeg, `-allowed_media_types video`): 0,9–1,6 s, 8 Versuche, keine Dekoderfehler. Mit beiden Tonspuren braucht ffmpeg 4–5,6 s, weil es auf die leere Gegensprech-Spur wartet (kein Vergleichswert zur Beta gemessen).
- Auflösungswächter meldet „Camera delivers 1920x1080 (1080p)“.
- Dämmerungsschwelle zeigt 10 lx; Grundlicht-Schalter: ein → Lampe meldet 10 %, aus → 0 %.
- Neue Entitäten zeigen Werte (WLAN-Signal 57 %, SD-Karte 120,5 GB frei, Status normal, Firmware 0.1.2.31, Lautstärke 80 %).
- Zeitstempel: über Home Assistant auf „Unten rechts“ gestellt, Lampe meldet Position 4, Uhr steht im Bild unten rechts; danach zurück auf „Oben links“.
- Lautstärke: per Diagnose-Schnittstelle 79 gesetzt und zurückgelesen, dann wieder 80. Die Zahl-Entität selbst ist nicht geschaltet worden.
- Rahmen: Im Modus `"ai"` bekommt eine Person einen roten Rahmen (Beobachtung des Besitzers, mehrfach), ein vorbeifahrendes Auto in einer Aufnahme keinen. Im Modus `"off"` lief um 17:56 eine Person durchs Bild: in 17 gespeicherten Bildern dieser 10 Sekunden kein Rahmen, vom Besitzer am Dashboard bestätigt. Ein Lichtwechsel des Flutlichts löst in keinem Modus einen Rahmen aus (je 48 Bilder).
- Nicht geklärt: ob `"off"` auch die Ereignis-Aufnahmen auf die SD-Karte abstellt. Zwischen 17:41 und 17:58 entstand gar kein SD-Ereignis, auch nicht beim Lauf im Modus `"ai"` gegen 17:54, deshalb sagt der Vergleich nichts aus. Die anderen ausprobierten Moduswerte (`pir`, `motion`, `md`) wurden gespeichert, aber nicht am Bild geprüft.

## Nicht geprüft

- Verzögerung am Dashboard (WebRTC über go2rtc) mit Sofortbild. Der Zuschauer bekommt bis zu 2 s ältere Bilder auf einen Schlag; ob der Browser das sofort aufholt, muss am Tablet oder im Browser gemessen werden. Falls nicht: `instant_start: false`.
- Standbild und durchgehender Strom bei einem echten Lampenausfall (bisher keiner aufgetreten).
- Auflösungswächter im Fehlerfall (Lampe liefert etwas anderes als angefragt).
- `audio_codec: none` an der Lampe, in HA und in Surveillance Station.
- Bewegungskorrektur mit der Sirene.

## Bauen

```bash
python scripts/make-gmt-addon.py gmt.1
```

```bash
bash scripts/build-gmt.sh
```

Das erste Skript leitet `ha-addon-gmt/` aus der Beta-App des Originals ab (Optionen, Übersetzungen, Basis-Abbild), das zweite baut `ha-addon-gmt/steinel-bridge`. Gebraucht werden Go ≥ 1.27, Zig als C-Compiler und das Nabto-SDK in `.sdk/` (siehe Kopf von `build-gmt.sh`). Die lokale HA-App ist das Beta-Abbild des Originals mit ausgetauschtem Programm; auf dem Pi wird nichts kompiliert.

Tests ohne C-Compiler:

```bash
CGO_ENABLED=0 go test ./pkg/h264x/ ./pkg/rtsp/ ./pkg/mqtt/ ./pkg/onvif/
```

## Auf einen neuen Stand des Originals umziehen

1. `git fetch origin` und prüfen, auf welchem Zweig der neue Code liegt (zuletzt `feature/l620-sofia-download`, nicht `main`).
2. Änderungen des Forks auf diesen Stand übertragen, Konflikte in `pkg/rtsp`, `pkg/webrtc`, `pkg/mqtt` lösen.
3. Beide Skripte oben laufen lassen, Ordner `ha-addon-gmt/` in die Samba-Freigabe für lokale Apps von Home Assistant kopieren (`\\<HA-Adresse>\addons\steinel_cam_bridge_gmt`), in HA den Store neu laden und die App aktualisieren.
