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
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
)

var engineListFlags = struct {
	engineType string
}{}

// engineListCmd represents the engineList command
var engineListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List the engines in the cache",
	Long: `Lists all versions of engine binaries/images in the cache.

If engine type is not specified, it defaults to all.`,
	Run: func(cmd *cobra.Command, args []string) {
		// unspecified type is valid
		engineType := engine.GetConfiguredTypeWithDefault(engineListFlags.engineType, engine.EngineTypeNone)

		var engineTypes []engine.EngineType
		if engine.EngineTypeNone == engineType {
			engineTypes = engine.EnumerateLibraries()
		} else {
			engineTypes = []engine.EngineType{engineType}
		}
		listEngines(engineTypes)
	},
}

func listEngines(engineTypes []engine.EngineType) {
	logger.Tracef("listing engines")
	var available []engine.EngineMetadata

	for _, e := range engineTypes {
		library := engine.GetLibrary(e)
		engines, err := library.List()
		if err != nil {
			logger.Fatal(err)
		}
		available = append(available, engines...)
	}

	var rows [][]string
	for _, metadata := range available {
		rows = append(rows, []string{string(metadata.EngineType), metadata.Version})
	}
	renderEngines(rows)
}

func renderEngines(rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Type", "Version"})
	table.Bulk(rows)
	table.Render()
}

func init() {
	engineListCmd.Flags().StringVarP(&engineListFlags.engineType, "engine-type", "t", "", "Imposter engine type (valid: docker,golang,jvm - default is all")
	registerEngineTypeCompletions(engineListCmd)
	engineCmd.AddCommand(engineListCmd)
}
