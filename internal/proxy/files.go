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

// extractSoapAction returns the SOAP action for the request, looking first
// at the SOAPAction header (SOAP 1.1) and falling back to the `action`
// parameter on the Content-Type header (SOAP 1.2). The value is unquoted.
// An empty string is returned when no action is present.
func extractSoapAction(req *http.Request) string {
	// SOAP 1.1: SOAPAction header
	if soapAction := strings.Trim(req.Header.Get("SOAPAction"), "\""); soapAction != "" {
		logger.Debugf("extracted SOAPAction from header: %q", soapAction)
		return soapAction
	}
	// SOAP 1.2: action parameter on the Content-Type header
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		logger.Debugf("failed to parse Content-Type %q: %v", contentType, err)
		return ""
	}
	if action := params["action"]; action != "" {
		logger.Debugf("extracted SOAPAction from Content-Type: %q", action)
		return action
	}
	return ""
}

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
		if options.Soap11Mode {
			logger.Debugf("SOAP 1.1 mode enabled - checking for SOAPAction header")
			if soapAction := extractSoapAction(req); soapAction != "" {
				logger.Debugf("Found SOAPAction: '%s', generating SOAP-aware filename", soapAction)
				sanitizedAction := strings.ReplaceAll(soapAction, "/", "_")
				sanitizedAction = strings.ReplaceAll(sanitizedAction, ":", "_")
				sanitizedAction = strings.ReplaceAll(sanitizedAction, ".", "_")
				respFileName = respFileName + "_" + sanitizedAction
				logger.Debugf("Generated SOAP filename: '%s'", respFileName)
			} else {
				logger.Debugf("No SOAPAction found, using standard filename")
			}
		}

	} else {
		parentDir = path.Join(dir, sanitisedParent)
		if err := ensureDirExists(parentDir); err != nil {
			return "", err
		}
		respFileName = req.Method + "-" + baseFileName
		if options.Soap11Mode {
			if soapAction := extractSoapAction(req); soapAction != "" {
				sanitizedAction := strings.ReplaceAll(soapAction, "/", "_")
				sanitizedAction = strings.ReplaceAll(sanitizedAction, ":", "_")
				sanitizedAction = strings.ReplaceAll(sanitizedAction, ".", "_")
				respFileName = respFileName + "_" + sanitizedAction
			}
		}
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
