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

var funcMap = template.FuncMap{
	"isInGroup":     isInGroup,
	"getBrightness": getBrightness,
	"ipToID":        ipToID,
}

var homeTemplate = template.Must(template.New("home").Funcs(funcMap).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>WiZ Controller</title>
	<link rel="stylesheet" href="/static/style.css">
	<script src="https://unpkg.com/htmx.org@2.0.0"></script>
</head>
<body>
	<div class="container">
		<div class="master-control">
			<label>All Living Room Lights</label>
			<div class="master-slider-container">
				<input type="range" min="0" max="100" value="0"
					hx-post="/api/lights/all/brightness"
					hx-trigger="change"
					hx-swap="innerHTML"
					hx-target=".light-grid"
					name="brightness" />
			</div>
			</div>

		<div class="light-grid" id="light-grid">
			{{range .Lights}}
			{{if not (isInGroup . $.Groups)}}
			<div class="card" id="light-{{ipToID .IP}}">
				<h5>{{.Name}}</h5>
				<div class="brightness-control">
					<label>Brightness</label>
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
			
			{{range .Groups}}
			<div class="card" id="group-{{.Name}}" hx-preserve>
				<h5>{{.Name}}</h5>
				<div class="brightness-control">
					<label>Brightness</label>
					<input type="range" min="0" max="100" value="{{getBrightness $.State (index .IPs 0)}}"
						hx-post="/api/groups/{{.Name}}/brightness"
						hx-trigger="change"
						hx-swap="innerHTML"
						hx-target="#group-{{.Name}} .brightness-control"
						name="brightness" />
				</div>
			</div>
			{{end}}
		</div>
	</div>
</body>
</html>`))

var lightCardTemplate = template.Must(template.New("light-card").Funcs(funcMap).Parse(`<div class="card" id="light-{{ipToID .IP}}">
	<h5>{{.Name}}</h5>
	<div class="brightness-control">
		<label>Brightness</label>
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
		<label>Brightness</label>
		<input type="range" min="0" max="100" value="{{$.Brightness}}"
			hx-post="/api/lights/{{.IP}}/brightness"
			hx-trigger="change"
			hx-swap="outerHTML"
			hx-target="#light-{{ipToID .IP}}"
			name="brightness" />
	</div>
</div>
{{end}}
{{end}}
{{range .Groups}}
<div class="card" id="group-{{.Name}}" hx-preserve>
	<h5>{{.Name}}</h5>
	<div class="brightness-control">
		<label>Brightness</label>
		<input type="range" min="0" max="100"
			hx-post="/api/groups/{{.Name}}/brightness"
			hx-trigger="change"
			hx-swap="innerHTML"
			hx-target="#group-{{.Name}} .brightness-control"
			name="brightness" />
	</div>
</div>
{{end}}`))

var groupLightCardsTemplate = template.Must(template.New("group-light-cards").Parse(`<div class="brightness-control">
	<label>Brightness</label>
	<input type="range" min="0" max="100" value="{{.Brightness}}"
		hx-post="/api/groups/{{.Group}}/brightness"
		hx-trigger="change"
		hx-swap="innerHTML"
		hx-target="#group-{{.Group}} .brightness-control"
		name="brightness" />
</div>`))