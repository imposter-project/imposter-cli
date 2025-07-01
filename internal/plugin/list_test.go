package plugin

import (
	"gatehill.io/imposter/internal/config"
	"gatehill.io/imposter/internal/engine"
	"os"
	"path/filepath"
	"testing"
)

func TestList(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugin-list-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	version := "4.2.2"

	// Set up isolated config directory
	oldDirPath := config.DirPath
	defer func() { config.DirPath = oldDirPath }()

	// Create plugin directory
	pluginDir, err := EnsurePluginDir(version)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		engineType    engine.EngineType
		version       string
		setupFiles    []string
		expectedCount int
		expectError   bool
	}{
		{
			name:          "empty plugin directory",
			engineType:    engine.EngineTypeDockerCore,
			version:       version,
			setupFiles:    []string{},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "single jvm plugin",
			engineType:    engine.EngineTypeDockerCore,
			version:       version,
			setupFiles:    []string{"imposter-plugin-store-redis.jar"},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "multiple jvm plugins",
			engineType:    engine.EngineTypeDockerCore,
			version:       version,
			setupFiles:    []string{"imposter-plugin-store-redis.jar", "imposter-plugin-js-graal.zip"},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "golang plugin",
			engineType:    engine.EngineTypeGolang,
			version:       version,
			setupFiles:    []string{"plugin-swaggerui"},
			expectedCount: 1,
			expectError:   false,
		},
		{
			name:          "mixed valid and invalid files",
			engineType:    engine.EngineTypeDockerCore,
			version:       version,
			setupFiles:    []string{"imposter-plugin-store-redis.jar", "not-a-plugin.txt", "imposter-plugin-js-graal.zip"},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name:          "empty setup for non-existent version",
			engineType:    engine.EngineTypeDockerCore,
			version:       "0.0.0",
			setupFiles:    []string{},
			expectedCount: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test files
			if tt.version == version {
				// Clean plugin directory first
				files, err := os.ReadDir(pluginDir)
				if err == nil {
					for _, file := range files {
						os.Remove(filepath.Join(pluginDir, file.Name()))
					}
				}

				// Create test plugin files
				for _, filename := range tt.setupFiles {
					filePath := filepath.Join(pluginDir, filename)
					file, err := os.Create(filePath)
					if err != nil {
						t.Fatal(err)
					}
					file.Close()
				}
			}

			// Test List function
			plugins, err := List(tt.engineType, tt.version)
			if (err != nil) != tt.expectError {
				t.Errorf("List() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError && len(plugins) != tt.expectedCount {
				t.Errorf("List() returned %d plugins, expected %d", len(plugins), tt.expectedCount)
			}

			// Verify plugin metadata structure
			for _, plugin := range plugins {
				if plugin.Version != tt.version {
					t.Errorf("Plugin version = %v, expected %v", plugin.Version, tt.version)
				}
				if plugin.EngineType != tt.engineType {
					t.Errorf("Plugin engine type = %v, expected %v", plugin.EngineType, tt.engineType)
				}
				if plugin.Name == "" {
					t.Error("Plugin name should not be empty")
				}
			}
		})
	}
}

func TestListVersionDirs(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugin-version-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	tests := []struct {
		name             string
		setupVersions    []string
		expectedVersions []string
		expectError      bool
	}{
		{
			name:             "no version directories",
			setupVersions:    []string{},
			expectedVersions: []string{},
			expectError:      false,
		},
		{
			name:             "single version directory",
			setupVersions:    []string{"4.2.2"},
			expectedVersions: []string{"4.2.2"},
			expectError:      false,
		},
		{
			name:             "multiple version directories",
			setupVersions:    []string{"4.2.2", "4.2.1", "latest"},
			expectedVersions: []string{"4.2.1", "4.2.2", "latest"}, // Should be sorted
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing plugin directories
			baseDir, err := getBasePluginDir()
			if err == nil {
				os.RemoveAll(baseDir)
			}

			// Setup version directories
			for _, version := range tt.setupVersions {
				_, err := EnsurePluginDir(version)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Test ListVersionDirs function
			versions, err := ListVersionDirs()
			if (err != nil) != tt.expectError {
				t.Errorf("ListVersionDirs() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError {
				if len(versions) != len(tt.expectedVersions) {
					t.Errorf("ListVersionDirs() returned %d versions, expected %d", len(versions), len(tt.expectedVersions))
					return
				}

				// Check that all expected versions are present
				versionMap := make(map[string]bool)
				for _, version := range versions {
					versionMap[version] = true
				}

				for _, expected := range tt.expectedVersions {
					if !versionMap[expected] {
						t.Errorf("Expected version %s not found in result", expected)
					}
				}
			}
		})
	}
}
