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
	"bytes"
	"gatehill.io/imposter/internal/engine"
	"gatehill.io/imposter/internal/engine/docker"
	"gatehill.io/imposter/internal/engine/golang"
	"gatehill.io/imposter/internal/engine/jvm"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func init() {
	docker.EnableEngine()
	jvm.EnableSingleJarEngine()
	golang.EnableEngine()
}

func Test_renderMocks_without_engine(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rows := [][]string{
		{"abc123", "test-mock", "8080", "healthy"},
	}
	renderMocks(rows, false)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "ID")
	require.Contains(t, output, "NAME")
	require.Contains(t, output, "PORT")
	require.Contains(t, output, "HEALTH")
	require.NotContains(t, output, "ENGINE")
	require.Contains(t, output, "abc123")
	require.Contains(t, output, "test-mock")
	require.Contains(t, output, "8080")
	require.Contains(t, output, "healthy")
}

func Test_renderMocks_with_engine(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rows := [][]string{
		{"abc123", "test-mock", "8080", "healthy", "docker"},
		{"def456", "jvm-mock", "9090", "unhealthy", "jvm"},
	}
	renderMocks(rows, true)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "ENGINE")
	require.Contains(t, output, "docker")
	require.Contains(t, output, "jvm")
}

func Test_renderMocks_empty(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderMocks(nil, false)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	require.Contains(t, output, "ID")
	require.NotContains(t, output, "ENGINE")
}

func Test_listMocksForEngine(t *testing.T) {
	tests := []struct {
		name       string
		engineType engine.EngineType
		showEngine bool
	}{
		{
			name:       "list docker mocks with engine column",
			engineType: engine.EngineTypeDockerCore,
			showEngine: true,
		},
		{
			name:       "list docker mocks without engine column",
			engineType: engine.EngineTypeDockerCore,
			showEngine: false,
		},
		{
			name:       "list jvm mocks",
			engineType: engine.EngineTypeJvmSingleJar,
			showEngine: true,
		},
		{
			name:       "list golang mocks",
			engineType: engine.EngineTypeGolang,
			showEngine: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, mockCount, _, err := listMocksForEngine(tt.engineType, false, tt.showEngine)
			require.NoError(t, err)
			require.GreaterOrEqual(t, mockCount, 0)
			for _, row := range rows {
				if tt.showEngine {
					require.Len(t, row, 5, "row should have 5 columns when showing engine")
					require.Equal(t, string(tt.engineType), row[4])
				} else {
					require.Len(t, row, 4, "row should have 4 columns when not showing engine")
				}
			}
		})
	}
}

func Test_listCmd_mutual_exclusivity(t *testing.T) {
	rootCmd.SetArgs([]string{"list", "-a", "-t", "docker"})
	err := rootCmd.Execute()
	require.Error(t, err, "should reject --all with --engine-type")
}

