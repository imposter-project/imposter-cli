package plugin

import (
	"fmt"
	"gatehill.io/imposter/internal/engine"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"
)

type pluginFileTemplate struct {
	PluginName string
	Ext        string
	OS         string
	Arch       string
}

// GetPluginLocalPath returns the plugin file name and the
// local file path for the specified plugin name, engine type, and version.
func GetPluginLocalPath(
	pluginName string,
	engineType engine.EngineType,
	version string,
) (fileName string, localPath string, err error) {
	pluginDir, err := EnsurePluginDir(version)
	if err != nil {
		return "", "", err
	}

	fileName, err = getPluginFileName(engineType, pluginName, false)
	if err != nil {
		return "", "", fmt.Errorf("error determining plugin local file name for %s: %s", engineType, err)
	}

	localPath = filepath.Join(pluginDir, fileName)
	return fileName, localPath, err
}

// getPluginRemoteFileName returns the remote file name for the specified plugin
func getPluginRemoteFileName(engineType engine.EngineType, pluginName string) (string, error) {
	fileName, err := getPluginFileName(engineType, pluginName, true)
	if err != nil {
		return "", fmt.Errorf("error determining plugin remote file name for %s: %s", engineType, err)
	}
	return fileName, nil
}

// isValidPluginFile checks if the given file path is a valid plugin file
// for the specified engine type. It returns true if valid, along with the plugin name.
func isValidPluginFile(candidateFilePath string, engineType engine.EngineType) (bool, string) {
	pluginConfig := determinePluginConfig(engineType)

	for _, ext := range pluginConfig.extensions {
		fileBasePattern := pluginConfig.localFileTemplate
		fileBasePattern = strings.ReplaceAll(fileBasePattern, "{{.PluginName}}", "([a-zA-Z0-9_-]+)")
		fileBasePattern = strings.ReplaceAll(fileBasePattern, "{{.Ext}}", ext)
		fileBasePattern = strings.ReplaceAll(fileBasePattern, "{{.OS}}", runtime.GOOS)
		fileBasePattern = strings.ReplaceAll(fileBasePattern, "{{.Arch}}", runtime.GOARCH)

		if matched, _ := regexp.Compile(fileBasePattern); matched != nil {
			matchedFile := matched.FindStringSubmatch(candidateFilePath)
			if len(matchedFile) > 1 {
				return true, matchedFile[1]
			} else {
				return false, ""
			}
		} else {
			return false, ""
		}
	}
	return false, ""
}

// getPluginFileName returns the plugin file name based on the engine type and plugin name.
func getPluginFileName(engineType engine.EngineType, pluginName string, remote bool) (string, error) {
	pluginConfig := determinePluginConfig(engineType)
	switch len(pluginConfig.extensions) {
	case 0:
		return "", fmt.Errorf("plugin extensions not specified for engine type: " + string(engineType))

	case 1:
		fileName := buildPluginFileName(pluginConfig, pluginName, pluginConfig.extensions[0], remote)
		return fileName, nil

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
					fileName := buildPluginFileName(pluginConfig, trimmedPluginName, ext, remote)
					return fileName, nil
				}
			}
			return "", fmt.Errorf("no matching plugin extension found for engine type: " + string(engineType) + " and plugin name: " + pluginName)

		} else {
			// use the default extension
			fileName := buildPluginFileName(pluginConfig, pluginName, pluginConfig.extensions[0], remote)
			return fileName, nil
		}
	}
}

// buildPluginFileName builds the plugin file name based on the file name template
func buildPluginFileName(
	pluginConfig pluginConfiguration,
	pluginName string,
	ext string,
	remote bool,
) string {
	var fileTemplate string
	if remote {
		fileTemplate = pluginConfig.remoteFileTemplate
	} else {
		fileTemplate = pluginConfig.localFileTemplate
	}
	tmpl, err := template.New("pluginFileName").Parse(fileTemplate)
	if err != nil {
		panic(fmt.Errorf("error parsing plugin file name template: %s", err))
	}

	tmplData := pluginFileTemplate{
		PluginName: pluginName,
		Ext:        ext,
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
	}
	var fileNameBuilder strings.Builder
	if err = tmpl.Execute(&fileNameBuilder, tmplData); err != nil {
		panic(fmt.Errorf("error executing plugin file name template: %s", err))
	}
	fileName := fileNameBuilder.String()
	logger.Tracef("plugin file name [remote=%v]: %s", remote, fileName)
	return fileName
}
