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
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"
)

func init() {
	logger.SetLevel(logrus.TraceLevel)
}

func Test_generateRespFileName(t *testing.T) {
	outputDir, err := os.MkdirTemp(os.TempDir(), "imposter-cli")
	if err != nil {
		panic(err)
	}
	rootUrl, _ := url.Parse("https://example.com")
	nestedUrl, _ := url.Parse("https://example.com/a/b.txt")

	type args struct {
		upstreamHost string
		dir          string
		options      RecorderOptions
		exchange     HttpExchange
		prefix       string
	}
	tests := []struct {
		name         string
		args         args
		wantRespFile string
		wantErr      bool
	}{
		{
			name: "root text file, no headers",
			args: args{
				upstreamHost: "example.com",
				dir:          outputDir,
				options:      RecorderOptions{FlatResponseFileStructure: false},
				exchange: HttpExchange{
					Request:         &http.Request{Method: "GET", URL: rootUrl},
					ResponseHeaders: &http.Header{},
				},
			},
			wantRespFile: path.Join(outputDir, "GET-index.txt"),
			wantErr:      false,
		},
		{
			name: "root text file with prefix",
			args: args{
				upstreamHost: "example.com",
				dir:          outputDir,
				options:      RecorderOptions{FlatResponseFileStructure: false},
				exchange: HttpExchange{
					Request:         &http.Request{Method: "GET", URL: rootUrl},
					ResponseHeaders: &http.Header{},
				},
				prefix: "foo-",
			},
			wantRespFile: path.Join(outputDir, "GET-foo-index.txt"),
			wantErr:      false,
		},
		{
			name: "root html file using content disposition",
			args: args{
				upstreamHost: "example.com",
				dir:          outputDir,
				options:      RecorderOptions{FlatResponseFileStructure: false},
				exchange: HttpExchange{
					Request: &http.Request{Method: "GET", URL: rootUrl},
					ResponseHeaders: &http.Header{
						"Content-Disposition": []string{"filename=example.html"},
					},
				},
			},
			wantRespFile: path.Join(outputDir, "GET-index.html"),
			wantErr:      false,
		},
		{
			name: "root html file using content type",
			args: args{
				upstreamHost: "example.com",
				dir:          outputDir,
				options:      RecorderOptions{FlatResponseFileStructure: false},
				exchange: HttpExchange{
					Request: &http.Request{Method: "GET", URL: rootUrl},
					ResponseHeaders: &http.Header{
						"Content-Type": []string{"text/html"},
					},
				},
			},
			wantRespFile: path.Join(outputDir, "GET-index.htm"),
			wantErr:      false,
		},
		{
			name: "nested url, hierarchical response file path",
			args: args{
				upstreamHost: "example.com",
				dir:          outputDir,
				options:      RecorderOptions{FlatResponseFileStructure: false},
				exchange: HttpExchange{
					Request:         &http.Request{Method: "GET", URL: nestedUrl},
					ResponseHeaders: &http.Header{},
				},
			},
			wantRespFile: path.Join(outputDir, "a/GET-b.txt"),
			wantErr:      false,
		},
		{
			name: "nested url, flat response file path",
			args: args{
				upstreamHost: "example.com",
				dir:          outputDir,
				options:      RecorderOptions{FlatResponseFileStructure: true},
				exchange: HttpExchange{
					Request:         &http.Request{Method: "GET", URL: nestedUrl},
					ResponseHeaders: &http.Header{},
				},
			},
			wantRespFile: path.Join(outputDir, "example.com-GET-a_b.txt"),
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRespFile, err := generateRespFileName(tt.args.upstreamHost, tt.args.dir, tt.args.options, tt.args.exchange, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateRespFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotRespFile != tt.wantRespFile {
				t.Errorf("generateRespFileName() gotRespFile = %v, want %v", gotRespFile, tt.wantRespFile)
			}
		})
	}
}
