package library

import (
	"fmt"
	"net/http"
)

var defaultConfig = DownloadConfig{
	LatestBaseUrlTemplate:    "https://github.com/outofcoffee/imposter/releases/latest/download",
	VersionedBaseUrlTemplate: "https://github.com/outofcoffee/imposter/releases/download/v%v",
}

func DownloadBinary(localPath string, remoteFileName string, version string) error {
	return DownloadBinaryWithFallback(localPath, remoteFileName, version, "")
}

func DownloadBinaryWithFallback(localPath string, remoteFileName string, version string, fallbackRemoteFileName string) error {
	return DownloadBinaryWithConfig(defaultConfig, localPath, remoteFileName, version, fallbackRemoteFileName)
}

func makeHttpRequest(url string, err error) (*http.Response, error) {
	logger.Debugf("downloading %v", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error downloading from: %v: %v", url, err)
	}
	return resp, nil
}
