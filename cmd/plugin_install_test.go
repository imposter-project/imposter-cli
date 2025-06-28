/*
Copyright © 2021 Pete Cornish <outofcoffee@gmail.com>

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
	"gatehill.io/imposter/internal/config"
	"gatehill.io/imposter/internal/engine"
	"gatehill.io/imposter/internal/plugin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func init() {
	configDir, err := os.MkdirTemp(os.TempDir(), "imposter-cli")
	if err != nil {
		panic(err)
	}
	config.DirPath = configDir
}

func Test_installPlugins(t *testing.T) {
	type args struct {
		argPlugins    []string
		configPlugins []string
		version       string
		saveDefault   bool
		engineType    engine.EngineType
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "install no plugins", args: args{argPlugins: nil, configPlugins: nil, engineType: engine.EngineTypeDockerCore, version: "4.2.2"}},
		{name: "install plugins from args", args: args{argPlugins: []string{"store-redis"}, configPlugins: nil, engineType: engine.EngineTypeDockerCore, version: "4.2.2"}},
		{name: "install plugins from config", args: args{argPlugins: nil, configPlugins: []string{"store-redis"}, engineType: engine.EngineTypeDockerCore, version: "4.2.2"}},
		{name: "install and save plugins as default", args: args{argPlugins: []string{"store-redis"}, configPlugins: nil, engineType: engine.EngineTypeDockerCore, version: "4.2.2", saveDefault: true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			viper.Set("plugins", tt.args.configPlugins)
			viper.Set("default.plugins", []string{})
			t.Cleanup(func() {
				viper.Set("plugins", nil)
			})
			installPlugins(tt.args.argPlugins, tt.args.engineType, tt.args.version, tt.args.saveDefault)

			if tt.args.saveDefault {
				defaultPlugins, err := plugin.ListDefaultPlugins()
				if err != nil {
					t.Fatal(err)
				}
				require.ElementsMatch(t, tt.args.argPlugins, defaultPlugins, "default plugins should be set")
			}
		})
	}
}
