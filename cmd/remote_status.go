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
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

// remoteStatusCmd represents the remoteStatus command
var remoteStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show remote status",
	Long:  `Shows the status of the remote for the active workspace.`,
	Run: func(cmd *cobra.Command, args []string) {
		var dir string
		if remoteFlags.path != "" {
			dir = remoteFlags.path
		} else {
			dir, _ = os.Getwd()
		}
		showRemoteStatus(dir)
	},
}

func init() {
	remoteCmd.AddCommand(remoteStatusCmd)
}

func showRemoteStatus(dir string) {
	active, r, err := remote.LoadActive(dir)
	if err != nil {
		logger.Fatalf("failed to load remote: %s", err)
	}

	status, err := (*r).GetStatus()
	if err != nil {
		logger.Fatalf("failed to get remote status: %s", err)
	}

	var lastModified string
	if status.LastModified > 0 {
		lastModified = fmt.Sprintf("%v", time.UnixMilli(int64(status.LastModified)))
	} else {
		lastModified = "never"
	}
	msg := fmt.Sprintf("Workspace '%s' remote status: %s\nLast modified: %s", active.Name, status.Status, lastModified)

	if strings.ToUpper(status.Status) == "ACTIVE" {
		endpoint, err := (*r).GetEndpoint()
		if err != nil {
			logger.Warnf("failed to get remote details: %s", err)
		} else {
			msg += fmt.Sprintf("\nBase URL: %s\nSpec: %s\nStatus: %s", endpoint.BaseUrl, endpoint.SpecUrl, endpoint.StatusUrl)
		}
	}
	logger.Info(msg)
}
