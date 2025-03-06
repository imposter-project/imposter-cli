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
	"fmt"
	"gatehill.io/imposter/internal/remote"
	"gatehill.io/imposter/internal/stringutil"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

// remoteSetTypeCmd represents the remoteSetType command
var remoteSetTypeCmd = &cobra.Command{
	Use:   "set-type [REMOTE_TYPE]",
	Short: "Set remote deployment type",
	Long:  `Sets the remote deployment type for the active workspace.`,
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return remote.ListTypes(), cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		remoteType := args[0]
		if stringutil.Contains(remote.ListTypes(), remoteType) {
			setRemoteType(getWorkspaceDir(), remoteType)
		} else {
			printRemoteSetTypeHelp(cmd)
		}
	},
}

func init() {
	remoteCmd.AddCommand(remoteSetTypeCmd)
}

func printRemoteSetTypeHelp(cmd *cobra.Command) {
	supported := strings.Join(remote.ListTypes(), ", ")
	fmt.Fprintf(os.Stderr, "%v\nSupported remote types: %s\n", cmd.UsageString(), supported)
	os.Exit(1)
}

func setRemoteType(dir string, remoteType string) {
	active, err := remote.SaveActiveRemoteType(dir, remoteType)
	if err != nil {
		logger.Fatalf("failed to set remote type: %s", err)
	}
	logger.Infof("set remote type to '%s' for remote: %s", remoteType, active.Name)
}
