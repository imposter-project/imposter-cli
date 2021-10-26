/*
Copyright © 2021 Pete Cornish <outofcoffee@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package jvm

import (
	"fmt"
	"gatehill.io/imposter/engine"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const binCacheDir = ".imposter/cache/"
const latestUrl = "https://github.com/outofcoffee/imposter/releases/latest/download/imposter.jar"
const versionedBaseUrlTemplate = "https://github.com/outofcoffee/imposter/releases/download/v%v/"

type EngineJarProvider struct {
	engine.ProviderOptions
	jarPath string
}

func GetProvider(version string) *EngineJarProvider {
	return &EngineJarProvider{
		ProviderOptions: engine.ProviderOptions{
			EngineType: engine.EngineTypeJvm,
			Version:    version,
		},
	}
}

func (d *EngineJarProvider) Provide(policy engine.PullPolicy) error {
	jarPath, err := ensureBinary(d.Version, policy)
	if err != nil {
		return err
	}
	d.jarPath = jarPath
	return nil
}

func (d *EngineJarProvider) Satisfied() bool {
	return d.jarPath != ""
}

func (d *EngineJarProvider) GetEngineType() engine.EngineType {
	return d.EngineType
}

func ensureBinary(version string, policy engine.PullPolicy) (string, error) {
	err, binCachePath := ensureBinCache()
	if err != nil {
		logrus.Fatal(err)
	}

	binFilePath := filepath.Join(binCachePath, fmt.Sprintf("imposter-%v.jar", version))
	if policy == engine.PullSkip {
		return binFilePath, nil
	}

	if policy == engine.PullIfNotPresent {
		if _, err = os.Stat(binFilePath); err != nil {
			if !os.IsNotExist(err) {
				return "", fmt.Errorf("failed to stat: %v: %v", binFilePath, err)
			}
		} else {
			logrus.Debugf("engine binary '%v' already present", version)
			logrus.Tracef("binary for version %v found at: %v", version, binFilePath)
			return binFilePath, nil
		}
	}

	if err := downloadBinary(binFilePath, version); err != nil {
		return "", fmt.Errorf("failed to fetch binary: %v", err)
	}
	logrus.Tracef("using imposter at: %v", binFilePath)
	return binFilePath, nil
}

func ensureBinCache() (error, string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %v", err), ""
	}

	binCachePath := filepath.Join(homeDir, binCacheDir)
	if _, err = os.Stat(binCachePath); err != nil {
		if os.IsNotExist(err) {
			logrus.Tracef("creating cache directory: %v", binCachePath)
			err := os.MkdirAll(binCachePath, 0700)
			if err != nil {
				return fmt.Errorf("failed to create cache directory: %v: %v", binCachePath, err), ""
			}
		} else {
			return fmt.Errorf("failed to stat: %v: %v", binCachePath, err), ""
		}
	}

	logrus.Tracef("ensured binary cache directory: %v", binCachePath)
	return nil, binCachePath
}

func downloadBinary(localPath string, version string) error {
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("error creating file: %v: %v", localPath, err)
	}
	defer file.Close()

	var url string
	var resp *http.Response
	if version == "latest" {
		url = latestUrl
		resp, err = makeHttpRequest(url, err)
		if err != nil {
			return err
		}

	} else {
		versionedBaseUrl := fmt.Sprintf(versionedBaseUrlTemplate, version)

		url := versionedBaseUrl + "imposter.jar"
		resp, err = makeHttpRequest(url, err)
		if err != nil {
			return err
		}

		// fallback to versioned binary filename
		if resp.StatusCode == 404 {
			logrus.Tracef("binary not found at: %v - retrying with versioned filename", url)
			url = versionedBaseUrl + fmt.Sprintf("imposter-%v.jar", version)
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

func makeHttpRequest(url string, err error) (*http.Response, error) {
	logrus.Debugf("downloading %v", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error downloading from: %v: %v", url, err)
	}
	return resp, nil
}

func GetJavaCmdPath() (string, error) {
	var binaryPathSuffix string
	if runtime.GOOS == "Windows" {
		binaryPathSuffix = ".exe"
	} else {
		binaryPathSuffix = ""
	}

	// prefer JAVA_HOME environment variable
	if javaHomeEnv, found := os.LookupEnv("JAVA_HOME"); found {
		return filepath.Join(javaHomeEnv, "/bin/java"+binaryPathSuffix), nil
	}

	if runtime.GOOS == "darwin" {
		command, stdout := exec.Command("/usr/libexec/java_home"), new(strings.Builder)
		command.Stdout = stdout
		err := command.Run()
		if err != nil {
			return "", fmt.Errorf("error determining JAVA_HOME: %v", err)
		}
		if command.ProcessState.Success() {
			return filepath.Join(strings.TrimSpace(stdout.String()), "/bin/java"+binaryPathSuffix), nil
		} else {
			return "", fmt.Errorf("failed to determine JAVA_HOME using libexec")
		}
	}

	// search for 'java' in the PATH
	javaPath, err := exec.LookPath("java")
	if err != nil {
		return "", fmt.Errorf("could not find 'java' in PATH: %v", err)
	}
	logrus.Tracef("using java: %v", javaPath)
	return javaPath, nil
}
