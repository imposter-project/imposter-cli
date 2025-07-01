package plugin

import (
	"gatehill.io/imposter/internal/config"
	"gatehill.io/imposter/internal/engine"
	"gatehill.io/imposter/internal/stringutil"
	"os"
	"path/filepath"
	"testing"
)

func TestUninstallPlugin(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugin-uninstall-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	version := "4.2.2"

	tests := []struct {
		name          string
		pluginName    string
		engineType    engine.EngineType
		version       string
		setupFile     bool
		expectRemoved bool
		expectError   bool
	}{
		{
			name:          "uninstall existing jvm plugin",
			pluginName:    "store-redis",
			engineType:    engine.EngineTypeDockerCore,
			version:       version,
			setupFile:     true,
			expectRemoved: true,
			expectError:   false,
		},
		{
			name:          "uninstall non-existent plugin",
			pluginName:    "non-existent",
			engineType:    engine.EngineTypeDockerCore,
			version:       version,
			setupFile:     false,
			expectRemoved: false,
			expectError:   false,
		},
		{
			name:          "uninstall existing golang plugin",
			pluginName:    "swaggerui",
			engineType:    engine.EngineTypeGolang,
			version:       version,
			setupFile:     true,
			expectRemoved: true,
			expectError:   false,
		},
		{
			name:          "uninstall plugin with zip extension",
			pluginName:    "js-graal:zip",
			engineType:    engine.EngineTypeDockerCore,
			version:       version,
			setupFile:     true,
			expectRemoved: true,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin file if needed
			var pluginFilePath string
			if tt.setupFile {
				_, pluginFilePath, err = GetPluginLocalPath(tt.pluginName, tt.engineType, tt.version)
				if err != nil {
					t.Fatal(err)
				}

				// Ensure directory exists
				dir := filepath.Dir(pluginFilePath)
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					t.Fatal(err)
				}

				// Create plugin file
				file, err := os.Create(pluginFilePath)
				if err != nil {
					t.Fatal(err)
				}
				file.Close()
			}

			// Test UninstallPlugin
			wasRemoved, err := UninstallPlugin(tt.pluginName, tt.engineType, tt.version)
			if (err != nil) != tt.expectError {
				t.Errorf("UninstallPlugin() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if wasRemoved != tt.expectRemoved {
				t.Errorf("UninstallPlugin() wasRemoved = %v, expectRemoved %v", wasRemoved, tt.expectRemoved)
			}

			// Verify file was removed if expected
			if tt.setupFile && tt.expectRemoved {
				if _, err := os.Stat(pluginFilePath); !os.IsNotExist(err) {
					t.Error("Plugin file should have been removed")
				}
			}
		})
	}
}

func TestUninstallPlugins(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugins-uninstall-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	version := "4.2.2"

	tests := []struct {
		name            string
		pluginsToRemove []string
		engineType      engine.EngineType
		version         string
		setupFiles      []string
		removeDefault   bool
		setupDefaults   []string
		expectedRemoved int
		expectError     bool
	}{
		{
			name:            "uninstall single existing plugin",
			pluginsToRemove: []string{"store-redis"},
			engineType:      engine.EngineTypeDockerCore,
			version:         version,
			setupFiles:      []string{"store-redis"},
			removeDefault:   false,
			expectedRemoved: 1,
			expectError:     false,
		},
		{
			name:            "uninstall multiple existing plugins",
			pluginsToRemove: []string{"store-redis", "js-graal:zip"},
			engineType:      engine.EngineTypeDockerCore,
			version:         version,
			setupFiles:      []string{"store-redis", "js-graal:zip"},
			removeDefault:   false,
			expectedRemoved: 2,
			expectError:     false,
		},
		{
			name:            "uninstall non-existent plugins",
			pluginsToRemove: []string{"non-existent1", "non-existent2"},
			engineType:      engine.EngineTypeDockerCore,
			version:         version,
			setupFiles:      []string{},
			removeDefault:   false,
			expectedRemoved: 0,
			expectError:     false,
		},
		{
			name:            "uninstall mixed existing and non-existent",
			pluginsToRemove: []string{"store-redis", "non-existent"},
			engineType:      engine.EngineTypeDockerCore,
			version:         version,
			setupFiles:      []string{"store-redis"},
			removeDefault:   false,
			expectedRemoved: 1,
			expectError:     false,
		},
		{
			name:            "uninstall with default removal",
			pluginsToRemove: []string{"store-redis"},
			engineType:      engine.EngineTypeDockerCore,
			version:         version,
			setupFiles:      []string{"store-redis"},
			removeDefault:   true,
			setupDefaults:   []string{"store-redis", "js-graal"},
			expectedRemoved: 1,
			expectError:     false,
		},
		{
			name:            "uninstall empty list",
			pluginsToRemove: []string{},
			engineType:      engine.EngineTypeDockerCore,
			version:         version,
			setupFiles:      []string{},
			removeDefault:   false,
			expectedRemoved: 0,
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup plugin files
			for _, pluginName := range tt.setupFiles {
				_, pluginFilePath, err := GetPluginLocalPath(pluginName, tt.engineType, tt.version)
				if err != nil {
					t.Fatal(err)
				}

				// Ensure directory exists
				dir := filepath.Dir(pluginFilePath)
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					t.Fatal(err)
				}

				// Create plugin file
				file, err := os.Create(pluginFilePath)
				if err != nil {
					t.Fatal(err)
				}
				file.Close()
			}

			// Setup default plugins
			if len(tt.setupDefaults) > 0 {
				err := writeDefaultPlugins(tt.setupDefaults)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Test UninstallPlugins
			removed, err := UninstallPlugins(tt.pluginsToRemove, tt.engineType, tt.version, tt.removeDefault)
			if (err != nil) != tt.expectError {
				t.Errorf("UninstallPlugins() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if removed != tt.expectedRemoved {
				t.Errorf("UninstallPlugins() removed = %v, expectedRemoved %v", removed, tt.expectedRemoved)
			}

			// Verify files were removed
			for _, pluginName := range tt.setupFiles {
				if stringutil.Contains(tt.pluginsToRemove, pluginName) {
					_, pluginFilePath, err := GetPluginLocalPath(pluginName, tt.engineType, tt.version)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := os.Stat(pluginFilePath); !os.IsNotExist(err) {
						t.Errorf("Plugin file %s should have been removed", pluginFilePath)
					}
				}
			}

			// Verify default plugins were updated if requested
			if tt.removeDefault && len(tt.setupDefaults) > 0 {
				defaults, err := ListDefaultPlugins()
				if err != nil {
					t.Fatal(err)
				}

				for _, removed := range tt.pluginsToRemove {
					if stringutil.Contains(defaults, removed) {
						t.Errorf("Plugin %s should have been removed from defaults", removed)
					}
				}
			}
		})
	}
}
