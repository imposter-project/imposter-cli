package plugin

import (
	"fmt"
	"gatehill.io/imposter/internal/engine"
	"os"
)

// List returns a list of available plugins for the specified engine type and version.
func List(engineType engine.EngineType, version string) ([]PluginMetadata, error) {
	pluginDir, err := getFullPluginDir(version)
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		return nil, fmt.Errorf("error reading plugin directory: %v: %v", pluginDir, err)
	}

	var available []PluginMetadata
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		validPlugin, pluginName := isValidPluginFile(file.Name(), engineType)
		if !validPlugin {
			logger.Tracef("ignoring file: %s, not a valid plugin file", file.Name())
			continue
		}
		available = append(available, PluginMetadata{
			Name:       pluginName,
			EngineType: engineType,
			Version:    version,
		})
	}
	return available, nil
}

// ListVersionDirs returns the names of the versioned directories under
// the plugin base dir. This is only the list of versions, not fully qualified
// paths.
func ListVersionDirs() ([]string, error) {
	basePluginDir, err := getBasePluginDir()
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(basePluginDir)
	if err != nil {
		return nil, fmt.Errorf("error reading plugin base directory: %v: %v", basePluginDir, err)
	}
	var dirs []string
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		dirs = append(dirs, file.Name())
	}
	return dirs, nil
}
