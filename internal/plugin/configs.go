package plugin

import (
	"gatehill.io/imposter/internal/engine"
	library2 "gatehill.io/imposter/internal/library"
)

type pluginConfiguration struct {
	downloadConfig library2.DownloadConfig

	// localFileTemplate uses gotpl format and supports the fields of pluginFileTemplate
	localFileTemplate string

	// localFileTemplate uses gotpl format and supports the fields of pluginFileTemplate
	remoteFileTemplate string

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
		extensions: []string{".zip"},
		// note we use .exe not .Ext because the template represents the binary inside the archive
		localFileTemplate:  `plugin-{{ .PluginName }}{{if eq .OS "windows"}}.exe{{end}}`,
		remoteFileTemplate: `plugin-{{ .PluginName }}_{{ .OS }}_{{ .Arch }}{{ .Ext }}`,
	},
	"*": {
		downloadConfig: library2.NewDownloadConfig(
			"https://github.com/imposter-project/imposter-jvm-engine/releases/latest/download",
			"https://github.com/imposter-project/imposter-jvm-engine/releases/download/v%v",
			false,
		),
		extensions:         []string{".jar", ".zip"},
		localFileTemplate:  `imposter-plugin-{{ .PluginName }}{{ .Ext }}`,
		remoteFileTemplate: `imposter-plugin-{{ .PluginName }}{{ .Ext }}`,
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
