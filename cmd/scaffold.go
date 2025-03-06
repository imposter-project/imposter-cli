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
	impostermodel2 "gatehill.io/imposter/internal/impostermodel"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var scaffoldFlags = struct {
	forceOverwrite    bool
	generateResources bool
	scriptEngine      string
}{}

// scaffoldCmd represents the up command
var scaffoldCmd = &cobra.Command{
	Use:     "scaffold [DIR]",
	Aliases: []string{"init"},
	Short:   "Create Imposter configuration",
	Long: `Creates Imposter configuration files. If one or more OpenAPI/Swagger
specification files are present, they are used as the basis for the generated
resources. If no specification files are present, a simple REST mock is created.

If DIR is not specified, the current working directory is used.`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var configDir string
		if len(args) == 0 {
			configDir, _ = os.Getwd()
		} else {
			configDir, _ = filepath.Abs(args[0])
		}
		scriptEngine := impostermodel2.ParseScriptEngine(scaffoldFlags.scriptEngine)
		impostermodel2.Create(configDir, scaffoldFlags.generateResources, scaffoldFlags.forceOverwrite, scriptEngine, false)
	},
}

func init() {
	scaffoldCmd.Flags().BoolVarP(&scaffoldFlags.forceOverwrite, "force-overwrite", "f", false, "Force overwrite of destination file(s) if already exist")
	scaffoldCmd.Flags().BoolVar(&scaffoldFlags.generateResources, "generate-resources", true, "Generate Imposter resources from OpenAPI paths")
	scaffoldCmd.Flags().StringVarP(&scaffoldFlags.scriptEngine, "script-engine", "s", "none", "Generate placeholder Imposter script (none|groovy|js)")
	rootCmd.AddCommand(scaffoldCmd)
}
