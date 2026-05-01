#!/usr/bin/env python3
"""Query the current state of all WiZ lights."""

import socket
import json

LIGHTS = [
    ("192.168.1.24", "Couch"),
    ("192.168.1.25", "Window corner"),
    ("192.168.1.26", "Record player"),
    ("192.168.1.27", "Behind desk"),
    ("192.168.1.28", "Bedroom1"),
    ("192.168.1.29", "Bedroom2"),
]

PORT = 38899
TIMEOUT = 2

KNOWN_SCENES = {
    6: "Cozy",
    11: "Warm white",
    12: "Daylight",
}

payload = json.dumps({"method": "getPilot", "params": {}}).encode()

print(f"{'Name':<16} {'IP':<15} {'State':<5} {'Scene':<20} {'Brightness':<12} {'Temp (K)'}")
print("-" * 80)

for ip, name in LIGHTS:
    try:
        sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        sock.settimeout(TIMEOUT)
        sock.sendto(payload, (ip, PORT))
        data, _ = sock.recvfrom(1024)
        r = json.loads(data)["result"]

        state = "on" if r.get("state") else "off"
        scene_id = r.get("sceneId", "N/A")
        scene_name = KNOWN_SCENES.get(scene_id, f"scene {scene_id}")
        brightness = r.get("dimming", "N/A")
        temp = r.get("temp", "—")

        print(f"{name:<16} {ip:<15} {state:<5} {scene_name:<20} {str(brightness)+'%':<12} {temp}")
    except Exception as e:
        print(f"{name:<16} {ip:<15} ERROR: {e}")
    finally:
        sock.close()
