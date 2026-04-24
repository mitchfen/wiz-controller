package handlers

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/mitchfen/wiz-controller/internal/services"
)

type Handlers struct {
	cfg    *services.Config
	wiz    *services.WizService
	mu     sync.RWMutex
	state  map[string]*services.LightState
	lights map[string]string // ip -> name mapping
}

func NewHandlers(cfg *services.Config, wiz *services.WizService) *Handlers {
	h := &Handlers{
		cfg:    cfg,
		wiz:    wiz,
		state:  make(map[string]*services.LightState),
		lights: make(map[string]string),
	}

	// Load initial state and build IP->name mapping
	for _, light := range cfg.Lights {
		h.lights[light.IP] = light.Name
		if state, err := wiz.GetLightState(light.IP); err == nil {
			h.state[light.IP] = state
		} else {
			log.Printf("Failed to get state for %s: %v", light.IP, err)
			h.state[light.IP] = &services.LightState{IsOn: false, Brightness: 0}
		}
	}

	return h
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	if err := homeTemplate.Execute(w, h.cfg.Lights); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

func (h *Handlers) SetBrightness(w http.ResponseWriter, r *http.Request) {
	ip := r.PathValue("ip")
	brightness, err := strconv.Atoi(r.FormValue("brightness"))
	if err != nil {
		http.Error(w, "Invalid brightness value", http.StatusBadRequest)
		return
	}

	if brightness < 0 || brightness > 100 {
		http.Error(w, "Brightness must be 0-100", http.StatusBadRequest)
		return
	}

	// Set light state
	if brightness == 0 {
		_ = h.wiz.SetLightState(ip, false)
	} else {
		_ = h.wiz.SetBrightness(ip, brightness)
	}

	// Update local state
	h.mu.Lock()
	h.state[ip] = &services.LightState{IsOn: brightness > 0, Brightness: brightness}
	h.mu.Unlock()

	// Return HTML for HTMX swap
	w.Header().Set("Content-Type", "text/html")
	if err := lightCardTemplate.Execute(w, map[string]interface{}{
		"IP":         ip,
		"Name":       h.lights[ip],
		"Brightness": brightness,
	}); err != nil {
		log.Printf("Template error: %v", err)
	}
}

func (h *Handlers) SetAllBrightness(w http.ResponseWriter, r *http.Request) {
	brightness, err := strconv.Atoi(r.FormValue("brightness"))
	if err != nil {
		http.Error(w, "Invalid brightness value", http.StatusBadRequest)
		return
	}

	if brightness < 0 || brightness > 100 {
		http.Error(w, "Brightness must be 0-100", http.StatusBadRequest)
		return
	}

	// Set all lights
	for _, light := range h.cfg.Lights {
		if brightness == 0 {
			_ = h.wiz.SetLightState(light.IP, false)
		} else {
			_ = h.wiz.SetBrightness(light.IP, brightness)
		}

		h.mu.Lock()
		h.state[light.IP] = &services.LightState{IsOn: brightness > 0, Brightness: brightness}
		h.mu.Unlock()
	}

	// Return all light cards for HTMX swap
	w.Header().Set("Content-Type", "text/html")
	if err := allLightCardsTemplate.Execute(w, map[string]interface{}{
		"Lights":     h.cfg.Lights,
		"Brightness": brightness,
	}); err != nil {
		log.Printf("Template error: %v", err)
	}
}

func (h *Handlers) getLightBrightness(ip string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if state, ok := h.state[ip]; ok && state.IsOn {
		return state.Brightness
	}
	return 0
}
