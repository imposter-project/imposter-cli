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

package cmd

import (
	"fmt"
	proxy2 "gatehill.io/imposter/internal/proxy"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var proxyFlags = struct {
	port                      int
	outputDir                 string
	rewrite                   bool
	captureRequestBody        bool
	captureRequestHeaders     bool
	ignoreDuplicateRequests   bool
	recordOnlyResponseHeaders []string
	flatResponseFileStructure bool
}{}

// proxyCmd represents the up command
var proxyCmd = &cobra.Command{
	Use:   "proxy [URL]",
	Short: "Proxy an endpoint and record HTTP exchanges",
	Long:  `Proxies an endpoint and records HTTP exchanges to file, in Imposter format.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		upstream := args[0]
		var outputDir string
		if proxyFlags.outputDir != "" {
			outputDir = proxyFlags.outputDir
		} else {
			workingDir, err := os.Getwd()
			if err != nil {
				panic(err)
			}
			outputDir = workingDir
		}
		options := proxy2.RecorderOptions{
			CaptureRequestBody:        proxyFlags.captureRequestBody,
			CaptureRequestHeaders:     proxyFlags.captureRequestHeaders,
			IgnoreDuplicateRequests:   proxyFlags.ignoreDuplicateRequests,
			RecordOnlyResponseHeaders: proxyFlags.recordOnlyResponseHeaders,
			FlatResponseFileStructure: proxyFlags.flatResponseFileStructure,
		}
		proxyUpstream(upstream, proxyFlags.port, outputDir, proxyFlags.rewrite, options)
	},
}

func init() {
	proxyCmd.Flags().IntVarP(&proxyFlags.port, "port", "p", 8080, "Port on which to listen")
	proxyCmd.Flags().StringVarP(&proxyFlags.outputDir, "output-dir", "o", "", "Directory in which HTTP exchanges are recorded (default: current working directory)")
	proxyCmd.Flags().BoolVar(&proxyFlags.captureRequestBody, "capture-request-body", false, "Capture the request body")
	proxyCmd.Flags().BoolVar(&proxyFlags.captureRequestHeaders, "capture-request-headers", false, "Capture the request headers")
	proxyCmd.Flags().BoolVarP(&proxyFlags.rewrite, "rewrite-urls", "r", false, "Rewrite upstream URL in response body to proxy URL")
	proxyCmd.Flags().BoolVarP(&proxyFlags.ignoreDuplicateRequests, "ignore-duplicate-requests", "i", true, "Ignore duplicate requests with same method and URI")
	proxyCmd.Flags().StringSliceVarP(&proxyFlags.recordOnlyResponseHeaders, "response-headers", "H", nil, "Record only these response headers")
	proxyCmd.Flags().BoolVar(&proxyFlags.flatResponseFileStructure, "flat", false, "Flatten the response file structure")
	rootCmd.AddCommand(proxyCmd)
}

func proxyUpstream(upstream string, port int, dir string, rewrite bool, options proxy2.RecorderOptions) {
	logger.Infof("starting proxy for upstream %s on port %v", upstream, port)
	recorderC, err := proxy2.StartRecorder(upstream, dir, options)
	if err != nil {
		logger.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/system/status", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprintf(writer, "ok\n")
	})
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		proxy2.Handle(upstream, writer, request, func(reqBody *[]byte, statusCode int, respBody *[]byte, respHeaders *http.Header) (*[]byte, *http.Header) {
			if rewrite {
				respBody = proxy2.Rewrite(respHeaders, respBody, upstream, port)
			}
			recorderC <- proxy2.HttpExchange{
				Request:         request,
				RequestBody:     reqBody,
				StatusCode:      statusCode,
				ResponseBody:    respBody,
				ResponseHeaders: respHeaders,
			}
			return respBody, respHeaders
		})
	})

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	if err != nil {
		logger.Fatal(err)
	}
}
