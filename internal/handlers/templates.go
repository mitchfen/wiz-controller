package handlers

import (
	"strings"
	"text/template"

	"github.com/mitchfen/wiz-controller/internal/services"
)

func isInGroup(light services.Light, groups []services.Group) bool {
	for _, group := range groups {
		for _, ip := range group.IPs {
			if ip == light.IP {
				return true
			}
		}
	}
	return false
}

func isInSpecificGroup(light services.Light, groups []services.Group, groupName string) bool {
	for _, group := range groups {
		if group.Name == groupName {
			for _, ip := range group.IPs {
				if ip == light.IP {
					return true
				}
			}
		}
	}
	return false
}

func getBrightness(state map[string]*services.LightState, ip string) int {
	if s, ok := state[ip]; ok && s.IsOn {
		return s.Brightness
	}
	return 0
}

// ipToID converts an IP address to a safe HTML/CSS identifier by replacing dots with dashes.
func ipToID(ip string) string {
	return strings.ReplaceAll(ip, ".", "-")
}

// groupToID converts a group name to a safe HTML/CSS identifier by replacing spaces with dashes and converting to lowercase.
func groupToID(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", "-"))
}

var funcMap = template.FuncMap{
	"isInGroup":         isInGroup,
	"isInSpecificGroup": isInSpecificGroup,
	"getBrightness":     getBrightness,
	"ipToID":            ipToID,
	"groupToID":         groupToID,
	"getGroupIPs":       getGroupIPs,
}

