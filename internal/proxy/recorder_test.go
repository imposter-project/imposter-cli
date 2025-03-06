package proxy

import (
	"fmt"
	"gatehill.io/imposter/internal/impostermodel"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"testing"
)

func Test_getResponseFile(t *testing.T) {
	outputDir, err := os.MkdirTemp(os.TempDir(), "imposter-cli")
	if err != nil {
		panic(err)
	}
	rootUrl, _ := url.Parse("https://example.com")

	responseBody := []byte("test")
	bodyHash := "a94a8fe5ccb19ba61c4c0873d391e987982fbbd3"

	type args struct {
		upstreamHost string
		dir          string
		options      RecorderOptions
		exchange     HttpExchange
		fileHashes   *map[string]string
		prefix       string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "empty response body",
			args: args{
				upstreamHost: "example.com",
				dir:          outputDir,
				options:      RecorderOptions{FlatResponseFileStructure: false},
				exchange: HttpExchange{
					Request:      &http.Request{Method: "GET", URL: rootUrl},
					ResponseBody: &[]byte{},
				},
				fileHashes: buildMap(outputDir, []string{}),
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "existing response body",
			args: args{
				upstreamHost: "example.com",
				dir:          outputDir,
				options:      RecorderOptions{FlatResponseFileStructure: false},
				exchange: HttpExchange{
					Request:         &http.Request{Method: "GET", URL: rootUrl},
					ResponseBody:    &responseBody,
					ResponseHeaders: &http.Header{},
				},
				fileHashes: buildMap(outputDir, []string{bodyHash}),
			},
			want:    path.Join(outputDir, "existing-file-0.txt"),
			wantErr: false,
		},
		{
			name: "new response body",
			args: args{
				upstreamHost: "example.com",
				dir:          outputDir,
				options:      RecorderOptions{FlatResponseFileStructure: false},
				exchange: HttpExchange{
					Request:         &http.Request{Method: "GET", URL: rootUrl},
					ResponseBody:    &responseBody,
					ResponseHeaders: &http.Header{},
				},
				fileHashes: buildMap(outputDir, []string{}),
			},
			want:    path.Join(outputDir, "GET-index.txt"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getResponseFile(tt.args.upstreamHost, tt.args.dir, tt.args.options, tt.args.exchange, tt.args.fileHashes, tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("getResponseFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getResponseFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildResource(t *testing.T) {
	outputDir, err := os.MkdirTemp(os.TempDir(), "imposter-cli")
	if err != nil {
		panic(err)
	}
	rootUrl, _ := url.Parse("https://example.com/")

	type args struct {
		dir      string
		options  RecorderOptions
		exchange HttpExchange
		respFile string
	}
	tests := []struct {
		name    string
		args    args
		want    impostermodel.Resource
		wantErr bool
	}{
		{
			name: "request with query params",
			args: args{
				dir: outputDir,
				options: RecorderOptions{
					CaptureRequestBody:    false,
					CaptureRequestHeaders: false,
				},
				exchange: HttpExchange{
					Request: &http.Request{
						Method: "GET",
						URL:    &url.URL{Path: "/test", RawQuery: "param1=value1&param2=value2"},
					},
					ResponseHeaders: &http.Header{},
				},
				respFile: "",
			},
			want: impostermodel.Resource{
				Path:   "/test",
				Method: "GET",
				QueryParams: &map[string]string{
					"param1": "value1",
					"param2": "value2",
				},
				Response: &impostermodel.ResponseConfig{
					StatusCode: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "request with headers",
			args: args{
				dir: outputDir,
				options: RecorderOptions{
					CaptureRequestBody:    false,
					CaptureRequestHeaders: true,
				},
				exchange: HttpExchange{
					Request: &http.Request{
						Method: "GET",
						URL:    rootUrl,
						Header: http.Header{
							"Header1": []string{"value1"},
							"Header2": []string{"value2"},
						},
					},
					ResponseHeaders: &http.Header{},
				},
				respFile: "",
			},
			want: impostermodel.Resource{
				Path:   "/",
				Method: "GET",
				RequestHeaders: &map[string]string{
					"Header1": "value1",
					"Header2": "value2",
				},
				Response: &impostermodel.ResponseConfig{
					StatusCode: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "request with body",
			args: args{
				dir: outputDir,
				options: RecorderOptions{
					CaptureRequestBody:    true,
					CaptureRequestHeaders: false,
				},
				exchange: HttpExchange{
					Request: &http.Request{
						Method: "POST",
						URL:    rootUrl,
						Header: http.Header{
							"Content-Type": []string{"application/json"},
						},
					},
					RequestBody: func() *[]byte {
						body := []byte(`{"key":"value"}`)
						return &body
					}(),
					ResponseHeaders: &http.Header{},
				},
				respFile: "",
			},
			want: impostermodel.Resource{
				Path:   "/",
				Method: "POST",
				RequestBody: &impostermodel.RequestBody{
					Value:    `{"key":"value"}`,
					Operator: "EqualTo",
				},
				Response: &impostermodel.ResponseConfig{
					StatusCode: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "request with unsupported content type",
			args: args{
				dir: outputDir,
				options: RecorderOptions{
					CaptureRequestBody:    true,
					CaptureRequestHeaders: false,
				},
				exchange: HttpExchange{
					Request: &http.Request{
						Method: "POST",
						URL:    rootUrl,
						Header: http.Header{
							"Content-Type": []string{"application/octet-stream"},
						},
					},
					RequestBody: func() *[]byte {
						body := []byte(`{"key":"value"}`)
						return &body
					}(),
					ResponseHeaders: &http.Header{},
				},
				respFile: "",
			},
			want: impostermodel.Resource{
				Path:   "/",
				Method: "POST",
				Response: &impostermodel.ResponseConfig{
					StatusCode: 0,
				},
			},
			wantErr: false,
		},
		{
			name: "request with response file",
			args: args{
				dir: outputDir,
				options: RecorderOptions{
					CaptureRequestBody:    false,
					CaptureRequestHeaders: false,
				},
				exchange: HttpExchange{
					Request: &http.Request{
						Method: "GET",
						URL:    rootUrl,
					},
					ResponseHeaders: &http.Header{},
				},
				respFile: path.Join(outputDir, "response.txt"),
			},
			want: impostermodel.Resource{
				Path:   "/",
				Method: "GET",
				Response: &impostermodel.ResponseConfig{
					StatusCode: 0,
					StaticFile: "response.txt",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildResource(tt.args.dir, tt.args.options, tt.args.exchange, tt.args.respFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildResource() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func buildMap(dir string, hashes []string) *map[string]string {
	m := make(map[string]string)
	for i, hash := range hashes {
		filename := fmt.Sprintf("existing-file-%d.txt", i)
		m[hash] = path.Join(dir, filename)
	}
	return &m
}
