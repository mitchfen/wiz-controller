#!/usr/bin/env python3
"""Set WiZ lights to warm (Cozy in living room, Warm white in bedroom) or daylight."""

import socket
import json
import sys

LIVING_ROOM_LIGHTS = [
    ("192.168.1.24", "Couch"),
    ("192.168.1.25", "Window corner"),
    ("192.168.1.26", "Record player"),
    ("192.168.1.27", "Behind desk"),
]

BEDROOM_LIGHTS = [
    ("192.168.1.28", "Bedroom1"),
    ("192.168.1.29", "Bedroom2"),
]

ALL_LIGHTS = LIVING_ROOM_LIGHTS + BEDROOM_LIGHTS

PORT = 38899
TIMEOUT = 2

# Scene mappings
SCENE_COZY = 6
SCENE_WARM_WHITE = 11
SCENE_DAYLIGHT = 12

PRESETS = {
    "warm": {
        "description": "Cozy in living room, Warm white in bedroom",
        "lights": {
            "living_room": SCENE_COZY,
            "bedroom": SCENE_WARM_WHITE,
        }
    },
    "daylight": {
        "description": "Daylight in all rooms",
        "lights": {
            "all": SCENE_DAYLIGHT,
        }
    }
}

def set_lights_to_scene(lights, scene_id):
    """Helper to set a list of lights to a scene."""
    payload = json.dumps({"method": "setPilot", "params": {"sceneId": scene_id}}).encode()
    success_count = 0
    
    for ip, name in lights:
        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
            sock.settimeout(TIMEOUT)
            sock.sendto(payload, (ip, PORT))
            data, _ = sock.recvfrom(1024)
            r = json.loads(data)
            
            success = r.get("result", {}).get("success", False)
            status = "✓ Success" if success else "✗ Failed"
            print(f"{name:<16} {ip:<15} Scene {scene_id:<5} {status:<20}")
            if success:
                success_count += 1
        except Exception as e:
            print(f"{name:<16} {ip:<15} Scene {scene_id:<5} ✗ Error: {e}")
        finally:
            sock.close()
    
    return success_count, len(lights)

def set_preset(preset_name):
    """Set lights using a preset configuration."""
    if preset_name not in PRESETS:
        print(f"❌ Error: Preset '{preset_name}' not found.")
        print("Available presets:")
        for pname in sorted(PRESETS.keys()):
            print(f"  {pname}: {PRESETS[pname]['description']}")
        return False
    
    preset = PRESETS[preset_name]
    print(f"Setting preset: {preset_name.upper()} ({preset['description']})...\n")
    print(f"{'Name':<16} {'IP':<15} {'Scene':<7} {'Result':<20}")
    print("-" * 58)
    
    total_success = 0
    total_lights = 0
    
    if "all" in preset["lights"]:
        # Apply same scene to all lights
        scene_id = preset["lights"]["all"]
        success, count = set_lights_to_scene(ALL_LIGHTS, scene_id)
        total_success += success
        total_lights += count
    else:
        # Apply different scenes to different room groups
        if "living_room" in preset["lights"]:
            scene_id = preset["lights"]["living_room"]
            success, count = set_lights_to_scene(LIVING_ROOM_LIGHTS, scene_id)
            total_success += success
            total_lights += count
        
        if "bedroom" in preset["lights"]:
            scene_id = preset["lights"]["bedroom"]
            success, count = set_lights_to_scene(BEDROOM_LIGHTS, scene_id)
            total_success += success
            total_lights += count
    
    print()
    print(f"Result: {total_success}/{total_lights} lights updated successfully")
    return total_success == total_lights

def main():
    if len(sys.argv) > 1:
        # Command-line argument provided
        arg = sys.argv[1].lower()
        set_preset(arg)
    else:
        # Interactive mode
        print("Available presets:")
        for pname in sorted(PRESETS.keys()):
            print(f"  {pname}: {PRESETS[pname]['description']}")
        print()
        
        while True:
            try:
                choice = input("Enter preset name (warm/daylight): ").strip().lower()
                
                if choice in PRESETS:
                    set_preset(choice)
                    break
                else:
                    print(f"❌ Preset '{choice}' not found. Try again.")
            except KeyboardInterrupt:
                print("\n\nCancelled.")
                sys.exit(0)

if __name__ == "__main__":
    main()
