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
	"path/filepath"
	"strconv"
)

var listFlags = struct {
	engineType     string
	healthExitCode bool
	quiet          bool
}{}

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List running mocks",
	Long: `Lists running Imposter mocks for the current engine type
and reports their health.`,
	Run: func(cmd *cobra.Command, args []string) {
		listMocks(engine.GetConfiguredType(listFlags.engineType), listFlags.quiet)
	},
}

func init() {
	listCmd.Flags().StringVarP(&listFlags.engineType, "engine-type", "t", "", "Imposter engine type (valid: docker,jvm - default \"docker\")")
	listCmd.Flags().BoolVarP(&listFlags.healthExitCode, "exit-code-health", "x", false, "Set exit code based on mock health")
	listCmd.Flags().BoolVarP(&listFlags.quiet, "quiet", "q", false, "Quieten output; only print ID")
	registerEngineTypeCompletions(listCmd)
	rootCmd.AddCommand(listCmd)
}

func listMocks(engineType engine.EngineType, quiet bool) {
	configDir := filepath.Join(os.TempDir(), "imposter-list")
	mockEngine := engine.BuildEngine(engineType, configDir, engine.StartOptions{})

	mocks, err := mockEngine.ListAllManaged()
	if err != nil {
		logger.Fatalf("failed to list mocks: %s", err)
	}

	var anyFailed = false
	var rows [][]string
	for _, mock := range mocks {
		engine.PopulateHealth(&mock)
		if quiet {
			os.Stdout.WriteString(mock.ID + "\n")
		} else {
			rows = append(rows, []string{mock.ID, mock.Name, strconv.Itoa(mock.Port), string(mock.Health)})
		}
		if mock.Health != engine.MockHealthHealthy {
			anyFailed = true
		}
	}
	if !quiet {
		renderMocks(rows)
	}

	if listFlags.healthExitCode {
		// if there is at least one mock, and all mocks are healthy, return status 0
		if len(mocks) > 0 && !anyFailed {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}

func renderMocks(rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name", "Port", "Health"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(rows)
	table.Render()
}
