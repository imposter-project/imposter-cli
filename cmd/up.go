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

package cmd

import (
	"fmt"
	config2 "gatehill.io/imposter/internal/config"
	"gatehill.io/imposter/internal/engine"
	"gatehill.io/imposter/internal/fileutil"
	"gatehill.io/imposter/internal/plugin"
	"gatehill.io/imposter/internal/stringutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

var upFlags = struct {
	deduplicate         string
	engineType          string
	engineVersion       string
	forcePull           bool
	port                int
	restartOnChange     bool
	scaffoldMissing     bool
	enablePlugins       bool
	ensurePlugins       bool
	enableFileCache     bool
	environment         []string
	dirMounts           []string
	recursiveConfigScan bool
	debugMode           bool
}{}

// upCmd represents the up command
var upCmd = &cobra.Command{
	Use:   "up [CONFIG_DIR]",
	Short: "Start live mocks of APIs",
	Long: `Starts a live mock of your APIs, using their Imposter configuration.

If CONFIG_DIR is not specified, the current working directory is used.`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		injectExplicitEnvironment(upFlags.environment)

		var configDir string
		if len(args) == 0 {
			configDir, _ = os.Getwd()
		} else {
			configDir, _ = filepath.Abs(args[0])
		}
		if err := config2.ValidateConfigExists(configDir, upFlags.scaffoldMissing); err != nil {
			logger.Fatal(err)
		}

		// Search for CLI config files in the mock config dir.
		config2.MergeCliConfigIfExists(configDir)

		var pullPolicy engine.PullPolicy
		if upFlags.forcePull {
			pullPolicy = engine.PullAlways
		} else {
			pullPolicy = engine.PullIfNotPresent
		}

		engineType := engine.GetConfiguredType(upFlags.engineType)
		lib := engine.GetLibrary(engineType)

		var version string
		if !lib.IsSealedDistro() {
			// only resolve version if not a sealed distro, to avoid prefs write
			version = engine.GetConfiguredVersion(upFlags.engineVersion, pullPolicy != engine.PullAlways)

			// only ensure (and potentially fetch) default plugins if not a sealed distro
			if upFlags.ensurePlugins && lib.ShouldEnsurePlugins() {
				_, err := plugin.EnsureConfiguredPlugins(engineType, version)
				if err != nil {
					logger.Fatal(err)
				}
			}
		}

		startOptions := engine.StartOptions{
			Port:            upFlags.port,
			Version:         version,
			PullPolicy:      pullPolicy,
			LogLevel:        config2.Config.LogLevel,
			ReplaceRunning:  true,
			Deduplicate:     upFlags.deduplicate,
			EnablePlugins:   upFlags.enablePlugins,
			EnableFileCache: upFlags.enableFileCache,
			Environment:     buildStartEnvironment(upFlags.environment),
			DirMounts:       upFlags.dirMounts,
			DebugMode:       upFlags.debugMode,
		}
		start(&lib, startOptions, configDir, upFlags.restartOnChange)
	},
}

func init() {
	upCmd.Flags().StringVarP(&upFlags.engineType, "engine-type", "t", "", "Imposter engine type (valid: docker,jvm - default \"docker\")")
	upCmd.Flags().StringVarP(&upFlags.engineVersion, "version", "v", "", "Imposter engine version (default \"latest\")")
	upCmd.Flags().IntVarP(&upFlags.port, "port", "p", 8080, "Port on which to listen")
	upCmd.Flags().BoolVar(&upFlags.forcePull, "pull", false, "Force engine pull")
	upCmd.Flags().BoolVar(&upFlags.restartOnChange, "auto-restart", true, "Automatically restart when config dir contents change")
	upCmd.Flags().BoolVarP(&upFlags.scaffoldMissing, "scaffold", "s", false, "Scaffold Imposter configuration for all OpenAPI files")
	upCmd.Flags().StringVar(&upFlags.deduplicate, "deduplicate", "", "Override deduplication ID for replacement of containers")
	upCmd.Flags().BoolVar(&upFlags.enablePlugins, "enable-plugins", true, "Enable plugins")
	upCmd.Flags().BoolVar(&upFlags.ensurePlugins, "install-default-plugins", true, "Install missing default plugins")
	upCmd.Flags().BoolVar(&upFlags.enableFileCache, "enable-file-cache", true, "Enable file cache")
	upCmd.Flags().StringArrayVarP(&upFlags.environment, "env", "e", []string{}, "Explicit environment variables to set")
	upCmd.Flags().StringArrayVar(&upFlags.dirMounts, "mount-dir", []string{}, "(Docker engine type only) Extra directory bind-mounts in the form HOST_PATH:CONTAINER_PATH (e.g. $HOME/somedir:/opt/imposter/somedir) or simply HOST_PATH, which will mount the directory at /opt/imposter/<dir>")
	upCmd.Flags().BoolVarP(&upFlags.recursiveConfigScan, "recursive-config-scan", "r", false, "Scan for config files in subdirectories")
	upCmd.Flags().BoolVar(&upFlags.debugMode, "debug-mode", false, fmt.Sprintf("Enable JVM debug mode and listen on port %v", engine.DefaultDebugPort))
	registerEngineTypeCompletions(upCmd)
	rootCmd.AddCommand(upCmd)
}

func injectExplicitEnvironment(cliEnvArgs []string) {
	for _, env := range cliEnvArgs {
		envParts := strings.Split(env, "=")
		if len(envParts) > 1 {
			_ = os.Setenv(envParts[0], envParts[1])
		}
	}
	if upFlags.recursiveConfigScan {
		_ = os.Setenv("IMPOSTER_CONFIG_SCAN_RECURSIVE", "true")
	}
}

func buildStartEnvironment(cliEnvArgs []string) []string {
	env := append([]string{}, cliEnvArgs...)

	// include environment variables from CLI config file, under the 'env' key, such as:
	// ```yaml
	// env:
	//   IMPOSTER_FOO: bar
	//   IMPOSTER_BAZ: qux
	// ```
	for k, v := range viper.GetStringMapString("env") {
		envKey := strings.ToUpper(k)
		// environment variables passed as command-line arguments take precedence
		// over those in the config file
		if !stringutil.ContainsPrefix(env, envKey+"=") {
			env = append(env, envKey+"="+v)
		}
	}

	return env
}

func start(lib *engine.EngineLibrary, startOptions engine.StartOptions, configDir string, restartOnChange bool) {
	provider := (*lib).GetProvider(startOptions.Version)
	mockEngine := provider.Build(configDir, startOptions)

	wg := &sync.WaitGroup{}
	trapExit(mockEngine, wg)
	success := mockEngine.Start(wg)

	if success && restartOnChange {
		dirUpdated := fileutil.WatchDir(configDir)
		go func() {
			for {
				<-dirUpdated
				logger.Infof("detected change in: %v - triggering restart", configDir)
				mockEngine.Restart(wg)
			}
		}()
	}

	wg.Wait()
	logger.Debug("shutting down")
}

// listen for an interrupt from the OS, then attempt engine cleanup
func trapExit(mockEngine engine.MockEngine, wg *sync.WaitGroup) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		println()
		mockEngine.StopImmediately(wg)
	}()
}
