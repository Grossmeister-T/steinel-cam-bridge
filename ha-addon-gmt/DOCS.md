# Steinel CAM Bridge GMT (lokale App)

Eigener Fork der Steinel CAM Bridge auf Basis von 1.5.2-beta.10. Was er anders macht, steht in `GMT-CHANGES.md` im Quellordner.

## Einrichten

1. Die bisherige Bridge-App stoppen. Beide nutzen dieselben Ports (8554, 8000) und dürfen nicht gleichzeitig mit der Lampe verbunden sein.
2. Optionen wie bei der bisherigen App eintragen (Kamera-IP, Modell `l625`, Auflösung, Treiber `cgo`).
3. QR-Code aus der STEINEL CAM App („Kamera teilen“) selbst in `qr_code` eintragen. Die App hat einen eigenen Datenordner und koppelt sich deshalb neu.
4. Starten. Die Geräte-ID bleibt gleich, die MQTT-Entitäten in Home Assistant bleiben erhalten.

## Zusätzliche Optionen

- `instant_start`: Sofortbild für neue Zuschauer (Standard aus; auf einem Tablet im WLAN wurde das Bild damit später flüssig).
- `hold_last_frame_seconds`: Standbild bei Aussetzern in Sekunden, 0 = aus (Standard 300).
- `audio_codec: none`: reiner Videostrom ohne Ton und ohne Gegensprechen.

## Zurück zur bisherigen App

Diese App stoppen, die bisherige starten. Deren Kopplung liegt unverändert in ihrem eigenen Datenordner.
