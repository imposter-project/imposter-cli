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
	"github.com/imposter-project/imposter-cli/internal/engine"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var downFlags = struct {
	engineType string
	all        bool
}{}

// downCmd represents the down command
var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop running mocks",
	Long:  `Stops running Imposter mocks for the current engine type.`,
	Run: func(cmd *cobra.Command, args []string) {
		if downFlags.all {
			stopAllEngines()
		} else {
			stopAll(engine.GetConfiguredType(downFlags.engineType))
		}
	},
}

func init() {
	downCmd.Flags().StringVarP(&downFlags.engineType, "engine-type", "t", "", "Imposter engine type (valid: docker,golang,jvm - default \"docker\")")
	downCmd.Flags().BoolVarP(&downFlags.all, "all", "a", false, "Stop mocks for all engine types")
	downCmd.MarkFlagsMutuallyExclusive("engine-type", "all")
	registerEngineTypeCompletions(downCmd)
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

func stopAll(engineType engine.EngineType) {
	logger.Info("stopping all managed mocks...")
	stopped, err := stopEngine(engineType)
	if err != nil {
		logger.Fatalf("failed to stop mocks: %s", err)
	}
	if stopped > 0 {
		logger.Infof("stopped %d managed mock(s)", stopped)
	} else {
		logger.Info("no managed mocks were found")
	}
}

func stopEngine(engineType engine.EngineType) (int, error) {
	configDir := filepath.Join(os.TempDir(), "imposter-down")
	mockEngine := engine.BuildEngine(engineType, configDir, engine.StartOptions{})
	stopped := mockEngine.StopAllManaged()
	return stopped, nil
}
