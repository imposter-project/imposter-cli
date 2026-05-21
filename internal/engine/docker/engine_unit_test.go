/*
Copyright © 2021 Pete Cornish <outofcoffee@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package docker

import (
	"reflect"
	"testing"

	"github.com/imposter-project/imposter-cli/internal/engine"
	"github.com/imposter-project/imposter-cli/internal/stringutil"
)

func TestBuildCmd(t *testing.T) {
	tests := []struct {
		name         string
		options      engine.StartOptions
		useEnvConfig bool
		want         []string
	}{
		{
			name:         "legacy 4.x emits configDir and listenPort flags",
			options:      engine.StartOptions{Port: 8080},
			useEnvConfig: false,
			want: []string{
				"--configDir=/opt/imposter/config",
				"--listenPort=8080",
			},
		},
		{
			name:         "5.x+ omits CLI flags",
			options:      engine.StartOptions{Port: 8080},
			useEnvConfig: true,
			want:         nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCmd(tt.options, tt.useEnvConfig)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildEnv(t *testing.T) {
	t.Run("legacy 4.x omits IMPOSTER_CONFIG_DIR and IMPOSTER_PORT", func(t *testing.T) {
		env := buildEnv(engine.StartOptions{Port: 8080, LogLevel: "DEBUG"}, false)
		if stringutil.ContainsPrefix(env, "IMPOSTER_CONFIG_DIR=") {
			t.Errorf("expected no IMPOSTER_CONFIG_DIR, got env: %v", env)
		}
		if stringutil.ContainsPrefix(env, "IMPOSTER_PORT=") {
			t.Errorf("expected no IMPOSTER_PORT, got env: %v", env)
		}
	})

	t.Run("5.x+ includes IMPOSTER_CONFIG_DIR and IMPOSTER_PORT", func(t *testing.T) {
		env := buildEnv(engine.StartOptions{Port: 9090, LogLevel: "DEBUG"}, true)
		if !stringutil.Contains(env, "IMPOSTER_CONFIG_DIR=/opt/imposter/config") {
			t.Errorf("expected IMPOSTER_CONFIG_DIR=/opt/imposter/config, got env: %v", env)
		}
		if !stringutil.Contains(env, "IMPOSTER_PORT=9090") {
			t.Errorf("expected IMPOSTER_PORT=9090, got env: %v", env)
		}
	})

	t.Run("file cache adds cache env vars regardless of version", func(t *testing.T) {
		env := buildEnv(engine.StartOptions{Port: 8080, LogLevel: "DEBUG", EnableFileCache: true}, true)
		if !stringutil.Contains(env, "IMPOSTER_CACHE_DIR=/tmp/imposter-cache") {
			t.Errorf("expected IMPOSTER_CACHE_DIR=/tmp/imposter-cache, got env: %v", env)
		}
		if !stringutil.Contains(env, "IMPOSTER_OPENAPI_REMOTE_FILE_CACHE=true") {
			t.Errorf("expected IMPOSTER_OPENAPI_REMOTE_FILE_CACHE=true, got env: %v", env)
		}
	})
}
