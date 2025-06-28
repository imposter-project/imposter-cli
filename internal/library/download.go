package library

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

type DownloadConfig struct {
	LatestBaseUrlTemplate    string
	VersionedBaseUrlTemplate string
}

// DownloadBinaryWithConfig downloads a binary file from a remote URL based on the provided configuration.
// It saves the file to the specified local path, using a temporary file during the download process.
func DownloadBinaryWithConfig(
	config DownloadConfig,
	localPath string,
	remoteFileName string,
	version string,
	fallbackRemoteFileName string,
) error {
	logger.Tracef("attempting to download %s version %s to %s", remoteFileName, version, localPath)
	tempFile, err := os.CreateTemp(os.TempDir(), path.Base(localPath))
	if err != nil {
		return fmt.Errorf("error creating temp file: %v: %v", localPath, err)
	}

	defer func() {
		_ = tempFile.Close()
		tempFilePath := tempFile.Name()
		if _, err := os.Stat(tempFilePath); err == nil {
			logger.Tracef("removing temp file: %s", tempFilePath)
			_ = os.Remove(tempFilePath)
		}
	}()

	url, resp, err := getHttpResponse(config, version, remoteFileName, fallbackRemoteFileName)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("error downloading from: %v: status code: %d", url, resp.StatusCode)
	}

	written, err := io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to temp file: %v: %v", tempFile.Name(), err)
	}
	if written == 0 {
		return fmt.Errorf("no data written to temp file: %v", tempFile.Name())
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("error closing temp file: %v: %v", tempFile.Name(), err)
	}
	if err := os.Rename(tempFile.Name(), localPath); err != nil {
		return fmt.Errorf("error renaming temp file to final destination: %v -> %v: %v", tempFile.Name(), localPath, err)
	}
	return err
}

// getHttpResponse constructs the URL based on the version and attempts to make an HTTP request.
// If the version is "latest", it uses the latest base URL template.
// If the version is not "latest", it uses the versioned base URL template.
// If the response status code is 404 and a fallback remote file name is provided, it retries with the fallback name.
// Returns the URL, the HTTP response, and any error encountered.
func getHttpResponse(
	config DownloadConfig,
	version string,
	remoteFileName string,
	fallbackRemoteFileName string,
) (url string, resp *http.Response, err error) {
	if version == "latest" {
		url = config.LatestBaseUrlTemplate + "/" + remoteFileName
		resp, err = makeHttpRequest(url, err)
		if err != nil {
			return "", nil, err
		}

	} else {
		versionedBaseUrl := fmt.Sprintf(config.VersionedBaseUrlTemplate, version)

		url = versionedBaseUrl + "/" + remoteFileName
		resp, err = makeHttpRequest(url, err)
		if err != nil {
			return "", nil, err
		}

		// fallback to versioned binary filename
		if resp.StatusCode == 404 && fallbackRemoteFileName != "" {
			logger.Tracef("binary not found at: %v - retrying with fallback filename", url)
			url = versionedBaseUrl + "/" + fallbackRemoteFileName
			resp, err = makeHttpRequest(url, err)
			if err != nil {
				return "", nil, err
			}
		}
	}
	return url, resp, nil
}
