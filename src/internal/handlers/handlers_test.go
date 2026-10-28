package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mitchfen/wiz-controller/internal/services"
)

func TestHomeReturnsImmediatelyWhenLightIsUnavailable(t *testing.T) {
	cfg := &services.Config{
		Lights: []services.Light{
			{IP: "127.0.0.1", Name: "Unavailable", PreferredScene: 11},
		},
		Groups: []services.Group{
			{Name: "Living Room", IPs: []string{"127.0.0.1"}},
			{Name: "Bedroom", IPs: []string{"127.0.0.1"}},
		},
	}
	h := NewHandlers(cfg, services.NewWizServiceWithPort(1))
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	start := time.Now()
	h.Home(response, request)
	elapsed := time.Since(start)

	if response.Code != http.StatusOK {
		t.Fatalf("Home status = %d, want %d", response.Code, http.StatusOK)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("Home took %v with an unavailable light, want less than 500ms", elapsed)
	}
	if !strings.Contains(response.Body.String(), `value="0"`) {
		t.Error("Home did not render the offline light at zero brightness")
	}
}

func TestHealth(t *testing.T) {
	h := NewHandlers(&services.Config{}, services.NewWizService())
	response := httptest.NewRecorder()

	h.Health(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("Health status = %d, want %d", response.Code, http.StatusOK)
	}
}

func TestSetBrightnessUpdatesCacheForHome(t *testing.T) {
	cfg := &services.Config{
		Lights: []services.Light{
			{IP: "127.0.0.1", Name: "Desk Light", PreferredScene: 11},
			{IP: "127.0.0.2", Name: "Bed Light", PreferredScene: 11},
		},
		Groups: []services.Group{
			{Name: "Living Room", IPs: []string{"127.0.0.1"}},
			{Name: "Bedroom", IPs: []string{"127.0.0.2"}},
		},
	}
	h := NewHandlers(cfg, services.NewWizServiceWithPort(1))

	// Initial Home render should have 0
	initResp := httptest.NewRecorder()
	h.Home(initResp, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(initResp.Body.String(), `value="0"`) {
		t.Errorf("Expected initial render to show 0, got %s", initResp.Body.String())
	}

	// Change brightness to 75
	postReq := httptest.NewRequest(http.MethodPost, "/api/lights/127.0.0.1/brightness", strings.NewReader("brightness=75"))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.SetPathValue("ip", "127.0.0.1")
	postResp := httptest.NewRecorder()
	h.SetBrightness(postResp, postReq)

	if postResp.Code != http.StatusOK {
		t.Fatalf("SetBrightness returned status %d", postResp.Code)
	}

	// Next Home render (e.g. from new tab or phone) must immediately show 75 without needing refresh
	homeResp := httptest.NewRecorder()
	h.Home(homeResp, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(homeResp.Body.String(), `value="75"`) {
		t.Errorf("Home did not reflect updated brightness 75 in cache")
	}
}

func TestSetGroupBrightnessUpdatesCacheForHome(t *testing.T) {
	cfg := &services.Config{
		Lights: []services.Light{
			{IP: "127.0.0.1", Name: "Living Light", PreferredScene: 11},
			{IP: "127.0.0.2", Name: "Bed1", PreferredScene: 11},
			{IP: "127.0.0.3", Name: "Bed2", PreferredScene: 11},
		},
		Groups: []services.Group{
			{Name: "Living Room", IPs: []string{"127.0.0.1"}},
			{Name: "Bedroom", IPs: []string{"127.0.0.2", "127.0.0.3"}},
		},
	}
	h := NewHandlers(cfg, services.NewWizServiceWithPort(1))

	// Change bedroom group brightness to 60
	postReq := httptest.NewRequest(http.MethodPost, "/api/groups/Bedroom/brightness", strings.NewReader("brightness=60"))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.SetPathValue("group", "Bedroom")
	postResp := httptest.NewRecorder()
	h.SetGroupBrightness(postResp, postReq)

	if postResp.Code != http.StatusOK {
		t.Fatalf("SetGroupBrightness returned status %d", postResp.Code)
	}

	// Home render should show 60 for bedroom group
	homeResp := httptest.NewRecorder()
	h.Home(homeResp, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(homeResp.Body.String(), `value="60"`) {
		t.Errorf("Home did not reflect updated group brightness 60 in cache")
	}
}

func TestStartPolling(t *testing.T) {
	cfg := &services.Config{
		Lights: []services.Light{
			{IP: "127.0.0.1", Name: "PolledLight", PreferredScene: 11},
		},
		Groups: []services.Group{},
	}
	h := NewHandlers(cfg, services.NewWizServiceWithPort(1))

	stop := h.StartPolling(50 * time.Millisecond)
	time.Sleep(120 * time.Millisecond)
	stop()
}
