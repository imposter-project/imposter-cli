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
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/imposter-project/imposter-cli/internal/fileutil"
	"github.com/imposter-project/imposter-cli/internal/logging"
	"github.com/imposter-project/imposter-cli/internal/openapi"
	"github.com/imposter-project/imposter-cli/internal/wsdl"
	"sigs.k8s.io/yaml"
)

type ConfigGenerationOptions struct {
	PluginName     string
	ScriptEngine   ScriptEngine
	ScriptFileName string
	SpecFilePath   string
	WSDLFilePath   string
}

var logger = logging.GetLogger()

func Create(configDir string, generateResources bool, forceOverwrite bool, scriptEngine ScriptEngine, requireSpecFiles bool) {
	openApiSpecs := openapi.DiscoverOpenApiSpecs(configDir)
	wsdlFiles := wsdl.DiscoverWSDLFiles(configDir)
	logger.Infof("found %d OpenAPI spec(s) and %d WSDL file(s)", len(openApiSpecs), len(wsdlFiles))

	specsFound := false

	if len(openApiSpecs) > 0 {
		specsFound = true
		logger.Tracef("using openapi plugin")
		for _, openApiSpec := range openApiSpecs {
			scriptFileName := getScriptFileName(openApiSpec, scriptEngine, forceOverwrite)
			writeOpenapiMockConfig(openApiSpec, generateResources, forceOverwrite, scriptEngine, scriptFileName)
		}
	}

	if len(wsdlFiles) > 0 {
		specsFound = true
		logger.Tracef("using soap plugin")
		for _, wsdlFile := range wsdlFiles {
			scriptFileName := getScriptFileName(wsdlFile, scriptEngine, forceOverwrite)
			writeWsdlMockConfig(wsdlFile, generateResources, forceOverwrite, scriptEngine, scriptFileName)
		}
	}

	if !specsFound {
		if !requireSpecFiles {
			logger.Infof("falling back to rest plugin")
			syntheticMockPath := path.Join(configDir, "mock.txt")
			_, responseFilePath := generateRestMockFiles(configDir)
			scriptFileName := getScriptFileName(syntheticMockPath, scriptEngine, forceOverwrite)
			writeRestMockConfig(syntheticMockPath, responseFilePath, generateResources, forceOverwrite, scriptEngine, scriptFileName)
		} else {
			logger.Fatalf("no OpenAPI or WSDL specs found in: %s", configDir)
		}
	}
}

func GenerateConfig(options ConfigGenerationOptions, resources []Resource) []byte {
	pluginConfig := PluginConfig{
		Plugin: options.PluginName,
	}
	if options.SpecFilePath != "" {
		pluginConfig.SpecFile = filepath.Base(options.SpecFilePath)
	}
	if options.WSDLFilePath != "" {
		pluginConfig.WSDLFile = filepath.Base(options.WSDLFilePath)
	}
	if len(resources) > 0 {
		pluginConfig.Resources = resources
	} else {
		if IsScriptEngineEnabled(options.ScriptEngine) {
			logger.Warn("script engine is enabled but no resources were present - skipping adding script step")
		}
	}

	config, err := yaml.Marshal(pluginConfig)
	if err != nil {
		logger.Fatalf("unable to marshal imposter config: %v", err)
	}
	return append([]byte(buildConfigHeader(options.PluginName)), config...)
}

// buildConfigHeader returns a YAML comment block describing the file and
// linking to the relevant docs.imposter.sh pages for the given plugin.
func buildConfigHeader(pluginName string) string {
	links := []string{
		"General configuration: https://docs.imposter.sh/configuration/",
	}
	switch pluginName {
	case "rest":
		links = append(links,
			"REST plugin: https://docs.imposter.sh/rest_plugin/",
			"Request matching: https://docs.imposter.sh/request_matching/",
		)
	case "openapi":
		links = append(links,
			"OpenAPI plugin: https://docs.imposter.sh/openapi_plugin/",
		)
	case "soap":
		links = append(links,
			"SOAP plugin: https://docs.imposter.sh/soap_plugin/",
		)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# Imposter mock configuration (%s plugin)\n", pluginName)
	b.WriteString("#\n")
	b.WriteString("# Reference docs:\n")
	for _, l := range links {
		fmt.Fprintf(&b, "#   - %s\n", l)
	}
	b.WriteString("#\n")
	b.WriteString("# Edit this file to customise your mock, then run: imposter up\n")
	b.WriteString("\n")
	return b.String()
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
