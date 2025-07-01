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
	"gatehill.io/imposter/internal/engine"
	"gatehill.io/imposter/internal/plugin"
	"github.com/spf13/cobra"
)

var pluginUninstallFlags = struct {
	engineVersion string
	removeDefault bool
}{}

// pluginUninstallCmd represents the pluginUninstall command
var pluginUninstallCmd = &cobra.Command{
	Use:     "uninstall [PLUGIN_NAME_1] [PLUGIN_NAME_N...]",
	Aliases: []string{"rm", "remove"},
	Short:   "Uninstall plugins",
	Long: `Uninstalls plugins for a specific engine version.

If version is not specified, it defaults to 'latest'.

Example 1: Uninstall named plugin

	imposter plugin uninstall store-redis

Example 2: Uninstall multiple plugins

	imposter plugin uninstall store-redis js-graal

Example 3: Uninstall plugin and remove from defaults

	imposter plugin uninstall store-redis --remove-default`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		engineType := engine.GetConfiguredType(pluginFlags.engineType)
		version := engine.GetConfiguredVersion(engineType, pluginUninstallFlags.engineVersion, true)
		uninstallPlugins(args, engineType, version, pluginUninstallFlags.removeDefault)
	},
}

func init() {
	pluginUninstallCmd.Flags().StringVarP(&pluginUninstallFlags.engineVersion, "version", "v", "", "Imposter engine version (default \"latest\")")
	pluginUninstallCmd.Flags().BoolVarP(&pluginUninstallFlags.removeDefault, "remove-default", "d", false, "Whether to remove the plugin from defaults")
	pluginCmd.AddCommand(pluginUninstallCmd)
}

func uninstallPlugins(plugins []string, engineType engine.EngineType, version string, removeDefault bool) {
	removed, err := plugin.UninstallPlugins(plugins, engineType, version, removeDefault)
	if err != nil {
		logger.Fatal(err)
	}
	if removed == 0 {
		logger.Infof("no plugins to uninstall")
	} else {
		logger.Infof("%d plugin(s) uninstalled", removed)
	}
}
