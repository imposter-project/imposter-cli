package plugin

import (
	library2 "gatehill.io/imposter/internal/library"
	"github.com/spf13/viper"
	"path/filepath"
)

const pluginBaseDir = ".imposter/plugins/"

func EnsurePluginDir(version string) (string, error) {
	fullPluginDir, err := getFullPluginDir(version)
	if err != nil {
		return "", err
	}
	err = library2.EnsureDir(fullPluginDir)
	if err != nil {
		return "", err
	}
	logger.Tracef("ensured plugin directory: %v", fullPluginDir)
	return fullPluginDir, nil
}

func getFullPluginDir(version string) (dir string, err error) {
	// use IMPOSTER_PLUGIN_DIR directly, if set
	fullPluginDir := viper.GetString("plugin.dir")
	if fullPluginDir == "" {
		basePluginDir, err := getBasePluginDir()
		if err != nil {
			return "", err
		}
		fullPluginDir = filepath.Join(basePluginDir, version)
	}
	return fullPluginDir, nil
}

func getBasePluginDir() (string, error) {
	return library2.EnsureDirUsingConfig("plugin.baseDir", pluginBaseDir)
}
