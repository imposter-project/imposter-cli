package plugin

import (
	"fmt"
	"gatehill.io/imposter/internal/config"
	"gatehill.io/imposter/internal/engine"
	library2 "gatehill.io/imposter/internal/library"
	"gatehill.io/imposter/internal/logging"
	"gatehill.io/imposter/internal/stringutil"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

type PluginMetadata struct {
	Name       string
	EngineType engine.EngineType
	Version    string
}

const pluginBaseDir = ".imposter/plugins/"
const defaultPluginsConfigKey = "default.plugins"

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
	_, pluginFilePath, err := GetPluginFilePath(pluginName, engineType, version)
	if err != nil {
		return err
	}
	if _, err := os.Stat(pluginFilePath); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("unable to stat plugin file: %s: %s", pluginFilePath, err)
		}
	} else {
		logger.Tracef("plugin %s version %s already exists at: %s", pluginName, version, pluginFilePath)
		return nil
	}
	logger.Debugf("plugin %s version %s is not installed", pluginName, version)
	err = downloadPlugin(engineType, pluginName, version)
	if err != nil {
		return err
	}
	return nil
}

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

func downloadPlugin(engineType engine.EngineType, pluginName string, version string) error {
	fullPluginFileName, pluginFilePath, err := GetPluginFilePath(pluginName, engineType, version)
	if err != nil {
		return err
	}

	pluginConfig := determinePluginConfig(engineType)
	downloadConfig := pluginConfig.downloadConfig
	err = library2.DownloadBinary(downloadConfig, pluginFilePath, fullPluginFileName, version)
	if err != nil {
		return err
	}

	logger.Infof("downloaded plugin %s version %s", pluginName, version)
	return nil
}

// GetPluginFilePath returns the full plugin file name and the
// plugin file path for the specified plugin name, engine type, and version.
func GetPluginFilePath(
	pluginName string,
	engineType engine.EngineType,
	version string,
) (fullPluginFileName string, pluginFilePath string, err error) {
	pluginDir, err := EnsurePluginDir(version)
	if err != nil {
		return "", "", err
	}

	fullPluginFileName, err = getFullPluginFileName(engineType, pluginName)
	if err != nil {
		return "", "", fmt.Errorf("error determining plugin file extension for %s: %s", engineType, err)
	}

	pluginFilePath = filepath.Join(pluginDir, fullPluginFileName)
	return fullPluginFileName, pluginFilePath, err
}

// addDefaultPlugins adds the provided plugins to the list of default
// plugins, if they are not already present, and writes the
// configuration file.
func addDefaultPlugins(plugins []string) error {
	existing, err := ListDefaultPlugins()
	if err != nil {
		return fmt.Errorf("failed to lead default plugins: %s", err)
	}
	combined := stringutil.CombineUnique(existing, plugins)
	if len(existing) == len(combined) {
		// none added
		return nil
	}
	return writeDefaultPlugins(combined)
}

func ListDefaultPlugins() ([]string, error) {
	v, err := parseConfigFile()
	if err != nil {
		return []string{}, err
	} else {
		return v.GetStringSlice(defaultPluginsConfigKey), nil
	}
}

func writeDefaultPlugins(plugins []string) error {
	v, err := parseConfigFile()
	if err != nil {
		return err
	}
	v.Set(defaultPluginsConfigKey, plugins)

	configDir, err := config.GetGlobalConfigDir()
	if err != nil {
		return err
	}
	configFilePath := filepath.Join(configDir, config.GlobalConfigFileName+".yaml")
	err = v.WriteConfigAs(configFilePath)
	if err != nil {
		return fmt.Errorf("error writing default plugin configuration to: %s: %s", configFilePath, err)
	}

	logger.Tracef("wrote default plugin configuration to: %s", configFilePath)
	return nil
}

func parseConfigFile() (*viper.Viper, error) {
	v := viper.New()
	configDir, err := config.GetGlobalConfigDir()
	if err != nil {
		return nil, err
	}
	v.AddConfigPath(configDir)
	v.SetConfigName(config.GlobalConfigFileName)

	// sink if does not exist
	_ = v.ReadInConfig()
	return v, nil
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
		err := UninstallPlugin(plugin, engineType, version)
		if err != nil {
			return removed, fmt.Errorf("error uninstalling plugin %s: %s", plugin, err)
		}
		logger.Debugf("plugin %s version %s is uninstalled", plugin, version)
		removed++
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
func UninstallPlugin(pluginName string, engineType engine.EngineType, version string) error {
	_, pluginFilePath, err := GetPluginFilePath(pluginName, engineType, version)
	if err != nil {
		return err
	}

	if _, err := os.Stat(pluginFilePath); err != nil {
		if os.IsNotExist(err) {
			logger.Debugf("plugin %s version %s is not installed at: %s", pluginName, version, pluginFilePath)
			return fmt.Errorf("plugin %s version %s is not installed", pluginName, version)
		}
		return fmt.Errorf("unable to stat plugin file: %s: %s", pluginFilePath, err)
	}

	err = os.Remove(pluginFilePath)
	if err != nil {
		return fmt.Errorf("failed to remove plugin file %s: %s", pluginFilePath, err)
	}

	logger.Infof("removed plugin %s version %s from: %s", pluginName, version, pluginFilePath)
	return nil
}

// removeDefaultPlugins removes the specified plugins from the default
// plugins configuration and writes the updated configuration file.
func removeDefaultPlugins(plugins []string) error {
	existing, err := ListDefaultPlugins()
	if err != nil {
		return fmt.Errorf("failed to load default plugins: %s", err)
	}

	var updated []string
	for _, plugin := range existing {
		if !stringutil.Contains(plugins, plugin) {
			updated = append(updated, plugin)
		}
	}

	if len(existing) == len(updated) {
		// none removed
		return nil
	}

	return writeDefaultPlugins(updated)
}
