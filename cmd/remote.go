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

var remoteFlags struct {
	path string
}

// remoteCmd represents the remote command
var remoteCmd = &cobra.Command{
	Use:   "remote",
	Short: "Remote management and deployment commands",
}

func init() {
	remoteCmd.PersistentFlags().StringVarP(&remoteFlags.path, "workspace", "w", "", "workspace path")

	rootCmd.AddCommand(remoteCmd)
}

func getWorkspaceDir() string {
	var dir string
	if remoteFlags.path != "" {
		dir = remoteFlags.path
	} else {
		dir, _ = os.Getwd()
	}
	return dir
}

func suggestWorkspaceNames() ([]string, cobra.ShellCompDirective) {
	if workspaces, err := workspace.List(getWorkspaceDir()); err == nil {
		var wsNames []string
		for _, w := range workspaces {
			wsNames = append(wsNames, w.Name)
		}
		return wsNames, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}
