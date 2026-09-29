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
