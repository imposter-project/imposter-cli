package native

import (
	"path/filepath"

	"github.com/imposter-project/imposter-cli/internal/engine"
	"github.com/spf13/viper"
)

var nativeInitialised = false

// EnableEngine registers the native engine implementation
func EnableEngine() {
	if !nativeInitialised {
		nativeInitialised = true

		// Accept the legacy "golang.binCache" config key as an alias for "native.binCache"
		// so existing config files and IMPOSTER_GOLANG_BINCACHE env vars keep working.
		viper.RegisterAlias("golang.binCache", "native.binCache")

		engine.RegisterLibrary(engine.EngineTypeNative, func() engine.EngineLibrary {
			return NewLibrary()
		})
		engine.RegisterEngine(engine.EngineTypeNative, func(configDir string, startOptions engine.StartOptions) engine.MockEngine {
			lib := NewLibrary()
			binCachePath, err := lib.ensureBinCache()
			if err != nil {
				providerLogger.Fatal(err)
			}
			versionedBinDir := filepath.Join(binCachePath, startOptions.Version)
			provider := NewProvider(startOptions.Version, versionedBinDir)
			return NewNativeMockEngine(configDir, startOptions, provider)
		})
	}
}
