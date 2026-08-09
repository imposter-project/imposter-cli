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
		{name: "major alias 4 uses CLI flags", version: "4", want: false},
		{name: "major alias 5 uses env vars", version: "5", want: true},
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
		{name: "major alias 3 has no derivation", version: "3", want: EngineTypeNone},
		{name: "major alias 4 has no derivation", version: "4", want: EngineTypeNone},
		{name: "major alias 5 derives native", version: "5", want: EngineTypeNative},
		{name: "major alias 6 derives native", version: "6", want: EngineTypeNative},
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

func TestParseMajorAlias(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    int64
		wantOk  bool
	}{
		{name: "bare major", version: "4", want: 4, wantOk: true},
		{name: "bare major 5", version: "5", want: 5, wantOk: true},
		{name: "multi-digit major", version: "10", want: 10, wantOk: true},
		{name: "full version is not an alias", version: "4.9.3", wantOk: false},
		{name: "partial version is not an alias", version: "4.9", wantOk: false},
		{name: "latest is not a major alias", version: "latest", wantOk: false},
		{name: "empty is not an alias", version: "", wantOk: false},
		{name: "prefixed major is not an alias", version: "v5", wantOk: false},
		{name: "negative is not an alias", version: "-5", wantOk: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := ParseMajorAlias(tt.version)
			if gotOk != tt.wantOk || (gotOk && got != tt.want) {
				t.Errorf("ParseMajorAlias(%q) = (%v, %v), want (%v, %v)", tt.version, got, gotOk, tt.want, tt.wantOk)
			}
		})
	}
}

func TestSelectHighestForMajor(t *testing.T) {
	releases := []release{
		{TagName: "v5.21.3"},
		{TagName: "v5.22.0", Prerelease: true},
		{TagName: "v5.22.1", Draft: true},
		{TagName: "not-a-version"},
		{TagName: "v5.21.10"},
		{TagName: "v4.9.3"},
	}

	tests := []struct {
		name          string
		releases      []release
		major         int64
		want          string
		wantOlderSeen bool
	}{
		{name: "highest of major", releases: releases, major: 5, want: "5.21.10", wantOlderSeen: true},
		{name: "older major", releases: releases, major: 4, want: "4.9.3", wantOlderSeen: false},
		{name: "no match", releases: releases, major: 6, want: "", wantOlderSeen: true},
		{name: "empty", releases: nil, major: 5, want: "", wantOlderSeen: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			highest, olderSeen := selectHighestForMajor(tt.releases, tt.major)
			var got string
			if highest != nil {
				got = highest.String()
			}
			if got != tt.want {
				t.Errorf("selectHighestForMajor() version = %q, want %q", got, tt.want)
			}
			if olderSeen != tt.wantOlderSeen {
				t.Errorf("selectHighestForMajor() sawOlderMajor = %v, want %v", olderSeen, tt.wantOlderSeen)
			}
		})
	}
}

func TestGetHighestVersionForMajor(t *testing.T) {
	engines := []EngineMetadata{
		{Version: "4.9.3"},
		{Version: "5.21.3"},
		{Version: "5.21.10"},
		{Version: "dev"},
	}

	tests := []struct {
		name  string
		major int64
		want  string
	}{
		{name: "highest 5.x", major: 5, want: "5.21.10"},
		{name: "highest 4.x", major: 4, want: "4.9.3"},
		{name: "none installed for major", major: 6, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetHighestVersionForMajor(engines, tt.major); got != tt.want {
				t.Errorf("GetHighestVersionForMajor(major=%d) = %q, want %q", tt.major, got, tt.want)
			}
		})
	}
}
