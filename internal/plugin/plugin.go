package plugin

import (
	"fmt"
	"gatehill.io/imposter/internal/engine"
	"gatehill.io/imposter/internal/logging"
	"gatehill.io/imposter/internal/stringutil"
	"github.com/spf13/viper"
	"os"
	"strings"
)

type PluginMetadata struct {
	Name       string
	EngineType engine.EngineType
	Version    string
}

var logger = logging.GetLogger()

func EnsurePlugins(plugins []string, engineType engine.EngineType, version string, saveDefault bool) (int, error) {
	logger.Tracef("ensuring %d plugins: %v", len(plugins), plugins)
	if len(plugins) == 0 {
		return 0, nil
	}
	for _, plugin := range plugins {
		err := EnsurePlugin(plugin, engineType, version)
		if err != nil {
			return 0, fmt.Errorf("error ensuring plugin %s: %s", plugin, err)
		}
		logger.Debugf("plugin %s version %s is installed", plugin, version)
	}
	if saveDefault {
		err := addDefaultPlugins(plugins)
		if err != nil {
			logger.Warnf("error setting plugins as default: %s", err)
		}
	}
	return len(plugins), nil
}

// EnsureConfiguredPlugins collects the plugins from both the global CLI
// config, as well those within the current configuration context, such
// as config files within the working directory
func EnsureConfiguredPlugins(engineType engine.EngineType, version string) (int, error) {
	// this includes the config from the current configuration context,
	// not just the global CLI config file, so it includes any
	// configuration in the working directory
	plugins := viper.GetStringSlice(defaultPluginsConfigKey)

	for _, plugin := range plugins {
		// work-around for https://github.com/spf13/viper/issues/380
		if strings.Contains(plugin, ",") {
			for _, p := range strings.Split(plugin, ",") {
				plugins = append(plugins, p)
			}
		} else {
			plugins = append(plugins, plugin)
		}
	}
	plugins = stringutil.Unique(plugins)

	logger.Tracef("found %d configured plugin(s): %v", len(plugins), plugins)
	return EnsurePlugins(plugins, engineType, version, false)
}

func EnsurePlugin(pluginName string, engineType engine.EngineType, version string) error {
	_, localFilePath, err := GetPluginLocalPath(pluginName, engineType, version)
	if err != nil {
		return err
	}
	if _, err := os.Stat(localFilePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("unable to stat plugin file: %s: %s", localFilePath, err)
		}
	} else {
		logger.Tracef("plugin %s version %s already exists at: %s", pluginName, version, localFilePath)
		return nil
	}
	logger.Debugf("plugin %s version %s is not installed", pluginName, version)
	err = downloadPlugin(engineType, pluginName, version)
	if err != nil {
		return err
	}
	return nil
}

// UninstallPlugins removes the specified plugins from disk and optionally
// from the default plugins configuration.
func UninstallPlugins(plugins []string, engineType engine.EngineType, version string, removeDefault bool) (int, error) {
	logger.Tracef("uninstalling %d plugins: %v", len(plugins), plugins)
	if len(plugins) == 0 {
		return 0, nil
	}

	var removed int
	for _, plugin := range plugins {
		wasInstalled, err := UninstallPlugin(plugin, engineType, version)
		if err != nil {
			return removed, fmt.Errorf("error uninstalling plugin %s: %s", plugin, err)
		}
		if wasInstalled {
			logger.Debugf("plugin %s version %s is uninstalled", plugin, version)
			removed++
		} else {
			logger.Debugf("plugin %s version %s was not installed", plugin, version)
		}
	}

	if removeDefault {
		err := removeDefaultPlugins(plugins)
		if err != nil {
			logger.Warnf("error removing plugins from default list: %s", err)
		}
	}

	return removed, nil
}

// UninstallPlugin removes a single plugin from disk.
// Returns true if the plugin was installed and removed, false if it wasn't installed.
func UninstallPlugin(pluginName string, engineType engine.EngineType, version string) (bool, error) {
	_, localFilePath, err := GetPluginLocalPath(pluginName, engineType, version)
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(localFilePath); err != nil {
		if os.IsNotExist(err) {
			logger.Debugf("plugin %s version %s is not installed at: %s", pluginName, version, localFilePath)
			return false, nil
		}
		return false, fmt.Errorf("unable to stat plugin file: %s: %s", localFilePath, err)
	}

	err = os.Remove(localFilePath)
	if err != nil {
		return false, fmt.Errorf("failed to remove plugin file %s: %s", localFilePath, err)
	}

	logger.Infof("removed plugin %s version %s from: %s", pluginName, version, localFilePath)
	return true, nil
}
