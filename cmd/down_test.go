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
	"github.com/imposter-project/imposter-cli/internal/engine"
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
			name:       "stop native engine",
			engineType: engine.EngineTypeNative,
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

func Test_downCmd_requires_id_or_all(t *testing.T) {
	// bare `imposter down` is fatal — capture via runWithRecovery
	rootCmd.SetArgs([]string{"down"})
	err := runWithRecovery(func() {
		_ = rootCmd.Execute()
	})
	require.Error(t, err, "should fail without an ID or --all")
}

func Test_downCmd_rejects_id_with_all(t *testing.T) {
	rootCmd.SetArgs([]string{"down", "--all", "abc123"})
	err := runWithRecovery(func() {
		_ = rootCmd.Execute()
	})
	require.Error(t, err, "should reject ID together with --all")
}

func Test_stopMockByID_unknown(t *testing.T) {
	err := runWithRecovery(func() {
		stopMockByID("definitely-not-a-real-mock-id")
	})
	require.Error(t, err, "should fail when no engine has a mock with the given id")
}
