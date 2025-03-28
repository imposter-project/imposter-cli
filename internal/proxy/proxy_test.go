package proxy

import (
	"fmt"
	"gatehill.io/imposter/internal/engine/enginetests"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

// waitForServerReady polls a healthcheck endpoint until it responds with a 200 status code
// or until the timeout is reached.
func waitForServerReady(t *testing.T, healthCheckURL string, serverName string) {
	healthClient := &http.Client{
		Timeout: 500 * time.Millisecond,
	}

	startTime := time.Now()
	timeout := 20 * time.Second

	t.Logf("Waiting for %s to start at %s...", serverName, healthCheckURL)

	for {
		elapsed := time.Since(startTime)
		if elapsed > timeout {
			t.Fatalf("%s failed to start within %s", serverName, timeout)
		}

		// Poll the health endpoint
		resp, err := healthClient.Get(healthCheckURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			t.Logf("%s started successfully after %s", serverName, elapsed)
			break
		}
		if resp != nil {
			resp.Body.Close()
		}

		// Small delay between polls
		time.Sleep(100 * time.Millisecond)
	}
}

func TestHandle(t *testing.T) {
	// Define test cases with different status codes, headers, and response bodies
	testCases := []struct {
		name            string
		statusCode      int
		responseBody    string
		requestPath     string
		requestMethod   string
		requestHeaders  map[string]string
		responseHeaders map[string]string
	}{
		{
			name:           "basic GET request",
			statusCode:     http.StatusOK,
			responseBody:   "hello world",
			requestPath:    "/",
			requestMethod:  "GET",
			requestHeaders: map[string]string{},
			responseHeaders: map[string]string{
				"Content-Type": "text/plain",
			},
		},
		{
			name:           "GET request with query params",
			statusCode:     http.StatusOK,
			responseBody:   "query param response",
			requestPath:    "/query?param1=value1&param2=value2",
			requestMethod:  "GET",
			requestHeaders: map[string]string{},
			responseHeaders: map[string]string{
				"Content-Type": "text/plain",
			},
		},
		{
			name:          "POST request with headers",
			statusCode:    http.StatusCreated,
			responseBody:  `{"id":"123","status":"created"}`,
			requestPath:   "/resources",
			requestMethod: "POST",
			requestHeaders: map[string]string{
				"Content-Type":    "application/json",
				"X-Custom-Header": "custom-value",
			},
			responseHeaders: map[string]string{
				"Content-Type": "application/json",
				"Location":     "/resources/123",
			},
		},
		{
			name:           "error response",
			statusCode:     http.StatusNotFound,
			responseBody:   "not found",
			requestPath:    "/missing",
			requestMethod:  "GET",
			requestHeaders: map[string]string{},
			responseHeaders: map[string]string{
				"Content-Type": "text/plain",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Start a test server that returns predetermined responses
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify that the request path matches what we expect
				if r.URL.String() != tc.requestPath {
					t.Errorf("Expected request path %s, got %s", tc.requestPath, r.URL.String())
				}

				// Verify request method
				if r.Method != tc.requestMethod {
					t.Errorf("Expected request method %s, got %s", tc.requestMethod, r.Method)
				}

				// Verify request headers (if specified)
				for headerName, expectedValue := range tc.requestHeaders {
					actualValue := r.Header.Get(headerName)
					if actualValue != expectedValue {
						t.Errorf("Expected request header %s=%s, got %s", headerName, expectedValue, actualValue)
					}
				}

				// Set response headers
				for headerName, headerValue := range tc.responseHeaders {
					w.Header().Set(headerName, headerValue)
				}

				// Set status code and write response body
				w.WriteHeader(tc.statusCode)
				_, err := w.Write([]byte(tc.responseBody))
				if err != nil {
					t.Fatalf("Failed to write response: %v", err)
				}
			}))
			defer server.Close()

			// Create a request that appears to come from a client
			req, err := http.NewRequest(tc.requestMethod, tc.requestPath, http.NoBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			// Set remote address to simulate a real client
			req.RemoteAddr = "127.0.0.1:12345"

			// Add request headers
			for headerName, headerValue := range tc.requestHeaders {
				req.Header.Set(headerName, headerValue)
			}

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Variables to capture response from the listener
			var capturedStatusCode int
			var capturedRespBody *[]byte
			var capturedRespHeaders *http.Header

			// Set up a listener function that captures the response and passes it through unchanged
			listenerFn := func(reqBody *[]byte, statusCode int, respBody *[]byte, respHeaders *http.Header) (*[]byte, *http.Header) {
				capturedStatusCode = statusCode
				capturedRespBody = respBody
				capturedRespHeaders = respHeaders
				return respBody, respHeaders
			}

			// Call the Handle function with our test server as upstream
			Handle(server.URL, rr, req, listenerFn)

			// Verify status code
			if capturedStatusCode != tc.statusCode {
				t.Errorf("Expected status code %d, got %d", tc.statusCode, capturedStatusCode)
			}

			// Verify response body
			expectedBody := []byte(tc.responseBody)
			if !reflect.DeepEqual(*capturedRespBody, expectedBody) {
				t.Errorf("Expected response body %s, got %s", expectedBody, *capturedRespBody)
			}

			// Verify proxied response headers (only check the ones we explicitly set)
			for headerName, expectedValue := range tc.responseHeaders {
				values, exists := (*capturedRespHeaders)[headerName]
				if !exists || len(values) == 0 {
					t.Errorf("Expected response header %s to exist", headerName)
					continue
				}
				actualValue := values[0]
				if actualValue != expectedValue {
					t.Errorf("Expected response header %s=%s, got %s", headerName, expectedValue, actualValue)
				}
			}

			// Verify status code in the actual response
			if rr.Code != tc.statusCode {
				t.Errorf("Expected response status code %d, got %d", tc.statusCode, rr.Code)
			}

			// Verify response body in the actual response
			if rr.Body.String() != tc.responseBody {
				t.Errorf("Expected response body %s, got %s", tc.responseBody, rr.Body.String())
			}

			// Verify headers in the actual response
			for headerName, expectedValue := range tc.responseHeaders {
				actualValue := rr.Header().Get(headerName)
				if actualValue != expectedValue {
					t.Errorf("Expected response header %s=%s, got %s", headerName, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestHandleEndToEnd(t *testing.T) {
	// Create a real HTTP server that we'll proxy
	mux := http.NewServeMux()
	// Add health check endpoint for polling
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"OK"}`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Custom-Header", "test-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})
	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Hello, JSON!"}`))
	})
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	})

	port := enginetests.GetFreePort()
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() {
		server.ListenAndServe()
	}()
	defer server.Close()

	// Poll the healthcheck endpoint instead of sleeping
	upstreamURL := fmt.Sprintf("http://localhost:%d", port)
	healthCheckURL := fmt.Sprintf("%s/health", upstreamURL)

	// Wait for the upstream server to become ready
	waitForServerReady(t, healthCheckURL, "Upstream server")

	// Create a proxy server
	proxyMux := http.NewServeMux()
	// Add health check for proxy
	proxyMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"OK"}`))
	})
	proxyMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		Handle(upstreamURL, w, r, func(reqBody *[]byte, statusCode int, respBody *[]byte, respHeaders *http.Header) (*[]byte, *http.Header) {
			// Pass through unchanged
			return respBody, respHeaders
		})
	})
	proxyPort := enginetests.GetFreePort()
	proxyServer := &http.Server{Addr: fmt.Sprintf(":%d", proxyPort), Handler: proxyMux}
	go func() {
		proxyServer.ListenAndServe()
	}()
	defer proxyServer.Close()

	// Poll the healthcheck endpoint for proxy instead of sleeping
	proxyURL := fmt.Sprintf("http://localhost:%d", proxyPort)
	proxyHealthCheckURL := fmt.Sprintf("%s/health", proxyURL)

	// Wait for the proxy server to become ready
	waitForServerReady(t, proxyHealthCheckURL, "Proxy server")

	// Test cases for different endpoints
	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
		expectedHeader map[string]string
	}{
		{
			name:           "Root path",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedBody:   "Hello, World!",
			expectedHeader: map[string]string{
				"Content-Type": "text/plain; charset=utf-8",
				// X-Custom-Header might be filtered by the proxy
			},
		},
		{
			name:           "JSON endpoint",
			path:           "/json",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Hello, JSON!"}`,
			expectedHeader: map[string]string{
				"Content-Type": "application/json; charset=utf-8",
			},
		},
		{
			name:           "Error endpoint",
			path:           "/error",
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error",
			expectedHeader: map[string]string{
				"Content-Type": "text/plain; charset=utf-8",
			},
		},
	}

	// Run tests
	client := &http.Client{Timeout: 5 * time.Second}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Send request to proxy
			req, err := http.NewRequest("GET", proxyURL+tt.path, http.NoBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			// Check response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}
			if string(body) != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, string(body))
			}

			// Check headers with special handling for the JSON endpoint
			// because of how Go's http.ServeContent might set Content-Type
			for key, value := range tt.expectedHeader {
				if resp.Header.Get(key) != value {
					t.Errorf("Expected header %s=%s, got %s", key, value, resp.Header.Get(key))
				}
			}
		})
	}
}
