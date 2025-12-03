package config

import "encoding/json"

type WailsProjectConfig struct {
	Info WailsProjectConfigInfo `json:"info"`
}

type WailsProjectConfigInfo struct {
	ProductVersion string `json:"productVersion"`
}

func ParseWailsProjectConfig(config []byte) (*WailsProjectConfig, error) {
	var wailsProjectConfig WailsProjectConfig
	err := json.Unmarshal(config, &wailsProjectConfig)
	if err != nil {
		return nil, err
	}
	return &wailsProjectConfig, nil
}
