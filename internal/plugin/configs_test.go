package plugin

import (
	"github.com/imposter-project/imposter-cli/internal/config"
	"github.com/imposter-project/imposter-cli/internal/engine"
	"github.com/spf13/viper"
	"os"
	"testing"
)

func TestEnsureConfiguredPlugins(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugin-configs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	tests := []struct {
		name              string
		configuredPlugins []string
		engineType        engine.EngineType
		version           string
		expectedCount     int
		expectError       bool
	}{
		{
			name:              "no configured plugins",
			configuredPlugins: []string{},
			engineType:        engine.EngineTypeDockerCore,
			version:           "4.9.1",
			expectedCount:     0,
			expectError:       false,
		},
		{
			name:              "single configured plugin",
			configuredPlugins: []string{"store-redis"},
			engineType:        engine.EngineTypeDockerCore,
			version:           "4.9.1",
			expectedCount:     1,
			expectError:       false,
		},
		{
			name:              "multiple configured plugins",
			configuredPlugins: []string{"store-redis", "js-nashorn"},
			engineType:        engine.EngineTypeDockerCore,
			version:           "4.9.1",
			expectedCount:     2,
			expectError:       false,
		},
		{
			name:              "duplicate plugins should be deduplicated",
			configuredPlugins: []string{"store-redis", "store-redis"},
			engineType:        engine.EngineTypeDockerCore,
			version:           "4.9.1",
			expectedCount:     1, // store-redis (deduplicated)
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup viper configuration using the new top-level key
			viper.Set(pluginsConfigKey, tt.configuredPlugins)
			defer viper.Set(pluginsConfigKey, nil) // Clean up

			// Test EnsureConfiguredPlugins
			count, err := EnsureConfiguredPlugins(tt.engineType, tt.version)
			if (err != nil) != tt.expectError {
				t.Errorf("EnsureConfiguredPlugins() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if count != tt.expectedCount {
				t.Errorf("EnsureConfiguredPlugins() count = %v, expectedCount %v", count, tt.expectedCount)
			}
		})
	}
}

func TestEnsureConfiguredPlugins_DeprecatedKey(t *testing.T) {
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugin-configs-deprecated-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	tests := []struct {
		name              string
		deprecatedPlugins []string
		expectedCount     int
	}{
		{
			name:              "deprecated key with plugins",
			deprecatedPlugins: []string{"store-redis"},
			expectedCount:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set(defaultPluginsConfigKey, tt.deprecatedPlugins)
			defer viper.Set(defaultPluginsConfigKey, nil)

			count, err := EnsureConfiguredPlugins(engine.EngineTypeDockerCore, "4.9.1")
			if err != nil {
				t.Errorf("EnsureConfiguredPlugins() error = %v", err)
				return
			}
			if count != tt.expectedCount {
				t.Errorf("EnsureConfiguredPlugins() count = %v, expectedCount %v", count, tt.expectedCount)
			}
		})
	}
}

func TestEnsureConfiguredPlugins_BothKeys(t *testing.T) {
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugin-configs-both-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	viper.Set(pluginsConfigKey, []string{"store-redis"})
	viper.Set(defaultPluginsConfigKey, []string{"store-redis"})
	defer func() {
		viper.Set(pluginsConfigKey, nil)
		viper.Set(defaultPluginsConfigKey, nil)
	}()

	count, err := EnsureConfiguredPlugins(engine.EngineTypeDockerCore, "4.9.1")
	if err != nil {
		t.Errorf("EnsureConfiguredPlugins() error = %v", err)
		return
	}
	if count != 1 {
		t.Errorf("EnsureConfiguredPlugins() count = %v, expected 1 (merged and deduplicated)", count)
	}
}
