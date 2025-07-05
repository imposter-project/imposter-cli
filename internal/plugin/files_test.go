package plugin

import (
	"fmt"
	"gatehill.io/imposter/internal/engine"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func Test_buildPluginFileName(t *testing.T) {
	tests := []struct {
		name         string
		pluginConfig pluginConfiguration
		pluginName   string
		ext          string
		remote       bool
		want         string
	}{
		{
			name: "golang plugin local file",
			pluginConfig: pluginConfiguration{
				localFileTemplate:  "plugin-{{ .PluginName }}",
				remoteFileTemplate: "plugin-{{ .PluginName }}_{{ .OS }}_{{ .Arch }}{{ .Ext }}",
				extensions:         []string{".zip"},
			},
			pluginName: "test-plugin",
			ext:        ".zip",
			remote:     false,
			want:       "plugin-test-plugin",
		},
		{
			name: "golang plugin remote file",
			pluginConfig: pluginConfiguration{
				localFileTemplate:  "plugin-{{ .PluginName }}",
				remoteFileTemplate: "plugin-{{ .PluginName }}_{{ .OS }}_{{ .Arch }}{{ .Ext }}",
				extensions:         []string{".zip"},
			},
			pluginName: "test-plugin",
			ext:        ".zip",
			remote:     true,
			want:       "plugin-test-plugin_" + runtime.GOOS + "_" + runtime.GOARCH + ".zip",
		},
		{
			name: "jvm plugin local file",
			pluginConfig: pluginConfiguration{
				localFileTemplate:  "imposter-plugin-{{ .PluginName }}{{ .Ext }}",
				remoteFileTemplate: "imposter-plugin-{{ .PluginName }}{{ .Ext }}",
				extensions:         []string{".jar", ".zip"},
			},
			pluginName: "store-redis",
			ext:        ".jar",
			remote:     false,
			want:       "imposter-plugin-store-redis.jar",
		},
		{
			name: "jvm plugin remote file",
			pluginConfig: pluginConfiguration{
				localFileTemplate:  "imposter-plugin-{{ .PluginName }}{{ .Ext }}",
				remoteFileTemplate: "imposter-plugin-{{ .PluginName }}{{ .Ext }}",
				extensions:         []string{".jar", ".zip"},
			},
			pluginName: "store-redis",
			ext:        ".jar",
			remote:     true,
			want:       "imposter-plugin-store-redis.jar",
		},
		{
			name: "plugin with empty name",
			pluginConfig: pluginConfiguration{
				localFileTemplate:  "plugin-{{ .PluginName }}",
				remoteFileTemplate: "plugin-{{ .PluginName }}_{{ .OS }}_{{ .Arch }}{{ .Ext }}",
				extensions:         []string{".zip"},
			},
			pluginName: "",
			ext:        ".zip",
			remote:     false,
			want:       "plugin-",
		},
		{
			name: "plugin with different extension",
			pluginConfig: pluginConfiguration{
				localFileTemplate:  "imposter-plugin-{{ .PluginName }}{{ .Ext }}",
				remoteFileTemplate: "imposter-plugin-{{ .PluginName }}{{ .Ext }}",
				extensions:         []string{".jar", ".zip"},
			},
			pluginName: "multi-ext",
			ext:        ".zip",
			remote:     false,
			want:       "imposter-plugin-multi-ext.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPluginFileName(tt.pluginConfig, tt.pluginName, tt.ext, tt.remote)
			if got != tt.want {
				t.Errorf("buildPluginFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_GetPluginLocalPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		pluginName string
		version    string
		engineType engine.EngineType
	}
	tests := []struct {
		name                   string
		args                   args
		wantFullPluginFileName string
		wantPluginFilePath     string
		wantErr                bool
	}{
		{
			name:                   "get plugin local path for jvm plugin",
			args:                   args{pluginName: "store-redis", engineType: engine.EngineTypeDockerCore, version: "4.2.2"},
			wantFullPluginFileName: "imposter-plugin-store-redis.jar",
			wantPluginFilePath:     filepath.Join(homeDir, pluginBaseDir, "4.2.2", "imposter-plugin-store-redis.jar"),
			wantErr:                false,
		},
		{
			name:                   "get plugin local path with zip suffix",
			args:                   args{pluginName: "js-graal:zip", engineType: engine.EngineTypeDockerCore, version: "4.2.2"},
			wantFullPluginFileName: "imposter-plugin-js-graal.zip",
			wantPluginFilePath:     filepath.Join(homeDir, pluginBaseDir, "4.2.2", "imposter-plugin-js-graal.zip"),
			wantErr:                false,
		},
		{
			name: "get plugin local path for golang plugin",
			args: args{pluginName: "swaggerui", engineType: engine.EngineTypeGolang, version: "1.2.2"},
			wantFullPluginFileName: func() string {
				if runtime.GOOS == "windows" {
					return "plugin-swaggerui.exe"
				}
				return "plugin-swaggerui"
			}(),
			wantPluginFilePath: func() string {
				fileName := "plugin-swaggerui"
				if runtime.GOOS == "windows" {
					fileName = "plugin-swaggerui.exe"
				}
				return filepath.Join(homeDir, pluginBaseDir, "1.2.2", fileName)
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFullPluginFileName, gotPluginFilePath, err := GetPluginLocalPath(tt.args.pluginName, tt.args.engineType, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPluginLocalPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFullPluginFileName != tt.wantFullPluginFileName {
				t.Errorf("GetPluginLocalPath() gotFullPluginFileName = %v, want %v", gotFullPluginFileName, tt.wantFullPluginFileName)
			}
			if gotPluginFilePath != tt.wantPluginFilePath {
				t.Errorf("GetPluginLocalPath() gotPluginFilePath = %v, want %v", gotPluginFilePath, tt.wantPluginFilePath)
			}
		})
	}
}

func Test_getPluginFileName(t *testing.T) {
	tests := []struct {
		name       string
		engineType engine.EngineType
		pluginName string
		remote     bool
		want       string
		wantErr    bool
	}{
		{
			name:       "golang plugin local name",
			engineType: engine.EngineTypeGolang,
			pluginName: "swaggerui",
			remote:     false,
			want: func() string {
				if runtime.GOOS == "windows" {
					return "plugin-swaggerui.exe"
				}
				return "plugin-swaggerui"
			}(),
			wantErr: false,
		},
		{
			name:       "golang plugin remote name",
			engineType: engine.EngineTypeGolang,
			pluginName: "swaggerui",
			remote:     true,
			want:       fmt.Sprintf("plugin-swaggerui_%s_%s.zip", runtime.GOOS, runtime.GOARCH),
			wantErr:    false,
		},
		{
			name:       "jvm plugin local name",
			engineType: engine.EngineTypeDockerCore,
			pluginName: "store-redis",
			remote:     false,
			want:       "imposter-plugin-store-redis.jar",
			wantErr:    false,
		},
		{
			name:       "jvm plugin remote name",
			engineType: engine.EngineTypeDockerCore,
			pluginName: "store-redis",
			remote:     true,
			want:       "imposter-plugin-store-redis.jar",
			wantErr:    false,
		},
		{
			name:       "jvm plugin with zip extension",
			engineType: engine.EngineTypeDockerCore,
			pluginName: "js-graal:zip",
			remote:     false,
			want:       "imposter-plugin-js-graal.zip",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPluginFileName(tt.engineType, tt.pluginName, tt.remote)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPluginFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getPluginFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isValidPluginFile(t *testing.T) {
	tests := []struct {
		name           string
		candidateFile  string
		engineType     engine.EngineType
		wantValid      bool
		wantPluginName string
	}{
		{
			name:           "invalid file extension",
			candidateFile:  "imposter-plugin-test.txt",
			engineType:     engine.EngineTypeDockerCore,
			wantValid:      false,
			wantPluginName: "",
		},
		{
			name:           "invalid file prefix",
			candidateFile:  "wrong-prefix-test.jar",
			engineType:     engine.EngineTypeDockerCore,
			wantValid:      false,
			wantPluginName: "",
		},
		{
			name: "valid golang plugin file",
			candidateFile: func() string {
				if runtime.GOOS == "windows" {
					return "plugin-swaggerui.exe"
				}
				return "plugin-swaggerui"
			}(),
			engineType:     engine.EngineTypeGolang,
			wantValid:      true,
			wantPluginName: "swaggerui",
		},
		{
			name:           "valid jvm plugin file",
			candidateFile:  "imposter-plugin-store-redis.jar",
			engineType:     engine.EngineTypeDockerCore,
			wantValid:      true,
			wantPluginName: "store-redis",
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
