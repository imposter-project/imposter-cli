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
	config2 "gatehill.io/imposter/internal/config"
	"gatehill.io/imposter/internal/engine"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

var bundleFlags = struct {
	engineType    string
	engineVersion string
	output        string
}{}

// bundleCmd represents the bundle command
var bundleCmd = &cobra.Command{
	Use:   "bundle [CONFIG_DIR]",
	Short: "Bundle configuration and mock engine",
	Long: `Bundles the mock engine and configuration into a single file,
appropriate for the specified engine type.

For example, a Docker image for the Docker engine type, or a ZIP file
for the AWS Lambda engine type.

If CONFIG_DIR is not specified, the current working directory is used.`,
	Args: cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		var configDir string
		if len(args) == 0 {
			configDir, _ = os.Getwd()
		} else {
			configDir, _ = filepath.Abs(args[0])
		}
		if err := config2.ValidateConfigExists(configDir, false); err != nil {
			logger.Fatal(err)
		}

		// Search for CLI config files in the mock config dir.
		config2.MergeCliConfigIfExists(configDir)

		engineType := engine.GetConfiguredType(bundleFlags.engineType)
		lib := engine.GetLibrary(engineType)

		if lib.IsSealedDistro() {
			logger.Fatal("cannot bundle a sealed distribution")
		}

		version := engine.GetConfiguredVersion(bundleFlags.engineVersion, true)

		bundle(&lib, version, configDir, getBundleDest(engineType))
	},
}

func init() {
	bundleCmd.Flags().StringVarP(&bundleFlags.output, "output", "o", "", "The destination to write the bundle to. If using the 'docker' engine type, this must be a valid image name. Otherwise, this must be a path to a writeable file. If not specified, a name is generated.")
	bundleCmd.Flags().StringVarP(&bundleFlags.engineType, "engine-type", "t", "", "Imposter engine type (valid: awslambda,docker,jvm)")
	bundleCmd.Flags().StringVarP(&bundleFlags.engineVersion, "version", "v", "", "Imposter engine version (default \"latest\")")

	_ = bundleCmd.MarkFlagRequired("engine-type")
	registerEngineTypeCompletions(bundleCmd, engine.EngineTypeAwsLambda)
	rootCmd.AddCommand(bundleCmd)
}

func getBundleDest(engineType engine.EngineType) string {
	var dest string
	if bundleFlags.output != "" {
		dest = bundleFlags.output
	} else {
		if engineType == engine.EngineTypeDockerCore ||
			engineType == engine.EngineTypeDockerAll ||
			engineType == engine.EngineTypeDockerDistroless {

			imageTag := time.Now().Format("20060102150405")
			dest = "imposter-bundle:" + imageTag

		} else {
			temp, err := os.CreateTemp(os.TempDir(), "imposter-bundle-*.zip")
			if err != nil {
				logger.Fatal(fmt.Errorf("failed to create temporary file: %w", err))
			}
			dest = temp.Name()
			_ = os.Remove(dest)
		}
	}
	return dest
}

func bundle(lib *engine.EngineLibrary, version string, configDir string, dest string) {
	provider := (*lib).GetProvider(version)
	logger.Debugf("creating %s bundle %s using version %s", provider.GetEngineType(), configDir, version)

	err := provider.Provide(engine.PullIfNotPresent)
	if err != nil {
		logger.Fatal(err)
	}

	err = provider.Bundle(configDir, dest)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Infof("created %s bundle: %s", provider.GetEngineType(), dest)
}
