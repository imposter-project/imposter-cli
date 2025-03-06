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
	"gatehill.io/imposter/internal/engine"
	"gatehill.io/imposter/internal/engine/enginetests"
	"gatehill.io/imposter/internal/proxy"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path"
	"testing"
	"time"
)

func init() {
	logger.SetLevel(logrus.TraceLevel)
}

func Test_proxyUpstream(t *testing.T) {
	type args struct {
		rewrite bool
		options proxy.RecorderOptions
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "proxy example.com, hierarchical response files",
			args: args{
				rewrite: false,
				options: proxy.RecorderOptions{
					FlatResponseFileStructure: false,
				},
			},
		},
		{
			name: "proxy example.com, flat response files",
			args: args{
				rewrite: false,
				options: proxy.RecorderOptions{
					FlatResponseFileStructure: true,
				},
			},
		},
	}
	for _, tt := range tests {
		server, upstream, upstreamPort, err := startUpstream()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			server.Close()
		})

		t.Run(tt.name, func(t *testing.T) {
			port := enginetests.GetFreePort()
			outputDir, err := os.MkdirTemp(os.TempDir(), "imposter-cli")
			if err != nil {
				panic(err)
			}

			go func() {
				proxyUpstream(upstream, port, outputDir, tt.args.rewrite, tt.args.options)
			}()
			if up := engine.WaitUntilUp(port, nil); !up {
				t.Fatalf("proxy did not come up on port %d", port)
			}

			if err := sendRequestToProxy(port); err != nil {
				t.Fatal(err)
			}

			upstreamHostAndPort := fmt.Sprintf("localhost-%d", upstreamPort)
			cfgFileName := upstreamHostAndPort + "-config.yaml"
			var indexFileName string
			if tt.args.options.FlatResponseFileStructure {
				indexFileName = upstreamHostAndPort + "-GET-index.txt"
			} else {
				indexFileName = "GET-index.txt"
			}

			if cfgExists := engine.WaitForOp(fmt.Sprintf("config file: %s", cfgFileName), 10*time.Second, nil, func() bool {
				if _, err = os.Stat(path.Join(outputDir, cfgFileName)); err != nil {
					return false
				}
				return true
			}); !cfgExists {
				t.Fatalf("config file not found")
			}

			if indexExists := engine.WaitForOp(fmt.Sprintf("index file: %s", indexFileName), 10*time.Second, nil, func() bool {
				if _, err = os.Stat(path.Join(outputDir, indexFileName)); err != nil {
					return false
				}
				return true
			}); !indexExists {
				t.Fatalf("index file not found")
			}
		})
	}
}

func startUpstream() (server *http.Server, url string, port int, err error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Disposition", "filename=hello.txt")
		writer.Write([]byte("hello world"))
	})
	port = enginetests.GetFreePort()
	server = &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() {
		server.ListenAndServe()
	}()
	url = fmt.Sprintf("http://localhost:%d", port)
	if up := engine.WaitForUrl("upstream", url, nil); !up {
		return nil, "", 0, fmt.Errorf("failed to start upstream on port %d", port)
	}
	return server, url, port, nil
}

func sendRequestToProxy(port int) error {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	url := fmt.Sprintf("http://localhost:%d", port)
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("request failed for proxy at %s: %s", url, err)
	}
	if _, err := io.ReadAll(resp.Body); err != nil {
		return fmt.Errorf("body read failed for proxy at %s: %s", url, err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode == 200 {
		logger.Tracef("proxy up at %s", url)
		return nil
	}
	return nil
}
