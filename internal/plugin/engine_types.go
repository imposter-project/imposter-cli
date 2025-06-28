package plugin

import (
	"gatehill.io/imposter/internal/engine"
	library2 "gatehill.io/imposter/internal/library"
	"gatehill.io/imposter/internal/stringutil"
	"strings"
)

// determineDownloadConfig returns the appropriate download configuration
// based on the engine type. For Golang plugins, it uses the Golang plugin repository,
// and extracts compressed plugins, while for others, it uses the
// JVM engine repository and does not extract compressed plugins.
func determineDownloadConfig(engineType engine.EngineType) library2.DownloadConfig {
	if engineType == engine.EngineTypeGolang {
		return library2.NewDownloadConfig(
			"https://github.com/imposter-project/imposter-go-plugins/releases/latest/download",
			"https://github.com/imposter-project/imposter-go-plugins/releases/download/v%v",
			true,
		)
	} else {
		return library2.NewDownloadConfig(
			"https://github.com/imposter-project/imposter-jvm-engine/releases/latest/download",
			"https://github.com/imposter-project/imposter-jvm-engine/releases/download/v%v",
			false,
		)
	}
}

// isValidPluginFile checks if the given file path is a valid plugin file
// for the specified engine type. It returns true if valid, along with the plugin name.
func isValidPluginFile(candidateFilePath string, engineType engine.EngineType) (bool, string) {
	pluginFileNamePrefix := determinePluginFileNamePrefix(engineType)
	if !strings.HasPrefix(candidateFilePath, pluginFileNamePrefix) {
		return false, ""
	}
	pluginName := strings.TrimPrefix(candidateFilePath, pluginFileNamePrefix)

	if engineType != engine.EngineTypeGolang {
		supportedPluginExtensions := []string{".jar", ".zip"}
		supportedSuffix := stringutil.GetMatchingSuffix(pluginName, supportedPluginExtensions)
		if supportedSuffix == "" {
			return false, ""
		}
		pluginName = strings.TrimSuffix(pluginName, supportedSuffix)
	}

	return true, pluginName
}

// determinePluginFileNamePrefix returns the prefix for plugin files based on the engine type.
func determinePluginFileNamePrefix(engineType engine.EngineType) string {
	if engineType == engine.EngineTypeGolang {
		return "plugin-"
	} else {
		return "imposter-plugin-"
	}
}
