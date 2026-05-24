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

package engine

import "testing"

func TestUsesEnvConfig(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{name: "3.x uses CLI flags", version: "3.40.0", want: false},
		{name: "4.x uses CLI flags", version: "4.9.1", want: false},
		{name: "4.x latest patch uses CLI flags", version: "4.99.99", want: false},
		{name: "5.0.0 uses env vars", version: "5.0.0", want: true},
		{name: "5.x uses env vars", version: "5.2.3", want: true},
		{name: "6.x uses env vars", version: "6.0.0", want: true},
		{name: "5.x pre-release uses env vars", version: "5.0.0-beta.1", want: true},
		{name: "unparseable falls back to env vars", version: "dev", want: true},
		{name: "empty version falls back to env vars", version: "", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UsesEnvConfig(tt.version); got != tt.want {
				t.Errorf("UsesEnvConfig(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestDeriveEngineTypeFromVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    EngineType
	}{
		{name: "3.x has no derivation", version: "3.40.0", want: EngineTypeNone},
		{name: "4.x has no derivation", version: "4.9.1", want: EngineTypeNone},
		{name: "5.0.0 derives native", version: "5.0.0", want: EngineTypeNative},
		{name: "5.x derives native", version: "5.2.3", want: EngineTypeNative},
		{name: "6.x derives native", version: "6.0.0", want: EngineTypeNative},
		{name: "5.x pre-release derives native", version: "5.0.0-beta.1", want: EngineTypeNative},
		{name: "latest keeps default", version: "latest", want: EngineTypeNone},
		{name: "empty keeps default", version: "", want: EngineTypeNone},
		{name: "unparseable keeps default", version: "dev", want: EngineTypeNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeriveEngineTypeFromVersion(tt.version); got != tt.want {
				t.Errorf("DeriveEngineTypeFromVersion(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}
