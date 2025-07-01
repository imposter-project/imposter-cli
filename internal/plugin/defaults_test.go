package plugin

import (
	"gatehill.io/imposter/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func TestListDefaultPlugins(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugin-defaults-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	tests := []struct {
		name            string
		configContent   string
		expectedPlugins []string
		expectError     bool
	}{
		{
			name:            "empty config",
			configContent:   "",
			expectedPlugins: []string{},
			expectError:     false,
		},
		{
			name: "config with plugins",
			configContent: `default:
  plugins:
    - store-redis
    - js-graal
`,
			expectedPlugins: []string{"store-redis", "js-graal"},
			expectError:     false,
		},
		{
			name: "config with empty plugins list",
			configContent: `default:
  plugins: []
`,
			expectedPlugins: []string{},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config file
			configFilePath := filepath.Join(configDir, "config.yaml")
			if tt.configContent != "" {
				err := os.WriteFile(configFilePath, []byte(tt.configContent), 0644)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				// Remove config file if it exists
				os.Remove(configFilePath)
			}

			plugins, err := ListDefaultPlugins()
			if (err != nil) != tt.expectError {
				t.Errorf("ListDefaultPlugins() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if len(plugins) != len(tt.expectedPlugins) {
				t.Errorf("ListDefaultPlugins() returned %d plugins, expected %d", len(plugins), len(tt.expectedPlugins))
				return
			}

			for i, plugin := range plugins {
				if plugin != tt.expectedPlugins[i] {
					t.Errorf("ListDefaultPlugins() plugin[%d] = %v, expected %v", i, plugin, tt.expectedPlugins[i])
				}
			}
		})
	}
}

