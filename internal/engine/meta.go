package engine

// getRepoNameForEngineType returns the GitHub repository name for the given engine type.
func getRepoNameForEngineType(engineType EngineType) string {
	switch engineType {
	case EngineTypeNative:
		return "imposter-go"
	default:
		return "imposter-jvm-engine"
	}
}

// getRepoNameForMajor returns the GitHub repository name releasing the given
// major version of the engine: 5 is the native engine and 4 is the JVM engine.
// This is independent of the configured engine type, as some engine types
// (such as awslambda) are built from both repositories.
func getRepoNameForMajor(major int64) string {
	if major >= 5 {
		return "imposter-go"
	}
	return "imposter-jvm-engine"
}
