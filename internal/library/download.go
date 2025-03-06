package library

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

type DownloadConfig struct {
	LatestBaseUrlTemplate    string
	VersionedBaseUrlTemplate string
}

func DownloadBinaryWithConfig(config DownloadConfig, localPath string, remoteFileName string, version string, fallbackRemoteFileName string) error {
	logger.Tracef("attempting to download %s version %s to %s", remoteFileName, version, localPath)
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("error creating file: %v: %v", localPath, err)
	}
	defer func() {
		_ = file.Close()
		if stat, err := os.Stat(localPath); err == nil && stat.Size() == 0 {
			logger.Tracef("removing empty file: %s", localPath)
			_ = os.Remove(localPath)
		}
	}()

	var url string
	var resp *http.Response
	if version == "latest" {
		url = config.LatestBaseUrlTemplate + "/" + remoteFileName
		resp, err = makeHttpRequest(url, err)
		if err != nil {
			return err
		}

	} else {
		versionedBaseUrl := fmt.Sprintf(config.VersionedBaseUrlTemplate, version)

		url := versionedBaseUrl + "/" + remoteFileName
		resp, err = makeHttpRequest(url, err)
		if err != nil {
			return err
		}

		// fallback to versioned binary filename
		if resp.StatusCode == 404 && fallbackRemoteFileName != "" {
			logger.Tracef("binary not found at: %v - retrying with fallback filename", url)
			url = versionedBaseUrl + "/" + fallbackRemoteFileName
			resp, err = makeHttpRequest(url, err)
			if err != nil {
				return err
			}
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("error downloading from: %v: status code: %d", url, resp.StatusCode)
	}
	defer resp.Body.Close()
	_, err = io.Copy(file, resp.Body)
	return err
}
