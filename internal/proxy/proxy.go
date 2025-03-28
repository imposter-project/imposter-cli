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
	"bytes"
	"fmt"
	"gatehill.io/imposter/internal/logging"
	"gatehill.io/imposter/internal/stringutil"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"net/url"
	"time"
)

type HttpExchange struct {
	Request         *http.Request
	RequestBody     *[]byte
	StatusCode      int
	ResponseBody    *[]byte
	ResponseHeaders *http.Header
}

var skipProxyHeaders = []string{
	"Accept-Encoding",

	// Hop-by-hop headers. These are removed in requests to the upstream or reponses to the client.
	// See "13.5.1 End-to-end and Hop-by-hop Headers" in http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"TE",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

var skipRecordHeaders = []string{
	"Accept-Ranges",
	"Age",
	"Cache-Control",
	"Content-Length",
	"Date",
	"Etag",
	"Expires",
	"Last-Modified",
	"Server",
	"Vary",
}

var logger = logging.GetLogger()

var transport *http.Transport

func init() {
	transport = &http.Transport{
		DisableCompression: true,
		MaxIdleConns:       viper.GetInt("proxy.maxIdleConns"),
		IdleConnTimeout:    viper.GetDuration("proxy.idleConnTimeout"),
	}
	logger.Tracef("initialised proxy transport: %+v", transport)
}

func Handle(
	upstream string,
	w http.ResponseWriter,
	req *http.Request,
	listener func(reqBody *[]byte, statusCode int, respBody *[]byte, respHeaders *http.Header) (*[]byte, *http.Header),
) {
	startTime := time.Now()

	client := req.RemoteAddr
	logger.Debugf("received request %v %v from client %v", req.Method, req.URL, client)

	path, queryString, clientReqHeaders, requestBody, err := parseRequest(req)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	statusCode, responseBody, respHeaders, err := forward(upstream, req.Method, path, queryString, clientReqHeaders, requestBody)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	responseBody, respHeaders = listener(requestBody, statusCode, responseBody, respHeaders)

	err = sendResponse(w, respHeaders, statusCode, responseBody, client)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	elapsed := time.Since(startTime)
	logger.Infof("proxied %s %v to upstream [status: %v, body %v bytes] for client %v in %v", req.Method, req.URL, statusCode, len(*responseBody), client, elapsed)
}

func parseRequest(req *http.Request) (path string, queryString string, headers *http.Header, body *[]byte, err error) {
	defer req.Body.Close()
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("error parsing request body: %v", err)
	}
	return req.URL.Path, req.URL.RawQuery, &req.Header, &requestBody, nil
}

func forward(
	upstream string,
	httpMethod string,
	path string,
	queryString string,
	clientRequestHeaders *http.Header,
	requestBody *[]byte,
) (statusCode int, responseBody *[]byte, upstreamRespHeaders *http.Header, err error) {
	logger.Debugf("invoking upstream %s with %s %s [body: %v bytes]", upstream, httpMethod, path, len(*requestBody))

	upstreamUrl, err := url.JoinPath(upstream, path)
	if queryString != "" {
		upstreamUrl += "?" + queryString
	}
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to build upstream URL: %v", err)
	}
	logger.Tracef("upstream url: %s", upstreamUrl)

	req, err := http.NewRequest(httpMethod, upstreamUrl, bytes.NewReader(*requestBody))
	upstreamReqHeaders := req.Header
	copyHeaders(clientRequestHeaders, &upstreamReqHeaders)

	client := &http.Client{Transport: transport}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()

	var respBody []byte
	if resp.Body != nil {
		respBody, err = io.ReadAll(resp.Body)
		if err != nil {
			return 0, nil, nil, fmt.Errorf("error parsing request body: %v", err)
		}
	}
	logger.Debugf("upstream responded to %s %s with status %d [body %v bytes]", httpMethod, upstreamUrl, resp.StatusCode, len(respBody))
	return resp.StatusCode, &respBody, &resp.Header, nil
}

func sendResponse(w http.ResponseWriter, headers *http.Header, statusCode int, body *[]byte, client string) (err error) {
	clientRespHeaders := w.Header()
	copyHeaders(headers, &clientRespHeaders)
	w.WriteHeader(statusCode)
	_, err = w.Write(*body)
	if err != nil {
		return fmt.Errorf("error writing response: %v", err)
	}

	logger.Debugf("wrote response [status: %v, body %v bytes] to client %v", statusCode, len(*body), client)
	return nil
}

// copyHeaders copies all headers from source to destination, unless the name
// of the header is a hop-by-hop header.
func copyHeaders(source *http.Header, destination *http.Header) {
	for headerName, headerValues := range *source {
		if !stringutil.Contains(skipProxyHeaders, headerName) {
			for _, headerValue := range headerValues {
				destination.Add(headerName, headerValue)
			}
		}
	}
}
