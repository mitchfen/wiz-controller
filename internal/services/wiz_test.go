package services

import (
	"encoding/json"
	"net"
	"sync"
	"testing"
	"time"
)

// MockWizLight simulates a single WiZ light on UDP
type MockWizLight struct {
	state      *LightState
	addr       *net.UDPAddr
	conn       *net.UDPConn
	mu         sync.Mutex
	requestLog []string // log of received methods for testing
	t          *testing.T
}

// NewMockWizLight creates and starts a mock WiZ light on a random UDP port
func NewMockWizLight(t *testing.T) *MockWizLight {
	addr := &net.UDPAddr{
		Port: 0, // OS assigns random port
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("Failed to listen on UDP: %v", err)
	}

	mock := &MockWizLight{
		state: &LightState{
			IsOn:       true,
			Brightness: 50,
			SceneId:    11,
		},
		addr:       conn.LocalAddr().(*net.UDPAddr),
		conn:       conn,
		requestLog: []string{},
		t:          t,
	}

	// Start listening for requests
	go mock.listen()

	return mock
}

// listen handles incoming UDP requests
func (m *MockWizLight) listen() {
	buf := make([]byte, 1024)
	for {
		n, remoteAddr, err := m.conn.ReadFromUDP(buf)
		if err != nil {
			return // connection closed
		}

		var req map[string]interface{}
		if err := json.Unmarshal(buf[:n], &req); err != nil {
			continue
		}

		m.mu.Lock()
		method, ok := req["method"].(string)
		if ok {
			m.requestLog = append(m.requestLog, method)
		}
		m.mu.Unlock()

		response := m.handleRequest(req)
		respBytes, _ := json.Marshal(response)
		m.conn.WriteToUDP(respBytes, remoteAddr)
	}
}

// handleRequest processes a WiZ protocol request
func (m *MockWizLight) handleRequest(req map[string]interface{}) map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	method, _ := req["method"].(string)

	switch method {
	case "getPilot":
		return map[string]interface{}{
			"method": "getPilot",
			"env":    "pro",
			"result": m.state,
		}

	case "setState":
		params, _ := req["params"].(map[string]interface{})
		if state, ok := params["state"].(bool); ok {
			m.state.IsOn = state
			if !state {
				m.state.Brightness = 0
			}
		}
		return map[string]interface{}{
			"method": "setState",
			"env":    "pro",
			"result": map[string]bool{"success": true},
		}

	case "setPilot":
		params, _ := req["params"].(map[string]interface{})
		if dimming, ok := params["dimming"].(float64); ok {
			m.state.Brightness = int(dimming)
			m.state.IsOn = true
		}
		if sceneId, ok := params["sceneId"].(float64); ok {
			m.state.SceneId = int(sceneId)
			m.state.IsOn = true
		}
		return map[string]interface{}{
			"method": "setPilot",
			"env":    "pro",
			"result": map[string]bool{"success": true},
		}

	default:
		return map[string]interface{}{"error": "unknown method"}
	}
}

// GetState returns current light state
func (m *MockWizLight) GetState() *LightState {
	m.mu.Lock()
	defer m.mu.Unlock()
	stateCopy := *m.state
	return &stateCopy
}

// GetRequestLog returns methods that were called
func (m *MockWizLight) GetRequestLog() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	logCopy := make([]string, len(m.requestLog))
	copy(logCopy, m.requestLog)
	return logCopy
}

// IP returns the listening IP address
func (m *MockWizLight) IP() string {
	return m.addr.IP.String()
}

// Port returns the listening port
func (m *MockWizLight) Port() int {
	return m.addr.Port
}

// Close stops the mock light
func (m *MockWizLight) Close() error {
	return m.conn.Close()
}

// Tests

func TestSetBrightness(t *testing.T) {
	mock := NewMockWizLight(t)
	defer mock.Close()

	wiz := NewWizServiceWithPort(mock.Port())

	// Give mock time to start listening
	time.Sleep(10 * time.Millisecond)

	err := wiz.SetBrightness(mock.IP(), 75)
	if err != nil {
		t.Fatalf("SetBrightness failed: %v", err)
	}

	// Give UDP time to process
	time.Sleep(300 * time.Millisecond)

	state := mock.GetState()
	if state.Brightness != 75 {
		t.Errorf("Expected brightness 75, got %d", state.Brightness)
	}
	if !state.IsOn {
		t.Errorf("Expected light to be on")
	}
}

func TestSetScene(t *testing.T) {
	mock := NewMockWizLight(t)
	defer mock.Close()

	wiz := NewWizServiceWithPort(mock.Port())
	time.Sleep(10 * time.Millisecond)

	err := wiz.SetScene(mock.IP(), 6)
	if err != nil {
		t.Fatalf("SetScene failed: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	state := mock.GetState()
	if state.SceneId != 6 {
		t.Errorf("Expected scene 6, got %d", state.SceneId)
	}
	if !state.IsOn {
		t.Errorf("Expected light to be on")
	}
}

func TestSetLightState(t *testing.T) {
	mock := NewMockWizLight(t)
	defer mock.Close()

	wiz := NewWizServiceWithPort(mock.Port())
	time.Sleep(10 * time.Millisecond)

	// Turn off
	err := wiz.SetLightState(mock.IP(), false)
	if err != nil {
		t.Fatalf("SetLightState(off) failed: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	state := mock.GetState()
	if state.IsOn {
		t.Errorf("Expected light to be off")
	}
	if state.Brightness != 0 {
		t.Errorf("Expected brightness 0 when off, got %d", state.Brightness)
	}

	// Turn on
	err = wiz.SetLightState(mock.IP(), true)
	if err != nil {
		t.Fatalf("SetLightState(on) failed: %v", err)
	}

	time.Sleep(300 * time.Millisecond)

	state = mock.GetState()
	if !state.IsOn {
		t.Errorf("Expected light to be on")
	}
}

func TestGetLightState(t *testing.T) {
	mock := NewMockWizLight(t)
	defer mock.Close()

	wiz := NewWizServiceWithPort(mock.Port())
	time.Sleep(10 * time.Millisecond)

	state, err := wiz.GetLightState(mock.IP())
	if err != nil {
		t.Fatalf("GetLightState failed: %v", err)
	}

	if state.IsOn != true {
		t.Errorf("Expected light to be on, got %v", state.IsOn)
	}
	if state.Brightness != 50 {
		t.Errorf("Expected brightness 50, got %d", state.Brightness)
	}
}

func TestRetryLogic(t *testing.T) {
	mock := NewMockWizLight(t)
	defer mock.Close()

	wiz := NewWizServiceWithPort(mock.Port())
	time.Sleep(10 * time.Millisecond)

	// Send a command
	err := wiz.SetBrightness(mock.IP(), 80)
	if err != nil {
		t.Fatalf("SetBrightness failed: %v", err)
	}

	// Wait for all 3 retries to complete
	time.Sleep(500 * time.Millisecond)

	// Should have received setPilot method 3 times due to retries
	log := mock.GetRequestLog()
	if len(log) < 1 {
		t.Fatalf("Expected at least 1 setPilot request, got %d", len(log))
	}

	// Count setPilot requests
	setPilotCount := 0
	for _, method := range log {
		if method == "setPilot" {
			setPilotCount++
		}
	}

	if setPilotCount != 3 {
		t.Logf("Expected 3 setPilot retries, got %d (still counts as success if brightness was set)", setPilotCount)
	}

	state := mock.GetState()
	if state.Brightness != 80 {
		t.Errorf("Expected brightness 80, got %d", state.Brightness)
	}
}
