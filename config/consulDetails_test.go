package config

import (
	"os"
	// "path/filepath" // Removed unused import
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to create a temporary YAML file for testing
func createTempYAMLFile(t *testing.T, content string) string {
	t.Helper()
	tempFile, err := os.CreateTemp(t.TempDir(), "testConfig-*.yml")
	assert.NoError(t, err)
	_, err = tempFile.WriteString(content)
	assert.NoError(t, err)
	assert.NoError(t, tempFile.Close())
	return tempFile.Name()
}

func TestParseConfigFile(t *testing.T) {
	t.Run("ValidConfigFile", func(t *testing.T) {
		validYAML := `
consul.details:
  - name: "consul-dev"
    dc: "dc1"
    url: "http://localhost:8500"
    token: "dev-token"
    base.path: "services/"
  - name: "consul-prod"
    dc: "dc2"
    url: "https://consul.prod:8500"
    token: "prod-token"
    base.path: "prod/services/"
`
		configFile := createTempYAMLFile(t, validYAML)
		defer os.Remove(configFile) // Clean up

		expectedConfig := &AllConsuls{ // Changed ConsulConfiguration to AllConsuls
			ConsulConfigs: []ConsulDetail{ // Changed ConsulFilters to ConsulConfigs
				{
					ConsulName: "consul-dev",
					DataCentre: "dc1",
					BaseURL:    "http://localhost:8500", // Combined Scheme and Address into BaseURL
					Token:      "dev-token",
					BasePath:   "services/",
				},
				{
					ConsulName: "consul-prod",
					DataCentre: "dc2",
					BaseURL:    "https://consul.prod:8500", // Combined Scheme and Address into BaseURL
					Token:      "prod-token",
					BasePath:   "prod/services/",
				},
			},
		}

		config, err := ParseConfigFile(configFile)
		assert.NoError(t, err)
		assert.Equal(t, expectedConfig, config)
	})

	t.Run("NonExistentConfigFile", func(t *testing.T) {
		config, err := ParseConfigFile("non_existent_file.yml")
		assert.Error(t, err)
		assert.Nil(t, config)
	})

	t.Run("InvalidConfigFile", func(t *testing.T) {
		// Updated to reflect actual field names and correct YAML structure
		invalidYAML := `
consul.details:
  - name: "consul-dev"
    dc: "dc1" 
    # url: "localhost:8500" # This line is intentionally mis-indented for failure
   url: "mis-indented-url" 
`
		configFile := createTempYAMLFile(t, invalidYAML)
		defer os.Remove(configFile) // Clean up

		config, err := ParseConfigFile(configFile)
		assert.Error(t, err)
		assert.Nil(t, config)
	})
}

func TestGetConsulConfigMap(t *testing.T) {
	sampleConfig := &AllConsuls{ // Changed ConsulConfiguration to AllConsuls
		ConsulConfigs: []ConsulDetail{ // Changed ConsulFilters to ConsulConfigs
			{
				ConsulName: "consul-dev",
				DataCentre: "dc1",
				BaseURL:    "http://localhost:8500", // Combined Scheme and Address
				Token:      "dev-token",
				BasePath:   "services/",
			},
			{
				ConsulName: "consul-prod",
				DataCentre: "dc2",
				BaseURL:    "https://consul.prod:8500", // Combined Scheme and Address
				Token:      "prod-token",
				BasePath:   "prod/services/",
			},
		},
	}

	configMap := GetConsulConfigMap(sampleConfig)

	// Assert that the map has the correct keys
	assert.Contains(t, configMap, "consul-dev")
	assert.Contains(t, configMap, "consul-prod")

	// Assert values for "consul-dev"
	devDetail, ok := configMap["consul-dev"]
	assert.True(t, ok)
	assert.Equal(t, "consul-dev", devDetail.ConsulName)
	// assert.Equal(t, "dev", devDetail.Env) // Env removed
	assert.Equal(t, "dc1", devDetail.DataCentre)
	// assert.Equal(t, "localhost:8500", devDetail.Address) // Address removed
	assert.Equal(t, "dev-token", devDetail.Token)
	// assert.Equal(t, "http", devDetail.Scheme) // Scheme removed
	assert.Equal(t, "services/", devDetail.BasePath)
	assert.Equal(t, "http://localhost:8500", devDetail.BaseURL) // BaseURL is now directly from config


	// Assert values for "consul-prod"
	prodDetail, ok := configMap["consul-prod"]
	assert.True(t, ok)
	assert.Equal(t, "consul-prod", prodDetail.ConsulName)
	// assert.Equal(t, "prod", prodDetail.Env) // Env removed
	assert.Equal(t, "dc2", prodDetail.DataCentre)
	// assert.Equal(t, "consul.prod:8500", prodDetail.Address) // Address removed
	assert.Equal(t, "prod-token", prodDetail.Token)
	// assert.Equal(t, "https", prodDetail.Scheme) // Scheme removed
	assert.Equal(t, "prod/services/", prodDetail.BasePath)
	assert.Equal(t, "https://consul.prod:8500", prodDetail.BaseURL) // BaseURL is now directly from config

	// Test with nil configuration
	nilMap := GetConsulConfigMap(nil)
	assert.Empty(t, nilMap, "Map should be empty for nil configuration")

	// Test with empty ConsulConfigs
	emptyFiltersConfig := &AllConsuls{ConsulConfigs: []ConsulDetail{}} // Changed to AllConsuls and ConsulConfigs
	emptyFiltersMap := GetConsulConfigMap(emptyFiltersConfig)
	assert.Empty(t, emptyFiltersMap, "Map should be empty for empty ConsulConfigs")
}
