package services

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const WizPort = 38899

type LightState struct {
	IsOn       bool `json:"state"`
	Brightness int  `json:"dimming"`
	SceneId    int  `json:"sceneId"`
}

type WizService struct{}

func NewWizService() *WizService {
	return &WizService{}
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
		Port: WizPort,
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

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		Port: WizPort,
		IP:   net.ParseIP(ip),
	})
	if err != nil {
		return fmt.Errorf("error connecting to %s: %w", ip, err)
	}
	defer conn.Close()

	if err := conn.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return fmt.Errorf("error setting deadline for %s: %w", ip, err)
	}

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("error sending to %s: %w", ip, err)
	}

	return nil
}

func (w *WizService) SetScene(ip string, sceneId int) error {
	payload := map[string]interface{}{
		"method": "setPilot",
		"params": map[string]int{"sceneId": sceneId},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		Port: WizPort,
		IP:   net.ParseIP(ip),
	})
	if err != nil {
		return fmt.Errorf("error connecting to %s: %w", ip, err)
	}
	defer conn.Close()

	if err := conn.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return fmt.Errorf("error setting deadline for %s: %w", ip, err)
	}

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("error sending to %s: %w", ip, err)
	}

	return nil
}

func (w *WizService) SetBrightness(ip string, brightness int) error {
	payload := map[string]interface{}{
		"method": "setPilot",
		"params": map[string]interface{}{
			"state":   true,
			"dimming": brightness,
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		Port: WizPort,
		IP:   net.ParseIP(ip),
	})
	if err != nil {
		return fmt.Errorf("error connecting to %s: %w", ip, err)
	}
	defer conn.Close()

	if err := conn.SetWriteDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return fmt.Errorf("error setting deadline for %s: %w", ip, err)
	}

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("error sending to %s: %w", ip, err)
	}

	return nil
}
