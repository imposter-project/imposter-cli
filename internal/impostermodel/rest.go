/*
Copyright © 2022 Pete Cornish <outofcoffee@gmail.com>

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

package impostermodel

import (
	"os"
	"path/filepath"
)

// generateRestMockFiles creates files for a rest mock, and returns
// the full path to the response file.
func generateRestMockFiles(configDir string) (readmeFilePath string, responseFilePath string) {
	return generateReadmeFile(configDir), generateResponseFile(configDir)
}

func generateReadmeFile(configDir string) string {
	readmeFile := filepath.Join(configDir, "README.md")
	configFile, err := os.Create(readmeFile)
	if err != nil {
		logger.Fatal(err)
	}
	defer configFile.Close()
	_, err = configFile.WriteString(`Imposter REST mock

Start the mock with:

    imposter up

The mock will be accessible at: http://localhost:8080

Example files:

- response.json
- mock-config.yaml
`)
	if err != nil {
		logger.Fatalf("failed to write readme file: %s: %s", readmeFile, err)
	}
	logger.Debugf("wrote readme file: %v", readmeFile)
	return readmeFile
}

func generateResponseFile(configDir string) string {
	responseFile := filepath.Join(configDir, "response.json")
	configFile, err := os.Create(responseFile)
	if err != nil {
		logger.Fatal(err)
	}
	defer configFile.Close()
	_, err = configFile.WriteString("{ \"hello\": \"world\" }\n")
	if err != nil {
		logger.Fatalf("failed to write response file: %s: %s", responseFile, err)
	}
	logger.Debugf("wrote response file: %v", responseFile)
	return responseFile
}

func writeRestMockConfig(mockConfigPath string, responseFilePath string, generateResources bool, forceOverwrite bool, scriptEngine ScriptEngine, scriptFileName string) {
	var resources []Resource
	if generateResources {
		resources = buildRestResources(responseFilePath, scriptEngine, scriptFileName)
	} else {
		logger.Debug("skipping resource generation")
	}
	options := ConfigGenerationOptions{
		PluginName:     "rest",
		ScriptEngine:   scriptEngine,
		ScriptFileName: scriptFileName,
	}
	writeMockConfigAdjacent(mockConfigPath, resources, forceOverwrite, options)
}

func buildRestResources(responseFilePath string, scriptEngine ScriptEngine, scriptFileName string) []Resource {
	resource := Resource{
		Path:   "/",
		Method: "GET",
		Response: &ResponseConfig{
			StatusCode: 200,
			StaticFile: filepath.Base(responseFilePath),
		},
	}
	if IsScriptEngineEnabled(scriptEngine) {
		resource.Response.ScriptFile = scriptFileName
	}
	return []Resource{resource}
}
