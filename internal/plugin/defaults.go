package plugin

import (
	"fmt"
	"gatehill.io/imposter/internal/config"
	"gatehill.io/imposter/internal/stringutil"
	"github.com/spf13/viper"
	"path/filepath"
)

const defaultPluginsConfigKey = "default.plugins"

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

// removeDefaultPlugins removes the specified plugins from the default
// plugins configuration and writes the updated configuration file.
func removeDefaultPlugins(plugins []string) error {
	existing, err := ListDefaultPlugins()
	if err != nil {
		return fmt.Errorf("failed to load default plugins: %s", err)
	}

	var updated []string
	var removed []string
	for _, plugin := range existing {
		if !stringutil.Contains(plugins, plugin) {
			updated = append(updated, plugin)
		} else {
			removed = append(removed, plugin)
		}
	}

	if len(existing) == len(updated) {
		// none removed
		return nil
	}

	err = writeDefaultPlugins(updated)
	if err != nil {
		return err
	}

	logger.Infof("removed %d plugin(s) from default list: %v", len(removed), removed)
	return nil
}
