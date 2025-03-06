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
	"gatehill.io/imposter/internal/remote"
	"gatehill.io/imposter/internal/workspace"
	"github.com/spf13/cobra"
	"os"
)

// workspaceShowCmd represents the workspaceShow command
var workspaceShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show active workspace",
	Long:  `Shows the active workspace, if set.`,
	Run: func(cmd *cobra.Command, args []string) {
		var dir string
		if workspaceFlags.path != "" {
			dir = workspaceFlags.path
		} else {
			dir, _ = os.Getwd()
		}
		printActiveWorkspace(dir)
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceShowCmd)
}

func printActiveWorkspace(dir string) {
	active, err := workspace.GetActive(dir)
	if err != nil {
		logger.Fatalf("failed to get active workspace: %s", err)
	}
	if active != nil {
		var remoteType string
		rem, err := remote.Load(dir, active)
		if err != nil {
			logger.Trace(err)
			remoteType = "unknown"
		} else {
			remoteType = (*rem).GetType()
		}
		fmt.Printf("Active workspace: %s\nRemote provider: %s\n", active.Name, remoteType)
	} else {
		fmt.Printf("No active workspace\n")
	}
}
