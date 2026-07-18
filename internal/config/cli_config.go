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

package config

import (
	"fmt"
	"github.com/coreos/go-semver/semver"
	"github.com/imposter-project/imposter-cli/internal/logging"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type CliConfig struct {
	Version  string
	LogLevel string
}

type ConfigPair struct {
	Key   string
	Value string
}

const DevCliVersion = "dev"

// Version holds the current version of the application.
// This will be set during build time using -ldflags.
var version = DevCliVersion

// The GlobalConfigFileName is the file name without the file extension.
const GlobalConfigFileName = "config"

// The LocalDirConfigFileName is the canonical local project config file name,
// without the file extension.
const LocalDirConfigFileName = "imposter-project"

// The LegacyLocalDirConfigFileName is the deprecated local project config file
// name, without the file extension. It is still honoured for backward
// compatibility, but takes lower precedence than LocalDirConfigFileName.
const LegacyLocalDirConfigFileName = ".imposter"

var logger = logging.GetLogger()

var (
	Config  CliConfig
	DirPath string
)

func init() {
	Config = CliConfig{
		Version:  DevCliVersion,
		LogLevel: "DEBUG",
	}
}

func GetGlobalConfigDir() (string, error) {
	if DirPath != "" {
		return DirPath, nil
	}
	return getDefaultGlobalConfigDir()
}

func getDefaultGlobalConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to determine user home directory: %s", err)
	}
	configFile := filepath.Join(homeDir, ".imposter")
	return configFile, nil
}

func MergeCliConfigIfExists(configDir string) {
	viper.AddConfigPath(configDir)

	// Merge the legacy config file first, so that the canonical file takes
	// precedence over it when both are present.
	viper.SetConfigName(LegacyLocalDirConfigFileName)
	if err := viper.MergeInConfig(); err == nil {
		logger.Warnf(
			"using deprecated local CLI config file: %v - rename it to %s.yaml (or .yml/.json)",
			viper.ConfigFileUsed(), LocalDirConfigFileName,
		)
	}

	// If a local CLI config file is found, read it in. This takes precedence
	// over the legacy file merged above.
	viper.SetConfigName(LocalDirConfigFileName)
	if err := viper.MergeInConfig(); err == nil {
		logger.Tracef("using local CLI config file: %v", viper.ConfigFileUsed())
	}

	// if a CLI version is specified - check it
	if requiredCliVersion := viper.GetString("cli.version"); requiredCliVersion != "" {
		if err := checkCliVersion(requiredCliVersion); err != nil {
			logger.Fatal(err)
		}
	}
}

func checkCliVersion(required string) error {
	if Config.Version == DevCliVersion {
		logger.Warnf("using dev CLI version - cannot check version constraint against %v", required)
		return nil
	}
	cliVer, err := semver.NewVersion(Config.Version)
	if err != nil {
		return fmt.Errorf("failed to parse CLI version: %v: %v", Config.Version, err)
	}
	reqVer, err := semver.NewVersion(required)
	if err != nil {
		return fmt.Errorf("failed to parse required CLI version: %v: %v", required, err)
	}
	if cliVer.Compare(*reqVer) >= 0 {
		logger.Tracef("CLI version requirement met [required: %v, current: %v]", required, Config.Version)
		return nil
	} else {
		return fmt.Errorf("CLI version requirement not met [required: %v, current: %v]", required, Config.Version)
	}
}

func ParseConfig(args []string) []ConfigPair {
	var pairs []ConfigPair
	for _, arg := range args {
		if !strings.Contains(arg, "=") {
			logger.Warnf("invalid config item: %s", arg)
			continue
		}
		splitArgs := strings.Split(arg, "=")
		pairs = append(pairs, ConfigPair{
			Key:   splitArgs[0],
			Value: strings.Trim(splitArgs[1], `"`),
		})
	}
	return pairs
}

// localConfigExtensions are the supported extensions for the local project
// config file, in preference order.
var localConfigExtensions = []string{"yaml", "yml", "json"}

// FindLocalConfigFile returns the path to an existing local project config file
// in configDir, preferring the canonical name over the legacy name (and yaml
// over other extensions). It returns an empty string if none exists.
func FindLocalConfigFile(configDir string) string {
	for _, name := range []string{LocalDirConfigFileName, LegacyLocalDirConfigFileName} {
		for _, ext := range localConfigExtensions {
			candidate := path.Join(configDir, name+"."+ext)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
		}
	}
	return ""
}

func WriteLocalConfigValue(configDir string, key string, value string) error {
	v := viper.New()

	// Update an existing config file if one is present (including the legacy
	// name), otherwise create a new one using the canonical name.
	localConfig := FindLocalConfigFile(configDir)
	if localConfig == "" {
		localConfig = path.Join(configDir, LocalDirConfigFileName+".yaml")
	}
	v.SetConfigFile(localConfig)

	// sink if does not exist
	_ = v.ReadInConfig()

	v.Set(key, value)
	err := v.WriteConfig()
	if err != nil {
		return fmt.Errorf("failed to write config file: %s: %v", localConfig, err)
	}

	logger.Tracef("wrote CLI config to: %s", localConfig)
	return nil
}

func SetCliConfig(logLevel string) {
	Config = CliConfig{
		Version:  version,
		LogLevel: logLevel,
	}
}
