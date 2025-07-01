package plugin

import (
	"gatehill.io/imposter/internal/config"
	"gatehill.io/imposter/internal/engine"
	"os"
	"testing"
)

func TestGetPluginRemoteFileName(t *testing.T) {
	tests := []struct {
		name       string
		engineType engine.EngineType
		pluginName string
		want       string
		wantErr    bool
	}{
		{
			name:       "jvm plugin remote filename",
			engineType: engine.EngineTypeDockerCore,
			pluginName: "store-redis",
			want:       "imposter-plugin-store-redis.jar",
			wantErr:    false,
		},
		{
			name:       "golang plugin remote filename",
			engineType: engine.EngineTypeGolang,
			pluginName: "swaggerui",
			want:       "plugin-swaggerui_darwin_arm64.zip",
			wantErr:    false,
		},
		{
			name:       "jvm plugin with zip extension",
			engineType: engine.EngineTypeDockerCore,
			pluginName: "js-graal:zip",
			want:       "imposter-plugin-js-graal.zip",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPluginRemoteFileName(tt.engineType, tt.pluginName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPluginRemoteFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getPluginRemoteFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidPluginFileEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		candidateFile  string
		engineType     engine.EngineType
		wantValid      bool
		wantPluginName string
	}{
		{
			name:           "empty filename",
			candidateFile:  "",
			engineType:     engine.EngineTypeDockerCore,
			wantValid:      false,
			wantPluginName: "",
		},
		{
			name:           "filename with spaces",
			candidateFile:  "imposter plugin store redis.jar",
			engineType:     engine.EngineTypeDockerCore,
			wantValid:      false,
			wantPluginName: "",
		},
		{
			name:           "correct prefix but wrong extension",
			candidateFile:  "imposter-plugin-test.exe",
			engineType:     engine.EngineTypeDockerCore,
			wantValid:      false,
			wantPluginName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValid, gotPluginName := isValidPluginFile(tt.candidateFile, tt.engineType)
			if gotValid != tt.wantValid {
				t.Errorf("isValidPluginFile() gotValid = %v, want %v", gotValid, tt.wantValid)
			}
			if gotPluginName != tt.wantPluginName {
				t.Errorf("isValidPluginFile() gotPluginName = %v, want %v", gotPluginName, tt.wantPluginName)
			}
		})
	}
}

func TestGetPluginFileNameErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		engineType  engine.EngineType
		pluginName  string
		remote      bool
		expectError bool
	}{
		{
			name:        "plugin with invalid extension format",
			engineType:  engine.EngineTypeDockerCore,
			pluginName:  "test:invalidext",
			remote:      false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getPluginFileName(tt.engineType, tt.pluginName, tt.remote)
			if (err != nil) != tt.expectError {
				t.Errorf("getPluginFileName() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestEnsurePluginsWithSaveDefault(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugin-save-default-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	// Test with empty plugins list to avoid network calls
	plugins := []string{}
	engineType := engine.EngineTypeDockerCore
	version := "4.2.2"

	// Test with saveDefault = true
	count, err := EnsurePlugins(plugins, engineType, version, true)
	if err != nil {
		t.Errorf("EnsurePlugins() with saveDefault=true error = %v", err)
		return
	}

	if count != 0 {
		t.Errorf("EnsurePlugins() returned count = %v, expected 0", count)
	}

	// Verify defaults functionality works (empty list)
	defaults, err := ListDefaultPlugins()
	if err != nil {
		t.Fatal(err)
	}

	if len(defaults) != 0 {
		t.Errorf("Expected 0 default plugins, got %d", len(defaults))
	}
}

func TestDownloadPluginDirectly(t *testing.T) {
	// Setup temporary config directory
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-plugin-download-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(configDir)
	config.DirPath = configDir

	// Test downloading a plugin with a version that doesn't exist to test error handling
	// This will timeout on download, so we expect an error
	err = downloadPlugin(engine.EngineTypeDockerCore, "non-existent-plugin", "0.0.0")
	if err == nil {
		t.Error("downloadPlugin() should have failed for non-existent plugin and version")
	}
}

func TestEnsurePluginDirError(t *testing.T) {
	// Test with empty version string
	_, err := EnsurePluginDir("")
	if err != nil {
		// This is expected to work, just return empty dir
		t.Logf("EnsurePluginDir with empty version returned error: %v", err)
	}

	// Test directory creation works
	dir, err := EnsurePluginDir("test-version")
	if err != nil {
		t.Errorf("EnsurePluginDir() should have succeeded: %v", err)
	}
	if dir == "" {
		t.Error("EnsurePluginDir() should return non-empty directory path")
	}
}