var homeTemplate = template.Must(template.New("home").Funcs(funcMap).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>WiZ Controller</title>
	<link rel="stylesheet" href="/static/style.css">
	<script src="https://unpkg.com/htmx.org@2.0.0"></script>
	<script>
		document.body.addEventListener('htmx:afterRequest', function(evt) {
			if (evt.detail.xhr.status === 200) {
				const button = evt.detail.target;
				if (button.classList.contains('scene-btn')) {
					const feedbackDiv = button.closest('.room-section').querySelector('.scene-feedback');
					const sceneName = button.textContent.trim();
					feedbackDiv.textContent = sceneName + ' scene activated!';
					feedbackDiv.style.display = 'block';
					setTimeout(() => {
						feedbackDiv.style.display = 'none';
					}, 2000);
				}
			}
		});

		function openColorPicker(roomName) {
			const modal = document.getElementById('color-modal-' + roomName);
			const colorInput = document.getElementById('color-input-' + roomName);
			modal.style.display = 'flex';
			colorInput.focus();
		}

		function closeColorPicker(roomName) {
			const modal = document.getElementById('color-modal-' + roomName);
			modal.style.display = 'none';
		}

		function applyColor(roomName) {
			const colorInput = document.getElementById('color-input-' + roomName);
			const color = colorInput.value;
			const rgb = hexToRgb(color);
			const url = '/api/groups/' + roomName + '/color';
			const body = 'r=' + rgb.r + '&g=' + rgb.g + '&b=' + rgb.b;
			
			fetch(url, {
				method: 'POST',
				headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
				body: body
			}).then(r => {
				if (r.ok) closeColorPicker(roomName);
			});
		}

		function hexToRgb(hex) {
			const result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
			return result ? {
				r: parseInt(result[1], 16),
				g: parseInt(result[2], 16),
				b: parseInt(result[3], 16)
			} : { r: 255, g: 255, b: 255 };
		}

		window.addEventListener('click', function(event) {
			const modals = document.querySelectorAll('.color-modal');
			modals.forEach(modal => {
				if (event.target === modal) {
					modal.style.display = 'none';
				}
			});
		});
	</script>
</head>
<body>
	<div class="container">
		<!-- LIVING ROOM SECTION -->
		<div class="room-section">
			<h2 class="room-title">Living Room</h2>
			
			<div class="room-controls">
				<div class="master-control" id="group-living-room">
					<div class="master-slider-container brightness-control">
						<input type="range" min="0" max="100" value="{{getBrightness $.State (index (getGroupIPs $.Groups "Living Room") 0)}}"
							hx-post="/api/groups/Living Room/brightness"
							hx-trigger="change"
							hx-swap="innerHTML"
							hx-target="#group-living-room .brightness-control"
							name="brightness" />
					</div>
				</div>
				
				<button type="button" class="color-picker-btn" title="Color picker" onclick="openColorPicker('Living Room')">🎨</button>
				
				<details class="scene-menu">
					<summary>Scenes</summary>
					<div class="scene-buttons">
						<button type="button" hx-post="/api/scenes/warm-living-room" hx-trigger="click" hx-swap="none" class="scene-btn">Warm</button>
						<button type="button" hx-post="/api/scenes/daylight-living-room" hx-trigger="click" hx-swap="none" class="scene-btn">Daylight</button>
					</div>
				</details>
			</div>

			<div class="scene-feedback" style="display: none;"></div>

			<div class="individual-lights">
				{{range .Lights}}
				{{if not (isInSpecificGroup . $.Groups "Bedroom")}}
				<div class="card" id="light-{{ipToID .IP}}">
					<h5>{{.Name}}</h5>
					<div class="brightness-control">
						<input type="range" min="0" max="100" value="{{getBrightness $.State .IP}}"
							hx-post="/api/lights/{{.IP}}/brightness"
							hx-trigger="change"
							hx-swap="outerHTML"
							hx-target="#light-{{ipToID .IP}}"
							name="brightness" />
					</div>
				</div>
				{{end}}
				{{end}}
			</div>
		</div>

		<!-- LIVING ROOM COLOR PICKER MODAL -->
		<div id="color-modal-Living Room" class="color-modal">
			<div class="color-modal-content">
				<div class="color-modal-header">
					<h3>Pick a Color</h3>
					<button type="button" class="color-modal-close" onclick="closeColorPicker('Living Room')">✕</button>
				</div>
				<div class="color-modal-body">
					<input type="color" id="color-input-Living Room" value="#ff0000" class="color-input" />
				</div>
				<div class="color-modal-footer">
					<button type="button" class="color-modal-btn color-modal-cancel" onclick="closeColorPicker('Living Room')">Cancel</button>
					<button type="button" class="color-modal-btn color-modal-apply" onclick="applyColor('Living Room')">Apply</button>
				</div>
			</div>
		</div>

		<!-- BEDROOM SECTION -->
		<div class="room-section">
			<h2 class="room-title">Bedroom</h2>
			
			<div class="room-controls">
				<div class="master-control" id="group-bedroom">
					<div class="master-slider-container brightness-control">
						<input type="range" min="0" max="100" value="{{getBrightness $.State (index (getGroupIPs $.Groups "Bedroom") 0)}}"
							hx-post="/api/groups/Bedroom/brightness"
							hx-trigger="change"
							hx-swap="innerHTML"
							hx-target="#group-bedroom .brightness-control"
							name="brightness" />
					</div>
				</div>
				
				<button type="button" class="color-picker-btn" title="Color picker" onclick="openColorPicker('Bedroom')">🎨</button>
				
				<details class="scene-menu">
					<summary>Scenes</summary>
					<div class="scene-buttons">
						<button type="button" hx-post="/api/scenes/warm-bedroom" hx-trigger="click" hx-swap="none" class="scene-btn">Warm</button>
						<button type="button" hx-post="/api/scenes/daylight-bedroom" hx-trigger="click" hx-swap="none" class="scene-btn">Daylight</button>
					</div>
				</details>
			</div>

			<div class="scene-feedback" style="display: none;"></div>
		</div>

		<!-- BEDROOM COLOR PICKER MODAL -->
		<div id="color-modal-Bedroom" class="color-modal">
			<div class="color-modal-content">
				<div class="color-modal-header">
					<h3>Pick a Color</h3>
					<button type="button" class="color-modal-close" onclick="closeColorPicker('Bedroom')">✕</button>
				</div>
				<div class="color-modal-body">
					<input type="color" id="color-input-Bedroom" value="#ff0000" class="color-input" />
				</div>
				<div class="color-modal-footer">
					<button type="button" class="color-modal-btn color-modal-cancel" onclick="closeColorPicker('Bedroom')">Cancel</button>
					<button type="button" class="color-modal-btn color-modal-apply" onclick="applyColor('Bedroom')">Apply</button>
				</div>
			</div>
		</div>
	</div>
</body>
</html>`))

var lightCardTemplate = template.Must(template.New("light-card").Funcs(funcMap).Parse(`<div class="card" id="light-{{ipToID .IP}}">
	<h5>{{.Name}}</h5>
	<div class="brightness-control">
		<input type="range" min="0" max="100" value="{{.Brightness}}"
			hx-post="/api/lights/{{.IP}}/brightness"
			hx-trigger="change"
			hx-swap="outerHTML"
			hx-target="#light-{{ipToID .IP}}"
			name="brightness" />
	</div>
</div>`))

var allLightCardsTemplate = template.Must(template.New("all-light-cards").Funcs(funcMap).Parse(`{{range .Lights}}
{{if not (isInGroup . $.Groups)}}
<div class="card" id="light-{{ipToID .IP}}">
	<h5>{{.Name}}</h5>
	<div class="brightness-control">
		<input type="range" min="0" max="100" value="{{$.Brightness}}"
			hx-post="/api/lights/{{.IP}}/brightness"
			hx-trigger="change"
			hx-swap="outerHTML"
			hx-target="#light-{{ipToID .IP}}"
			name="brightness" />
	</div>
</div>
{{end}}
{{end}}`))

var groupLightCardsTemplate = template.Must(template.New("group-light-cards").Funcs(funcMap).Parse(`<div class="brightness-control">
	<input type="range" min="0" max="100" value="{{.Brightness}}"
		hx-post="/api/groups/{{.Group}}/brightness"
		hx-trigger="change"
		hx-swap="innerHTML"
		hx-target="#group-{{groupToID .Group}} .brightness-control"
		name="brightness" />
</div>`))