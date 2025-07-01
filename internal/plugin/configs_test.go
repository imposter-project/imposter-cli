package plugin

import (
	"gatehill.io/imposter/internal/config"
	"gatehill.io/imposter/internal/engine"
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
			version:           "4.2.2",
			expectedCount:     0,
			expectError:       false,
		},
		{
			name:              "single configured plugin",
			configuredPlugins: []string{"store-redis"},
			engineType:        engine.EngineTypeDockerCore,
			version:           "4.2.2",
			expectedCount:     1,
			expectError:       false,
		},
		{
			name:              "multiple configured plugins",
			configuredPlugins: []string{"store-redis", "js-nashorn"},
			engineType:        engine.EngineTypeDockerCore,
			version:           "4.2.2",
			expectedCount:     2,
			expectError:       false,
		},
		{
			name:              "duplicate plugins should be deduplicated",
			configuredPlugins: []string{"store-redis", "store-redis"},
			engineType:        engine.EngineTypeDockerCore,
			version:           "4.2.2",
			expectedCount:     1, // store-redis (deduplicated)
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup viper configuration
			viper.Set(defaultPluginsConfigKey, tt.configuredPlugins)
			defer viper.Set(defaultPluginsConfigKey, nil) // Clean up

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
