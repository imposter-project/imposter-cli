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

package impostermodel

import (
	"gatehill.io/imposter/internal/fileutil"
	"gatehill.io/imposter/internal/logging"
	"gatehill.io/imposter/internal/openapi"
	"os"
	"path"
	"path/filepath"
	"sigs.k8s.io/yaml"
)

type ConfigGenerationOptions struct {
	PluginName     string
	ScriptEngine   ScriptEngine
	ScriptFileName string
	SpecFilePath   string
}

var logger = logging.GetLogger()

func Create(configDir string, generateResources bool, forceOverwrite bool, scriptEngine ScriptEngine, requireOpenApi bool) {
	openApiSpecs := openapi.DiscoverOpenApiSpecs(configDir)
	logger.Infof("found %d OpenAPI spec(s)", len(openApiSpecs))

	if len(openApiSpecs) > 0 {
		logger.Tracef("using openapi plugin")
		for _, openApiSpec := range openApiSpecs {
			scriptFileName := getScriptFileName(openApiSpec, scriptEngine, forceOverwrite)
			writeOpenapiMockConfig(openApiSpec, generateResources, forceOverwrite, scriptEngine, scriptFileName)
		}
	} else if !requireOpenApi {
		logger.Infof("falling back to rest plugin")
		syntheticMockPath := path.Join(configDir, "mock.txt")
		_, responseFilePath := generateRestMockFiles(configDir)
		scriptFileName := getScriptFileName(syntheticMockPath, scriptEngine, forceOverwrite)
		writeRestMockConfig(syntheticMockPath, responseFilePath, generateResources, forceOverwrite, scriptEngine, scriptFileName)
	} else {
		logger.Fatalf("no OpenAPI specs found in: %s", configDir)
	}
}

func GenerateConfig(options ConfigGenerationOptions, resources []Resource) []byte {
	pluginConfig := PluginConfig{
		Plugin: options.PluginName,
	}
	if options.SpecFilePath != "" {
		pluginConfig.SpecFile = filepath.Base(options.SpecFilePath)
	}
	if len(resources) > 0 {
		pluginConfig.Resources = resources
	} else {
		if IsScriptEngineEnabled(options.ScriptEngine) {
			pluginConfig.Response = &ResponseConfig{
				ScriptFile: options.ScriptFileName,
			}
		}
	}

	config, err := yaml.Marshal(pluginConfig)
	if err != nil {
		logger.Fatalf("unable to marshal imposter config: %v", err)
	}
	return config
}

func writeMockConfigAdjacent(anchorFilePath string, resources []Resource, forceOverwrite bool, options ConfigGenerationOptions) {
	configFilePath := fileutil.GenerateFilePathAdjacentToFile(anchorFilePath, "-config.yaml", forceOverwrite)
	writeMockConfig(configFilePath, resources, forceOverwrite, options)
}

func writeMockConfig(configFilePath string, resources []Resource, forceOverwrite bool, options ConfigGenerationOptions) {
	configFile, err := os.Create(configFilePath)
	if err != nil {
		logger.Fatal(err)
	}
	defer configFile.Close()

	config := GenerateConfig(options, resources)
	_, err = configFile.Write(config)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Infof("wrote Imposter config: %v", configFilePath)
}
