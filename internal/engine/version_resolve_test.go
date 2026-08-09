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

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

// useTempPrefs points the prefs store at a temporary directory, so version
// cache reads and writes do not touch the user's own prefs file.
func useTempPrefs(t *testing.T) {
	t.Helper()
	viper.Set("prefs.dir", t.TempDir())
	t.Cleanup(func() {
		viper.Set("prefs.dir", nil)
	})
}

// countingLookup returns a lookup function yielding the given version, along
// with a pointer to the number of times it was invoked.
func countingLookup(version string) (func() (string, error), *int) {
	var calls int
	return func() (string, error) {
		calls++
		return version, nil
	}, &calls
}

func TestResolveLatestUsesFreshCache(t *testing.T) {
	useTempPrefs(t)
	storeCached("docker", "4.9.3", time.Now().Unix())

	lookup, calls := countingLookup("9.9.9")
	got, err := resolveLatest("docker", true, lookup)
	if err != nil {
		t.Fatalf("resolveLatest() error = %v", err)
	}
	if got != "4.9.3" {
		t.Errorf("resolveLatest() = %q, want cached %q", got, "4.9.3")
	}
	if *calls != 0 {
		t.Errorf("lookup called %d times, want 0 when a fresh cached value exists", *calls)
	}
}

func TestResolveLatestIgnoresStaleCache(t *testing.T) {
	useTempPrefs(t)
	stale := time.Now().Unix() - checkThresholdSeconds - 1
	storeCached("docker", "4.9.3", stale)

	lookup, calls := countingLookup("4.9.4")
	got, err := resolveLatest("docker", true, lookup)
	if err != nil {
		t.Fatalf("resolveLatest() error = %v", err)
	}
	if got != "4.9.4" {
		t.Errorf("resolveLatest() = %q, want looked up %q", got, "4.9.4")
	}
	if *calls != 1 {
		t.Errorf("lookup called %d times, want 1 when the cached value is stale", *calls)
	}

	// the newly resolved version should have been cached
	if cached := loadCached("docker", time.Now().Unix()); cached != "4.9.4" {
		t.Errorf("cached value = %q, want %q", cached, "4.9.4")
	}
}

func TestResolveLatestSkipsCacheWhenNotAllowed(t *testing.T) {
	useTempPrefs(t)
	storeCached("docker", "4.9.3", time.Now().Unix())

	lookup, calls := countingLookup("4.9.4")
	got, err := resolveLatest("docker", false, lookup)
	if err != nil {
		t.Fatalf("resolveLatest() error = %v", err)
	}
	if got != "4.9.4" {
		t.Errorf("resolveLatest() = %q, want looked up %q", got, "4.9.4")
	}
	if *calls != 1 {
		t.Errorf("lookup called %d times, want 1 when the cache is not allowed", *calls)
	}
}

func TestResolveLatestErrorsWhenLookupFails(t *testing.T) {
	tests := []struct {
		name        string
		allowCached bool
		wantErr     string
	}{
		{name: "cache not allowed", allowCached: false, wantErr: "failed to fetch latest version from API"},
		{name: "cache allowed but empty", allowCached: true, wantErr: "no cached version found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useTempPrefs(t)

			_, err := resolveLatest("docker", tt.allowCached, func() (string, error) {
				return "", fmt.Errorf("simulated API failure")
			})
			if err == nil {
				t.Fatal("resolveLatest() error = nil, want an error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("resolveLatest() error = %q, want it to contain %q", err, tt.wantErr)
			}
		})
	}
}

// TestResolveLatestScopesCache checks that each scope keeps its own cached
// version, so a version alias resolved against one repository cannot return
// the version cached for another.
func TestResolveLatestScopesCache(t *testing.T) {
	useTempPrefs(t)
	now := time.Now().Unix()
	storeCached("imposter-jvm-engine", "4.9.3", now)
	storeCached("imposter-go", "5.21.3", now)
	storeCached("docker", "4.9.1", now)

	tests := []struct {
		scope string
		want  string
	}{
		{scope: "imposter-jvm-engine", want: "4.9.3"},
		{scope: "imposter-go", want: "5.21.3"},
		{scope: "docker", want: "4.9.1"},
		{scope: "native", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.scope, func(t *testing.T) {
			if got := loadCached(tt.scope, now); got != tt.want {
				t.Errorf("loadCached(%q) = %q, want %q", tt.scope, got, tt.want)
			}
		})
	}
}

