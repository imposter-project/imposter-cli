package config

import (
	"github.com/imposter-project/imposter-cli/internal/logging"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"testing"
)

func Test_checkCliVersion(t *testing.T) {
	logger = logging.GetLogger()

	tests := []struct {
		name          string
		configVersion string
		required      string
		wantErr       bool
	}{
		{
			name:          "dev version",
			configVersion: DevCliVersion,
			required:      "1.0.0",
			wantErr:       false,
		},
		{
			name:          "version meets requirement",
			configVersion: "1.2.0",
			required:      "1.0.0",
			wantErr:       false,
		},
		{
			name:          "version does not meet requirement",
			configVersion: "0.9.0",
			required:      "1.0.0",
			wantErr:       true,
		},
		{
			name:          "invalid config version",
			configVersion: "invalid",
			required:      "1.0.0",
			wantErr:       true,
		},
		{
			name:          "invalid required version",
			configVersion: "1.0.0",
			required:      "invalid",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Config.Version = tt.configVersion
			err := checkCliVersion(tt.required)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkCliVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_MergeCliConfigIfExists(t *testing.T) {
	logger = logging.GetLogger()

	const legacyName = LegacyLocalDirConfigFileName + ".yaml"
	const canonicalName = LocalDirConfigFileName + ".yaml"

	tests := []struct {
		name        string
		files       map[string]string
		wantSetting string
	}{
		{
			name:        "no local config",
			files:       map[string]string{},
			wantSetting: "",
		},
		{
			name:        "canonical config only",
			files:       map[string]string{canonicalName: "engine: native\n"},
			wantSetting: "native",
		},
		{
			name:        "legacy config only",
			files:       map[string]string{legacyName: "engine: docker\n"},
			wantSetting: "docker",
		},
		{
			name: "canonical takes precedence over legacy",
			files: map[string]string{
				legacyName:    "engine: docker\n",
				canonicalName: "engine: native\n",
			},
			wantSetting: "native",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// isolate global viper state between cases
			viper.Reset()
			t.Cleanup(viper.Reset)

			configDir := t.TempDir()
			for name, content := range tt.files {
				if err := os.WriteFile(filepath.Join(configDir, name), []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}

			MergeCliConfigIfExists(configDir)

			if got := viper.GetString("engine"); got != tt.wantSetting {
				t.Errorf("engine setting = %q, want %q", got, tt.wantSetting)
			}
		})
	}
}

func Test_FindLocalConfigFile(t *testing.T) {
	const legacyName = LegacyLocalDirConfigFileName + ".yaml"
	const canonicalName = LocalDirConfigFileName + ".yaml"

	tests := []struct {
		name     string
		files    []string
		wantFile string
	}{
		{
			name:     "none present",
			files:    nil,
			wantFile: "",
		},
		{
			name:     "canonical present",
			files:    []string{canonicalName},
			wantFile: canonicalName,
		},
		{
			name:     "legacy present",
			files:    []string{legacyName},
			wantFile: legacyName,
		},
		{
			name:     "canonical preferred over legacy",
			files:    []string{legacyName, canonicalName},
			wantFile: canonicalName,
		},
		{
			name:     "yaml preferred over json",
			files:    []string{LocalDirConfigFileName + ".json", canonicalName},
			wantFile: canonicalName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configDir := t.TempDir()
			for _, name := range tt.files {
				if err := os.WriteFile(filepath.Join(configDir, name), []byte("engine: docker\n"), 0644); err != nil {
					t.Fatal(err)
				}
			}

			got := FindLocalConfigFile(configDir)
			want := ""
			if tt.wantFile != "" {
				want = filepath.Join(configDir, tt.wantFile)
			}
			if got != want {
				t.Errorf("FindLocalConfigFile() = %q, want %q", got, want)
			}
		})
	}
}
