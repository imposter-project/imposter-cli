package library

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestGetDirPath(t *testing.T) {
	// Create temporary home directory
	tempHome, err := os.MkdirTemp("", "home_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempHome)

	// Set up test home directory
	t.Setenv("HOME", tempHome)

	tests := []struct {
		name         string
		settingsKey  string
		configValue  string
		homeSubDir   string
		wantContains string
		wantErr      bool
	}{
		{
			name:         "use config value",
			settingsKey:  "test.dir",
			configValue:  "/custom/path",
			homeSubDir:   ".imposter/test",
			wantContains: "/custom/path",
			wantErr:      false,
		},
		{
			name:         "use home subdirectory",
			settingsKey:  "test.unused",
			configValue:  "",
			homeSubDir:   ".imposter/test",
			wantContains: filepath.Join(tempHome, ".imposter/test"),
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up viper config for this test
			viper.Set(tt.settingsKey, tt.configValue)
			defer viper.Set(tt.settingsKey, nil)

			got, err := GetDirPath(tt.settingsKey, tt.homeSubDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDirPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantContains {
				t.Errorf("GetDirPath() = %v, want %v", got, tt.wantContains)
			}
		})
	}
}

func TestEnsureDir(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "ensure_dir_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
	}{
		{
			name: "create new directory",
			setup: func(t *testing.T) string {
				return filepath.Join(tempDir, "new_dir")
			},
			wantErr: false,
		},
		{
			name: "use existing directory",
			setup: func(t *testing.T) string {
				dir := filepath.Join(tempDir, "existing_dir")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantErr: false,
		},
		{
			name: "error on file path",
			setup: func(t *testing.T) string {
				path := filepath.Join(tempDir, "file.txt")
				if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dirPath := tt.setup(t)
			err := EnsureDir(dirPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify directory exists and is a directory
				info, err := os.Stat(dirPath)
				if err != nil {
					t.Errorf("Failed to stat created directory: %v", err)
					return
				}
				if !info.IsDir() {
					t.Error("Created path is not a directory")
				}
			}
		})
	}
}

func TestEnsureDirUsingConfig(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "ensure_dir_config_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set up test home directory
	t.Setenv("HOME", tempDir)

	tests := []struct {
		name        string
		settingsKey string
		configValue string
		homeSubDir  string
		wantErr     bool
		verifyDir   bool
	}{
		{
			name:        "create directory from config",
			settingsKey: "test.configdir",
			configValue: filepath.Join(tempDir, "config_dir"),
			homeSubDir:  ".imposter/test",
			wantErr:     false,
			verifyDir:   true,
		},
		{
			name:        "create directory from home",
			settingsKey: "test.homedir",
			configValue: "",
			homeSubDir:  ".imposter/home_dir",
			wantErr:     false,
			verifyDir:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up viper config for this test
			viper.Set(tt.settingsKey, tt.configValue)
			defer viper.Set(tt.settingsKey, nil)

			got, err := EnsureDirUsingConfig(tt.settingsKey, tt.homeSubDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureDirUsingConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.verifyDir {
				// Verify directory exists and is a directory
				info, err := os.Stat(got)
				if err != nil {
					t.Errorf("Failed to stat created directory: %v", err)
					return
				}
				if !info.IsDir() {
					t.Error("Created path is not a directory")
				}
			}
		})
	}
}
