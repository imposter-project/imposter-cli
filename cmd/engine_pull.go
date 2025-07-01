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
	"github.com/spf13/cobra"
)

var enginePullFlags = struct {
	engineType    string
	engineVersion string
	forcePull     bool
}{}

// enginePullCmd represents the enginePull command
var enginePullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull the engine into the cache",
	Long: `Pulls a specified version of the engine binary/image into the cache.

If version is not specified, it defaults to 'latest'.`,
	Run: func(cmd *cobra.Command, args []string) {
		var pullPolicy engine.PullPolicy
		if enginePullFlags.forcePull {
			pullPolicy = engine.PullAlways
		} else {
			pullPolicy = engine.PullIfNotPresent
		}
		engineType := engine.GetConfiguredType(enginePullFlags.engineType)
		version := engine.GetConfiguredVersion(engineType, enginePullFlags.engineVersion, pullPolicy != engine.PullAlways)
		pull(version, engineType, pullPolicy)
	},
}

func pull(version string, engineType engine.EngineType, pullPolicy engine.PullPolicy) {
	downloader := engine.GetLibrary(engineType).GetProvider(version)
	err := downloader.Provide(pullPolicy)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("pulled engine version %v", version)
}

func init() {
	enginePullCmd.Flags().StringVarP(&enginePullFlags.engineType, "engine-type", "t", "", "Imposter engine type (valid: docker,golang,jvm - default \"docker\")")
	enginePullCmd.Flags().StringVarP(&enginePullFlags.engineVersion, "version", "v", "", "Imposter engine version (default \"latest\")")
	enginePullCmd.Flags().BoolVarP(&enginePullFlags.forcePull, "force", "f", false, "Force engine pull")
	registerEngineTypeCompletions(enginePullCmd)
	engineCmd.AddCommand(enginePullCmd)
}
