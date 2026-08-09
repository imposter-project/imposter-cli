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
		engineType  EngineType
	}
	tests := []struct {
		name             string
		args             args
		configureVersion string
		want             string
	}{
		{name: "return overridden version", args: args{engineType: EngineTypeDockerCore, override: "2.0.0", allowCached: false}, want: "2.0.0"},
		{name: "return configured version", args: args{engineType: EngineTypeDockerCore, override: "", allowCached: false}, configureVersion: "1.2.3", want: "1.2.3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.configureVersion != "" {
				viper.Set("version", tt.configureVersion)
			}
			t.Cleanup(func() {
				viper.Set("version", nil)
			})
			if got := GetConfiguredVersion(tt.args.engineType, tt.args.override, tt.args.allowCached); got != tt.want {
				t.Errorf("GetConfiguredVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetConfiguredVersionUnresolved(t *testing.T) {
	tests := []struct {
		name     string
		override string
		want     string
	}{
		{name: "latest alias left unresolved", override: "latest", want: "latest"},
		{name: "major alias left unresolved", override: "5", want: "5"},
		{name: "no version defaults to latest alias", override: "", want: "latest"},
		{name: "explicit version returned as-is", override: "5.21.3", want: "5.21.3"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetConfiguredVersionOrResolve(EngineTypeDockerCore, tt.override, false, false)
			if got != tt.want {
				t.Errorf("GetConfiguredVersionOrResolve() = %v, want %v", got, tt.want)
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
		{name: "normalise legacy golang override to native", args: args{override: "golang"}, want: EngineTypeNative},
		{name: "normalise legacy golang config to native", args: args{override: ""}, configureType: "golang", want: EngineTypeNative},
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

func TestGetConfiguredTypeWithVersion(t *testing.T) {
	type args struct {
		typeOverride    string
		versionOverride string
	}
	tests := []struct {
		name             string
		args             args
		configureType    string
		configureVersion string
		want             EngineType
	}{
		{name: "explicit type wins over version", args: args{typeOverride: "docker", versionOverride: "5.0.0"}, want: EngineTypeDockerCore},
		{name: "configured type wins over version", args: args{versionOverride: "5.0.0"}, configureType: "jvm", want: EngineTypeJvmSingleJar},
		{name: "5.x version override derives native", args: args{versionOverride: "5.0.0"}, want: EngineTypeNative},
		{name: "5.x configured version derives native", configureVersion: "5.2.3", want: EngineTypeNative},
		{name: "4.x version keeps default", args: args{versionOverride: "4.9.0"}, want: defaultEngineType},
		{name: "major alias 5 derives native", args: args{versionOverride: "5"}, want: EngineTypeNative},
		{name: "configured major alias 5 derives native", configureVersion: "5", want: EngineTypeNative},
		{name: "major alias 4 keeps default", args: args{versionOverride: "4"}, want: defaultEngineType},
		{name: "latest keeps default", args: args{versionOverride: "latest"}, want: defaultEngineType},
		{name: "no version, no type, returns default", want: defaultEngineType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.configureType != "" {
				viper.Set("engine", tt.configureType)
			}
			if tt.configureVersion != "" {
				viper.Set("version", tt.configureVersion)
			}
			t.Cleanup(func() {
				viper.Set("engine", nil)
				viper.Set("version", nil)
			})
			if got := GetConfiguredTypeWithVersion(tt.args.typeOverride, tt.args.versionOverride); got != tt.want {
				t.Errorf("GetConfiguredTypeWithVersion() = %v, want %v", got, tt.want)
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
		options    StartOptions
		envOptions EnvOptions
		env        []string
	}
	tests := []struct {
		name         string
		args         args
		wantPrefixes []string
	}{
		{
			name:         "should include home",
			args:         args{options: StartOptions{LogLevel: "WARN"}, envOptions: EnvOptions{IncludeHome: true, IncludePath: true}, env: []string{"HOME=/home/example"}},
			wantPrefixes: []string{"HOME="},
		},
		{
			name:         "should exclude home",
			args:         args{options: StartOptions{LogLevel: "WARN"}, envOptions: EnvOptions{IncludeHome: false, IncludePath: false}, env: []string{"HOME=/home/example"}},
			wantPrefixes: []string{""},
		},
		{
			name:         "should set log level",
			args:         args{options: StartOptions{LogLevel: "WARN"}, envOptions: EnvOptions{IncludeHome: false, IncludePath: false}, env: []string{}},
			wantPrefixes: []string{"IMPOSTER_LOG_LEVEL=WARN"},
		},
		{
			name:         "should pass through imposter env var",
			args:         args{options: StartOptions{LogLevel: "WARN"}, envOptions: EnvOptions{IncludeHome: false, IncludePath: false}, env: []string{"IMPOSTER_TEST=foo"}},
			wantPrefixes: []string{"IMPOSTER_TEST=foo"},
		},
		{
			name:         "should pass through log level env var",
			args:         args{options: StartOptions{LogLevel: "WARN"}, envOptions: EnvOptions{IncludeHome: false, IncludePath: false}, env: []string{"IMPOSTER_LOG_LEVEL=ERROR"}},
			wantPrefixes: []string{"IMPOSTER_LOG_LEVEL=ERROR"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildEnvFromParent(tt.args.env, tt.args.options, tt.args.envOptions)

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
