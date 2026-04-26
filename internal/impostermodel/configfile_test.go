/*
Copyright © 2026 Pete Cornish <outofcoffee@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package impostermodel

import (
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

func TestGenerateConfig_includesHeaderAndDocsLinks(t *testing.T) {
	tests := []struct {
		plugin       string
		expectedLink string
	}{
		{plugin: "rest", expectedLink: "https://docs.imposter.sh/rest_plugin/"},
		{plugin: "openapi", expectedLink: "https://docs.imposter.sh/openapi_plugin/"},
		{plugin: "soap", expectedLink: "https://docs.imposter.sh/soap_plugin/"},
	}

	for _, tt := range tests {
		t.Run(tt.plugin, func(t *testing.T) {
			out := GenerateConfig(ConfigGenerationOptions{PluginName: tt.plugin}, nil)
			s := string(out)

			if !strings.HasPrefix(s, "# Imposter mock configuration") {
				t.Fatalf("expected header comment at start of file, got:\n%s", s)
			}
			if !strings.Contains(s, "https://docs.imposter.sh/configuration/") {
				t.Errorf("expected general configuration docs link in header")
			}
			if !strings.Contains(s, tt.expectedLink) {
				t.Errorf("expected plugin docs link %s in header", tt.expectedLink)
			}

			// The body must still be valid YAML containing the plugin field.
			var parsed PluginConfig
			if err := yaml.Unmarshal(out, &parsed); err != nil {
				t.Fatalf("generated config is not valid YAML: %v", err)
			}
			if parsed.Plugin != tt.plugin {
				t.Errorf("expected plugin %q in unmarshalled config, got %q", tt.plugin, parsed.Plugin)
			}
		})
	}
}
