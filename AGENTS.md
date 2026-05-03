# wiz-controller — Agent Skill File

This file provides context for AI agents working on this project.

---

## What This Project Is

A Go web app that controls WiZ smart lights on the local network. It serves an HTMX-powered UI where anyone on the network can adjust brightness. Communication with lights happens over **UDP on port 38899** using the WiZ JSON-RPC protocol.

Build: `go build -o bin/wiz-controller ./cmd/wiz-controller`
Run: `./bin/wiz-controller` (reads `config.json` by default, or set `CONFIG_PATH` env var)
Port: 80 by default, override with `PORT` env var

**For testing from a laptop:** Start the app in background with `detach: true` on port 8080 using `PORT=8080` env var so the user can access it from their machine while the app runs independently on the headless server.

---

## Light Inventory

| IP             | Name          | Location / Notes                              | Preferred Scene |
|----------------|---------------|-----------------------------------------------|-----------------|
| 192.168.1.24   | Couch         | Living room — individual fixture              | 11              |
| 192.168.1.25   | Window corner | Living room — individual fixture              | 11              |
| 192.168.1.26   | Record player | Living room — individual fixture              | 11              |
| 192.168.1.27   | Behind desk   | Living room — individual fixture              | 11              |
| 192.168.1.28   | Bedroom1      | Bedroom — **inside same fixture as Bedroom2** | 11              |
| 192.168.1.29   | Bedroom2      | Bedroom — **inside same fixture as Bedroom1** | 11              |

### Key distinctions
- **Living room lights (.24–.27)**: older generation, individual fixtures, Scene 11 is the preferred warm default.
- **Bedroom lights (.28–.29)**: newer generation, physically inside one shared fixture, always controlled together as the `Bedroom` group. Their `getPilot` response includes a `temp` field (color temperature in Kelvin, e.g. 4200K) which older lights do not return. Preferred scene is also 11.

---

## WiZ UDP Protocol

All communication is fire-and-forget UDP. Send a JSON payload, read the response.

**Port:** `38899`  
**Timeout used in this project:** 2 seconds

### Get current state (`getPilot`)

Request:
```json
{"method": "getPilot", "params": {}}
```

Example response (living room light on Scene 11):
```json
{
  "method": "getPilot",
  "env": "pro",
  "result": {
    "mac": "444f8e1b6ae6",
    "rssi": -66,
    "state": true,
    "sceneId": 11,
    "dimming": 89
  }
}
```

Example response (bedroom light on Scene 11 with color temperature):
```json
{
  "method": "getPilot",
  "env": "pro",
  "result": {
    "mac": "cc40859cce96",
    "rssi": -60,
    "state": true,
    "sceneId": 11,
    "temp": 4200,
    "dimming": 100
  }
}
```

Example response (light with manual RGB color set — no active scene):
```json
{
  "result": {
    "state": true,
    "sceneId": 0,
    "r": 255,
    "g": 200,
    "b": 100,
    "c": 0,
    "w": 0,
    "dimming": 50
  }
}
```

### Color/state fields in `result`

| Field     | Type    | Range     | Meaning                                              |
|-----------|---------|-----------|------------------------------------------------------|
| `state`   | bool    | —         | On/off                                               |
| `dimming` | int     | 10–100    | Brightness percent                                   |
| `sceneId` | int     | 0–32+     | Built-in scene; 0 = manual/custom color              |
| `temp`    | int     | 2200–6500 | Color temperature in Kelvin (newer lights only)      |
| `r`       | int     | 0–255     | Red (present only when manual RGB is active)         |
| `g`       | int     | 0–255     | Green (present only when manual RGB is active)       |
| `b`       | int     | 0–255     | Blue (present only when manual RGB is active)        |
| `c`       | int     | 0–255     | Cold white channel (present only when manually set)  |
| `w`       | int     | 0–255     | Warm white channel (present only when manually set)  |

### Set state (`setPilot`)

Turn off:
```json
{"method": "setState", "params": {"state": false}}
```

Set brightness (also turns on):
```json
{"method": "setPilot", "params": {"state": true, "dimming": 80}}
```

Set scene:
```json
{"method": "setPilot", "params": {"sceneId": 11}}
```

Set color temperature:
```json
{"method": "setPilot", "params": {"state": true, "ct": 3000, "dimming": 80}}
```

Set RGB color:
```json
{"method": "setPilot", "params": {"state": true, "r": 255, "g": 200, "b": 100, "dimming": 80}}
```

Success response:
```json
{"method": "setPilot", "env": "pro", "result": {"success": true}}
```

---

## Project Structure

```
wiz-controller/
├── cmd/wiz-controller/main.go          # Entry point, routes
├── config.json                         # Light IPs, names, scenes, groups
├── internal/
│   ├── services/
│   │   ├── config.go                   # Config loading, Light/Group/Config structs
│   │   └── wiz.go                      # UDP comms: GetLightState, SetBrightness, SetScene, SetLightState
│   └── handlers/
│       ├── handlers.go                 # HTTP handlers
│       └── templates.go                # Go HTML templates + template func map
├── static/style.css
└── Dockerfile
```

### HTTP Routes

| Method | Path                              | Handler              | Notes                                      |
|--------|-----------------------------------|----------------------|--------------------------------------------|
| GET    | `/`                               | `Home`               | Renders full UI                            |
| POST   | `/api/lights/{ip}/brightness`     | `SetBrightness`      | Sets one light's brightness                |
| POST   | `/api/lights/all/brightness`      | `SetAllBrightness`   | Sets all **non-grouped** lights            |
| POST   | `/api/groups/{group}/brightness`  | `SetGroupBrightness` | Sets all lights in a named group           |
| POST   | `/api/sync-scenes`                | `SyncScenes`         | Resets all lights to their preferred scene |

---

## Config File Format (`config.json`)

```json
{
  "WizLights": {
    "Ips": ["192.168.1.24", ...],
    "Names": ["Couch", ...],
    "PreferredScenes": [11, 11, 11, 11, 11, 11],
    "Groups": [
      { "name": "Bedroom", "ips": ["192.168.1.28", "192.168.1.29"] }
    ]
  }
}
```

- Arrays are positionally aligned: `Ips[i]`, `Names[i]`, `PreferredScenes[i]` all refer to the same light.
- Groups cause those lights to be hidden from individual sliders in the UI and exposed as a single group slider.
- Default scene is 11 if `PreferredScenes` length doesn't match `Ips`.

---

## Design Decisions & History

- **Scene 11** is the preferred warm scene for all lights. It does not expose RGB/temp values in `getPilot` — those are controlled internally by the scene.
- Known scene numbers: **6** = Cozy, **11** = Warm white, **12** = Daylight.
- The owner wanted to match a specific "warm" light color from the proprietary WiZ app. A future improvement is to fetch the current color/state from lights and display it in the UI so the owner can reproduce settings without switching back to the proprietary app.
- `LightState` struct in `wiz.go` currently only captures `IsOn` and `Brightness`. It **does not yet capture** `r`, `g`, `b`, `c`, `w`, `temp`, or `sceneId` — extending it is a known future task.
- The `allLightCardsTemplate` and home template exclude grouped lights from individual controls using the `isInGroup` template func.
- HTMX is used for all UI interactions — no custom JS.
