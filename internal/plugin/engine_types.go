package plugin

import (
	"fmt"
	"gatehill.io/imposter/internal/engine"
	library2 "gatehill.io/imposter/internal/library"
	"gatehill.io/imposter/internal/stringutil"
	"runtime"
	"strings"
)

type pluginConfiguration struct {
	downloadConfig library2.DownloadConfig
	fileNamePrefix string

	// addOsAndArch indicates whether the plugin file name should include the OS and architecture.
	addOsAndArch bool

	// extensions is the list of supported file extensions.
	// note that the first extension is the default.
	extensions []string
}

var pluginConfigs = map[string]pluginConfiguration{
	"golang": {
		downloadConfig: library2.NewDownloadConfig(
			"https://github.com/imposter-project/imposter-go-plugins/releases/latest/download",
			"https://github.com/imposter-project/imposter-go-plugins/releases/download/v%v",
			true,
		),
		extensions:     []string{".zip"},
		fileNamePrefix: "plugin-",
		addOsAndArch:   true,
	},
	"*": {
		downloadConfig: library2.NewDownloadConfig(
			"https://github.com/imposter-project/imposter-jvm-engine/releases/latest/download",
			"https://github.com/imposter-project/imposter-jvm-engine/releases/download/v%v",
			false,
		),
		extensions:     []string{".jar", ".zip"},
		fileNamePrefix: "imposter-plugin-",
		addOsAndArch:   false,
	},
}

// determinePluginConfig returns the plugin configuration based on the engine type.
func determinePluginConfig(engineType engine.EngineType) pluginConfiguration {
	switch engineType {
	case engine.EngineTypeGolang:
		return pluginConfigs["golang"]
	default:
		return pluginConfigs["*"]
	}
}

// isValidPluginFile checks if the given file path is a valid plugin file
// for the specified engine type. It returns true if valid, along with the plugin name.
func isValidPluginFile(candidateFilePath string, engineType engine.EngineType) (bool, string) {
	pluginConfig := determinePluginConfig(engineType)
	if !strings.HasPrefix(candidateFilePath, pluginConfig.fileNamePrefix) {
		return false, ""
	}
	pluginName := strings.TrimPrefix(candidateFilePath, pluginConfig.fileNamePrefix)

	if len(pluginConfig.extensions) > 1 {
		supportedSuffix := stringutil.GetMatchingSuffix(pluginName, pluginConfig.extensions)
		if supportedSuffix == "" {
			return false, ""
		}
		pluginName = strings.TrimSuffix(pluginName, supportedSuffix)
	}

	return true, pluginName
}

// getFullPluginFileName returns the full plugin file name based on the engine type and plugin name.
func getFullPluginFileName(engineType engine.EngineType, pluginName string) (string, error) {
	pluginConfig := determinePluginConfig(engineType)
	switch len(pluginConfig.extensions) {
	case 0:
		return "", fmt.Errorf("plugin extensions not specified for engine type: " + string(engineType))

	case 1:
		fullPluginFileName := buildPluginFileName(pluginConfig, pluginName, pluginConfig.extensions[0])
		return fullPluginFileName, nil

	default:
		// Multiple extensions are supported, so we need to check the plugin name
		if strings.Contains(pluginName, ":") {
			// The format is indicated by the presence of a colon in the plugin name.
			// JVM/Docker archive format plugins use .zip extension, supported in JVM/Docker engine since v3.35.0,
			// as well as the default .jar extension.
			for _, ext := range pluginConfig.extensions {
				extAsSuffix := ":" + ext[1:]
				if strings.HasSuffix(pluginName, extAsSuffix) {
					trimmedPluginName := strings.TrimSuffix(pluginName, extAsSuffix)
					fullPluginFileName := buildPluginFileName(pluginConfig, trimmedPluginName, ext)
					return fullPluginFileName, nil
				}
			}
			return "", fmt.Errorf("no matching plugin extension found for engine type: " + string(engineType) + " and plugin name: " + pluginName)

		} else {
			// use the default extension
			fullPluginFileName := buildPluginFileName(pluginConfig, pluginName, pluginConfig.extensions[0])
			return fullPluginFileName, nil
		}
	}
}

func buildPluginFileName(pluginConfig pluginConfiguration, pluginName string, ext string) string {
	osAndArch := ""
	if pluginConfig.addOsAndArch {
		osAndArch = fmt.Sprintf("_%s_%s", runtime.GOOS, runtime.GOARCH)
	}
	fullPluginFileName := fmt.Sprintf("%s%s%s%s", pluginConfig.fileNamePrefix, pluginName, osAndArch, ext)
	return fullPluginFileName
}
