package impostermodel

func writeGrpcMockConfig(protoFilePath string, forceOverwrite bool, scriptEngine ScriptEngine, scriptFileName string) {
	options := ConfigGenerationOptions{
		PluginName:     "grpc",
		ScriptEngine:   scriptEngine,
		ScriptFileName: scriptFileName,
		ProtoFilePath:  protoFilePath,
	}
	writeMockConfigAdjacent(protoFilePath, nil, forceOverwrite, options)
}
