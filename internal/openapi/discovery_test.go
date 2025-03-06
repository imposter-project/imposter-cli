package openapi

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestDiscoverOpenApiSpecs(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "openapi_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	files := map[string]string{
		"valid_openapi.json": `{
			"openapi": "3.0.0",
			"info": {
				"title": "Test API",
				"version": "1.0.0"
			}
		}`,
		"valid_swagger.json": `{
			"swagger": "2.0",
			"info": {
				"title": "Test API",
				"version": "1.0.0"
			}
		}`,
		"valid_openapi.yaml": `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
`,
		"valid_swagger.yml": `
swagger: "2.0"
info:
  title: Test API
  version: 1.0.0
`,
		"invalid_spec.json": `{
			"notOpenApi": true,
			"info": {
				"title": "Test API",
				"version": "1.0.0"
			}
		}`,
		"invalid_spec.yaml": `
notOpenApi: true
info:
  title: Test API
  version: 1.0.0
`,
		"invalid_yaml.yaml": `{
			this is not valid yaml
		}`,
		"not_json.json": `this is not json`,
	}

	for name, content := range files {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Run discovery
	specs := DiscoverOpenApiSpecs(tempDir)

	// Sort the results for consistent comparison
	sort.Strings(specs)

	// Expected valid specs
	want := []string{
		filepath.Join(tempDir, "valid_openapi.json"),
		filepath.Join(tempDir, "valid_openapi.yaml"),
		filepath.Join(tempDir, "valid_swagger.json"),
		filepath.Join(tempDir, "valid_swagger.yml"),
	}
	sort.Strings(want)

	// Compare results
	if len(specs) != len(want) {
		t.Errorf("DiscoverOpenApiSpecs() found %d specs, want %d", len(specs), len(want))
	}

	for i := range want {
		if i >= len(specs) {
			t.Errorf("Missing expected spec: %s", want[i])
			continue
		}
		if specs[i] != want[i] {
			t.Errorf("Spec mismatch at index %d: got %s, want %s", i, specs[i], want[i])
		}
	}
}

func TestLoadYamlAsJson(t *testing.T) {
	// Create temporary test file
	tempFile, err := os.CreateTemp("", "yaml_test_*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	tests := []struct {
		name     string
		content  string
		wantErr  bool
		wantJSON string
	}{
		{
			name: "valid yaml",
			content: `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
`,
			wantErr:  false,
			wantJSON: `{"info":{"title":"Test API","version":"1.0.0"},"openapi":"3.0.0"}`,
		},
		{
			name: "invalid yaml",
			content: `
foo: [this is not closed
bar: : : invalid colons
  - unbalanced indent
      wrong`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(tempFile.Name(), []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			got, err := loadYamlAsJson(tempFile.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("loadYamlAsJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got) != tt.wantJSON {
				t.Errorf("loadYamlAsJson() = %v, want %v", string(got), tt.wantJSON)
			}
		})
	}
}

func TestIsOpenApiSpec(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name: "valid openapi spec",
			content: `{
				"openapi": "3.0.0",
				"info": {
					"title": "Test API",
					"version": "1.0.0"
				}
			}`,
			want: true,
		},
		{
			name: "valid swagger spec",
			content: `{
				"swagger": "2.0",
				"info": {
					"title": "Test API",
					"version": "1.0.0"
				}
			}`,
			want: true,
		},
		{
			name: "not an openapi spec",
			content: `{
				"notOpenApi": true,
				"info": {
					"title": "Test API",
					"version": "1.0.0"
				}
			}`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isOpenApiSpec([]byte(tt.content)); got != tt.want {
				t.Errorf("isOpenApiSpec() = %v, want %v", got, tt.want)
			}
		})
	}
}
