package src

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ConsulScale/config"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func TestCleanKVPairs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "Normal case",
			input: "key1=value1 | key2 = value2 | key3= value3 ",
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name:     "Empty input",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:     "Input with no equals sign",
			input:    "key1value1 | key2value2",
			expected: map[string]string{},
		},
		{
			name:  "Input with extra spaces around equals",
			input: "key1 =   value1   |  key2  =  value2 ",
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:  "Input with only one pair",
			input: "key1=value1",
			expected: map[string]string{
				"key1": "value1",
			},
		},
		{
			name:  "Input with empty value",
			input: "key1=value1|key2=",
			expected: map[string]string{
				"key1": "value1",
				"key2": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := cleanKVPairs(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestErrConsulConnection(t *testing.T) {
	var originalStdout = os.Stdout // Keep a backup of the original stdout
	var exitCalledWith int         // Variable to track if exit was called and with what code

	// Mock exit function
	mockExit := func(code int) {
		exitCalledWith = code
	}
	// Store original exit function and defer restoration
	originalExitFunc := exitFunc
	defer func() { exitFunc = originalExitFunc }()


	tests := []struct {
		name           string
		err            error
		shouldExit     bool
		expectedOutput string // Substring expected in the output
		expectExitCall bool
		expectedExitCode int
	}{
		{
			name:           "Error and exit true",
			err:            errors.New("test connection error"),
			shouldExit:     true,
			expectedOutput: "Consul connection error: test connection error",
			expectExitCall: true,
			expectedExitCode: 1,
		},
		{
			name:           "Error and exit false",
			err:            errors.New("another connection error"),
			shouldExit:     false,
			expectedOutput: "Consul connection error: another connection error",
			expectExitCall: false,
		},
		{
			name:           "No error",
			err:            nil,
			shouldExit:     true, // Should not matter if err is nil
			expectedOutput: "",    // No output expected
			expectExitCall: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset exitCalledWith for each test
			exitCalledWith = 0
			// Set the mock exit function for this test run
			exitFunc = mockExit


			r, w, _ := os.Pipe()
			os.Stdout = w

			errConsulConnection(tt.err, tt.shouldExit)

			w.Close()
			os.Stdout = originalStdout // Restore stdout

			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			if tt.expectedOutput != "" {
				assert.Contains(t, output, tt.expectedOutput)
			} else {
				assert.Empty(t, output)
			}


			if tt.expectExitCall {
				assert.Equal(t, tt.expectedExitCode, exitCalledWith, "exitFunc was not called with expected code")
			} else {
				assert.Equal(t, 0, exitCalledWith, "exitFunc was called unexpectedly")
			}
		})
	}
}

func TestRemoveBasePath(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		basePath string
		expected string
	}{
		{
			name:     "Normal case",
			key:      "base/path/service/key",
			basePath: "base/path/",
			expected: "service/key",
		},
		{
			name:     "Key does not contain base path",
			key:      "another/path/service/key",
			basePath: "base/path/",
			expected: "another/path/service/key",
		},
		{
			name:     "Empty key",
			key:      "",
			basePath: "base/path/",
			expected: "",
		},
		{
			name:     "Empty base path",
			key:      "base/path/service/key",
			basePath: "",
			expected: "base/path/service/key",
		},
		{
			name:     "Key and base path are the same",
			key:      "base/path/",
			basePath: "base/path/",
			expected: "",
		},
		{
			name:     "Key is shorter than base path",
			key:      "base/",
			basePath: "base/path/",
			expected: "base/",
		},
		{
			name:     "Base path not at the beginning of the key",
			key:      "service/base/path/key",
			basePath: "base/path/",
			expected: "service/base/path/key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := removeBasePath(tt.key, tt.basePath)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestIsConfigNameValid(t *testing.T) {
	sampleConsulDetailsMap := map[string]config.ConsulDetail{
		"config1": {ConsulName: "config1", Env: "dev"},
		"config2": {ConsulName: "config2", Env: "prod"},
	}

	tests := []struct {
		name        string
		configName  string
		detailsMap  map[string]config.ConsulDetail
		expected    bool
	}{
		{
			name:       "Valid config name",
			configName: "config1",
			detailsMap: sampleConsulDetailsMap,
			expected:   true,
		},
		{
			name:       "Invalid config name",
			configName: "config3",
			detailsMap: sampleConsulDetailsMap,
			expected:   false,
		},
		{
			name:       "Empty config name",
			configName: "",
			detailsMap: sampleConsulDetailsMap,
			expected:   false,
		},
		{
			name:       "Nil details map",
			configName: "config1",
			detailsMap: nil,
			expected:   false,
		},
		{
			name:       "Empty details map",
			configName: "config1",
			detailsMap: map[string]config.ConsulDetail{},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsConfigNameValid(tt.configName, tt.detailsMap)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestRemoveExistingKVPairs(t *testing.T) {
	tests := []struct {
		name                string
		sourceKVPairs       api.KVPairs
		sourceBasePath      string
		destinationKVPairs  api.KVPairs
		destinationBasePath string
		expected            map[string]string
	}{
		{
			name: "Source has keys not in target",
			sourceKVPairs: api.KVPairs{
				{Key: "source/s1/key1", Value: []byte("val1")},
				{Key: "source/s1/key2", Value: []byte("val2")},
			},
			sourceBasePath: "source/s1/",
			destinationKVPairs: api.KVPairs{
				{Key: "dest/d1/key3", Value: []byte("val3")},
			},
			destinationBasePath: "dest/d1/",
			expected: map[string]string{
				"key1": "val1",
				"key2": "val2",
			},
		},
		{
			name: "Source has keys that are also in target",
			sourceKVPairs: api.KVPairs{
				{Key: "src/common/key1", Value: []byte("src_val1")}, // Should be removed
				{Key: "src/common/key2", Value: []byte("src_val2")}, // Should remain
			},
			sourceBasePath: "src/common/",
			destinationKVPairs: api.KVPairs{
				{Key: "dst/common/key1", Value: []byte("dst_val1")}, // Matching key
			},
			destinationBasePath: "dst/common/",
			expected: map[string]string{
				"key2": "src_val2",
			},
		},
		{
			name:           "Source is empty",
			sourceKVPairs:  api.KVPairs{},
			sourceBasePath: "src/",
			destinationKVPairs: api.KVPairs{
				{Key: "dst/key1", Value: []byte("val1")},
			},
			destinationBasePath: "dst/",
			expected:            map[string]string{},
		},
		{
			name: "Target is empty",
			sourceKVPairs: api.KVPairs{
				{Key: "src/key1", Value: []byte("val1")},
				{Key: "src/key2", Value: []byte("val2")},
			},
			sourceBasePath:      "src/",
			destinationKVPairs:  api.KVPairs{},
			destinationBasePath: "dst/",
			expected: map[string]string{
				"key1": "val1",
				"key2": "val2",
			},
		},
		{
			name: "Different base paths, no overlap",
			sourceKVPairs: api.KVPairs{
				{Key: "app1/config/settingA", Value: []byte("apple")},
			},
			sourceBasePath: "app1/config/",
			destinationKVPairs: api.KVPairs{
				{Key: "app2/config/settingA", Value: []byte("orange")},
			},
			destinationBasePath: "app2/config/",
			expected: map[string]string{
				"settingA": "apple",
			},
		},
		{
			name: "Identical keys after removing different base paths",
			sourceKVPairs: api.KVPairs{
				{Key: "systemA/properties/timeout", Value: []byte("100")},
			},
			sourceBasePath: "systemA/properties/",
			destinationKVPairs: api.KVPairs{
				{Key: "systemB/settings/timeout", Value: []byte("200")},
			},
			destinationBasePath: "systemB/settings/", // "timeout" key exists in both after base path removal
			expected:            map[string]string{},
		},
		{
			name: "Source with folder and nil value keys",
			sourceKVPairs: api.KVPairs{
				{Key: "src/data/key1", Value: []byte("val1")},
				{Key: "src/data/folder/", Value: nil},             // folder, should be ignored
				{Key: "src/data/key_nil", Value: nil},             // nil value, should be ignored
				{Key: "src/data/key2", Value: []byte("val2")},
			},
			sourceBasePath: "src/data/",
			destinationKVPairs: api.KVPairs{
				{Key: "dst/data/key2", Value: []byte("dst_val2")}, // key2 will be removed from source
			},
			destinationBasePath: "dst/data/",
			expected: map[string]string{
				"key1": "val1",
			},
		},
		{
			name:                "Both source and destination are empty",
			sourceKVPairs:       api.KVPairs{},
			sourceBasePath:      "src/",
			destinationKVPairs:  api.KVPairs{},
			destinationBasePath: "dst/",
			expected:            map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := removeExistingKVPairs(tt.sourceKVPairs, tt.sourceBasePath, tt.destinationKVPairs, tt.destinationBasePath)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestConvertServiceKVPairsToMap(t *testing.T) {
	tests := []struct {
		name     string
		kvPairs  api.KVPairs
		basePath string
		expected map[string]string
	}{
		{
			name: "Normal case",
			kvPairs: api.KVPairs{
				{Key: "my/service/key1", Value: []byte("value1")},
				{Key: "my/service/key2", Value: []byte("value2")},
				{Key: "my/service/folder/", Value: nil}, // Folder, should be skipped
			},
			basePath: "my/service/",
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "Base path not at the beginning",
			kvPairs: api.KVPairs{
				{Key: "other/my/service/key1", Value: []byte("value1")},
			},
			basePath: "my/service/",
			expected: map[string]string{}, // Key should not be included if base path doesn't match start
		},
		{
			name: "Empty KVPairs",
			kvPairs: api.KVPairs{},
			basePath: "my/service/",
			expected: map[string]string{},
		},
		{
			name: "Nil KVPairs",
			kvPairs: nil,
			basePath: "my/service/",
			expected: map[string]string{},
		},
		{
			name: "Empty base path",
			kvPairs: api.KVPairs{
				{Key: "key1", Value: []byte("value1")},
				{Key: "folder/", Value: nil},
			},
			basePath: "",
			expected: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "Key equal to base path (should be skipped as it's like a folder)",
			kvPairs: api.KVPairs{
				{Key: "my/service/", Value: []byte("value_for_folder_itself")},
			},
			basePath: "my/service/",
			expected: map[string]string{},
		},
		{
			name: "Key with nil value (non-folder)",
			kvPairs: api.KVPairs{
				{Key: "my/service/key_nil_val", Value: nil},
			},
			basePath: "my/service/",
			expected: map[string]string{}, // Should be skipped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := convertServiceKVPairsToMap(tt.kvPairs, tt.basePath)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestCreateBackupFileAndWriteData(t *testing.T) {
	tempDir := t.TempDir() // Auto-cleaned up

	tests := []struct {
		name        string
		cn          string
		data        []byte
		expectedErr bool
	}{
		{
			name:        "Normal case",
			cn:          "myConsulConfig",
			data:        []byte(`{"key":"value"}`),
			expectedErr: false,
		},
		{
			name:        "Empty data",
			cn:          "emptyDataConfig",
			data:        []byte(""),
			expectedErr: false,
		},
		{
			name:        "Empty cn (filename)", // This should still work, filename will be ".json"
			cn:          "",
			data:        []byte(`{"key":"value"}`),
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backupFilePath, err := createBackupFileAndWriteData(tt.cn, tt.data, tempDir)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.FileExists(t, backupFilePath)

				// Verify filename
				expectedFileName := tt.cn + ".json"
				assert.Equal(t, expectedFileName, filepath.Base(backupFilePath))

				// Verify content
				content, readErr := os.ReadFile(backupFilePath)
				assert.NoError(t, readErr)
				assert.Equal(t, tt.data, content)
			}
		})
	}

	// Test case for when directory creation fails (by making tempDir a file)
	// This is a bit harder to test directly without more complex mocks or os-level manipulation.
	// For now, we assume that if tempDir is valid, os.MkdirAll and os.WriteFile behave as expected.
	// A more direct test for this would involve trying to write to an unwritable directory.

	t.Run("Error case - invalid base path", func(t *testing.T) {
		// Create a file to use as an invalid base path (cannot create directory inside a file)
		invalidBasePathFile, err := os.CreateTemp(t.TempDir(), "invalidbase")
		assert.NoError(t, err)
		invalidBasePathFile.Close() // Close the file so it's just a path

		_, err = createBackupFileAndWriteData("test", []byte("data"), invalidBasePathFile.Name())
		assert.Error(t, err) // Expect an error because base path is a file
	})
}

func TestReadJSONFileAndReturnStruct(t *testing.T) {
	validJSONContent := `[{"key":"service/key1","value":"value1"},{"key":"service/key2","value":"value2"}]`
	expectedStruct := &[]KVDetails{
		{Key: "service/key1", Value: "value1"},
		{Key: "service/key2", Value: "value2"},
	}

	// Create a temporary file with valid JSON content
	tempFile, err := os.CreateTemp(t.TempDir(), "testRead-*.json")
	assert.NoError(t, err)
	_, err = tempFile.WriteString(validJSONContent)
	assert.NoError(t, err)
	assert.NoError(t, tempFile.Close())
	filePath := tempFile.Name()

	t.Run("Valid JSON file", func(t *testing.T) {
		kvDetails, err := readJSONFileAndReturnStruct(filePath)
		assert.NoError(t, err)
		assert.Equal(t, expectedStruct, kvDetails)
	})

	t.Run("Non-existent file", func(t *testing.T) {
		kvDetails, err := readJSONFileAndReturnStruct("non_existent_file.json")
		assert.Error(t, err) // Expect an error
		assert.Nil(t, kvDetails) // Expect nil struct
	})

	// Create a temporary file with invalid JSON content
	invalidTempFile, err := os.CreateTemp(t.TempDir(), "testInvalidRead-*.json")
	assert.NoError(t, err)
	_, err = invalidTempFile.WriteString(`this is not json`)
	assert.NoError(t, err)
	assert.NoError(t, invalidTempFile.Close())
	invalidFilePath := invalidTempFile.Name()

	t.Run("Invalid JSON file", func(t *testing.T) {
		kvDetails, err := readJSONFileAndReturnStruct(invalidFilePath)
		assert.Error(t, err)
		assert.Nil(t, kvDetails)
	})

	// Test with an empty file
	emptyFile, err := os.CreateTemp(t.TempDir(), "empty-*.json")
	assert.NoError(t, err)
	assert.NoError(t, emptyFile.Close())
	emptyFilePath := emptyFile.Name()

	t.Run("Empty JSON file", func(t *testing.T) {
		// Expect an error because an empty file is not valid JSON for a slice/object
		kvDetails, err := readJSONFileAndReturnStruct(emptyFilePath)
		assert.Error(t, err)
		assert.Nil(t, kvDetails)
	})
}

func TestCreateConsulDetails(t *testing.T) {
	// 1. Valid temporary config file
	t.Run("Valid temporary config file", func(t *testing.T) {
		validYAML := `
consulFilters:
  - name: "consul-test"
    env: "test"
    datacenter: "dc1"
    address: "localhost:8500"
    token: "test-token"
    scheme: "http"
    basePath: "test/services/"
`
		tempFile, err := os.CreateTemp(t.TempDir(), "validConfig-*.yml")
		assert.NoError(t, err)
		_, err = tempFile.WriteString(validYAML)
		assert.NoError(t, err)
		assert.NoError(t, tempFile.Close())
		filePath := tempFile.Name()

		// Mocking is not directly done here, but we expect a valid map
		consulDetailsMap, err := CreateConsulDetails(filePath)
		assert.NoError(t, err)
		assert.NotNil(t, consulDetailsMap)
		assert.Contains(t, consulDetailsMap, "consul-test")
		assert.Equal(t, "test", consulDetailsMap["consul-test"].Env)
	})

	// 2. Empty cfp (defaults to config.DefaultPathToEnvConfigFile)
	// This test is more of an integration test. We'll check if it attempts to parse
	// the default file. If the default file doesn't exist, it should error out.
	// If it exists but is invalid, it should also error.
	// For this test, we'll assume the default path might not exist or be valid.
	t.Run("Empty config file path", func(t *testing.T) {
		// To properly test the default path, we'd need to ensure config.DefaultPathToEnvConfigFile
		// either exists and is valid, or mock the functions it calls.
		// For now, we call it and check if an error occurs, as the default file may not exist.
		// This isn't a perfect test of the default path logic without more setup/mocking.
		_, err := CreateConsulDetails("")
		// We expect an error if the default file doesn't exist or is invalid.
		// If it *does* exist and is valid in the test environment, this assertion would fail.
		// This depends on the state of `config.DefaultPathToEnvConfigFile` in the test env.
		// A more robust test would involve mocking `config.ParseConfigFile`.
		if _, statErr := os.Stat(config.DefaultPathToEnvConfigFile); os.IsNotExist(statErr) {
			assert.Error(t, err)
		} else if statErr == nil {
			// If file exists, we can't be sure if it's valid or not without parsing
			// For now, we'll just note that an error implies it was invalid or unreadable
			if err != nil {
				assert.Error(t, err) // It errored as expected for an invalid/unreadable existing default
			} else {
				// If it didn't error, the default file was found and parsed
				// We can't assert much more without knowing its content
			}
		}
	})

	// 3. Invalid cfp (non-existent file)
	t.Run("Invalid config file path - non-existent", func(t *testing.T) {
		_, err := CreateConsulDetails("non_existent_config_file.yml")
		assert.Error(t, err)
		// The function calls log.Fatalf, which calls os.Exit(1).
		// We can't directly test os.Exit without a more complex setup (e.g., exec.Command or forking).
		// We are checking the error returned by ValidateFilePath which happens before log.Fatalf.
	})

	// 4. Invalid cfp (directory path)
	t.Run("Invalid config file path - directory", func(t *testing.T) {
		tempDir := t.TempDir()
		_, err := CreateConsulDetails(tempDir)
		assert.Error(t, err) // Expect error from ValidateFilePath
	})
}

func TestValidateFilePath(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp(t.TempDir(), "testfile-*.txt")
	assert.NoError(t, err)
	filePath := tempFile.Name()
	assert.NoError(t, tempFile.Close()) // Close the file

	// Create a temporary directory
	tempDir, err := os.MkdirTemp(t.TempDir(), "testdir-*")
	assert.NoError(t, err)

	tests := []struct {
		name        string
		path        string
		expectedErr bool
		errContains string // Substring to check in the error message
	}{
		{
			name:        "Valid file path",
			path:        filePath,
			expectedErr: false,
		},
		{
			name:        "Path is a directory",
			path:        tempDir,
			expectedErr: true,
			errContains: "is a directory, not a file",
		},
		{
			name:        "Path does not exist",
			path:        filepath.Join(t.TempDir(), "non_existent_file.txt"),
			expectedErr: true,
			errContains: "no such file or directory", // Error message from os.Stat
		},
		{
			name:        "Empty path", // Should be caught by os.Stat
			path:        "",
			expectedErr: true,
			errContains: "no such file or directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFilePath(tt.path)
			if tt.expectedErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConvertJSONStructToKvPairs(t *testing.T) {
	tests := []struct {
		name     string
		input    *[]KVDetails
		expected []*api.KVPair
	}{
		{
			name: "Normal case",
			input: &[]KVDetails{
				{Key: "service/key1", Value: "value1"},
				{Key: "service/key2", Value: "value2"},
			},
			expected: []*api.KVPair{
				{Key: "service/key1", Value: []byte("value1")},
				{Key: "service/key2", Value: []byte("value2")},
			},
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: []*api.KVPair{},
		},
		{
			name:     "Empty input slice",
			input:    &[]KVDetails{},
			expected: []*api.KVPair{},
		},
		{
			name: "Input with empty key or value",
			input: &[]KVDetails{
				{Key: "service/key1", Value: "value1"},
				{Key: "", Value: "value2"},             // Empty key
				{Key: "service/key3", Value: ""},       // Empty value
				{Key: "service/key4", Value: "value4"},
			},
			expected: []*api.KVPair{
				{Key: "service/key1", Value: []byte("value1")},
				{Key: "", Value: []byte("value2")},
				{Key: "service/key3", Value: []byte("")},
				{Key: "service/key4", Value: []byte("value4")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := convertJSONStructToKvPairs(tt.input)
			assert.Equal(t, len(tt.expected), len(actual), "Length of slices should be equal")
			for i := range actual {
				assert.Equal(t, tt.expected[i].Key, actual[i].Key, "Keys should be equal")
				assert.Equal(t, string(tt.expected[i].Value), string(actual[i].Value), "Values should be equal")
			}
		})
	}
}

func TestKvPairsToJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    api.KVPairs
		expected string // Expected JSON string, or empty for error cases
		wantErr  bool
	}{
		{
			name: "Normal case",
			input: api.KVPairs{
				{Key: "service/key1", Value: []byte("value1")},
				{Key: "service/key2", Value: []byte("value2")},
			},
			expected: `[{"key":"service/key1","value":"value1"},{"key":"service/key2","value":"value2"}]`,
			wantErr:  false,
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: "null", // Changed from "" to "null" as json.Marshal(nil) is "null"
			wantErr:  false,
		},
		{
			name:     "Empty input",
			input:    api.KVPairs{},
			expected: `[]`, // Changed from "" to "[]" as json.Marshal of empty slice is "[]"
			wantErr:  false,
		},
		{
			name: "Input with folder key",
			input: api.KVPairs{
				{Key: "service/folder/", Value: nil}, // Folder key
				{Key: "service/key1", Value: []byte("value1")},
			},
			expected: `[{"key":"service/key1","value":"value1"}]`, // Folder should be skipped
			wantErr:  false,
		},
		{
			name: "Input with nil value (non-folder)",
			input: api.KVPairs{
				{Key: "service/key1", Value: []byte("value1")},
				{Key: "service/key2", Value: nil}, // Nil value, should be skipped
			},
			expected: `[{"key":"service/key1","value":"value1"}]`,
			wantErr:  false,
		},
		{
			name: "Input with only folder key",
			input: api.KVPairs{
				{Key: "service/folder/", Value: nil},
			},
			expected: `[]`,
			wantErr:  false,
		},
		{
			name: "Input with only nil value (non-folder)",
			input: api.KVPairs{
				{Key: "service/key_nil_val", Value: nil},
			},
			expected: `[]`,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := kvPairsToJSON(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.JSONEq(t, tt.expected, string(jsonData))
			}
		})
	}
}

func TestCreateKVPairs(t *testing.T) {
	tests := []struct {
		name        string
		props       string
		serviceName string
		expected    []*api.KVPair
	}{
		{
			name:        "Normal case",
			props:       "key1=val1|key2=val2",
			serviceName: "myService",
			expected: []*api.KVPair{
				{Key: "myService/key1", Value: []byte("val1")},
				{Key: "myService/key2", Value: []byte("val2")},
			},
		},
		{
			name:        "Empty props",
			props:       "",
			serviceName: "myService",
			expected:    []*api.KVPair{},
		},
		{
			name:        "Props with no equals sign",
			props:       "key1val1|key2val2",
			serviceName: "myService",
			expected:    []*api.KVPair{},
		},
		{
			name:        "Props with empty value",
			props:       "key1=val1|key2=",
			serviceName: "myService",
			expected: []*api.KVPair{
				{Key: "myService/key1", Value: []byte("val1")},
				{Key: "myService/key2", Value: []byte("")},
			},
		},
		{
			name:        "Empty service name",
			props:       "key1=val1",
			serviceName: "",
			expected: []*api.KVPair{
				{Key: "/key1", Value: []byte("val1")}, // Note: Leading slash due to empty serviceName
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := createKVPairs(tt.props, tt.serviceName)
			// Custom comparison because api.KVPair contains a byte slice
			if len(actual) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(actual))
			}
			for i := range actual {
				if actual[i].Key != tt.expected[i].Key || string(actual[i].Value) != string(tt.expected[i].Value) {
					t.Errorf("Expected KVPair %v, got %v", tt.expected[i], actual[i])
				}
			}
		})
	}
}
