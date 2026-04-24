package services

import (
	"encoding/json"
	"fmt"
	"os"
)

type Light struct {
	IP   string
	Name string
}

type Config struct {
	Lights []Light
}

func LoadConfig(filepath string) (*Config, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var rawConfig struct {
		WizLights struct {
			IPs   []string `json:"Ips"`
			Names []string `json:"Names"`
		} `json:"WizLights"`
	}

	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if len(rawConfig.WizLights.IPs) == 0 {
		return nil, fmt.Errorf("no light IPs configured")
	}

	ips := rawConfig.WizLights.IPs
	names := rawConfig.WizLights.Names

	// Generate default names if mismatch or missing
	if len(names) != len(ips) {
		names = make([]string, len(ips))
		for i := range names {
			names[i] = fmt.Sprintf("Light %d", i+1)
		}
	}

	lights := make([]Light, len(ips))
	for i := range ips {
		lights[i] = Light{IP: ips[i], Name: names[i]}
	}

	return &Config{Lights: lights}, nil
}
