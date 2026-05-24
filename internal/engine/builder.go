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
	"github.com/imposter-project/imposter-cli/internal/logging"
	"github.com/imposter-project/imposter-cli/internal/stringutil"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type EngineType string

const (
	EngineTypeNone             EngineType = ""
	EngineTypeAwsLambda        EngineType = "awslambda"
	EngineTypeDockerCore       EngineType = "docker"
	EngineTypeDockerAll        EngineType = "docker-all"
	EngineTypeDockerDistroless EngineType = "docker-distroless"
	EngineTypeJvmSingleJar     EngineType = "jvm"
	EngineTypeJvmUnpacked      EngineType = "unpacked"
	EngineTypeNative           EngineType = "native"
)
const defaultEngineType = EngineTypeDockerCore

// engineTypeAliases maps deprecated user-facing engine type names to their canonical form.
// Aliases are accepted on user input (CLI flag, env var, config file) but not surfaced in
// help text or shell completions.
var engineTypeAliases = map[EngineType]EngineType{
	"golang": EngineTypeNative,
}

func normaliseEngineType(t EngineType) EngineType {
	if canon, ok := engineTypeAliases[t]; ok {
		return canon
	}
	return t
}

var logger = logging.GetLogger()

var (
	libraries = make(map[EngineType]func() EngineLibrary)
	engines   = make(map[EngineType]func(configDir string, startOptions StartOptions) MockEngine)
)

func RegisterLibrary(engineType EngineType, b func() EngineLibrary) {
	libraries[engineType] = b
}

func RegisterEngine(engineType EngineType, b func(configDir string, startOptions StartOptions) MockEngine) {
	engines[engineType] = b
}

func EnumerateLibraries() []EngineType {
	var all []EngineType
	for key := range libraries {
		all = append(all, key)
	}
	return all
}

func GetLibrary(engineType EngineType) EngineLibrary {
	if err := validateEngineType(engineType); err != nil {
		logger.Fatal(err)
	}
	library := libraries[engineType]
	if library == nil {
		logger.Fatalf("unregistered engine type: %v", engineType)
	}
	logger.Tracef("using %s library", engineType)
	return library()
}

// BuildEngine is a convenience function that gets the library for the given engine type,
// obtains a provider for the version specified in the start options, then invokes
// the provider's builder function.
//
// Note that the provider's Provide() function is not invoked explicitly, although it may
// be invoked implicitly from the builder function.
func BuildEngine(engineType EngineType, configDir string, startOptions StartOptions) MockEngine {
	lib := GetLibrary(engineType)
	provider := lib.GetProvider(startOptions.Version)
	return provider.Build(configDir, startOptions)
}

// build validates the engine type against those supported, then invokes the
// associated engine builder function.
func build(engineType EngineType, configDir string, startOptions StartOptions) MockEngine {
	if err := validateEngineType(engineType); err != nil {
		logger.Fatal(err)
	}
	eng := engines[engineType]
	if eng == nil {
		logger.Fatalf("unregistered engine type: %v", engineType)
	}
	logger.Tracef("using %s engine", engineType)
	return eng(configDir, startOptions)
}

func validateEngineType(engineType EngineType) error {
	switch engineType {
	case EngineTypeAwsLambda, EngineTypeDockerCore, EngineTypeDockerAll, EngineTypeDockerDistroless, EngineTypeJvmSingleJar, EngineTypeJvmUnpacked, EngineTypeNative:
		return nil
	}
	return fmt.Errorf("unsupported engine type: %v", engineType)
}

// IsDockerEngine reports whether the engine type is one of the
// container-based docker variants.
func IsDockerEngine(engineType EngineType) bool {
	switch engineType {
	case EngineTypeDockerCore, EngineTypeDockerAll, EngineTypeDockerDistroless:
		return true
	}
	return false
}

func GetConfiguredType(override string) EngineType {
	return GetConfiguredTypeWithDefault(override, defaultEngineType)
}

// GetConfiguredTypeWithVersion is like GetConfiguredType but also takes an
// engine-version override so that a pinned version can imply an engine type
// when no explicit one is configured. CLI commands that expose a --version
// flag should pass its value as versionOverride.
func GetConfiguredTypeWithVersion(typeOverride string, versionOverride string) EngineType {
	return getConfiguredType(typeOverride, versionOverride, defaultEngineType)
}

func GetConfiguredTypeWithDefault(override string, defaultType EngineType) EngineType {
	return getConfiguredType(override, "", defaultType)
}

func getConfiguredType(typeOverride string, versionOverride string, defaultType EngineType) EngineType {
	explicit := stringutil.GetFirstNonEmpty(
		typeOverride,
		viper.GetString("engine"),
	)
	if explicit != "" {
		return normaliseEngineType(EngineType(explicit))
	}
	// No explicit engine type configured. If the user has pinned a specific
	// engine version we can sometimes derive the engine type from it (e.g.
	// 5.x implies the native engine). "latest" intentionally does not derive
	// — callers keep the supplied default until "latest" is re-pointed at v5.
	version := stringutil.GetFirstNonEmpty(versionOverride, viper.GetString("version"))
	if derived := DeriveEngineTypeFromVersion(version); derived != EngineTypeNone {
		return derived
	}
	return defaultType
}

func GetConfiguredVersion(engineType EngineType, override string, allowCached bool) string {
	return GetConfiguredVersionOrResolve(engineType, override, allowCached, true)
}

func GetConfiguredVersionOrResolve(engineType EngineType, override string, allowCached bool, resolveIfLatest bool) string {
	version := stringutil.GetFirstNonEmpty(
		override,
		viper.GetString("version"),
		"latest",
	)
	if version == "latest" && resolveIfLatest {
		latest, err := ResolveLatestToVersion(engineType, allowCached)
		if err != nil {
			panic(err)
		}
		version = latest
	}
	return version
}

func SanitiseVersionOutput(s string) string {
	var remove = []string{
		"Version:",
		"WARNING: sun.reflect.Reflection.getCallerClass is not supported. This will impact performance.",
	}
	for _, r := range remove {
		s = strings.ReplaceAll(s, r, "")
	}
	return strings.TrimSpace(s)
}

func BuildEnv(options StartOptions, envOptions EnvOptions) []string {
	env := buildEnvFromParent(os.Environ(), options, envOptions)
	if options.DebugMode {
		env = append(env, fmt.Sprintf("JAVA_TOOL_OPTIONS=-agentlib:jdwp=transport=dt_socket,server=y,suspend=n,address=0.0.0.0:%v", DefaultDebugPort))
	}
	return env
}

func buildEnvFromParent(parentEnv []string, options StartOptions, envOptions EnvOptions) []string {
	env := options.Environment

	for _, e := range parentEnv {
		if strings.HasPrefix(e, "IMPOSTER_") ||
			strings.HasPrefix(e, "JAVA_TOOL_OPTIONS=") ||
			(envOptions.IncludeHome && strings.HasPrefix(e, "HOME=")) ||
			(envOptions.IncludePath && strings.HasPrefix(e, "PATH=")) {

			// explicit environment takes precedence over parent
			key := strings.Split(e, "=")[0]
			if !stringutil.ContainsPrefix(env, key+"=") {
				env = append(env, e)
			}
		}
	}

	if !stringutil.ContainsPrefix(env, "IMPOSTER_LOG_LEVEL=") {
		env = append(env, "IMPOSTER_LOG_LEVEL="+strings.ToUpper(options.LogLevel))
	}

	return env
}

func (e *EngineMetadata) Build(configDir string, startOptions StartOptions) MockEngine {
	return build(e.EngineType, configDir, startOptions)
}
