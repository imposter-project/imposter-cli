package golang

import (
	"fmt"
	"gatehill.io/imposter/internal/library"
	"os"
	"path/filepath"

	"gatehill.io/imposter/internal/engine"
)

const binCacheDir = ".imposter/engines/golang"

// Library implements the engine.EngineLibrary interface for the golang engine
type Library struct {
	binCache string
}

// NewLibrary creates a new instance of the golang engine library
func NewLibrary() *Library {
	return &Library{}
}

func (l *Library) CheckPrereqs() (bool, []string) {
	return true, nil
}

func (l *Library) List() ([]engine.EngineMetadata, error) {
	binCachePath, err := l.ensureBinCache()
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(binCachePath)
	if err != nil {
		return nil, fmt.Errorf("error reading binary cache directory: %v: %v", binCachePath, err)
	}
	var available []engine.EngineMetadata
	for _, file := range files {
		if file.IsDir() {
			available = append(available, engine.EngineMetadata{
				EngineType: engine.EngineTypeGolang,
				Version:    file.Name(),
			})
		}
	}
	return available, nil
}

func (l *Library) GetProvider(version string) engine.Provider {
	binCachePath, err := l.ensureBinCache()
	if err != nil {
		providerLogger.Fatal(err)
	}
	versionedBinDir := filepath.Join(binCachePath, version)
	return NewProvider(version, versionedBinDir)
}

func (l *Library) IsSealedDistro() bool {
	return false
}

func (l *Library) ShouldEnsurePlugins() bool {
	return true
}

func (l *Library) ensureBinCache() (string, error) {
	return library.EnsureDirUsingConfig("golang.binCache", binCacheDir)
}
