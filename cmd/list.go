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
	all            bool
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
		if listFlags.all {
			listAllMocks(listFlags.quiet)
		} else {
			listMocks(engine.GetConfiguredType(listFlags.engineType), listFlags.quiet, false)
		}
	},
}

func init() {
	listCmd.Flags().StringVarP(&listFlags.engineType, "engine-type", "t", "", "Imposter engine type (valid: docker,golang,jvm - default \"docker\")")
	listCmd.Flags().BoolVarP(&listFlags.all, "all", "a", false, "List mocks for all engine types")
	listCmd.Flags().BoolVarP(&listFlags.healthExitCode, "exit-code-health", "x", false, "Set exit code based on mock health")
	listCmd.Flags().BoolVarP(&listFlags.quiet, "quiet", "q", false, "Quieten output; only print ID")
	listCmd.MarkFlagsMutuallyExclusive("engine-type", "all")
	registerEngineTypeCompletions(listCmd)
	rootCmd.AddCommand(listCmd)
}

func listAllMocks(quiet bool) {
	var allRows [][]string
	var totalMocks int
	var anyFailed bool

	for _, engineType := range allEngineTypes {
		var rows [][]string
		var mocks int
		var failed bool
		err := runWithRecovery(func() {
			var e error
			rows, mocks, failed, e = listMocksForEngine(engineType, quiet, true)
			if e != nil {
				logger.Warnf("failed to list %s mocks: %s", engineType, e)
			}
		})
		if err != nil {
			logger.Warnf("failed to list %s mocks: %s", engineType, err)
			continue
		}
		allRows = append(allRows, rows...)
		totalMocks += mocks
		if failed {
			anyFailed = true
		}
	}
	if !quiet {
		renderMocks(allRows, true)
	}

	if listFlags.healthExitCode {
		if totalMocks > 0 && !anyFailed {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}

func listMocks(engineType engine.EngineType, quiet bool, showEngine bool) {
	rows, mockCount, anyFailed, err := listMocksForEngine(engineType, quiet, showEngine)
	if err != nil {
		logger.Fatalf("failed to list mocks: %s", err)
	}
	if !quiet {
		renderMocks(rows, showEngine)
	}

	if listFlags.healthExitCode {
		if mockCount > 0 && !anyFailed {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}
}

func listMocksForEngine(engineType engine.EngineType, quiet bool, showEngine bool) (rows [][]string, mockCount int, anyFailed bool, err error) {
	configDir := filepath.Join(os.TempDir(), "imposter-list")
	mockEngine := engine.BuildEngine(engineType, configDir, engine.StartOptions{})

	mocks, err := mockEngine.ListAllManaged()
	if err != nil {
		return nil, 0, false, err
	}

	for _, mock := range mocks {
		engine.PopulateHealth(&mock)
		if quiet {
			os.Stdout.WriteString(mock.ID + "\n")
		} else {
			row := []string{mock.ID, mock.Name, strconv.Itoa(mock.Port), string(mock.Health)}
			if showEngine {
				row = append(row, string(engineType))
			}
			rows = append(rows, row)
		}
		if mock.Health != engine.MockHealthHealthy {
			anyFailed = true
		}
	}
	return rows, len(mocks), anyFailed, nil
}

func renderMocks(rows [][]string, showEngine bool) {
	table := tablewriter.NewWriter(os.Stdout)
	header := []string{"ID", "Name", "Port", "Health"}
	if showEngine {
		header = append(header, "Engine")
	}
	table.Header(header)
	table.Bulk(rows)
	table.Render()
}
