#!/usr/bin/env bash
set -e

LIVING_ROOM_IPS=(
  "192.168.1.24"
  "192.168.1.25"
  "192.168.1.26"
  "192.168.1.27"
)
LIVING_ROOM_NAMES=(
  "Couch"
  "Window corner"
  "Record player"
  "Behind desk"
)

BEDROOM_IPS=(
  "192.168.1.28"
  "192.168.1.29"
)
BEDROOM_NAMES=(
  "Bedroom1"
  "Bedroom2"
)

PORT=38899

# Scene mappings
SCENE_COZY=6
SCENE_WARM_WHITE=11
SCENE_DAYLIGHT=12

set_light_scene() {
  local ip=$1
  local name=$2
  local scene_id=$3
  
  local payload="{\"method\":\"setPilot\",\"params\":{\"sceneId\":$scene_id}}"
  local response=$(echo "$payload" | nc -u -w 2 "$ip" "$PORT" 2>/dev/null)
  
  local success="false"
  if [ -n "$response" ]; then
    local success_val=$(echo "$response" | grep -o '"success":[a-z]*' | cut -d: -f2)
    if [ "$success_val" = "true" ]; then
      success="true"
    fi
  fi
  
  if [ "$success" = "true" ]; then
    printf "%-16s %-15s Scene %-5s %s\n" "$name" "$ip" "$scene_id" "✓ Success"
    return 0
  else
    printf "%-16s %-15s Scene %-5s %s\n" "$name" "$ip" "$scene_id" "✗ Failed"
    return 1
  fi
}

apply_preset() {
  local preset=$1
  local total_success=0
  local total_lights=0
  
  printf "%-16s %-15s %-11s %s\n" "Name" "IP" "Scene" "Result"
  printf "%s\n" "--------------------------------------------------------"
  
  if [ "$preset" = "warm" ]; then
    # Living room: Cozy (6)
    for i in "${!LIVING_ROOM_IPS[@]}"; do
      if set_light_scene "${LIVING_ROOM_IPS[i]}" "${LIVING_ROOM_NAMES[i]}" "$SCENE_COZY"; then
        total_success=$((total_success + 1))
      fi
      total_lights=$((total_lights + 1))
    done
    # Bedroom: Warm white (11)
    for i in "${!BEDROOM_IPS[@]}"; do
      if set_light_scene "${BEDROOM_IPS[i]}" "${BEDROOM_NAMES[i]}" "$SCENE_WARM_WHITE"; then
        total_success=$((total_success + 1))
      fi
      total_lights=$((total_lights + 1))
    done
  elif [ "$preset" = "daylight" ]; then
    # All lights: Daylight (12)
    # Living room
    for i in "${!LIVING_ROOM_IPS[@]}"; do
      if set_light_scene "${LIVING_ROOM_IPS[i]}" "${LIVING_ROOM_NAMES[i]}" "$SCENE_DAYLIGHT"; then
        total_success=$((total_success + 1))
      fi
      total_lights=$((total_lights + 1))
    done
    # Bedroom
    for i in "${!BEDROOM_IPS[@]}"; do
      if set_light_scene "${BEDROOM_IPS[i]}" "${BEDROOM_NAMES[i]}" "$SCENE_DAYLIGHT"; then
        total_success=$((total_success + 1))
      fi
      total_lights=$((total_lights + 1))
    done
  fi
  
  echo ""
  echo "Result: $total_success/$total_lights lights updated successfully"
  
  if [ "$total_success" -eq "$total_lights" ]; then
    return 0
  else
    return 1
  fi
}

show_presets() {
  echo "Available presets:"
  echo "  warm: Cozy in living room, Warm white in bedroom"
  echo "  daylight: Daylight in all rooms"
}

# Main execution
preset_arg=""
if [ -n "$1" ]; then
  preset_arg=$(echo "$1" | tr '[:upper:]' '[:lower:]')
fi

if [ -n "$preset_arg" ]; then
  if [ "$preset_arg" = "warm" ] || [ "$preset_arg" = "daylight" ]; then
    apply_preset "$preset_arg"
  else
    echo "❌ Error: Preset '$preset_arg' not found."
    show_presets
    exit 1
  fi
else
  show_presets
  echo ""
  while true; do
    read -r -p "Enter preset name (warm/daylight): " choice
    choice=$(echo "$choice" | tr -d '[:space:]' | tr '[:upper:]' '[:lower:]')
    if [ "$choice" = "warm" ] || [ "$choice" = "daylight" ]; then
      apply_preset "$choice"
      break
    else
      echo "❌ Preset '$choice' not found. Try again."
    fi
  done
fi
