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
