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
	impostermodel2 "gatehill.io/imposter/internal/impostermodel"
	"gatehill.io/imposter/internal/stringutil"
	"github.com/google/uuid"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type RecorderOptions struct {
	CaptureRequestBody        bool
	CaptureRequestHeaders     bool
	IgnoreDuplicateRequests   bool
	RecordOnlyResponseHeaders []string
	FlatResponseFileStructure bool
}

func StartRecorder(upstream string, dir string, options RecorderOptions) (chan HttpExchange, error) {
	upstreamHost, err := formatUpstreamHostPort(upstream)
	if err != nil {
		return nil, err
	}
	configFile := path.Join(dir, upstreamHost+"-config.yaml")
	if _, err := os.Stat(configFile); err == nil {
		return nil, fmt.Errorf("config file %s already exists", configFile)
	}

	var resources []impostermodel2.Resource
	genOptions := impostermodel2.ConfigGenerationOptions{PluginName: "rest"}

	var requestHashes []string
	responseHashes := make(map[string]string)

	recordC := make(chan HttpExchange)
	go func() {
		for {
			exchange := <-recordC

			var responseFilePrefix string
			requestHash := getRequestHash(exchange.Request)
			if stringutil.Contains(requestHashes, requestHash) {
				if options.IgnoreDuplicateRequests {
					logger.Debugf("skipping recording of duplicate request %s %v", exchange.Request.Method, exchange.Request.URL)
					continue
				}
				responseFilePrefix = uuid.New().String() + "-"
			} else {
				responseFilePrefix = ""
			}
			requestHashes = append(requestHashes, requestHash)

			resource, err := record(upstreamHost, dir, &responseHashes, responseFilePrefix, exchange, options)
			if err != nil {
				logger.Warn(err)
				continue
			}
			resources = append(resources, *resource)

			if err := updateConfigFile(exchange, genOptions, resources, configFile); err != nil {
				logger.Warn(err)
			}
		}
	}()

	return recordC, nil
}

func formatUpstreamHostPort(upstream string) (string, error) {
	upstreamUrl, err := url.Parse(upstream)
	if err != nil {
		return "", fmt.Errorf("failed to parse upstream URL: %v", err)
	}
	host := upstreamUrl.Host
	if !strings.Contains(host, ":") {
		return host, nil
	} else {
		hostOnly, port, err := net.SplitHostPort(host)
		if err != nil {
			return "", fmt.Errorf("failed to parse split upstream host/port: %v", err)
		}
		if port != "" {
			hostOnly += "-" + port
		}
		return hostOnly, nil
	}
}

func record(
	upstreamHost string,
	dir string,
	responseHashes *map[string]string,
	prefix string,
	exchange HttpExchange,
	options RecorderOptions,
) (resource *impostermodel2.Resource, err error) {
	respFile, err := getResponseFile(upstreamHost, dir, options, exchange, responseHashes, prefix)
	if err != nil {
		return nil, err
	}
	r, err := buildResource(dir, options, exchange, respFile)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// getResponseFile checks if there is a response body. If not, an empty string is returned.
// If a body is not empty, the file hashes are checked for the hash of the response body to
// see if it has already been written. If not, a new file is written and its hash stored in the map.
func getResponseFile(
	upstreamHost string,
	dir string,
	options RecorderOptions,
	exchange HttpExchange,
	fileHashes *map[string]string,
	prefix string,
) (string, error) {
	req := exchange.Request
	respBody := *exchange.ResponseBody
	if len(respBody) == 0 {
		logger.Debugf("empty response body for %s %v", req.Method, req.URL)
		return "", nil
	}
	bodyHash := stringutil.Sha1hash(respBody)

	if existing := (*fileHashes)[bodyHash]; existing != "" {
		logger.Debugf("reusing identical response file %s for %s %v", existing, req.Method, req.URL)
		return existing, nil

	} else {
		respFile, err := generateRespFileName(upstreamHost, dir, options, exchange, prefix)
		if err != nil {
			return "", err
		}
		if err = os.WriteFile(respFile, respBody, 0644); err != nil {
			return "", fmt.Errorf("failed to write response file %s for %s %v: %v", respFile, req.Method, req.URL, err)
		}
		logger.Debugf("wrote response file %s for %s %v [%d bytes]", respFile, req.Method, req.URL, len(respBody))
		(*fileHashes)[bodyHash] = respFile
		return respFile, nil
	}
}

func buildResource(
	dir string,
	options RecorderOptions,
	exchange HttpExchange,
	respFile string,
) (impostermodel2.Resource, error) {
	req := *exchange.Request
	response := &impostermodel2.ResponseConfig{
		StatusCode: exchange.StatusCode,
	}
	if len(respFile) > 0 {
		relResponseFile, err := filepath.Rel(dir, respFile)
		if err != nil {
			return impostermodel2.Resource{}, fmt.Errorf("failed to get relative path for response file: %s: %v", respFile, err)
		}
		response.StaticFile = relResponseFile
	}
	resource := impostermodel2.Resource{
		Path:     req.URL.Path,
		Method:   req.Method,
		Response: response,
	}
	if len(req.URL.Query()) > 0 {
		queryParams := make(map[string]string)
		for qk, qvs := range req.URL.Query() {
			if len(qvs) > 0 {
				queryParams[qk] = qvs[0]
			}
		}
		resource.QueryParams = &queryParams
	}
	if options.CaptureRequestHeaders && len(req.Header) > 0 {
		headers := make(map[string]string)
		for headerName, headerValues := range req.Header {
			shouldSkip := stringutil.Contains(skipProxyHeaders, headerName) || stringutil.Contains(skipRecordHeaders, headerName)
			if !shouldSkip && len(headerValues) > 0 {
				headers[headerName] = headerValues[0]
			}
		}
		resource.RequestHeaders = &headers
	}
	if options.CaptureRequestBody && exchange.RequestBody != nil {
		contentType := req.Header.Get("Content-Type")
		if !isTextContentType(contentType) {
			logger.Debugf("unsupported content type '%s' for capture - skipping request body capture", contentType)
		} else {
			reqBody := *exchange.RequestBody
			if len(reqBody) > 0 {
				resource.RequestBody = &impostermodel2.RequestBody{
					Value:    string(reqBody),
					Operator: "EqualTo",
				}
			}
		}
	}
	if len(*exchange.ResponseHeaders) > 0 {
		headers := make(map[string]string)
		for headerName, headerValues := range *exchange.ResponseHeaders {
			shouldSkip := stringutil.Contains(skipProxyHeaders, headerName) || stringutil.Contains(skipRecordHeaders, headerName)
			if !shouldSkip &&
				(options.RecordOnlyResponseHeaders == nil) || stringutil.Contains(options.RecordOnlyResponseHeaders, headerName) {

				if len(headerValues) > 0 {
					headers[headerName] = headerValues[0]
				}
			}
		}
		resource.Response.Headers = &headers
	}
	return resource, nil
}

// getRequestHash generates a hash for a request based on the HTTP method and the URL. It does
// not take into consideration request headers.
func getRequestHash(req *http.Request) string {
	return stringutil.Sha1hashString(req.Method + req.URL.String())
}

func updateConfigFile(exchange HttpExchange, options impostermodel2.ConfigGenerationOptions, resources []impostermodel2.Resource, configFile string) error {
	req := exchange.Request
	config := impostermodel2.GenerateConfig(options, resources)
	err := os.WriteFile(configFile, config, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file %s for %s %v: %v", configFile, req.Method, req.URL, err)
	}
	logger.Debugf("wrote config file %s for %s %v", configFile, req.Method, req.URL)
	return nil
}
