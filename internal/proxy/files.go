/*
Copyright © 2022 Pete Cornish <outofcoffee@gmail.com>

Licensed under the Apache License, Proxy 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proxy

import (
	"fmt"
	"gatehill.io/imposter/internal/stringutil"
	"github.com/google/uuid"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
)

// generateRespFileName returns a unique filename for the given response
func generateRespFileName(
	upstreamHost string,
	dir string,
	options RecorderOptions,
	exchange HttpExchange,
	prefix string,
) (respFile string, err error) {
	req := exchange.Request
	sanitisedParent := strings.TrimPrefix(path.Dir(req.URL.EscapedPath()), "/")
	if sanitisedParent == "." {
		sanitisedParent = ""
	}

	baseFileName := path.Base(req.URL.EscapedPath())
	if baseFileName == "/" || baseFileName == "." {
		baseFileName = "index"
	}
	baseFileName = prefix + baseFileName

	var parentDir, respFileName string
	if options.FlatResponseFileStructure {
		flatParent := strings.ReplaceAll(sanitisedParent, "/", "_")
		if len(flatParent) > 0 {
			flatParent += "_"
		}
		parentDir = dir
		respFileName = upstreamHost + "-" + req.Method + "-" + flatParent + baseFileName

	} else {
		parentDir = path.Join(dir, sanitisedParent)
		if err := ensureDirExists(parentDir); err != nil {
			return "", err
		}
		respFileName = req.Method + "-" + baseFileName
	}

	var suffix string
	if path.Ext(baseFileName) == "" {
		suffix = getFileExtension(exchange.ResponseHeaders)
	} else {
		suffix = ""
	}
	respFile = path.Join(parentDir, respFileName+suffix)

	if _, err = os.Stat(respFile); err == nil {
		// already exists - add url hash
		suffix = "_" + stringutil.Sha1hashString(req.URL.String()) + suffix
		respFile = path.Join(parentDir, respFileName+suffix)
	}
	if _, err = os.Stat(respFile); err == nil {
		// already exists - add uuid
		suffix = "_" + uuid.New().String() + suffix
		respFile = path.Join(parentDir, respFileName+suffix)
	}

	return respFile, nil
}

func getFileExtension(respHeaders *http.Header) string {
	if contentDisp := respHeaders.Get("Content-Disposition"); contentDisp != "" {
		directives := strings.Split(contentDisp, ";")
		for _, directive := range directives {
			directive = strings.TrimSpace(directive)
			if strings.HasPrefix(directive, "filename=") {
				filename := strings.TrimPrefix(directive, "filename=")
				return path.Ext(filename)
			}
		}
	}

	if contentType := respHeaders.Get("Content-Type"); contentType != "" {
		if extensions, err := mime.ExtensionsByType(contentType); err == nil && len(extensions) > 0 {
			return extensions[0]
		}
	}
	return ".txt"
}

func ensureDirExists(dir string) error {
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0700)
			if err != nil {
				return fmt.Errorf("failed to create response file dir: %s: %v", dir, err)
			}
		} else {
			return fmt.Errorf("failed to stat response file dir: %s: %v", dir, err)
		}
	}
	return nil
}
