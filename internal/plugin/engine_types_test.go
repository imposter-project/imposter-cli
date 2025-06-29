package plugin

import (
	"gatehill.io/imposter/internal/platform"
	"testing"
)

func Test_buildPluginFileName(t *testing.T) {
	tests := []struct {
		name         string
		pluginConfig pluginConfiguration
		pluginName   string
		ext          string
		want         string
	}{
		{
			name: "golang plugin without OS/arch",
			pluginConfig: pluginConfiguration{
				fileNamePrefix: "plugin-",
				addOsAndArch:   false,
			},
			pluginName: "test-plugin",
			ext:        ".zip",
			want:       "plugin-test-plugin.zip",
		},
		{
			name: "golang plugin with OS/arch",
			pluginConfig: pluginConfiguration{
				fileNamePrefix: "plugin-",
				addOsAndArch:   true,
			},
			pluginName: "test-plugin",
			ext:        ".zip",
			want:       "plugin-test-plugin_" + getCurrentOSAndArch() + ".zip",
		},
		{
			name: "jvm plugin without OS/arch",
			pluginConfig: pluginConfiguration{
				fileNamePrefix: "imposter-plugin-",
				addOsAndArch:   false,
			},
			pluginName: "store-redis",
			ext:        ".jar",
			want:       "imposter-plugin-store-redis.jar",
		},
		{
			name: "jvm plugin with OS/arch",
			pluginConfig: pluginConfiguration{
				fileNamePrefix: "imposter-plugin-",
				addOsAndArch:   true,
			},
			pluginName: "store-redis",
			ext:        ".jar",
			want:       "imposter-plugin-store-redis_" + getCurrentOSAndArch() + ".jar",
		},
		{
			name: "plugin with empty name",
			pluginConfig: pluginConfiguration{
				fileNamePrefix: "plugin-",
				addOsAndArch:   false,
			},
			pluginName: "",
			ext:        ".zip",
			want:       "plugin-.zip",
		},
		{
			name: "plugin with special characters in name",
			pluginConfig: pluginConfiguration{
				fileNamePrefix: "plugin-",
				addOsAndArch:   false,
			},
			pluginName: "test-plugin_v1.0.0",
			ext:        ".jar",
			want:       "plugin-test-plugin_v1.0.0.jar",
		},
		{
			name: "plugin with different extension",
			pluginConfig: pluginConfiguration{
				fileNamePrefix: "imposter-plugin-",
				addOsAndArch:   false,
			},
			pluginName: "multi-ext",
			ext:        ".zip",
			want:       "imposter-plugin-multi-ext.zip",
		},
		{
			name: "plugin with empty prefix",
			pluginConfig: pluginConfiguration{
				fileNamePrefix: "",
				addOsAndArch:   false,
			},
			pluginName: "no-prefix",
			ext:        ".zip",
			want:       "no-prefix.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPluginFileName(tt.pluginConfig, tt.pluginName, tt.ext)
			if got != tt.want {
				t.Errorf("buildPluginFileName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to get current OS and architecture for testing
func getCurrentOSAndArch() string {
	os, arch := platform.GetPlatform()
	return os + "_" + arch
}
