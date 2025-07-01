package plugin

import (
	"fmt"
	"gatehill.io/imposter/internal/engine"
	library2 "gatehill.io/imposter/internal/library"
)

// downloadPlugin downloads the specified plugin for the given engine type and version.
func downloadPlugin(engineType engine.EngineType, pluginName string, version string) error {
	_, localFilePath, err := GetPluginLocalPath(pluginName, engineType, version)
	if err != nil {
		return fmt.Errorf("error determining local file path for plugin %s: %s", pluginName, err)
	}

	remoteFileName, err := getPluginRemoteFileName(engineType, pluginName)
	if err != nil {
		return fmt.Errorf("error determining remote file name for plugin %s: %s", pluginName, err)
	}

	pluginConfig := determinePluginConfig(engineType)
	downloadConfig := pluginConfig.downloadConfig
	err = library2.DownloadBinary(downloadConfig, localFilePath, remoteFileName, version)
	if err != nil {
		return err
	}

	logger.Infof("downloaded plugin %s version %s", pluginName, version)
	return nil
}
