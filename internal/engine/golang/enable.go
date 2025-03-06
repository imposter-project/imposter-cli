package golang

import (
	"path/filepath"

	"gatehill.io/imposter/internal/engine"
)

var golangInitialised = false

// EnableEngine registers the Golang engine implementation
func EnableEngine() {
	if !golangInitialised {
		golangInitialised = true

		engine.RegisterLibrary(engine.EngineTypeGolang, func() engine.EngineLibrary {
			return NewLibrary()
		})
		engine.RegisterEngine(engine.EngineTypeGolang, func(configDir string, startOptions engine.StartOptions) engine.MockEngine {
			lib := NewLibrary()
			binCachePath, err := lib.ensureBinCache()
			if err != nil {
				providerLogger.Fatal(err)
			}
			versionedBinDir := filepath.Join(binCachePath, startOptions.Version)
			provider := NewProvider(startOptions.Version, versionedBinDir)
			return NewGolangMockEngine(configDir, startOptions, provider)
		})
	}
}
