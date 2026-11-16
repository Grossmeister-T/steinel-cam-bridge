#!/usr/bin/env python3
"""Derive the local GMT add-on files (ha-addon-gmt/) from the upstream beta add-on.

Usage: python scripts/make-gmt-addon.py <gmt-suffix>      e.g. gmt.1
Keeps upstream's options in sync and adds the fork's own ones.
"""
import re
import shutil
import sys
from pathlib import Path

root = Path(__file__).resolve().parent.parent
src = root / "ha-addon-beta"
dst = root / "ha-addon-gmt"
suffix = sys.argv[1] if len(sys.argv) > 1 else "gmt.1"

(dst / "translations").mkdir(parents=True, exist_ok=True)
for name in ("icon.png", "logo.png"):
    shutil.copyfile(src / name, dst / name)

# ---- config.yaml ----
s = (src / "config.yaml").read_text(encoding="utf-8")
base_version = re.search(r'^version: "(.*)"', s, re.M).group(1)
version = f"{base_version}-{suffix}"


def sub(pattern, repl, text, count=1):
    new, n = re.subn(pattern, repl, text, count=count, flags=re.M)
    assert n >= 1, pattern
    return new


s = sub(r'^name: ".*"', 'name: "Steinel CAM Bridge GMT"', s)
s = sub(r'^description: ".*"', 'description: "Eigener Fork der Steinel CAM Bridge: Sofortbild, Standbild bei Aussetzern, echte Auflösung, Ton abschaltbar, Grundlicht-Schalter"', s)
s = sub(r'^version: ".*"', f'version: "{version}"', s)
s = sub(r'^slug: ".*"', 'slug: "steinel_cam_bridge_gmt"', s)
s = sub(r'^arch:\n(  - .*\n)+', "arch:\n  - aarch64\n", s)
s = sub(r'^environment:\n  IS_BETA: "true"\n', "", s)
s = sub(r'^image: .*\n', "", s)
s = sub(r'^  audio_codec: "aac"\n', '  audio_codec: "aac"\n  instant_start: false\n  hold_last_frame_seconds: 300\n  debug_api: false\n', s)
s = sub(r'^      audio_codec: "list\(aac\|pcmu\)\?"', '      audio_codec: "list(aac|pcmu|none)?"', s)
s = sub(r'^  audio_codec: "list\(aac\|pcmu\)"', '  audio_codec: "list(aac|pcmu|none)"\n  instant_start: "bool?"\n  hold_last_frame_seconds: "int(0,3600)?"\n  debug_api: "bool?"', s)
(dst / "config.yaml").write_text(s, encoding="utf-8", newline="\n")

# ---- Dockerfile ----
(dst / "Dockerfile").write_text(
    "# Local Home Assistant add-on of the GMT fork.\n"
    "# The image is the upstream beta image (launcher, FFmpeg, runtime) with the\n"
    "# bridge binary replaced by the fork's build. The binary is cross-compiled on\n"
    "# the PC (scripts/build-gmt.sh), so nothing is compiled on the Pi.\n"
    f"FROM ghcr.io/afrouper/steinel-cam-bridge:{base_version}\n"
    "COPY --chmod=0755 steinel-bridge /app/steinel-bridge\n",
    encoding="utf-8", newline="\n")

# ---- translations ----
extra = {
    "de": (
        '  instant_start:\n'
        '    name: "Sofortbild für neue Zuschauer"\n'
        '    description: "Schickt jedem neuen RTSP-Zuschauer sofort das aktuelle Schlüsselbild samt Folgebildern, statt bis zu 2 Sekunden auf das nächste Schlüsselbild der Kamera zu warten. Standard aus: Der Zuschauer bekommt bis zu 2 Sekunden Video auf einen Schlag; auf einem Tablet im WLAN wurde das Bild dadurch später flüssig. Sinnvoll für Rekorder und Abnehmer am Kabel."\n'
        '  hold_last_frame_seconds:\n'
        '    name: "Standbild bei Aussetzern (Sekunden)"\n'
        '    description: "Liefert die Kamera kein Bild (Verbindungsabriss), wird das letzte Schlüsselbild so viele Sekunden lang weitergesendet, damit Home Assistant und Rekorder verbunden bleiben und ohne Neuaufbau weiterlaufen. Danach bleibt der Strom still, damit ein echter Ausfall auffällt. 0 schaltet das Standbild ab. Nur L 625."\n'
        '  debug_api:\n'
        '    name: "Diagnose-Schnittstelle"\n'
        '    description: "Schaltet POST /api/debug/command frei, über das einzelne Befehle direkt an die Kamera gehen (nur L 625). Nur zur Fehlersuche einschalten und danach wieder ausschalten."\n',
        "Audio-Format des RTSP/ONVIF-Streams: 'aac' (transkodiert, empfohlen), 'pcmu' (Rohdaten) oder 'none' (kein Ton, kein Gegensprechen, reiner Videostrom).",
    ),
    "en": (
        '  instant_start:\n'
        '    name: "Instant start for new viewers"\n'
        '    description: "Sends every new RTSP viewer the current keyframe and the frames after it right away instead of waiting up to 2 seconds for the camera\'s next keyframe. Off by default: the viewer receives up to 2 seconds of video at once; on a Wi-Fi tablet this delayed smooth playback. Useful for recorders and wired clients."\n'
        '  hold_last_frame_seconds:\n'
        '    name: "Hold last frame during outages (seconds)"\n'
        '    description: "While the camera delivers no video (connection drop) the last keyframe is repeated for this many seconds so Home Assistant and recorders stay connected and resume without reconnecting. After that the stream goes silent so a real outage is noticed. 0 disables it. L 625 only."\n'
        '  debug_api:\n'
        '    name: "Diagnostic API"\n'
        '    description: "Enables POST /api/debug/command, which passes single commands straight to the camera (L 625 only). Switch on for troubleshooting only and off again afterwards."\n',
        "Audio format of the RTSP/ONVIF stream: 'aac' (transcoded, recommended), 'pcmu' (raw) or 'none' (no audio, no talk-back, video only).",
    ),
}
for lang, (block, codec_desc) in extra.items():
    t = (src / "translations" / f"{lang}.yaml").read_text(encoding="utf-8")
    m = re.search(r"^  rtsp_port:\n", t, re.M)
    assert m, lang
    t = t[:m.start()] + block + t[m.start():]
    t, n = re.subn(r'(  audio_codec:\n    name: .*\n    description: ")[^"]*"', lambda mo: mo.group(1) + codec_desc + '"', t)
    assert n == 1, lang
    (dst / "translations" / f"{lang}.yaml").write_text(t, encoding="utf-8", newline="\n")

print(f"ha-addon-gmt ready: {version} (base image {base_version})")
