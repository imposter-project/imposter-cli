/*
Copyright © 2023 Pete Cornish <outofcoffee@gmail.com>

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
	"gatehill.io/imposter/internal/remote"
	"github.com/spf13/cobra"
	"os"
)

// remoteUndeployCmd represents the remoteUndeploy command
var remoteUndeployCmd = &cobra.Command{
	Use:   "undeploy",
	Short: "Undeploy active workspace",
	Long:  `Undeploys the active workspace from the remote.`,
	Run: func(cmd *cobra.Command, args []string) {
		var dir string
		if remoteFlags.path != "" {
			dir = remoteFlags.path
		} else {
			dir, _ = os.Getwd()
		}
		remoteUndeploy(dir)
	},
}

func init() {
	remoteCmd.AddCommand(remoteUndeployCmd)
}

func remoteUndeploy(dir string) {
	active, r, err := remote.LoadActive(dir)
	if err != nil {
		logger.Fatalf("failed to load remote: %s", err)
	}
	logger.Infof("undeploying '%s' from %s remote", active.Name, active.RemoteType)

	err = (*r).Undeploy()
	if err != nil {
		logger.Fatalf("failed to undeploy from remote: %s", err)
	}
	logger.Infof("workspace '%s' is undeployed from remote", active.Name)
}
