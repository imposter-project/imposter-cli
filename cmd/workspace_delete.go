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
	"github.com/spf13/cobra"
	"os"
)

// workspaceDeleteCmd represents the workspaceDelete command
var workspaceDeleteCmd = &cobra.Command{
	Use:   "delete [WORKSPACE_NAME]",
	Short: "Delete a workspace",
	Long:  `Deletes a workspace, if it exists.`,
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return suggestWorkspaceNames()
	},
	Run: func(cmd *cobra.Command, args []string) {
		var dir string
		if workspaceFlags.path != "" {
			dir = workspaceFlags.path
		} else {
			dir, _ = os.Getwd()
		}
		name := args[0]
		deleteWorkspace(dir, name)
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceDeleteCmd)
}

func deleteWorkspace(dir string, name string) {
	err := workspace.Delete(dir, name)
	if err != nil {
		logger.Fatalf("failed to delete workspace: %s", err)
	}
	logger.Infof("deleted workspace '%s'", name)
}
