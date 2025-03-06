/*
Copyright © 2023 Pete Cornish <outofcoffee@gmail.com>

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

package awslambda

import (
	"archive/zip"
	"bytes"
	"fmt"
	"gatehill.io/imposter/internal/fileutil"
	"io"
	"os"
	"path"
	"strings"
)

func (p *LambdaProvider) Bundle(configDir string, dest string) error {
	deploymentPackage, err := CreateDeploymentPackage(p.Version, configDir)
	if err != nil {
		return fmt.Errorf("failed to create bundle: %v", err)
	}

	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("destination bundle file already exists: %s", dest)
	}

	err = os.WriteFile(dest, *deploymentPackage, 0644)
	if err != nil {
		return fmt.Errorf("error writing bundle file: %s: %v", dest, err)
	}
	return nil
}

func CreateDeploymentPackage(version string, dir string) (*[]byte, error) {
	binaryPath, err := checkOrDownloadBinary(version)
	if err != nil {
		return nil, err
	}
	local, err := fileutil.ListFiles(dir, false)
	if err != nil {
		return nil, err
	}
	pkg, err := addFilesToZip(binaryPath, local)
	if err != nil {
		return nil, err
	}
	logger.Debug("created deployment package")
	contents := pkg.Bytes()
	return &contents, nil
}

func addFilesToZip(zipPath string, files []string) (*bytes.Buffer, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source zip: %s: %v", zipPath, err)
	}
	defer zr.Close()

	dst := new(bytes.Buffer)
	zw := zip.NewWriter(dst)
	defer zw.Close()

	// copy existing
	for _, zipItem := range zr.File {
		zipItemReader, err := zipItem.OpenRaw()
		if err != nil {
			return nil, err
		}
		header := zipItem.FileHeader

		// work-around for https://github.com/golang/go/issues/54801
		if strings.HasSuffix(header.Name, "/") {
			continue
		}
		targetItem, err := zw.CreateRaw(&header)
		_, err = io.Copy(targetItem, zipItemReader)
		if err != nil {
			return nil, fmt.Errorf("failed to copy zip item: %s: %v", header.Name, err)
		}
	}

	logger.Infof("bundling %d files from workspace", len(files))
	for _, localFile := range files {
		logger.Tracef("bundling %s", localFile)
		f, err := zw.Create(path.Join("config", path.Base(localFile)))
		if err != nil {
			return nil, err
		}
		contents, err := fileutil.ReadFile(localFile)
		if err != nil {
			return nil, err
		}
		if _, err = f.Write(*contents); err != nil {
			return nil, err
		}
	}

	return dst, nil
}
