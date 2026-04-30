package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/mitchfen/wiz-controller/internal/services"
)

func fetchAllStates(wiz *services.WizService, lights []services.Light) map[string]*services.LightState {
	state := make(map[string]*services.LightState, len(lights))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, light := range lights {
		wg.Add(1)
		go func(l services.Light) {
			defer wg.Done()
			s, err := wiz.GetLightState(l.IP)
			if err != nil {
				log.Printf("Failed to get state for %s: %v", l.IP, err)
				s = &services.LightState{IsOn: false, Brightness: 0}
			}
			mu.Lock()
			state[l.IP] = s
			mu.Unlock()
		}(light)
	}

	wg.Wait()
	return state
}

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

	// Build IP->name mapping
	for _, light := range cfg.Lights {
		h.lights[light.IP] = light.Name
	}

	// Load initial state concurrently
	for ip, s := range fetchAllStates(wiz, cfg.Lights) {
		h.state[ip] = s
	}

	return h
}

func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	state := fetchAllStates(h.wiz, h.cfg.Lights)

	if err := homeTemplate.Execute(w, map[string]interface{}{
		"Lights": h.cfg.Lights,
		"Groups": h.cfg.Groups,
		"State":  state,
	}); err != nil {
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

func (h *Handlers) SetGroupBrightness(w http.ResponseWriter, r *http.Request) {
	groupName := r.PathValue("group")
	brightness, err := strconv.Atoi(r.FormValue("brightness"))
	if err != nil {
		http.Error(w, "Invalid brightness value", http.StatusBadRequest)
		return
	}

	if brightness < 0 || brightness > 100 {
		http.Error(w, "Brightness must be 0-100", http.StatusBadRequest)
		return
	}

	// Find group and set all lights in it
	for _, group := range h.cfg.Groups {
		if group.Name == groupName {
			for _, ip := range group.IPs {
				if brightness == 0 {
					_ = h.wiz.SetLightState(ip, false)
				} else {
					_ = h.wiz.SetBrightness(ip, brightness)
				}

				h.mu.Lock()
				h.state[ip] = &services.LightState{IsOn: brightness > 0, Brightness: brightness}
				h.mu.Unlock()
			}
			break
		}
	}

	// Return group cards for HTMX swap
	w.Header().Set("Content-Type", "text/html")
	if err := groupLightCardsTemplate.Execute(w, map[string]interface{}{
		"Group":      groupName,
		"IPs":        getGroupIPs(h.cfg.Groups, groupName),
		"Brightness": brightness,
	}); err != nil {
		log.Printf("Template error: %v", err)
	}
}

func (h *Handlers) SyncScenes(w http.ResponseWriter, r *http.Request) {
	for _, light := range h.cfg.Lights {
		if light.PreferredScene > 0 {
			_ = h.wiz.SetScene(light.IP, light.PreferredScene)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"status": "ok", "message": "All lights synced to preferred scenes"}`)
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

	// Set all lights except those in groups
	for _, light := range h.cfg.Lights {
		if !isInGroup(light, h.cfg.Groups) {
			if brightness == 0 {
				_ = h.wiz.SetLightState(light.IP, false)
			} else {
				_ = h.wiz.SetBrightness(light.IP, brightness)
			}

			h.mu.Lock()
			h.state[light.IP] = &services.LightState{IsOn: brightness > 0, Brightness: brightness}
			h.mu.Unlock()
		}
	}

	// Return all light cards for HTMX swap
	w.Header().Set("Content-Type", "text/html")
	if err := allLightCardsTemplate.Execute(w, map[string]interface{}{
		"Lights":     h.cfg.Lights,
		"Groups":     h.cfg.Groups,
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

func getGroupIPs(groups []services.Group, groupName string) []string {
	for _, group := range groups {
		if group.Name == groupName {
			return group.IPs
		}
	}
	return []string{}
}