// TestGetConfiguredVersionResolvesAliases checks alias resolution end to end,
// priming the version cache so no API call is made. The repository a major
// alias resolves against is determined by the alias, not the engine type.
func TestGetConfiguredVersionResolvesAliases(t *testing.T) {
	tests := []struct {
		name       string
		engineType EngineType
		override   string
		want       string
	}{
		{name: "alias 5 resolves against the native repo", engineType: EngineTypeDockerCore, override: "5", want: "5.21.3"},
		{name: "alias 5 resolves against the native repo for lambda", engineType: EngineTypeAwsLambda, override: "5", want: "5.21.3"},
		{name: "alias 4 resolves against the jvm repo", engineType: EngineTypeAwsLambda, override: "4", want: "4.9.3"},
		{name: "latest resolves against the engine type", engineType: EngineTypeDockerCore, override: "latest", want: "4.9.1"},
		{name: "explicit version is not resolved", engineType: EngineTypeDockerCore, override: "4.8.0", want: "4.8.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useTempPrefs(t)
			now := time.Now().Unix()
			storeCached("imposter-jvm-engine", "4.9.3", now)
			storeCached("imposter-go", "5.21.3", now)
			storeCached(string(EngineTypeDockerCore), "4.9.1", now)

			if got := GetConfiguredVersion(tt.engineType, tt.override, true); got != tt.want {
				t.Errorf("GetConfiguredVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckMajorAliasSupported(t *testing.T) {
	tests := []struct {
		name       string
		engineType EngineType
		alias      string
		major      int64
		wantErr    string
	}{
		{name: "jvm engine rejects alias 5", engineType: EngineTypeJvmSingleJar, alias: "5", major: 5, wantErr: "the JVM engine is version 4 and below"},
		{name: "unpacked engine rejects alias 5", engineType: EngineTypeJvmUnpacked, alias: "5", major: 5, wantErr: "the JVM engine is version 4 and below"},
		{name: "native engine rejects alias 4", engineType: EngineTypeNative, alias: "4", major: 4, wantErr: "the native engine is version 5 and above"},
		{name: "jvm engine accepts alias 4", engineType: EngineTypeJvmSingleJar, alias: "4", major: 4},
		{name: "native engine accepts alias 5", engineType: EngineTypeNative, alias: "5", major: 5},
		// docker images and lambda are built from both engine lines
		{name: "docker engine accepts alias 4", engineType: EngineTypeDockerCore, alias: "4", major: 4},
		{name: "docker engine accepts alias 5", engineType: EngineTypeDockerCore, alias: "5", major: 5},
		{name: "lambda accepts alias 4", engineType: EngineTypeAwsLambda, alias: "4", major: 4},
		{name: "lambda accepts alias 5", engineType: EngineTypeAwsLambda, alias: "5", major: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckMajorAliasSupported(tt.engineType, tt.alias, tt.major)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("CheckMajorAliasSupported() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckMajorAliasSupported() error = nil, want it to contain %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("CheckMajorAliasSupported() error = %q, want it to contain %q", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("alias '%s'", tt.alias)) {
				t.Errorf("CheckMajorAliasSupported() error = %q, want it to name the alias", err)
			}
		})
	}
}

// TestGetConfiguredVersionRejectsUnsupportedAlias checks that an alias naming
// the wrong engine line fails before any attempt to fetch that version.
func TestGetConfiguredVersionRejectsUnsupportedAlias(t *testing.T) {
	useTempPrefs(t)

	oldExitFunc := logger.ExitFunc
	logger.ExitFunc = func(code int) {
		panic(fmt.Sprintf("fatal exit with code %d", code))
	}
	t.Cleanup(func() {
		logger.ExitFunc = oldExitFunc
	})

	defer func() {
		if r := recover(); r == nil {
			t.Error("GetConfiguredVersion() returned normally, want a fatal exit")
		}
	}()

	// resolving would need the API, so a fatal exit here also proves the
	// check happens before the version is looked up
	GetConfiguredVersion(EngineTypeJvmSingleJar, "5", true)
}

func TestGetRepoNameForMajor(t *testing.T) {
	tests := []struct {
		major int64
		want  string
	}{
		{major: 4, want: "imposter-jvm-engine"},
		{major: 5, want: "imposter-go"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("v%d", tt.major), func(t *testing.T) {
			if got := getRepoNameForMajor(tt.major); got != tt.want {
				t.Errorf("getRepoNameForMajor(%d) = %q, want %q", tt.major, got, tt.want)
			}
		})
	}
}

func TestGetRepoNameForEngineType(t *testing.T) {
	tests := []struct {
		engineType EngineType
		want       string
	}{
		{engineType: EngineTypeNative, want: "imposter-go"},
		{engineType: EngineTypeDockerCore, want: "imposter-jvm-engine"},
		{engineType: EngineTypeJvmSingleJar, want: "imposter-jvm-engine"},
		{engineType: EngineTypeAwsLambda, want: "imposter-jvm-engine"},
	}
	for _, tt := range tests {
		t.Run(string(tt.engineType), func(t *testing.T) {
			if got := getRepoNameForEngineType(tt.engineType); got != tt.want {
				t.Errorf("getRepoNameForEngineType(%q) = %q, want %q", tt.engineType, got, tt.want)
			}
		})
	}
}

func TestFetchLatestFromApi(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		want       string
		wantErr    string
	}{
		{name: "tag name with v prefix", statusCode: http.StatusOK, body: `{"tag_name":"v5.21.3"}`, want: "5.21.3"},
		{name: "tag name without v prefix", statusCode: http.StatusOK, body: `{"tag_name":"5.21.3"}`, want: "5.21.3"},
		{name: "error status code", statusCode: http.StatusNotFound, body: `{}`, wantErr: "status code: 404"},
		{name: "unparseable body", statusCode: http.StatusOK, body: `not json`, wantErr: "cannot unmarshall response body"},
		{name: "missing tag name", statusCode: http.StatusOK, body: `{}`, wantErr: "no tag name in response body"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			got, err := fetchLatestFromApi(server.URL)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("fetchLatestFromApi() error = nil, want it to contain %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("fetchLatestFromApi() error = %q, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("fetchLatestFromApi() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("fetchLatestFromApi() = %q, want %q", got, tt.want)
			}
		})
	}
}
