package engine

import (
	"github.com/spf13/viper"
	"strings"
	"testing"
)

func TestGetConfiguredVersion(t *testing.T) {
	type args struct {
		override    string
		allowCached bool
	}
	tests := []struct {
		name             string
		args             args
		configureVersion string
		want             string
	}{
		{name: "return overridden version", args: args{override: "2.0.0", allowCached: false}, want: "2.0.0"},
		{name: "return configured version", args: args{override: "", allowCached: false}, configureVersion: "1.2.3", want: "1.2.3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.configureVersion != "" {
				viper.Set("version", tt.configureVersion)
			}
			t.Cleanup(func() {
				viper.Set("version", nil)
			})
			if got := GetConfiguredVersion(tt.args.override, tt.args.allowCached); got != tt.want {
				t.Errorf("GetConfiguredVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetConfiguredType(t *testing.T) {
	type args struct {
		override string
	}
	tests := []struct {
		name          string
		args          args
		configureType string
		want          EngineType
	}{
		{name: "return overridden engine type", args: args{override: "docker"}, want: "docker"},
		{name: "return configured engine type", args: args{override: ""}, configureType: "jvm", want: "jvm"},
		{name: "return default engine type", args: args{override: ""}, want: defaultEngineType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.configureType != "" {
				viper.Set("engine", tt.configureType)
			}
			t.Cleanup(func() {
				viper.Set("engine", nil)
			})
			if got := GetConfiguredType(tt.args.override); got != tt.want {
				t.Errorf("GetConfiguredType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitiseVersionOutput(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "remove version", args: args{s: "Version: 1.2.3"}, want: "1.2.3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitiseVersionOutput(tt.args.s); got != tt.want {
				t.Errorf("SanitiseVersionOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildEnvFromParent(t *testing.T) {
	type args struct {
		options     StartOptions
		includeHome bool
		env         []string
	}
	tests := []struct {
		name         string
		args         args
		wantPrefixes []string
	}{
		{name: "should include home", args: args{options: StartOptions{LogLevel: "WARN"}, includeHome: true, env: []string{"HOME=/home/example"}}, wantPrefixes: []string{"HOME="}},
		{name: "should exclude home", args: args{options: StartOptions{LogLevel: "WARN"}, includeHome: false, env: []string{"HOME=/home/example"}}, wantPrefixes: []string{""}},
		{name: "should set log level", args: args{options: StartOptions{LogLevel: "WARN"}, includeHome: false, env: []string{}}, wantPrefixes: []string{"IMPOSTER_LOG_LEVEL=WARN"}},
		{name: "should pass through imposter env var", args: args{options: StartOptions{LogLevel: "WARN"}, includeHome: false, env: []string{"IMPOSTER_TEST=foo"}}, wantPrefixes: []string{"IMPOSTER_TEST=foo"}},
		{name: "should pass through log level env var", args: args{options: StartOptions{LogLevel: "WARN"}, includeHome: false, env: []string{"IMPOSTER_LOG_LEVEL=ERROR"}}, wantPrefixes: []string{"IMPOSTER_LOG_LEVEL=ERROR"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildEnvFromParent(tt.args.env, tt.args.options, tt.args.includeHome)

			found := false
			for _, prefix := range tt.wantPrefixes {
				for _, env := range got {
					if strings.HasPrefix(env, prefix) {
						found = true
					}
				}
			}
			if !found {
				t.Errorf("buildEnvFromParent() = %v, wantPrefixes %v", got, tt.wantPrefixes)
			}
		})
	}
}
