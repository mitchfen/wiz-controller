package services

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const WizPort = 38899
const sendRetries = 3
const retrySendDelay = 50 * time.Millisecond

type LightState struct {
	IsOn       bool `json:"state"`
	Brightness int  `json:"dimming"`
	SceneId    int  `json:"sceneId"`
}

type WizService struct {
	port int
}

func NewWizService() *WizService {
	return &WizService{port: WizPort}
}

// NewWizServiceWithPort creates a WizService with a custom port (for testing)
func NewWizServiceWithPort(port int) *WizService {
	return &WizService{port: port}
}

// sendCommandMultiple sends the same payload 3 times with small delays between sends
func (w *WizService) sendCommandMultiple(ip string, payload map[string]interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 0; attempt < sendRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retrySendDelay)
		}

		conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
			Port: w.port,
			IP:   net.ParseIP(ip),
		})
		if err != nil {
			lastErr = err
			continue
		}

		if err := conn.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
			conn.Close()
			lastErr = err
			continue
		}

		_, err = conn.Write(data)
		conn.Close()
		if err != nil {
			lastErr = err
			continue
		}

		// Successfully sent, don't retry
		return nil
	}

	// All retries failed
	if lastErr != nil {
		return fmt.Errorf("error sending to %s after %d attempts: %w", ip, sendRetries, lastErr)
	}
	return fmt.Errorf("error sending to %s after %d attempts", ip, sendRetries)
}

func (w *WizService) GetLightState(ip string) (*LightState, error) {
	payload := map[string]interface{}{
		"method": "getPilot",
		"params": map[string]interface{}{},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		Port: w.port,
		IP:   net.ParseIP(ip),
	})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return nil, err
	}

	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}

	var response struct {
		Result LightState `json:"result"`
	}
	if err := json.Unmarshal(buf[:n], &response); err != nil {
		return nil, err
	}

	if !response.Result.IsOn {
		response.Result.Brightness = 0
	}

	return &response.Result, nil
}

func (w *WizService) SetLightState(ip string, state bool) error {
	payload := map[string]interface{}{
		"method": "setState",
		"params": map[string]bool{"state": state},
	}
	return w.sendCommandMultiple(ip, payload)
}

func (w *WizService) SetScene(ip string, sceneId int) error {
	payload := map[string]interface{}{
		"method": "setPilot",
		"params": map[string]int{"sceneId": sceneId},
	}
	return w.sendCommandMultiple(ip, payload)
}

func (w *WizService) SetBrightness(ip string, brightness int) error {
	payload := map[string]interface{}{
		"method": "setPilot",
		"params": map[string]interface{}{
			"state":   true,
			"dimming": brightness,
		},
	}
	return w.sendCommandMultiple(ip, payload)
}
