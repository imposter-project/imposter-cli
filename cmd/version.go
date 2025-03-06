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
	"gatehill.io/imposter/internal/config"
	"gatehill.io/imposter/internal/engine"
	"github.com/spf13/cobra"
)

type outputFormat string

const (
	outputFormatPlain outputFormat = "plain"
	outputFormatJson  outputFormat = "json"
)

var versionFlags = struct {
	engineType string
	format     string
}{}

// versionCmd represents the up command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints version information",
	Long:  `Prints the version of the CLI and engine, if available.`,
	Run: func(cmd *cobra.Command, args []string) {
		engineType := engine.GetConfiguredType(versionFlags.engineType)
		var format outputFormat
		if versionFlags.format != "" {
			format = outputFormat(versionFlags.format)
		} else {
			format = outputFormatPlain
		}
		println(describeVersions(engineType, format))
	},
}

func init() {
	versionCmd.Flags().StringVarP(&versionFlags.engineType, "engine-type", "t", "", "Imposter engine type (valid: docker,jvm - default \"docker\")")
	versionCmd.Flags().StringVarP(&versionFlags.format, "output-format", "o", "", "Output format (valid: plain,json - default \"plain\")")
	registerEngineTypeCompletions(versionCmd)
	rootCmd.AddCommand(versionCmd)
}

func describeVersions(engineType engine.EngineType, format outputFormat) string {
	output := formatProperty(format, "imposter-cli", config.Config.Version, false)

	library := engine.GetLibrary(engineType)
	engines, err := library.List()
	if err != nil {
		logger.Fatal(err)
	}
	if len(engines) == 0 {
		output += formatProperty(format, "imposter-engine", "none", true)
	} else {
		engineConfigVersion := engine.GetConfiguredVersionOrResolve("", true, false)
		if engineConfigVersion == "latest" {
			engineConfigVersion = engine.GetHighestVersion(engines)
		}
		output += formatProperty(format, "imposter-engine", engineConfigVersion, false)
		output += formatProperty(format, "engine-output", getInstalledEngineVersion(engineType, engineConfigVersion), true)
	}

	switch format {
	case outputFormatPlain:
		return output
	case outputFormatJson:
		return fmt.Sprintf("{\n%s}", output)
	default:
		panic(fmt.Errorf("unsupported output format: %s", format))
	}
}

func formatProperty(format outputFormat, key string, value string, lastProp bool) string {
	var formatted string
	switch format {
	case outputFormatPlain:
		formatted = fmt.Sprintf("%s %s", key, value)
	case outputFormatJson:
		formatted = fmt.Sprintf(`  "%s": "%s"`, key, value)
		if !lastProp {
			formatted += ","
		}
	default:
		panic(fmt.Errorf("unsupported output format: %s", format))
	}
	return formatted + "\n"
}

func getInstalledEngineVersion(engineType engine.EngineType, version string) string {
	mockEngine := engine.BuildEngine(engineType, "", engine.StartOptions{
		Version:  version,
		LogLevel: "INFO",
	})
	versionString, err := mockEngine.GetVersionString()
	if err != nil {
		logger.Warn(err)
		return "error"
	}
	return versionString
}
