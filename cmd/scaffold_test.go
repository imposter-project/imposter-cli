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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/imposter-project/imposter-cli/internal/fileutil"
	impostermodel2 "github.com/imposter-project/imposter-cli/internal/impostermodel"
	"github.com/sirupsen/logrus"
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
		copyWsdl          bool
		copyProto         bool
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
		{
			name: "generate wsdl mock with resources no script",
			args: args{
				generateResources: true,
				forceOverwrite:    true,
				scriptEngine:      impostermodel2.ScriptEngineNone,
				anchorFileName:    "pet_service",
				copyWsdl:          true,
				checkResponseFile: false,
			},
		},
		{
			name: "generate wsdl mock with resources with script",
			args: args{
				generateResources: true,
				forceOverwrite:    true,
				scriptEngine:      impostermodel2.ScriptEngineJavaScript,
				anchorFileName:    "pet_service",
				copyWsdl:          true,
				checkResponseFile: false,
			},
		},
		{
			name: "generate wsdl mock no resources no script",
			args: args{
				generateResources: false,
				forceOverwrite:    true,
				scriptEngine:      impostermodel2.ScriptEngineNone,
				anchorFileName:    "pet_service",
				copyWsdl:          true,
				checkResponseFile: false,
			},
		},
		{
			name: "generate grpc mock no script",
			args: args{
				generateResources: true,
				forceOverwrite:    true,
				scriptEngine:      impostermodel2.ScriptEngineNone,
				anchorFileName:    "pet_store",
				copyProto:         true,
				checkResponseFile: false,
			},
		},
		{
			name: "generate grpc mock with script",
			args: args{
				generateResources: true,
				forceOverwrite:    true,
				scriptEngine:      impostermodel2.ScriptEngineJavaScript,
				anchorFileName:    "pet_store",
				copyProto:         true,
				checkResponseFile: false,
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
			if tt.args.copyWsdl {
				err = fileutil.CopyFile(filepath.Join(testConfigPath, "pet_service.wsdl"), filepath.Join(configDir, "pet_service.wsdl"))
				if err != nil {
					t.Fatal(err)
				}
			}
			if tt.args.copyProto {
				err = fileutil.CopyFile(filepath.Join(workingDir, "testdata_grpc", "pet_store.proto"), filepath.Join(configDir, "pet_store.proto"))
				if err != nil {
					t.Fatal(err)
				}
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

			dotImposterPath := filepath.Join(configDir, ".imposter.yaml")
			if !doesFileExist(dotImposterPath) {
				t.Fatalf(".imposter.yaml should exist")
			}
			dotImposterContent, err := os.ReadFile(dotImposterPath)
			if err != nil {
				t.Fatal(err)
			}
			content := string(dotImposterContent)
			if tt.args.copyProto {
				if !strings.Contains(content, "version: 5-beta") {
					t.Fatalf(".imposter.yaml should contain version: 5-beta for grpc, got:\n%s", content)
				}
				if !strings.Contains(content, "- grpc") {
					t.Fatalf(".imposter.yaml should contain grpc plugin for grpc, got:\n%s", content)
				}
			} else {
				if !strings.Contains(content, "version: latest") {
					t.Fatalf(".imposter.yaml should contain version: latest, got:\n%s", content)
				}
			}
			if !strings.Contains(content, "IMPOSTER_LOG_LEVEL: DEBUG") {
				t.Fatalf(".imposter.yaml should contain IMPOSTER_LOG_LEVEL: DEBUG, got:\n%s", content)
			}
			if !strings.Contains(content, "# or pin to a particular version") {
				t.Fatalf(".imposter.yaml should contain version comment, got:\n%s", content)
			}
			if !strings.Contains(content, "# See https://docs.imposter.sh/environment_variables/") {
				t.Fatalf(".imposter.yaml should contain env docs comment, got:\n%s", content)
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
