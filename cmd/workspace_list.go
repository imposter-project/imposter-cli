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
	"gatehill.io/imposter/internal/workspace"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"os"
)

// workspaceListCmd represents the workspaceList command
var workspaceListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all workspaces",
	Long:    `Lists all workspaces, showing the active workspace, if set.`,
	Run: func(cmd *cobra.Command, args []string) {
		var dir string
		if workspaceFlags.path != "" {
			dir = workspaceFlags.path
		} else {
			dir, _ = os.Getwd()
		}
		listWorkspaces(dir)
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceListCmd)
}

func listWorkspaces(dir string) {
	workspaces, err := workspace.List(dir)
	if err != nil {
		logger.Fatalf("failed to list workspaces: %s", err)
	}
	active, err := workspace.GetActive(dir)
	if err != nil {
		logger.Fatalf("failed to list workspaces: %s", err)
	}
	var activeName string
	if active != nil {
		activeName = active.Name
	}

	var rows [][]string
	for _, w := range workspaces {
		var activeStatus string
		if w.Name == activeName {
			activeStatus = "active"
		}
		rows = append(rows, []string{w.Name, activeStatus})
	}
	renderWorkspaces(rows)
}

func renderWorkspaces(rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Workspace", "Status"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(rows)
	table.Render()
}
