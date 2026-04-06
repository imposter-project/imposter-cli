package impostermodel

import (
	wsdlparser "github.com/outofcoffee/go-wsdl-parser"
)

func writeWsdlMockConfig(wsdlFilePath string, generateResources bool, forceOverwrite bool, scriptEngine ScriptEngine, scriptFileName string) {
	var resources []Resource
	if generateResources {
		resources = buildWsdlResources(wsdlFilePath, scriptEngine, scriptFileName)
	} else {
		logger.Debug("skipping resource generation")
	}
	options := ConfigGenerationOptions{
		PluginName:     "soap",
		ScriptEngine:   scriptEngine,
		ScriptFileName: scriptFileName,
		WSDLFilePath:   wsdlFilePath,
	}
	writeMockConfigAdjacent(wsdlFilePath, resources, forceOverwrite, options)
}

func buildWsdlResources(wsdlFilePath string, scriptEngine ScriptEngine, scriptFileName string) []Resource {
	parser, err := wsdlparser.NewWSDLParser(wsdlFilePath)
	if err != nil {
		logger.Fatalf("unable to parse WSDL file: %v: %v", wsdlFilePath, err)
	}

	var resources []Resource
	for _, op := range parser.GetOperations() {
		resource := Resource{
			Method:    "POST",
			Operation: op.Name,
			Response: &ResponseConfig{
				StatusCode: 200,
			},
		}
		if IsScriptEngineEnabled(scriptEngine) {
			resource.Steps = &[]StepConfig{{Type: StepTypeScript, File: scriptFileName}}
		}
		resources = append(resources, resource)
	}

	logger.Debugf("generated %d resources from WSDL", len(resources))
	return resources
}
