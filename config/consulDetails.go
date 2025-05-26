package config

import (
	"fmt" // Added fmt import
	"os"

	"gopkg.in/yaml.v2"
)

/*
	@author: Aviral Nigam
	@github: https://github.com/iAviPro
	@date: 15 Jun, 2020
*/

// ConsulDetail : Struct for Environment Details
type ConsulDetail struct {
	ConsulName string `yaml:"name"`
	DataCentre string `yaml:"dc"`
	BaseURL    string `yaml:"url"`
	BasePath   string `yaml:"base.path"`
	Token      string `yaml:"token"`
}

// AllConsuls : Single struct for all the environments
type AllConsuls struct {
	ConsulConfigs []ConsulDetail `yaml:"consul.details"`
}

// DefaultPathToEnvConfigFile : If path is not provided by the user then default file is used.
const DefaultPathToEnvConfigFile string = "./config/consulConfig.yml"

// ParseConfigFile : based on the path provided
func ParseConfigFile(configFilePath string) (*AllConsuls, error) {
	env := &AllConsuls{}
	var filePath string
	if configFilePath == "" {
		filePath = DefaultPathToEnvConfigFile
	} else {
		filePath = configFilePath
	}
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Init new YAML decode
	yd := yaml.NewDecoder(file)
	// Start YAML decoding from file
	if err := yd.Decode(&env); err != nil {
		return nil, err
	}
	if env == nil || len(env.ConsulConfigs) == 0 {
		return nil, fmt.Errorf("no consul configurations found or failed to parse from file: %s", filePath)
	}

	return env, nil
}

// GetConsulConfigMap : Get Map of AllConsuls
func GetConsulConfigMap(consuls *AllConsuls) map[string]ConsulDetail {
	consulsMap := make(map[string]ConsulDetail)
	if consuls == nil || consuls.ConsulConfigs == nil {
		return consulsMap // Return empty map if input is nil or contains no configs
	}
	for _, c := range consuls.ConsulConfigs {
		consulsMap[c.ConsulName] = c
	}
	return consulsMap
}
