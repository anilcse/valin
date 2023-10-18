package main

import (
	"encoding/json"
	"os"
)

// ReadConfig reads configuration from a JSON file
func ReadConfig(filename string) ([]NetworkConfig, error) {
	configFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	var networkConfigs []NetworkConfig
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&networkConfigs); err != nil {
		return nil, err
	}

	return networkConfigs, nil
}
