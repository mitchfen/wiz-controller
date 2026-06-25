#!/usr/bin/env bash
set -e

IPS=(
  "192.168.1.24"
  "192.168.1.25"
  "192.168.1.26"
  "192.168.1.27"
  "192.168.1.28"
  "192.168.1.29"
)
NAMES=(
  "Couch"
  "Window corner"
  "Record player"
  "Behind desk"
  "Bedroom1"
  "Bedroom2"
)

PORT=38899

get_scene_name() {
  local scene_id=$1
  case "$scene_id" in
    6) echo "Cozy" ;;
    11) echo "Warm white" ;;
    12) echo "Daylight" ;;
    "") echo "N/A" ;;
    *) echo "scene $scene_id" ;;
  esac
}

printf "%-16s %-15s %-5s %-20s %-12s %s\n" "Name" "IP" "State" "Scene" "Brightness" "Temp (K)"
printf "%s\n" "--------------------------------------------------------------------------------"

for i in "${!IPS[@]}"; do
  ip="${IPS[i]}"
  name="${NAMES[i]}"
  
  response=$(echo '{"method":"getPilot","params":{}}' | nc -u -w 2 "$ip" "$PORT" 2>/dev/null)
  
  if [ -z "$response" ]; then
    printf "%-16s %-15s %s\n" "$name" "$ip" "ERROR: timeout or no response"
    continue
  fi
  
  state_val=$(echo "$response" | grep -o '"state":[a-z]*' | cut -d: -f2)
  if [ "$state_val" = "true" ]; then
    state="on"
  else
    state="off"
  fi
  
  scene_id=$(echo "$response" | grep -o '"sceneId":[0-9]*' | cut -d: -f2)
  scene_name=$(get_scene_name "$scene_id")
  
  dimming=$(echo "$response" | grep -o '"dimming":[0-9]*' | cut -d: -f2)
  if [ -n "$dimming" ]; then
    brightness="${dimming}%"
  else
    brightness="N/A"
  fi
  
  temp=$(echo "$response" | grep -o '"temp":[0-9]*' | cut -d: -f2)
  if [ -z "$temp" ]; then
    temp="—"
  fi
  
  printf "%-16s %-15s %-5s %-20s %-12s %s\n" "$name" "$ip" "$state" "$scene_name" "$brightness" "$temp"
done
