/*
Copyright © 2021 Pete Cornish <outofcoffee@gmail.com>

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

package cmd

import (
	"fmt"
	"gatehill.io/imposter/internal/engine"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_describeVersions(t *testing.T) {
	type args struct {
		engineType engine.EngineType
		version    string
		format     outputFormat
		cliOnly    bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "print explicit version with docker engine",
			args: args{
				engineType: engine.EngineTypeDockerCore,
				version:    "4.2.2",
				cliOnly:    false,
				format:     outputFormatPlain,
			},
		},
		{
			name: "print explicit version with jvm engine",
			args: args{
				engineType: engine.EngineTypeJvmSingleJar,
				version:    "4.2.2",
				cliOnly:    false,
				format:     outputFormatPlain,
			},
		},
		{
			name: "print latest version",
			args: args{
				engineType: engine.EngineTypeDockerCore,
				version:    "latest",
				cliOnly:    false,
				format:     outputFormatPlain,
			},
		},
		{
			name: "print explicit version in JSON format",
			args: args{
				engineType: engine.EngineTypeDockerCore,
				version:    "4.2.2",
				cliOnly:    false,
				format:     outputFormatJson,
			},
		},
		{
			name: "print CLI version only in JSON format",
			args: args{
				engineType: engine.EngineTypeDockerCore,
				version:    "4.2.2",
				cliOnly:    true,
				format:     outputFormatJson,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("version", tt.args.version)

			var expectedVersion string
			if tt.args.version == "latest" {
				latestVersion, err := engine.ResolveLatestToVersion(true)
				if err != nil {
					t.Fatal(err)
				}
				library := engine.GetLibrary(tt.args.engineType)
				provider := library.GetProvider(latestVersion)
				if err := provider.Provide(engine.PullIfNotPresent); err != nil {
					t.Fatal(err)
				}
				expectedVersion = latestVersion
			} else {
				expectedVersion = tt.args.version
			}

			var want string
			if tt.args.cliOnly {
				if tt.args.format == outputFormatPlain {
					want = `imposter-cli dev
`

				} else {
					want = `{
  "imposter-cli": "dev"
}`
				}
			} else {
				if tt.args.format == outputFormatPlain {
					want = fmt.Sprintf(`imposter-cli dev
imposter-engine %[1]s
engine-output %[1]s
`, expectedVersion)

				} else {
					want = fmt.Sprintf(`{
  "imposter-cli": "dev",
  "imposter-engine": "%[1]s",
  "engine-output": "%[1]s"
}`, expectedVersion)
				}
			}

			got := describeVersions(tt.args.engineType, tt.args.cliOnly, tt.args.format)
			require.Equal(t, want, got, "version should match")
		})
	}
}
