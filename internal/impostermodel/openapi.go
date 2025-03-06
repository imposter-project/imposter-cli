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
	"gatehill.io/imposter/internal/openapi"
	"sort"
	"strconv"
	"strings"
)

type ResourceGenerationOptions struct {
	ScriptEngine   ScriptEngine
	ScriptFileName string
}

func writeOpenapiMockConfig(specFilePath string, generateResources bool, forceOverwrite bool, scriptEngine ScriptEngine, scriptFileName string) {
	var resources []Resource
	if generateResources {
		resources = buildOpenapiResources(specFilePath, scriptEngine, scriptFileName)
	} else {
		logger.Debug("skipping resource generation")
	}
	options := ConfigGenerationOptions{
		PluginName:     "openapi",
		ScriptEngine:   scriptEngine,
		ScriptFileName: scriptFileName,
		SpecFilePath:   specFilePath,
	}
	writeMockConfigAdjacent(specFilePath, resources, forceOverwrite, options)
}

func buildOpenapiResources(specFilePath string, scriptEngine ScriptEngine, scriptFileName string) []Resource {
	resources := GenerateResourcesFromSpec(specFilePath, ResourceGenerationOptions{
		ScriptEngine:   scriptEngine,
		ScriptFileName: scriptFileName,
	})
	logger.Debugf("generated %d resources from spec", len(resources))
	return resources
}

func GenerateResourcesFromSpec(specFilePath string, options ResourceGenerationOptions) []Resource {
	var resources []Resource
	partialSpec, err := openapi.Parse(specFilePath)
	if err != nil {
		logger.Fatalf("unable to parse openapi spec: %v: %v", specFilePath, err)
	}
	if partialSpec != nil {
		for path, pathDetail := range partialSpec.Paths {
			for verb, resp := range pathDetail {
				resource := Resource{
					Path:   path,
					Method: strings.ToUpper(verb),
					Response: &ResponseConfig{
						StatusCode: chooseOpStatusCode(resp),
					},
				}
				if IsScriptEngineEnabled(options.ScriptEngine) {
					resource.Response.ScriptFile = options.ScriptFileName
				}
				resources = append(resources, resource)
			}
		}

	}
	return resources
}

func chooseOpStatusCode(resp openapi.Operation) int {
	if len(resp.Responses) == 0 {
		logger.Tracef("no responses found for openapi operation - guessing 200 status code")
		return 200
	}
	var statusCodes []int
	for statusCode := range resp.Responses {
		if sc, err := strconv.Atoi(statusCode); err == nil && sc >= 200 {
			statusCodes = append(statusCodes, sc)
		}
	}
	sort.Ints(statusCodes)
	if len(statusCodes) > 0 {
		return statusCodes[0]
	}

	logger.Tracef("unable to determine status code found for openapi operation - guessing 200")
	return 200
}
