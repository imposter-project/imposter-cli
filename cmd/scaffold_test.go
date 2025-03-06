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
	"gatehill.io/imposter/internal/fileutil"
	impostermodel2 "gatehill.io/imposter/internal/impostermodel"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	logger.SetLevel(logrus.TraceLevel)
}

func Test_createMockConfig(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	testConfigPath := filepath.Join(workingDir, "/testdata")

	type args struct {
		generateResources bool
		forceOverwrite    bool
		scriptEngine      impostermodel2.ScriptEngine
		copySpecs         bool
		anchorFileName    string
		checkResponseFile bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "generate openapi mock no resources no script",
			args: args{
				generateResources: false,
				forceOverwrite:    true,
				scriptEngine:      impostermodel2.ScriptEngineNone,
				anchorFileName:    "order_service",
				copySpecs:         true,
				checkResponseFile: false,
			},
		},
		{
			name: "generate openapi mock no resources with script",
			args: args{
				generateResources: false,
				forceOverwrite:    true,
				scriptEngine:      impostermodel2.ScriptEngineJavaScript,
				anchorFileName:    "order_service",
				copySpecs:         true,
				checkResponseFile: false,
			},
		},
		{
			name: "generate openapi mock with resources no script",
			args: args{
				generateResources: true,
				forceOverwrite:    true,
				scriptEngine:      impostermodel2.ScriptEngineNone,
				anchorFileName:    "order_service",
				copySpecs:         true,
				checkResponseFile: false,
			},
		},
		{
			name: "generate openapi mock with resources with script",
			args: args{
				generateResources: true,
				forceOverwrite:    true,
				scriptEngine:      impostermodel2.ScriptEngineJavaScript,
				anchorFileName:    "order_service",
				copySpecs:         true,
				checkResponseFile: false,
			},
		},
		{
			name: "generate rest mock with resources no script",
			args: args{
				generateResources: true,
				forceOverwrite:    true,
				scriptEngine:      impostermodel2.ScriptEngineNone,
				anchorFileName:    "mock",
				copySpecs:         false,
				checkResponseFile: true,
			},
		},
		{
			name: "generate rest mock with resources with script",
			args: args{
				generateResources: true,
				forceOverwrite:    true,
				scriptEngine:      impostermodel2.ScriptEngineJavaScript,
				anchorFileName:    "mock",
				copySpecs:         false,
				checkResponseFile: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configDir, err := os.MkdirTemp(os.TempDir(), "specs")
			if err != nil {
				t.Fatal(err)
			}
			if tt.args.copySpecs {
				prepTestData(t, configDir, testConfigPath)
			}
			impostermodel2.Create(configDir, tt.args.generateResources, tt.args.forceOverwrite, tt.args.scriptEngine, false)

			if !doesFileExist(filepath.Join(configDir, tt.args.anchorFileName+"-config.yaml")) {
				t.Fatalf("imposter config file should exist")
			}
			if tt.args.checkResponseFile && !doesFileExist(filepath.Join(configDir, "response.json")) {
				t.Fatalf("response file should exist")
			}

			scriptPath := filepath.Join(configDir, tt.args.anchorFileName+".js")
			if impostermodel2.IsScriptEngineEnabled(tt.args.scriptEngine) {
				if !doesFileExist(scriptPath) {
					t.Fatalf("script file should exist")
				}
			} else {
				if doesFileExist(scriptPath) {
					t.Fatalf("script file should not exist")
				}
			}
		})
	}
}

func doesFileExist(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func prepTestData(t *testing.T, configDir string, src string) {
	err := fileutil.CopyDirShallow(src, configDir)
	if err != nil {
		t.Fatal(err)
	}
}
