/*
Copyright © 2021-2023 Pete Cornish <outofcoffee@gmail.com>

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
	"gatehill.io/imposter/internal/config"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// localConfigCmd represents the localConfig command
var localConfigCmd = &cobra.Command{
	Use:   "config [key=value]",
	Short: "Set CLI config for working directory",
	Long:  `Sets CLI configuration for the working directory.`,
	Args:  cobra.MinimumNArgs(0),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var formattedKeys []string
		for _, k := range listSupportedLocalKeys() {
			formattedKeys = append(formattedKeys, k+"=VAL")
		}
		return formattedKeys, cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			configDir, _ := os.Getwd()
			setLocalConfig(cmd, args, configDir)
		} else {
			printLocalConfigHelp(cmd)
		}
	},
}

func init() {
	rootCmd.AddCommand(localConfigCmd)
}

func printLocalConfigHelp(cmd *cobra.Command) {
	supported := strings.Join(listSupportedLocalKeys(), ", ")
	fmt.Fprintf(os.Stderr, "%v\nSupported config keys: %s\n", cmd.UsageString(), supported)
	os.Exit(1)
}

func setLocalConfig(cmd *cobra.Command, args []string, configDir string) {
	parsed := config.ParseConfig(args)
	if len(parsed) == 0 {
		printLocalConfigHelp(cmd)
	} else {
		for _, pair := range parsed {
			err := config.WriteLocalConfigValue(configDir, pair.Key, pair.Value)
			if err != nil {
				panic(err)
			}
		}
		logger.Infof("set CLI config for: %s", configDir)
	}
}

func listSupportedLocalKeys() []string {
	return []string{
		"engine",
		"version",
	}
}
