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
	"gatehill.io/imposter/internal/engine"
	"github.com/spf13/cobra"
	"strings"
)

const reportTemplate = `
SUMMARY
%[1]v

DOCKER ENGINE
%[2]v

JVM ENGINE
%[3]v
`

// doctorCmd represents the doctor command
var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check prerequisites for running Imposter",
	Long: `Checks prerequisites for running Imposter, including those needed
by the engines.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("running check up...")
		println(checkPrereqs())
	},
}

func checkPrereqs() string {
	dockerOk, dockerMsgs := engine.GetLibrary(engine.EngineTypeDockerCore).CheckPrereqs()
	jvmOk, jvmMsgs := engine.GetLibrary(engine.EngineTypeJvmSingleJar).CheckPrereqs()

	var summary string
	if dockerOk || jvmOk {
		var hints []string
		if dockerOk {
			hints = append(hints, "'--engine-type docker'")
		}
		if jvmOk {
			hints = append(hints, "'--engine-type jvm'")
		}
		summary = fmt.Sprintf("🚀 You should be able to run Imposter, as you have support for one or more engines.\nPass %s when running 'imposter up' to select a supported engine type.", strings.Join(hints, " or "))
	} else {
		summary = "😭 You may not be able to run Imposter, as you do not have support for at least one engine."
	}
	return fmt.Sprintf(reportTemplate, summary, strings.Join(dockerMsgs, "\n"), strings.Join(jvmMsgs, "\n"))
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
