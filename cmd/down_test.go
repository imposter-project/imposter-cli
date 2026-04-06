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
	"gatehill.io/imposter/internal/engine"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_stopEngine(t *testing.T) {
	tests := []struct {
		name       string
		engineType engine.EngineType
	}{
		{
			name:       "stop docker engine",
			engineType: engine.EngineTypeDockerCore,
		},
		{
			name:       "stop jvm engine",
			engineType: engine.EngineTypeJvmSingleJar,
		},
		{
			name:       "stop golang engine",
			engineType: engine.EngineTypeGolang,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stopped, err := stopEngine(tt.engineType)
			require.NoError(t, err)
			require.GreaterOrEqual(t, stopped, 0)
		})
	}
}

func Test_downCmd_mutual_exclusivity(t *testing.T) {
	rootCmd.SetArgs([]string{"down", "-a", "-t", "docker"})
	err := rootCmd.Execute()
	require.Error(t, err, "should reject --all with --engine-type")
}
