package jvm

import (
	"testing"

	"gatehill.io/imposter/internal/engine/procutil"
)

func Test_readArg(t *testing.T) {
	type args struct {
		cmdline  []string
		longArg  string
		shortArg string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "combined form",
			args: args{
				cmdline:  []string{"--arg=VAL"},
				longArg:  "arg",
				shortArg: "a",
			},
			want: "VAL",
		},
		{
			name: "separate form",
			args: args{
				cmdline:  []string{"--arg", "VAL"},
				longArg:  "arg",
				shortArg: "a",
			},
			want: "VAL",
		},
		{
			name: "separate form with short arg",
			args: args{
				cmdline:  []string{"-a", "VAL"},
				longArg:  "arg",
				shortArg: "a",
			},
			want: "VAL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := procutil.ReadArg(tt.args.cmdline, tt.args.longArg, tt.args.shortArg); got != tt.want {
				t.Errorf("readArg() = %v, want %v", got, tt.want)
			}
		})
	}
}
