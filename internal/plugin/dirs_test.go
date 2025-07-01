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
