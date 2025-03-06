package cmd

import (
	"gatehill.io/imposter/internal/config"
	"github.com/spf13/cobra"
	"os"
	"path"
	"testing"
)

func Test_setLocalConfig(t *testing.T) {
	type args struct {
		cmd  *cobra.Command
		args []string
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
		fileShouldExist bool
		want            string
	}{
		{
			name: "set local config",
			args: args{
				cmd:  &cobra.Command{},
				args: []string{"foo=bar"},
			},
			wantErr:         false,
			fileShouldExist: true,
			want:            "foo: bar\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp(os.TempDir(), "imposter-cli")
			if err != nil {
				panic(err)
			}

			setLocalConfig(tt.args.cmd, tt.args.args, tempDir)
			cfg, err := os.ReadFile(path.Join(tempDir, config.LocalDirConfigFileName+".yaml"))
			if (err != nil) != tt.wantErr {
				t.Errorf("setLocalConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.fileShouldExist && err != nil {
				t.Errorf("setLocalConfig() file should exist but does not")
				return
			}
			if tt.want != string(cfg) {
				t.Errorf("file contents do not match expected - got %s, want %s", string(cfg), tt.want)
				return
			}
		})
	}
}
