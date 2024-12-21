package configuration

import (
	"encoding/json"
	"fmt"
	"os"
)

type Service struct {
	Name string `json:"name"`
	StartScript []string `json:"start-script"`
}

type Configuration struct {
	OwnAgent string `json:"own-agent"`
	Agents []string `json:"agents"`
	Services []Service `json:"services"`
}

func ReadConfiguration() (Configuration, error) {
	data, err := os.ReadFile("../../configuration.json")
	if err != nil {
		return Configuration{}, fmt.Errorf("unable to read configuration file: %v", err)
	}

	var configuration Configuration
	err = json.Unmarshal(data, &configuration)
	if err != nil {
		return Configuration{}, fmt.Errorf("unable to parse configuration file: %v", err)
	}
	return configuration, nil
}