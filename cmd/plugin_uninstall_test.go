/*
Copyright © 2021 Pete Cornish <outofcoffee@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"gatehill.io/imposter/internal/config"
	"gatehill.io/imposter/internal/engine"
	"gatehill.io/imposter/internal/plugin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func Test_uninstallPlugins(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-cli-uninstall")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	type args struct {
		plugins       []string
		version       string
		removeDefault bool
		engineType    engine.EngineType
		setupPlugins  []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "uninstall single plugin",
			args: args{
				plugins:      []string{"store-redis"},
				setupPlugins: []string{"store-redis"},
				engineType:   engine.EngineTypeDockerCore,
				version:      "4.2.2",
			},
		},
		{
			name: "uninstall multiple plugins",
			args: args{
				plugins:      []string{"store-redis", "js-graal"},
				setupPlugins: []string{"store-redis", "js-graal"},
				engineType:   engine.EngineTypeDockerCore,
				version:      "4.2.2",
			},
		},
		{
			name: "uninstall and remove from defaults",
			args: args{
				plugins:       []string{"store-redis"},
				setupPlugins:  []string{"store-redis"},
				engineType:    engine.EngineTypeDockerCore,
				version:       "4.2.2",
				removeDefault: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: install plugins first
			viper.Set("default.plugins", tt.args.setupPlugins)
			t.Cleanup(func() {
				viper.Set("default.plugins", nil)
			})

			// Create fake plugin files to simulate installed plugins
			for _, pluginName := range tt.args.setupPlugins {
				_, pluginFilePath, err := plugin.GetPluginLocalPath(pluginName, tt.args.engineType, tt.args.version)
				if err != nil {
					t.Fatal(err)
				}

				// Ensure directory exists
				dir := filepath.Dir(pluginFilePath)
				err = os.MkdirAll(dir, 0755)
				if err != nil {
					t.Fatal(err)
				}

				// Create fake plugin file
				file, err := os.Create(pluginFilePath)
				if err != nil {
					t.Fatal(err)
				}
				file.Close()
			}

			// Test uninstall
			uninstallPlugins(tt.args.plugins, tt.args.engineType, tt.args.version, tt.args.removeDefault)

			// Verify plugins are removed from disk
			for _, pluginName := range tt.args.plugins {
				_, pluginFilePath, err := plugin.GetPluginLocalPath(pluginName, tt.args.engineType, tt.args.version)
				if err != nil {
					t.Fatal(err)
				}

				if _, err := os.Stat(pluginFilePath); !os.IsNotExist(err) {
					t.Errorf("plugin file should be removed: %s", pluginFilePath)
				}
			}

			// Verify default plugins configuration
			if tt.args.removeDefault {
				defaultPlugins, err := plugin.ListDefaultPlugins()
				if err != nil {
					t.Fatal(err)
				}
				for _, pluginName := range tt.args.plugins {
					require.NotContains(t, defaultPlugins, pluginName, "plugin should be removed from defaults")
				}
			}
		})
	}
}

func Test_uninstallNonExistentPlugin(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-cli-uninstall-nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	// Test uninstalling a plugin that doesn't exist
	wasInstalled, err := plugin.UninstallPlugin("nonexistent-plugin", engine.EngineTypeDockerCore, "4.2.2")
	require.NoError(t, err, "should not return error for non-existent plugin")
	require.False(t, wasInstalled, "should return false for non-existent plugin")
}

func Test_uninstallNonInstalledPluginFromDefaults(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-cli-uninstall-defaults")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	// Set up a plugin in defaults but not installed locally
	viper.Set("default.plugins", []string{"swaggerui"})
	t.Cleanup(func() {
		viper.Set("default.plugins", nil)
	})

	// Test uninstalling a plugin that's in defaults but not installed
	uninstallPlugins([]string{"swaggerui"}, engine.EngineTypeGolang, "1.2.4", true)

	// Verify plugin was removed from defaults
	defaultPlugins, err := plugin.ListDefaultPlugins()
	if err != nil {
		t.Fatal(err)
	}
	require.NotContains(t, defaultPlugins, "swaggerui", "plugin should be removed from defaults")
}

func Test_uninstallMultiplePluginsFromDefaults(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-cli-uninstall-multi-defaults")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	// Create a config file with multiple default plugins
	configFilePath := filepath.Join(configDir, "config.yaml")
	configContent := `default:
  plugins:
    - swaggerui
    - store-redis
    - js-graal
`
	err = os.WriteFile(configFilePath, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test uninstalling some plugins that are in defaults but not installed
	uninstallPlugins([]string{"swaggerui", "store-redis"}, engine.EngineTypeGolang, "1.2.4", true)

	// Verify specified plugins were removed from defaults
	defaultPlugins, err := plugin.ListDefaultPlugins()
	if err != nil {
		t.Fatal(err)
	}

	// Check that the removed plugins are no longer in defaults
	require.NotContains(t, defaultPlugins, "swaggerui", "swaggerui should be removed from defaults")
	require.NotContains(t, defaultPlugins, "store-redis", "store-redis should be removed from defaults")

	// Check that js-graal remains (the logic should preserve plugins not requested for removal)
	require.Equal(t, []string{"js-graal"}, defaultPlugins, "js-graal should remain in defaults")
}
