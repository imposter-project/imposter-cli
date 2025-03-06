package openapi

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "openapi_parser_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name          string
		content       string
		wantErr       bool
		wantPaths     []string
		wantMethods   map[string][]string            // path -> methods
		wantResponses map[string]map[string][]string // path -> method -> status codes
	}{
		{
			name: "valid openapi with paths and responses",
			content: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /pets:
    get:
      description: List all pets
      responses:
        "200":
          description: A list of pets
          content:
            application/json:
              schema:
                type: array
    post:
      description: Create a pet
      responses:
        "201":
          description: Pet created
        "400":
          description: Invalid input
  /pets/{id}:
    get:
      description: Get a pet by ID
      responses:
        "200":
          description: Pet found
        "404":
          description: Pet not found
`,
			wantErr:   false,
			wantPaths: []string{"/pets", "/pets/{id}"},
			wantMethods: map[string][]string{
				"/pets":      {"get", "post"},
				"/pets/{id}": {"get"},
			},
			wantResponses: map[string]map[string][]string{
				"/pets": {
					"get":  {"200"},
					"post": {"201", "400"},
				},
				"/pets/{id}": {
					"get": {"200", "404"},
				},
			},
		},
		{
			name: "invalid yaml syntax",
			content: `
openapi: 3.0.0
paths:
  - this is invalid yaml
    get:
      responses:
        not valid
`,
			wantErr: true,
		},
		{
			name: "empty spec",
			content: `
openapi: 3.0.0
info:
  title: Empty API
  version: 1.0.0
`,
			wantErr:       false,
			wantPaths:     []string{},
			wantMethods:   map[string][]string{},
			wantResponses: map[string]map[string][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			specFile := filepath.Join(tempDir, "test-spec.yaml")
			if err := os.WriteFile(specFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			// Parse the spec
			model, err := Parse(specFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Verify paths exist
			if len(model.Paths) != len(tt.wantPaths) {
				t.Errorf("Parse() got %d paths, want %d", len(model.Paths), len(tt.wantPaths))
			}

			// Check each path has expected methods
			for path, methods := range tt.wantMethods {
				operations, exists := model.Paths[path]
				if !exists {
					t.Errorf("Parse() missing path %s", path)
					continue
				}

				for _, method := range methods {
					operation, exists := operations[method]
					if !exists {
						t.Errorf("Parse() missing method %s for path %s", method, path)
						continue
					}

					// Check responses
					wantResponses := tt.wantResponses[path][method]
					if len(operation.Responses) != len(wantResponses) {
						t.Errorf("Parse() got %d responses for %s %s, want %d",
							len(operation.Responses), method, path, len(wantResponses))
					}

					for _, code := range wantResponses {
						if _, exists := operation.Responses[code]; !exists {
							t.Errorf("Parse() missing response code %s for %s %s",
								code, method, path)
						}
					}
				}
			}
		})
	}
}
