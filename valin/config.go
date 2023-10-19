package main

import (
	"encoding/json"
	"os"
)

// ReadConfig reads configuration from a JSON file
func ReadConfig(filename string) (Configuration, error) {
	var conf Configuration
	configFile, err := os.Open(filename)
	if err != nil {
		return conf, err
	}
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&conf); err != nil {
		return conf, err
	}

	return conf, nil
}
