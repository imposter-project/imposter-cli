package plugin

import (
	"fmt"
	"gatehill.io/imposter/internal/config"
	library2 "gatehill.io/imposter/internal/library"
	"gatehill.io/imposter/internal/logging"
	"gatehill.io/imposter/internal/stringutil"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
)

type PluginMetadata struct {
	Name    string
	Version string
}

const pluginBaseDir = ".imposter/plugins/"
const defaultPluginsConfigKey = "default.plugins"
const pluginFileNamePrefix = "imposter-plugin-"

var supportedPluginExtensions = []string{".jar", ".zip"}

var logger = logging.GetLogger()

func EnsurePlugins(plugins []string, version string, saveDefault bool) (int, error) {
	logger.Tracef("ensuring %d plugins: %v", len(plugins), plugins)
	if len(plugins) == 0 {
		return 0, nil
	}
	for _, plugin := range plugins {
		err := EnsurePlugin(plugin, version)
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
func EnsureConfiguredPlugins(version string) (int, error) {
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
	return EnsurePlugins(plugins, version, false)
}

func EnsurePlugin(pluginName string, version string) error {
	_, pluginFilePath, err := getPluginFilePath(pluginName, version)
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
	err = downloadPlugin(pluginName, version)
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

func downloadPlugin(pluginName string, version string) error {
	fullPluginFileName, pluginFilePath, err := getPluginFilePath(pluginName, version)
	if err != nil {
		return err
	}
	err = library2.DownloadBinary(pluginFilePath, fullPluginFileName, version)
	if err != nil {
		return err
	}
	logger.Infof("downloaded plugin %s version %s", pluginName, version)
	return nil
}

func getPluginFilePath(pluginName string, version string) (fullPluginFileName string, pluginFilePath string, err error) {
	pluginDir, err := EnsurePluginDir(version)
	if err != nil {
		return "", "", err
	}

	// archive format plugins use .zip extension
	// supported since engine v3.35.0
	var pluginExtension string
	if strings.HasSuffix(pluginName, ":zip") {
		pluginName = strings.TrimSuffix(pluginName, ":zip")
		pluginExtension = "zip"
	} else {
		pluginExtension = "jar"
	}

	fullPluginFileName = fmt.Sprintf("%s%s.%s", pluginFileNamePrefix, pluginName, pluginExtension)
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

func List(version string) ([]PluginMetadata, error) {
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
		supportedSuffix := stringutil.GetMatchingSuffix(file.Name(), supportedPluginExtensions)
		if supportedSuffix == "" || file.IsDir() {
			continue
		}
		pluginName := strings.TrimPrefix(strings.TrimSuffix(file.Name(), supportedSuffix), pluginFileNamePrefix)
		available = append(available, PluginMetadata{
			Name:    pluginName,
			Version: version,
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
