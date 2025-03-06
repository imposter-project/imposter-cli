package plugin

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	logger.SetLevel(logrus.TraceLevel)
}

func TestEnsurePluginDir(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "ensure plugin base dir", args: args{version: "1.2.3"}, want: filepath.Join(homeDir, pluginBaseDir, "1.2.3"), wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EnsurePluginDir(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsurePluginDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EnsurePluginDir() got = %v, want %v", got, tt.want)
			}
			stat, err := os.Stat(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsurePluginDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !stat.IsDir() {
				t.Errorf("EnsurePluginDir() path '%s' is not a directory", got)
			}
		})
	}
}

func TestEnsurePlugin(t *testing.T) {
	type args struct {
		pluginName string
		version    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "fetch plugin", args: args{pluginName: "store-redis", version: "4.2.2"}, wantErr: false},
		{name: "fetch nonexistent plugin version", args: args{pluginName: "store-redis", version: "0.0.0"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := EnsurePlugin(tt.args.pluginName, tt.args.version); (err != nil) != tt.wantErr {
				t.Errorf("EnsurePlugin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnsurePlugins(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		plugins []string
		wantErr bool
	}{
		{name: "no op if no plugins configured", args: args{version: "4.2.2"}, plugins: nil, wantErr: false},
		{name: "fetch configured plugins", args: args{version: "4.2.2"}, plugins: []string{"store-redis"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ensured, err := EnsurePlugins(tt.plugins, tt.args.version, false)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsurePlugins() error = %v, wantErr %v", err, tt.wantErr)
			}
			if ensured != len(tt.plugins) {
				t.Errorf("EnsurePlugins() wanted %d plugins, ensured: %d", len(tt.plugins), ensured)
			}
		})
	}
}

func Test_getPluginFilePath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	type args struct {
		pluginName string
		version    string
	}
	tests := []struct {
		name                   string
		args                   args
		wantFullPluginFileName string
		wantPluginFilePath     string
		wantErr                bool
	}{
		{
			name:                   "get plugin file path",
			args:                   args{pluginName: "store-redis", version: "4.2.2"},
			wantFullPluginFileName: "imposter-plugin-store-redis.jar",
			wantPluginFilePath:     filepath.Join(homeDir, pluginBaseDir, "4.2.2", "imposter-plugin-store-redis.jar"),
			wantErr:                false,
		},
		{
			name:                   "get plugin file path with zip suffix",
			args:                   args{pluginName: "js-graal:zip", version: "4.2.2"},
			wantFullPluginFileName: "imposter-plugin-js-graal.zip",
			wantPluginFilePath:     filepath.Join(homeDir, pluginBaseDir, "4.2.2", "imposter-plugin-js-graal.zip"),
			wantErr:                false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFullPluginFileName, gotPluginFilePath, err := getPluginFilePath(tt.args.pluginName, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPluginFilePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFullPluginFileName != tt.wantFullPluginFileName {
				t.Errorf("getPluginFilePath() gotFullPluginFileName = %v, want %v", gotFullPluginFileName, tt.wantFullPluginFileName)
			}
			if gotPluginFilePath != tt.wantPluginFilePath {
				t.Errorf("getPluginFilePath() gotPluginFilePath = %v, want %v", gotPluginFilePath, tt.wantPluginFilePath)
			}
		})
	}
}
