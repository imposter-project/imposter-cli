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

package golang

import (
	"gatehill.io/imposter/internal/engine"
	"gatehill.io/imposter/internal/engine/enginetests"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"testing"
)

var engineBuilder = func(tt enginetests.EngineTestScenario) engine.MockEngine {
	return engine.BuildEngine("golang", tt.Fields.ConfigDir, tt.Fields.Options)
}

func init() {
	logger.SetLevel(logrus.TraceLevel)
	EnableEngine()
}

func TestEngine_StartStop(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	testConfigPath := filepath.Join(workingDir, "../enginetests/testdata")

	tests := []enginetests.EngineTestScenario{
		{
			Name: "start golang engine",
			Fields: enginetests.EngineTestFields{
				ConfigDir: testConfigPath,
				Options: engine.StartOptions{
					Port:       enginetests.GetFreePort(),
					Version:    "0.16.0",
					PullPolicy: engine.PullIfNotPresent,
					LogLevel:   "DEBUG",
				},
			},
		},
	}
	enginetests.StartStop(t, tests, engineBuilder)
}

func TestEngine_Restart(t *testing.T) {
	logger.SetLevel(logrus.TraceLevel)
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	testConfigPath := filepath.Join(workingDir, "../enginetests/testdata")

	tests := []enginetests.EngineTestScenario{
		{
			Name: "restart golang engine",
			Fields: enginetests.EngineTestFields{
				ConfigDir: testConfigPath,
				Options: engine.StartOptions{
					Port:       enginetests.GetFreePort(),
					Version:    "0.16.0",
					PullPolicy: engine.PullIfNotPresent,
					LogLevel:   "DEBUG",
				},
			},
		},
	}
	enginetests.Restart(t, tests, engineBuilder)
}

func TestEngine_List(t *testing.T) {
	logger.SetLevel(logrus.TraceLevel)
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	testConfigPath := filepath.Join(workingDir, "../enginetests/testdata")

	tests := []enginetests.EngineTestScenario{
		{
			Name: "list golang engine",
			Fields: enginetests.EngineTestFields{
				ConfigDir: testConfigPath,
				Options: engine.StartOptions{
					Port:       enginetests.GetFreePort(),
					Version:    "0.16.0",
					PullPolicy: engine.PullIfNotPresent,
					LogLevel:   "DEBUG",
				},
				// skip port check as port can't be determined from environment variables
				SkipCheckPort: true,
			},
		},
	}
	enginetests.List(t, tests, engineBuilder)
}