func TestAddDefaultPlugins(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugin-defaults-add-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	tests := []struct {
		name            string
		existingPlugins []string
		pluginsToAdd    []string
		expectedPlugins []string
	}{
		{
			name:            "add to empty list",
			existingPlugins: []string{},
			pluginsToAdd:    []string{"store-redis"},
			expectedPlugins: []string{"store-redis"},
		},
		{
			name:            "add to existing list",
			existingPlugins: []string{"js-graal"},
			pluginsToAdd:    []string{"store-redis"},
			expectedPlugins: []string{"js-graal", "store-redis"},
		},
		{
			name:            "add duplicate plugin",
			existingPlugins: []string{"store-redis"},
			pluginsToAdd:    []string{"store-redis"},
			expectedPlugins: []string{"store-redis"},
		},
		{
			name:            "add multiple plugins",
			existingPlugins: []string{"js-graal"},
			pluginsToAdd:    []string{"store-redis", "swaggerui"},
			expectedPlugins: []string{"js-graal", "store-redis", "swaggerui"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup existing plugins
			if len(tt.existingPlugins) > 0 {
				err := writeDefaultPlugins(tt.existingPlugins)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Add new plugins
			err := addDefaultPlugins(tt.pluginsToAdd)
			if err != nil {
				t.Errorf("addDefaultPlugins() error = %v", err)
				return
			}

			// Verify result
			plugins, err := ListDefaultPlugins()
			if err != nil {
				t.Fatal(err)
			}

			if len(plugins) != len(tt.expectedPlugins) {
				t.Errorf("After addDefaultPlugins() got %d plugins, expected %d", len(plugins), len(tt.expectedPlugins))
				return
			}

			// Check that all expected plugins are present (order might vary due to deduplication)
			pluginMap := make(map[string]bool)
			for _, plugin := range plugins {
				pluginMap[plugin] = true
			}

			for _, expected := range tt.expectedPlugins {
				if !pluginMap[expected] {
					t.Errorf("Expected plugin %s not found in result", expected)
				}
			}
		})
	}
}

func TestRemoveDefaultPlugins(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugin-defaults-remove-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	tests := []struct {
		name            string
		existingPlugins []string
		pluginsToRemove []string
		expectedPlugins []string
	}{
		{
			name:            "remove from empty list",
			existingPlugins: []string{},
			pluginsToRemove: []string{"store-redis"},
			expectedPlugins: []string{},
		},
		{
			name:            "remove existing plugin",
			existingPlugins: []string{"store-redis", "js-graal"},
			pluginsToRemove: []string{"store-redis"},
			expectedPlugins: []string{"js-graal"},
		},
		{
			name:            "remove non-existing plugin",
			existingPlugins: []string{"store-redis"},
			pluginsToRemove: []string{"js-graal"},
			expectedPlugins: []string{"store-redis"},
		},
		{
			name:            "remove multiple plugins",
			existingPlugins: []string{"store-redis", "js-graal", "swaggerui"},
			pluginsToRemove: []string{"store-redis", "js-graal"},
			expectedPlugins: []string{"swaggerui"},
		},
		{
			name:            "remove all plugins",
			existingPlugins: []string{"store-redis", "js-graal"},
			pluginsToRemove: []string{"store-redis", "js-graal"},
			expectedPlugins: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup existing plugins
			if len(tt.existingPlugins) > 0 {
				err := writeDefaultPlugins(tt.existingPlugins)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Remove plugins
			err := removeDefaultPlugins(tt.pluginsToRemove)
			if err != nil {
				t.Errorf("removeDefaultPlugins() error = %v", err)
				return
			}

			// Verify result
			plugins, err := ListDefaultPlugins()
			if err != nil {
				t.Fatal(err)
			}

			if len(plugins) != len(tt.expectedPlugins) {
				t.Errorf("After removeDefaultPlugins() got %d plugins, expected %d", len(plugins), len(tt.expectedPlugins))
				return
			}

			// Check that all expected plugins are present
			pluginMap := make(map[string]bool)
			for _, plugin := range plugins {
				pluginMap[plugin] = true
			}

			for _, expected := range tt.expectedPlugins {
				if !pluginMap[expected] {
					t.Errorf("Expected plugin %s not found in result", expected)
				}
			}
		})
	}
}

func TestWriteDefaultPlugins(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugin-write-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	tests := []struct {
		name           string
		pluginsToWrite []string
	}{
		{
			name:           "write empty list",
			pluginsToWrite: []string{},
		},
		{
			name:           "write single plugin",
			pluginsToWrite: []string{"store-redis"},
		},
		{
			name:           "write multiple plugins",
			pluginsToWrite: []string{"store-redis", "js-graal", "swaggerui"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := writeDefaultPlugins(tt.pluginsToWrite)
			if err != nil {
				t.Errorf("writeDefaultPlugins() error = %v", err)
				return
			}

			// Verify the plugins were written correctly
			plugins, err := ListDefaultPlugins()
			if err != nil {
				t.Fatal(err)
			}

			if len(plugins) != len(tt.pluginsToWrite) {
				t.Errorf("writeDefaultPlugins() wrote %d plugins, expected %d", len(plugins), len(tt.pluginsToWrite))
				return
			}

			for i, plugin := range plugins {
				if plugin != tt.pluginsToWrite[i] {
					t.Errorf("writeDefaultPlugins() plugin[%d] = %v, expected %v", i, plugin, tt.pluginsToWrite[i])
				}
			}
		})
	}
}

func TestParseConfigFile(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugin-parse-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	// Test with non-existent config file
	viper, err := parseConfigFile()
	if err != nil {
		t.Errorf("parseConfigFile() with non-existent file error = %v", err)
	}
	if viper == nil {
		t.Error("parseConfigFile() returned nil viper instance")
	}

	// Test with existing config file
	configContent := `default:
  plugins:
    - test-plugin
`
	configFilePath := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configFilePath, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	viper, err = parseConfigFile()
	if err != nil {
		t.Errorf("parseConfigFile() with existing file error = %v", err)
	}
	if viper == nil {
		t.Error("parseConfigFile() returned nil viper instance")
	}

	// Verify the config was parsed correctly
	plugins := viper.GetStringSlice("default.plugins")
	if len(plugins) != 1 || plugins[0] != "test-plugin" {
		t.Errorf("parseConfigFile() parsed plugins incorrectly: got %v, expected [test-plugin]", plugins)
	}
}
