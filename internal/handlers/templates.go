package handlers

import (
	"text/template"
)

var homeTemplate = template.Must(template.New("home").Parse(`<!DOCTYPE html>
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
			<label>All Lights</label>
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
			{{range .}}
			<div class="card" id="light-{{.IP}}">
				<h5>{{.Name}}</h5>
				<div class="brightness-control">
					<label>Brightness</label>
					<input type="range" min="0" max="100" value="0"
						hx-post="/api/lights/{{.IP}}/brightness"
						hx-trigger="change"
						hx-swap="outerHTML"
						name="brightness" />
				</div>
			</div>
			{{end}}
		</div>
	</div>
</body>
</html>`))

var lightCardTemplate = template.Must(template.New("light-card").Parse(`<div class="card" id="light-{{.IP}}">
	<h5>{{.Name}}</h5>
	<div class="brightness-control">
		<label>Brightness</label>
		<input type="range" min="0" max="100" value="{{.Brightness}}"
			hx-post="/api/lights/{{.IP}}/brightness"
			hx-trigger="change"
			hx-swap="outerHTML"
			name="brightness" />
	</div>
</div>`))

var allLightCardsTemplate = template.Must(template.New("all-light-cards").Parse(`{{range .Lights}}<div class="card" id="light-{{.IP}}">
	<h5>{{.Name}}</h5>
	<div class="brightness-control">
		<label>Brightness</label>
		<input type="range" min="0" max="100" value="{{$.Brightness}}"
			hx-post="/api/lights/{{.IP}}/brightness"
			hx-trigger="change"
			hx-swap="outerHTML"
			name="brightness" />
	</div>
</div>{{end}}`))
