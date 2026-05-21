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
	"os"
	"path/filepath"
	"strings"

	"github.com/imposter-project/imposter-cli/internal/engine"
	"github.com/spf13/cobra"
)

var downFlags = struct {
	all bool
}{}

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down [ID]",
	Short: "Stop a running mock by ID, or all mocks with --all",
	Long: `Stops a single running Imposter mock identified by ID, or all
managed mocks across every engine type with --all.

Use 'imposter ls' to discover the IDs of running mocks.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if downFlags.all {
			if len(args) > 0 {
				logger.Fatal("cannot specify both --all and a mock ID")
			}
			stopAllEngines()
			return
		}
		if len(args) == 0 {
			logger.Fatal("a mock ID is required (or use --all to stop all mocks); see 'imposter ls' for IDs")
		}
		stopMockByID(args[0])
	},
}

func init() {
	downCmd.Flags().BoolVarP(&downFlags.all, "all", "a", false, "Stop all managed mocks across all engine types")
	rootCmd.AddCommand(downCmd)
}

func stopAllEngines() {
	logger.Info("stopping all managed mocks for all engine types...")
	totalStopped := 0
	for _, engineType := range allEngineTypes {
		logger.Infof("checking %s engine...", engineType)
		var stopped int
		err := runWithRecovery(func() {
			var e error
			stopped, e = stopEngine(engineType)
			if e != nil {
				logger.Warnf("failed to stop %s mocks: %s", engineType, e)
			}
		})
		if err != nil {
			logger.Warnf("failed to stop %s mocks: %s", engineType, err)
			continue
		}
		totalStopped += stopped
	}
	if totalStopped > 0 {
		logger.Infof("stopped %d managed mock(s) in total", totalStopped)
	} else {
		logger.Info("no managed mocks were found")
	}
}

// stopMockByID searches every engine type for a managed mock with the
// given ID and stops it.
func stopMockByID(id string) {
	var engineErrors []string
	for _, engineType := range allEngineTypes {
		var stopped bool
		var stopErr error
		err := runWithRecovery(func() {
			mockEngine := engine.BuildEngine(engineType, filepath.Join(os.TempDir(), "imposter-down"), engine.StartOptions{})
			stopped, stopErr = mockEngine.StopManaged(id)
		})
		if err != nil {
			engineErrors = append(engineErrors, fmt.Sprintf("%s: %v", engineType, err))
			continue
		}
		if stopErr != nil {
			engineErrors = append(engineErrors, fmt.Sprintf("%s: %v", engineType, stopErr))
			continue
		}
		if stopped {
			logger.Infof("stopped mock %s (%s engine)", id, engineType)
			return
		}
	}
	if len(engineErrors) == len(allEngineTypes) {
		logger.Fatalf("failed to query any engine: %s", strings.Join(engineErrors, "; "))
	}
	logger.Fatalf("no managed mock found with ID %q (run 'imposter ls' to see running mocks)", id)
}

func stopEngine(engineType engine.EngineType) (int, error) {
	configDir := filepath.Join(os.TempDir(), "imposter-down")
	mockEngine := engine.BuildEngine(engineType, configDir, engine.StartOptions{})
	stopped := mockEngine.StopAllManaged()
	return stopped, nil
}
